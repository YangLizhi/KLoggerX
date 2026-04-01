package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"kloggerx-server/internal/model"
	miniosvc "kloggerx-server/internal/pkg/minio"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/repository/mysql"
)

func UploadFile(file *multipart.FileHeader, uploaderID uint) (*model.FileRecord, string, error) {
	src, err := file.Open()
	if err != nil {
		return nil, "", err
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("uploads/%s/%s%s", time.Now().Format("2006/01/02"), utils.GenerateRandomString(16), ext)

	if err := miniosvc.Upload(objectName, src, file.Size, file.Header.Get("Content-Type")); err != nil {
		return nil, "", err
	}

	url, err := miniosvc.GetPresignedURL(objectName)
	if err != nil {
		return nil, "", err
	}

	record := model.FileRecord{
		Name:       file.Filename,
		Path:       objectName,
		Size:       file.Size,
		MimeType:   file.Header.Get("Content-Type"),
		UploaderID: uploaderID,
	}
	mysql.DB.Create(&record)
	return &record, url, nil
}

func GetFilePreviewURL(fileID uint) (string, error) {
	var f model.FileRecord
	if err := mysql.DB.First(&f, fileID).Error; err != nil {
		return "", err
	}
	return miniosvc.GetPresignedURL(f.Path)
}

// DownloadFileStream returns the original filename and a ReadCloser for streaming the file contents.
func DownloadFileStream(fileID uint) (string, string, io.ReadCloser, error) {
	var f model.FileRecord
	if err := mysql.DB.First(&f, fileID).Error; err != nil {
		return "", "", nil, err
	}
	rc, err := miniosvc.GetObject(f.Path)
	if err != nil {
		return "", "", nil, err
	}
	return f.Name, f.MimeType, rc, nil
}

// GetFileStreamByPath returns a ReadCloser for a file by its storage object name.
func GetFileStreamByPath(objectName string) (io.ReadCloser, error) {
	return miniosvc.GetObject(objectName)
}

func DeleteFile(fileID uint) error {
	var f model.FileRecord
	if err := mysql.DB.First(&f, fileID).Error; err != nil {
		return err
	}
	miniosvc.Delete(f.Path)
	return mysql.DB.Delete(&f).Error
}

