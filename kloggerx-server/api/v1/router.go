package v1

import (
	"kloggerx-server/internal/middleware"
	"kloggerx-server/internal/ws"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, hub *ws.Hub) {
	api := r.Group("/api/v1")

	// User - public
	api.POST("/user/login", UserLogin)
	api.POST("/user/register", UserRegister)

	// Authenticated routes
	auth := api.Group("", middleware.Auth())
	{
		// User
		auth.GET("/user/info", GetUserInfo)
		auth.POST("/user/update", UpdateUserInfo)
		auth.POST("/user/change-password", ChangeUserPassword)
		auth.GET("/user/list", GetUserList)

		// Document
		auth.GET("/document/tree", GetDocumentTree)
		auth.GET("/document/:id", GetDocumentDetail)
		auth.POST("/document/create", CreateDocument)
		auth.POST("/document/:id/update", UpdateDocument)
		auth.POST("/document/:id/content", SaveDocumentContent)
		auth.POST("/document/:id/delete", DeleteDocument)
		auth.POST("/document/:id/restore", RestoreDocument)
		auth.DELETE("/document/:id", PermanentDeleteDocument)
		auth.POST("/document/:id/move", MoveDocument)
		auth.POST("/document/:id/copy", CopyDocument)
		auth.POST("/document/:id/pin", PinDocument)
		auth.POST("/document/:id/favorite", FavoriteDocument)
		auth.POST("/document/:id/transfer", TransferOwnership)
		auth.GET("/document/recycle-bin", GetRecycleBin)
		auth.GET("/document/favorites", GetFavorites)
		auth.GET("/document/pinned", GetPinnedDocuments)
		auth.GET("/document/recent", GetRecentDocuments)
		auth.GET("/document/search", SearchDocuments)
		auth.GET("/document/:id/versions", GetDocumentVersions)
		auth.POST("/document/:id/rollback", RollbackVersion)
		auth.POST("/document/import", ImportDocument)
		auth.GET("/document/:id/export", ExportDocument)
		auth.GET("/document/:id/file-preview", GetDocumentFilePreview)
		auth.GET("/document/:id/download", DownloadDocumentFile)

		// Permission
		auth.GET("/auth/document/:id", GetDocumentPermissions)
		auth.POST("/auth/permission/set", SetPermission)
		auth.POST("/auth/permission/remove", RemovePermission)
		auth.GET("/auth/share/:id", GetShareSetting)
		auth.POST("/auth/share/update", UpdateShareSetting)
		auth.GET("/auth/check/:id", CheckPermission)

		// Knowledge
		auth.GET("/knowledge/list", GetKnowledgeBaseList)
		auth.GET("/knowledge/:id", GetKnowledgeBaseDetail)
		auth.POST("/knowledge/create", CreateKnowledgeBase)
		auth.POST("/knowledge/:id/update", UpdateKnowledgeBase)
		auth.DELETE("/knowledge/:id", DeleteKnowledgeBase)
		auth.GET("/knowledge/:id/tree", GetKnowledgeBaseTree)
		auth.POST("/knowledge/:id/member/add", AddKnowledgeMember)
		auth.POST("/knowledge/:id/member/remove", RemoveKnowledgeMember)
		auth.GET("/knowledge/:id/members", GetKnowledgeMembers)
		auth.GET("/knowledge/search", SearchKnowledge)
		auth.POST("/knowledge/:id/publish", PublishDocument)
		// Knowledge sources (folder/document linkage)
		auth.GET("/knowledge/:id/sources", GetKnowledgeSources)
		auth.POST("/knowledge/:id/source/add", AddKnowledgeSource)
		auth.DELETE("/knowledge/:id/source/:srcId", RemoveKnowledgeSource)
		auth.POST("/knowledge/:id/source/:srcId/sync", SyncKnowledgeSource)
		// Knowledge AI chat
		auth.POST("/knowledge/:id/chat", KnowledgeChat)
		auth.POST("/knowledge/chat/global", KnowledgeChatGlobal)
		// Knowledge RAPTOR and embedding
		auth.POST("/knowledge/:id/raptor/build", BuildRaptorTree)
		auth.GET("/knowledge/:id/raptor/stats", GetRaptorTreeStats)
		auth.GET("/knowledge/:id/embedding/status", GetEmbeddingStatus)
		auth.POST("/knowledge/:id/embedding/rebuild", RebuildEmbeddings)

		// Templates
		auth.GET("/template/list", GetTemplates)
		auth.GET("/template/:id", GetTemplateDetail)
		auth.POST("/template/:id/use", UseTemplate)

		// Collaborate
		auth.POST("/collaborate/comment/add", AddComment)
		auth.GET("/collaborate/comment/:id", GetComments)
		auth.POST("/collaborate/comment/:id/resolve", ResolveComment)
		auth.POST("/collaborate/comment/:id/delete", DeleteComment)
		auth.GET("/collaborate/notifications", GetNotifications)
		auth.POST("/collaborate/notification/:id/read", MarkNotificationRead)
		auth.POST("/collaborate/notification/read-all", MarkAllNotificationsRead)
		auth.POST("/collaborate/notification/create", CreateSystemNotification)

		// File
		auth.POST("/file/upload", UploadFile)
		auth.GET("/file/:id/preview", GetFilePreviewURL)
		auth.POST("/file/:id/delete", DeleteFile)

		// Storage Statistics
		auth.GET("/storage/stats", GetStorageStats)
		auth.GET("/storage/usage", GetStorageUsage)

		// User Storage Settings
		auth.GET("/user/storage-settings", GetUserStorageSettings)
		auth.POST("/user/storage-settings", SaveUserStorageSettings)

		// Remote Storage File Operations (for all authenticated users)
		auth.GET("/remote-storage/:id/files", ListRemoteStorageFiles)
		auth.GET("/remote-storage/:id/download", DownloadRemoteFile)
		auth.POST("/remote-storage/:id/upload", UploadRemoteFile)
		auth.POST("/remote-storage/:id/mkdir", CreateRemoteFolder)
		auth.DELETE("/remote-storage/:id/file", DeleteRemoteFile)
		auth.POST("/remote-storage/:id/test", TestRemoteStorageConnect)
		auth.POST("/remote-storage/:id/disconnect", DisconnectRemoteStorageHandler)

		// Operation Logs
		auth.GET("/document/:id/logs", GetOperationLogs)

		// OnlyOffice
		auth.POST("/onlyoffice/config", GetOnlyOfficeConfig)
		auth.GET("/onlyoffice/server-url", GetOnlyOfficeServerURL)

		// Admin routes (require admin role)
		admin := auth.Group("", middleware.AdminOnly())
		{
			// User Management
			admin.GET("/admin/users", AdminGetUserList)
			admin.POST("/admin/users", AdminCreateUser)
			admin.PUT("/admin/users/:id", AdminUpdateUser)
			admin.DELETE("/admin/users/:id", AdminDeleteUser)
			admin.POST("/admin/users/:id/reset-password", AdminResetUserPassword)

			// Department Management
			admin.GET("/admin/departments", AdminGetDepartmentTree)
			admin.POST("/admin/departments", AdminCreateDepartment)
			admin.PUT("/admin/departments/:id", AdminUpdateDepartment)
			admin.DELETE("/admin/departments/:id", AdminDeleteDepartment)
			admin.GET("/admin/departments/:id/members", AdminGetDepartmentMembers)

			// Settings
			admin.GET("/admin/settings/ai", GetAIModelSettings)
			admin.POST("/admin/settings/ai", SaveAIModelSettings)
			admin.POST("/admin/settings/ai/detect", DetectAIModels)
			admin.POST("/admin/settings/ai/test", TestAIModel)
			admin.GET("/admin/settings/storage", GetStorageSettings)
			admin.POST("/admin/settings/storage", SaveStorageSettings)
			admin.POST("/admin/settings/storage/test", TestRemoteStorage)

			// Storage Statistics (Admin)
			admin.GET("/admin/storage/usage", GetAdminStorageUsage)

			// Remote Storage Management
			admin.GET("/admin/remote-storages", ListRemoteStorages)
			admin.POST("/admin/remote-storages", CreateRemoteStorage)
			admin.PUT("/admin/remote-storages/:id", UpdateRemoteStorage)
			admin.DELETE("/admin/remote-storages/:id", DeleteRemoteStorage)
			admin.POST("/admin/remote-storages/:id/test", TestRemoteStorageConnection)
			admin.POST("/admin/remote-storages/:id/connect", ConnectRemoteStorage)
			admin.POST("/admin/remote-storages/:id/disconnect", DisconnectRemoteStorage)

			// LDAP/AD
			admin.POST("/admin/ldap/test", TestLdapConnection)
			admin.POST("/admin/ldap/users", FetchLdapUsers)
			admin.POST("/admin/ldap/import", ImportLdapUsers)
		}
	}

	// OnlyOffice callback (no auth required - OnlyOffice server calls this)
	api.POST("/onlyoffice/callback/:docId", OnlyOfficeCallback)
	api.GET("/onlyoffice/download/:docId", DownloadDocumentForOnlyOffice)

	// WebSocket
	r.GET("/ws/doc/:id", func(c *gin.Context) {
		ws.HandleWebSocket(hub, c)
	})
	r.GET("/ws/yjs/:id", func(c *gin.Context) {
		ws.HandleYjsWebSocket(hub, c)
	})
}
