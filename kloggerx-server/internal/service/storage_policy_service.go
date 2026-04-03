package service

import (
	"encoding/json"
	"log"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

// StartStoragePolicyJobs starts background jobs for storage policy enforcement
func StartStoragePolicyJobs() {
	// Run cleanup every hour
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			runStoragePolicyCleanup()
		}
	}()

	// Run once at startup
	go runStoragePolicyCleanup()
}

func runStoragePolicyCleanup() {
	// Get storage policy settings
	policy := getStoragePolicy()

	// Clean old document versions
	if policy.VersionRetention > 0 {
		cleanOldVersions(policy.VersionRetention)
	}

	// Clean old recycle bin items
	if policy.RecycleRetention > 0 {
		cleanRecycleBin(policy.RecycleRetention)
	}
}

type StoragePolicyConfig struct {
	VersionRetention   int
	RecycleRetention   int
	LargeFileThreshold int
	AutoCleanCache     bool
}

func getStoragePolicy() StoragePolicyConfig {
	var setting model.SystemSetting
	if err := mysql.DB.Where("`key` = ?", "storage_settings").First(&setting).Error; err != nil {
		return StoragePolicyConfig{
			VersionRetention:   20,
			RecycleRetention:   30,
			LargeFileThreshold: 100,
			AutoCleanCache:     true,
		}
	}

	// Parse JSON
	var data map[string]interface{}
	if err := parseJSON(setting.Value, &data); err != nil {
		return StoragePolicyConfig{
			VersionRetention:   20,
			RecycleRetention:   30,
			LargeFileThreshold: 100,
			AutoCleanCache:     true,
		}
	}

	policy := StoragePolicyConfig{}
	if policyData, ok := data["storagePolicy"].(map[string]interface{}); ok {
		if v, ok := policyData["versionRetention"].(float64); ok {
			policy.VersionRetention = int(v)
		}
		if v, ok := policyData["recycleRetention"].(float64); ok {
			policy.RecycleRetention = int(v)
		}
		if v, ok := policyData["largeFileThreshold"].(float64); ok {
			policy.LargeFileThreshold = int(v)
		}
		if v, ok := policyData["autoCleanCache"].(bool); ok {
			policy.AutoCleanCache = v
		}
	}

	return policy
}

func cleanOldVersions(retentionCount int) {
	// Get all documents
	var documents []model.Document
	mysql.DB.Where("type != ?", "folder").Find(&documents)

	for _, doc := range documents {
		// Get versions for this document
		var versions []model.DocumentVersion
		mysql.DB.Where("document_id = ?", doc.ID).Order("created_at DESC").Find(&versions)

		// If more versions than retention, delete old ones
		if len(versions) > retentionCount {
			versionsToDelete := versions[retentionCount:]
			for _, v := range versionsToDelete {
				mysql.DB.Delete(&v)
			}
			log.Printf("[StoragePolicy] Cleaned %d old versions for document %d", len(versionsToDelete), doc.ID)
		}
	}
}

func cleanRecycleBin(retentionDays int) {
	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Find deleted documents older than retention period
	var documents []model.Document
	mysql.DB.Where("is_deleted = ? AND deleted_at < ?", true, cutoffDate).Find(&documents)

	// Permanently delete them
	for _, doc := range documents {
		// Delete associated versions
		mysql.DB.Where("document_id = ?", doc.ID).Delete(&model.DocumentVersion{})

		// Delete associated permissions
		mysql.DB.Where("document_id = ?", doc.ID).Delete(&model.Permission{})

		// Delete the document
		mysql.DB.Unscoped().Delete(&doc)
	}

	if len(documents) > 0 {
		log.Printf("[StoragePolicy] Permanently deleted %d documents from recycle bin", len(documents))
	}
}

func parseJSON(jsonStr string, target interface{}) error {
	return json.Unmarshal([]byte(jsonStr), target)
}
