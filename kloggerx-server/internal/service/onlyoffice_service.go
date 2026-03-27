package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"kloggerx-server/config"
	"kloggerx-server/internal/model"
	miniosvc "kloggerx-server/internal/pkg/minio"
	"kloggerx-server/internal/repository/mysql"

	"github.com/golang-jwt/jwt/v5"
)

// OnlyOfficeConfig represents the editor configuration sent to frontend
type OnlyOfficeConfig struct {
	Document    OnlyOfficeDocument `json:"document"`
	DocumentType string            `json:"documentType"`
	EditorConfig OnlyOfficeEditor  `json:"editorConfig"`
	Token       string             `json:"token"`
}

type OnlyOfficeDocument struct {
	Key      string                 `json:"key"`
	Title    string                 `json:"title"`
	URL      string                 `json:"url"`
	FileType string                 `json:"fileType"`
}

type OnlyOfficeEditor struct {
	CallbackURL string                   `json:"callbackUrl"`
	Mode        string                   `json:"mode"`
	User        OnlyOfficeUser           `json:"user"`
	Customization *OnlyOfficeCustomization `json:"customization,omitempty"`
	Lang        string                   `json:"lang"`
}

type OnlyOfficeUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OnlyOfficeCustomization struct {
	Autosave  bool `json:"autosave"`
	Chat      bool `json:"chat"`
	Comments  bool `json:"comments"`
}

// OnlyOfficeCallback represents the callback from OnlyOffice
type OnlyOfficeCallback struct {
	Key        string `json:"key"`
	Status     int    `json:"status"`
	URL        string `json:"url"`
	ChangesURL string `json:"changesurl"`
	History    *struct {
		ServerVersion string `json:"serverVersion"`
		Changes       []struct {
			ServerVersion string `json:"serverVersion"`
			Created       string `json:"created"`
			User          struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"user"`
		} `json:"changes"`
	} `json:"history"`
	Changes     []interface{} `json:"changes"`
	Users       []string      `json:"users"`
	Actions     []interface{} `json:"actions"`
	LastSave    string        `json:"lastsave"`
	NotModified bool          `json:"notmodified"`
}

// Callback status constants
const (
	StatusEditing       = 1
	StatusReadyForSave  = 2
	StatusErrorSaving   = 3
	StatusClosedNoChanges = 4
	StatusForceSave     = 6
	StatusErrorForceSave = 7
)

// GetOnlyOfficeConfig generates editor configuration with JWT token
func GetOnlyOfficeConfig(docID, userID uint, username string, mode string) (*OnlyOfficeConfig, error) {
	doc, err := GetDocumentDetail(docID, userID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// Determine file type (for uploaded files, infer from content metadata)
	fileType := resolveOnlyOfficeFileType(doc)
	if fileType == "" {
		return nil, fmt.Errorf("unsupported document type: %s", doc.Type)
	}

	// Generate unique document key (changes when content is updated)
	docKey := generateDocumentKey(docID, doc.Version)

	// Build signed file URL for OnlyOffice to download
	fileURL := GetDocumentDownloadURL(docID)

	// Build callback URL for OnlyOffice to save changes
	callbackURL := fmt.Sprintf("%s/api/v1/onlyoffice/callback/%d", config.Cfg.OnlyOffice.CallbackURL, docID)

	// Editor mode: edit or view
	editorMode := "view"
	if mode == "edit" {
		editorMode = "edit"
	}

	// Build document config
	document := OnlyOfficeDocument{
		Key:      docKey,
		Title:    doc.Title,
		URL:      fileURL,
		FileType: fileType,
	}

	// Build editor config
	editor := OnlyOfficeEditor{
		CallbackURL: callbackURL,
		Mode:        editorMode,
		User: OnlyOfficeUser{
			ID:   fmt.Sprintf("%d", userID),
			Name: username,
		},
		Lang: "zh-CN",
		Customization: &OnlyOfficeCustomization{
			Autosave:  true,
			Chat:      true,
			Comments:  true,
		},
	}

	// Build JWT token for the config
	token, err := signOnlyOfficeToken(document, getOnlyOfficeDocumentType(fileType), editor)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &OnlyOfficeConfig{
		Document:     document,
		DocumentType: getOnlyOfficeDocumentType(fileType),
		EditorConfig: editor,
		Token:        token,
	}, nil
}

// signOnlyOfficeToken creates a JWT token for OnlyOffice authentication
func signOnlyOfficeToken(document OnlyOfficeDocument, documentType string, editor OnlyOfficeEditor) (string, error) {
	claims := jwt.MapClaims{
		"document":     document,
		"documentType": documentType,
		"editorConfig": editor,
		"iat":          time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Cfg.OnlyOffice.JWTSecret))
}

// VerifyOnlyOfficeCallback verifies the JWT token from OnlyOffice callback
func VerifyOnlyOfficeCallback(tokenString string) (*OnlyOfficeCallback, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Cfg.OnlyOffice.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Extract the payload
		payload, err := json.Marshal(claims["payload"])
		if err != nil {
			return nil, err
		}

		var callback OnlyOfficeCallback
		if err := json.Unmarshal(payload, &callback); err != nil {
			return nil, err
		}

		return &callback, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// VerifyOnlyOfficeSignature verifies HMAC signature for download requests
func VerifyOnlyOfficeSignature(docID uint, signature string) bool {
	expectedSig := generateDownloadSignature(docID)
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// generateDownloadSignature creates an HMAC signature for file download
func generateDownloadSignature(docID uint) string {
	key := []byte(config.Cfg.OnlyOffice.JWTSecret)
	message := []byte(fmt.Sprintf("download:%d", docID))
	h := hmac.New(sha256.New, key)
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

// HandleOnlyOfficeCallback processes the callback from OnlyOffice
func HandleOnlyOfficeCallback(docID uint, callback *OnlyOfficeCallback) error {
	switch callback.Status {
	case StatusReadyForSave, StatusForceSave:
		// Document is ready to be saved
		if callback.URL == "" {
			return fmt.Errorf("no URL provided for saving document")
		}

		// Download the updated document from OnlyOffice
		content, err := downloadDocumentFromOnlyOffice(callback.URL)
		if err != nil {
			return fmt.Errorf("failed to download document: %w", err)
		}

		// Save to database
		if err := SaveDocumentContent(docID, 0, string(content)); err != nil {
			return fmt.Errorf("failed to save document content: %w", err)
		}

		return nil

	case StatusEditing:
		// Document is being edited, no action needed
		return nil

	case StatusClosedNoChanges:
		// Document closed without changes
		return nil

	case StatusErrorSaving, StatusErrorForceSave:
		return fmt.Errorf("onlyoffice error saving document: status %d", callback.Status)

	default:
		return fmt.Errorf("unknown callback status: %d", callback.Status)
	}
}

// downloadDocumentFromOnlyOffice downloads the document from OnlyOffice server
func downloadDocumentFromOnlyOffice(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// generateDocumentKey creates a unique key for the document
// The key changes when the document content is updated
func generateDocumentKey(docID uint, version int) string {
	key := fmt.Sprintf("kloggerx_%d_v%d_%d", docID, version, time.Now().UnixNano()/1000000)
	h := sha256.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))[:20]
}

// resolveOnlyOfficeFileType resolves OnlyOffice file type from document model
// For uploaded files (doc.Type == "file"), it infers type from multiple metadata fields.
func resolveOnlyOfficeFileType(doc *model.Document) string {
	if doc == nil {
		return ""
	}

	if doc.Type != "file" {
		return getOnlyOfficeFileType(doc.Type)
	}

	// Uploaded file: infer real type from content metadata (priority: fileType > fileName > filePath)
	var contentData struct {
		FileType string `json:"fileType"`
		FileName string `json:"fileName"`
		FilePath string `json:"filePath"`
	}
	if err := json.Unmarshal([]byte(doc.Content), &contentData); err == nil {
		if ft := mapUploadedFileTypeToOOExt(contentData.FileType); ft != "" {
			return ft
		}
		if ft := inferOOExtFromPathOrName(contentData.FileName); ft != "" {
			return ft
		}
		if ft := inferOOExtFromPathOrName(contentData.FilePath); ft != "" {
			return ft
		}
	}

	// Last fallback for uploaded file: prefer spreadsheet-safe default?
	// Keep docx as legacy default to avoid breaking existing behavior when metadata is absent.
	return "docx"
}

// getOnlyOfficeFileType maps internal document type to OnlyOffice file extension
func getOnlyOfficeFileType(docType string) string {
	switch docType {
	case "doc", "document":
		return "docx"
	case "sheet", "spreadsheet":
		return "xlsx"
	case "slide", "presentation":
		return "pptx"
	case "file":
		return "docx"
	default:
		return ""
	}
}

func mapUploadedFileTypeToOOExt(fileType string) string {
	switch strings.ToLower(strings.TrimSpace(fileType)) {
	case "word", "doc", "docx":
		return "docx"
	case "excel", "sheet", "xlsx", "xls", "csv":
		return "xlsx"
	case "ppt", "pptx", "presentation", "slide":
		return "pptx"
	case "pdf":
		return "pdf"
	default:
		return ""
	}
}

func mapFileExtToOOExt(ext string) string {
	switch strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ext), ".")) {
	case "docx", "doc", "odt", "rtf", "txt", "html", "htm":
		return "docx"
	case "xlsx", "xls", "ods", "csv":
		return "xlsx"
	case "pptx", "ppt", "odp", "ppsx":
		return "pptx"
	case "pdf":
		return "pdf"
	default:
		return ""
	}
}

func inferOOExtFromPathOrName(s string) string {
	value := strings.TrimSpace(s)
	if value == "" {
		return ""
	}
	dot := strings.LastIndex(value, ".")
	if dot < 0 || dot == len(value)-1 {
		return ""
	}
	return mapFileExtToOOExt(value[dot+1:])
}

// getOnlyOfficeDocumentType returns the document type for OnlyOffice
func getOnlyOfficeDocumentType(fileType string) string {
	switch fileType {
	case "docx", "doc", "odt", "rtf", "txt", "html", "htm", "mht", "pdf", "djvu", "fb2", "epub", "xps":
		return "word"
	case "xlsx", "xls", "ods", "csv":
		return "cell"
	case "pptx", "ppt", "odp", "ppsx":
		return "slide"
	default:
		return "word"
	}
}

// GetDocumentDownloadURL generates a signed download URL for OnlyOffice
func GetDocumentDownloadURL(docID uint) string {
	sig := generateDownloadSignature(docID)
	return fmt.Sprintf("%s/api/v1/onlyoffice/download/%d?sig=%s", config.Cfg.OnlyOffice.FileBaseURL, docID, sig)
}

// GetDocumentContentForOnlyOffice retrieves document content for OnlyOffice download
func GetDocumentContentForOnlyOffice(docID uint) ([]byte, string, error) {
	var doc model.Document
	if err := mysql.DB.First(&doc, docID).Error; err != nil {
		return nil, "", err
	}

	// Parse content to check if it's an uploaded file
	var contentData struct {
		Type     string `json:"type"`
		FileType string `json:"fileType"`
		FileName string `json:"fileName"`
		FilePath string `json:"filePath"`
	}

	if err := json.Unmarshal([]byte(doc.Content), &contentData); err == nil && contentData.FilePath != "" {
		log.Printf("[OnlyOffice][content] docID=%d uploaded file metadata fileType=%q fileName=%q filePath=%q", docID, contentData.FileType, contentData.FileName, contentData.FilePath)
		// This is an uploaded file - read from storage
		content, err := miniosvc.GetFileContent(contentData.FilePath)
		if err != nil {
			log.Printf("[OnlyOffice][content] read storage failed docID=%d filePath=%q err=%v", docID, contentData.FilePath, err)
			return nil, "", fmt.Errorf("failed to read file from storage: %w", err)
		}

		// Determine file type from metadata with robust fallback
		fileType := mapUploadedFileTypeToOOExt(contentData.FileType)
		if fileType == "" {
			fileType = inferOOExtFromPathOrName(contentData.FileName)
		}
		if fileType == "" {
			fileType = inferOOExtFromPathOrName(contentData.FilePath)
		}
		if fileType == "" {
			fileType = "docx"
		}
		log.Printf("[OnlyOffice][content] docID=%d resolved fileType=%q size=%d", docID, fileType, len(content))

		return content, fileType, nil
	}

	// Get file type from document type
	fileType := getOnlyOfficeFileType(doc.Type)
	if fileType == "" {
		fileType = "docx"
	}

	// Return content (for inline documents)
	content := []byte(doc.Content)
	if content == nil || len(content) == 0 {
		// Return empty document template
		content = []byte("{}")
	}

	return content, fileType, nil
}

// ParseAuthorizationHeader extracts JWT token from Authorization header
func ParseAuthorizationHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

// EncodeBase64URL encodes string to base64 URL-safe format
func EncodeBase64URL(data string) string {
	return base64.URLEncoding.EncodeToString([]byte(data))
}

// DecodeBase64URL decodes base64 URL-safe string
func DecodeBase64URL(data string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
