package service

import (
	"syscall"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

// StorageStats represents storage statistics
type StorageStats struct {
	Total      int64 `json:"total"`      // Total disk space in bytes
	Used       int64 `json:"used"`       // Used disk space in bytes
	Available  int64 `json:"available"`  // Available disk space in bytes
	FileCount  int64 `json:"fileCount"`  // Total file count
	UsedByUser int64 `json:"usedByUser"` // Used by current user
}

// StorageUsageByType represents storage usage broken down by file type
type StorageUsageByType struct {
	Doc    int64 `json:"doc"`    // Documents
	Sheet  int64 `json:"sheet"`  // Spreadsheets
	Slide  int64 `json:"slide"`  // Presentations
	Image  int64 `json:"image"`  // Images
	Other  int64 `json:"other"`  // Other files
}

// GetDiskUsage returns disk usage statistics for the given path
func GetDiskUsage(path string) (total, used, available int64, err error) {
	var stat syscall.Statfs_t
	err = syscall.Statfs(path, &stat)
	if err != nil {
		return 0, 0, 0, err
	}
	
	// Total disk space
	total = int64(stat.Blocks) * int64(stat.Bsize)
	// Available disk space (for non-root users)
	available = int64(stat.Bavail) * int64(stat.Bsize)
	// Used disk space
	used = total - available
	
	return total, used, available, nil
}

// GetSystemStorageStats returns system-wide storage statistics
func GetSystemStorageStats() (*StorageStats, error) {
	// Get disk usage for uploads directory
	total, used, available, err := GetDiskUsage("./uploads")
	if err != nil {
		// Fallback to current directory
		total, used, available, err = GetDiskUsage(".")
		if err != nil {
			return nil, err
		}
	}
	
	// Count total files in database
	var fileCount int64
	mysql.DB.Model(&model.Document{}).Where("type = ?", "file").Count(&fileCount)
	
	// Also count FileRecord
	var fileRecordCount int64
	mysql.DB.Model(&model.FileRecord{}).Count(&fileRecordCount)
	fileCount += fileRecordCount
	
	return &StorageStats{
		Total:     total,
		Used:      used,
		Available: available,
		FileCount: fileCount,
	}, nil
}

// GetUserStorageUsage returns storage usage for a specific user
func GetUserStorageUsage(userID uint) (*StorageUsageByType, int64, int, error) {
	usage := &StorageUsageByType{}
	var totalSize int64
	var totalCount int
	
	// Get documents owned by user
	var documents []model.Document
	mysql.DB.Where("owner_id = ? AND is_deleted = ?", userID, false).Find(&documents)
	
	for _, doc := range documents {
		size := doc.FileSize
		totalSize += size
		totalCount++
		
		// Classify by file extension
		ext := doc.FileExt
		switch {
		case isDocType(ext):
			usage.Doc += size
		case isSheetType(ext):
			usage.Sheet += size
		case isSlideType(ext):
			usage.Slide += size
		case isImageType(ext):
			usage.Image += size
		default:
			usage.Other += size
		}
	}
	
	return usage, totalSize, totalCount, nil
}

// GetAllStorageUsage returns storage usage for all users (admin view)
func GetAllStorageUsage() (*StorageUsageByType, int64, int, error) {
	usage := &StorageUsageByType{}
	var totalSize int64
	var totalCount int
	
	// Get all non-deleted documents
	var documents []model.Document
	mysql.DB.Where("is_deleted = ?", false).Find(&documents)
	
	for _, doc := range documents {
		size := doc.FileSize
		totalSize += size
		totalCount++
		
		ext := doc.FileExt
		switch {
		case isDocType(ext):
			usage.Doc += size
		case isSheetType(ext):
			usage.Sheet += size
		case isSlideType(ext):
			usage.Slide += size
		case isImageType(ext):
			usage.Image += size
		default:
			usage.Other += size
		}
	}
	
	return usage, totalSize, totalCount, nil
}

// UpdateStorageUsageCache updates the storage_usage table for a user
func UpdateStorageUsageCache(userID uint) error {
	usage, _, _, err := GetUserStorageUsage(userID)
	if err != nil {
		return err
	}
	
	fileTypes := map[string]int64{
		"doc":   usage.Doc,
		"sheet": usage.Sheet,
		"slide": usage.Slide,
		"image": usage.Image,
		"other": usage.Other,
	}
	
	for fileType, totalSize := range fileTypes {
		var storageUsage model.StorageUsage
		result := mysql.DB.Where("user_id = ? AND file_type = ?", userID, fileType).First(&storageUsage)
		
		if result.Error != nil {
			// Create new record
			storageUsage = model.StorageUsage{
				UserID:    userID,
				FileType:  fileType,
				TotalSize: totalSize,
				FileCount: 0, // Will be updated separately if needed
			}
			mysql.DB.Create(&storageUsage)
		} else {
			// Update existing record
			mysql.DB.Model(&storageUsage).Update("total_size", totalSize)
		}
	}
	
	return nil
}

// Helper functions to classify file types
func isDocType(ext string) bool {
	docExts := map[string]bool{
		"doc": true, "docx": true, "pdf": true, "txt": true,
		"md": true, "rtf": true, "odt": true, "wps": true,
	}
	return docExts[ext]
}

func isSheetType(ext string) bool {
	sheetExts := map[string]bool{
		"xls": true, "xlsx": true, "csv": true, "ods": true,
		"et": true,
	}
	return sheetExts[ext]
}

func isSlideType(ext string) bool {
	slideExts := map[string]bool{
		"ppt": true, "pptx": true, "odp": true, "dps": true,
	}
	return slideExts[ext]
}

func isImageType(ext string) bool {
	imageExts := map[string]bool{
		"jpg": true, "jpeg": true, "png": true, "gif": true,
		"bmp": true, "svg": true, "webp": true, "ico": true,
	}
	return imageExts[ext]
}
