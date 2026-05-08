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

	// Authenticated routes (全局限流: 每秒30请求，突发50)
	auth := api.Group("", middleware.Auth(), middleware.RateLimit(30, 50), middleware.AuditLog())
	{
		// User
		auth.GET("/user/info", GetUserInfo)
		auth.POST("/user/update", UpdateUserInfo)
		auth.POST("/user/change-password", ChangeUserPassword)
		auth.GET("/user/list", GetUserList)
		auth.GET("/users/search", SearchUsers)

		// Document
		auth.GET("/document/tree", GetDocumentTree)
		auth.GET("/document/:id", GetDocumentDetail)
		auth.POST("/document/create", CreateDocument)
		auth.POST("/document/:id/update", UpdateDocument)
		auth.POST("/document/:id/content", SaveDocumentContent)
		auth.POST("/document/:id/delete", DeleteDocument)
		auth.POST("/document/:id/restore", RestoreDocument)
		auth.DELETE("/document/:id", PermanentDeleteDocument)
		auth.DELETE("/document/:id/permanent", PermanentDelete)
		auth.POST("/document/batch-restore", BatchRestore)
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
		auth.POST("/document/search", EnhancedSearchDocuments)
		auth.GET("/search/suggestions", SearchSuggestions)
		auth.GET("/document/:id/versions", GetDocumentVersions)
		auth.GET("/document/:id/versions/diff", GetVersionDiff)
		auth.POST("/document/:id/rollback", RollbackVersion)
		auth.POST("/document/import", ImportDocument)
		auth.GET("/document/:id/export", ExportDocument)
		auth.GET("/document/:id/file-preview", GetDocumentFilePreview)
		auth.GET("/document/:id/download", DownloadDocumentFile)
		auth.POST("/document/:id/save-as-template", SaveDocumentAsTemplate)
		auth.POST("/document/shortcut", AddDocumentShortcut)
		auth.POST("/document/migrate", MigrateDocuments)

		// Permission
		auth.GET("/auth/document/:id", GetDocumentPermissions)
		auth.POST("/auth/permission/set", SetPermission)
		auth.POST("/auth/permission/remove", RemovePermission)
		auth.GET("/auth/share/:id", GetShareSetting)
		auth.POST("/auth/share/update", UpdateShareSetting)
		auth.GET("/auth/check/:id", CheckPermission)

		// Knowledge Conversations
		auth.GET("/knowledge/conversations", GetConversationListHandler)
		auth.POST("/knowledge/conversations", CreateConversationHandler)
		auth.GET("/knowledge/conversations/:id", GetConversationDetailHandler)
		auth.PUT("/knowledge/conversations/:id", UpdateConversationHandler)
		auth.DELETE("/knowledge/conversations/:id", DeleteConversationHandler)
		auth.DELETE("/knowledge/conversations", BatchDeleteConversationsHandler)
		auth.GET("/knowledge/conversations/:id/export", ExportConversationHandler)
		auth.POST("/knowledge/conversations/:id/share", CreateShareLinkHandler)
		auth.DELETE("/knowledge/conversations/:id/share", RevokeShareLinkHandler)

		// Knowledge
		auth.GET("/knowledge/list", GetKnowledgeBaseList)
		auth.GET("/knowledge/:id", GetKnowledgeBaseDetail)
		auth.POST("/knowledge/create", CreateKnowledgeBase)
		auth.POST("/knowledge/:id/update", UpdateKnowledgeBase)
		auth.DELETE("/knowledge/:id", DeleteKnowledgeBase)
		auth.GET("/knowledge/:id/tree", GetKnowledgeBaseTree)
		auth.GET("/knowledge/:id/export", ExportKnowledgeBase)
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
		// Knowledge AI chat (AI限流: 每秒2请求，突发5)
		auth.POST("/knowledge/:id/chat", middleware.AIRateLimit(), KnowledgeChat)
		auth.POST("/knowledge/chat/global", middleware.AIRateLimit(), KnowledgeChatGlobal)
		auth.POST("/knowledge/chat/stream", middleware.AIRateLimit(), StreamChatHandler)
		// Knowledge Graph (重型任务并发限制: 最多3个)
		auth.POST("/knowledge/:id/graph/build", middleware.HeavyTaskLimit(3), BuildKnowledgeGraph)
		auth.GET("/knowledge/:id/graph", GetKnowledgeGraph)
		auth.GET("/knowledge/:id/graph/status", GetKnowledgeGraphStatus)
		// Knowledge RAPTOR and embedding (重型任务并发限制: 最多3个)
		auth.POST("/knowledge/:id/raptor/build", middleware.HeavyTaskLimit(3), BuildRaptorTree)
		auth.GET("/knowledge/:id/raptor/stats", GetRaptorTreeStats)
		auth.GET("/knowledge/:id/embedding/status", GetEmbeddingStatus)
		auth.POST("/knowledge/:id/embedding/rebuild", middleware.HeavyTaskLimit(3), RebuildEmbeddings)
		// Knowledge Feedback
		auth.POST("/knowledge/messages/:id/feedback", SubmitFeedbackHandler)
		auth.GET("/knowledge/messages/:id/feedback", GetFeedbackHandler)

		// Templates
		auth.GET("/template/list", GetTemplates)
		auth.GET("/template/:id", GetTemplateDetail)
		auth.POST("/template/:id/use", UseTemplate)
		auth.POST("/template/:id/favorite", FavoriteTemplateHandler)
		auth.DELETE("/template/:id/favorite", UnfavoriteTemplateHandler)

		// Collaborate
		auth.POST("/collaborate/comment/add", AddComment)
		auth.GET("/collaborate/comment/:id", GetComments)
		auth.POST("/collaborate/comment/:id/resolve", ResolveComment)
		auth.POST("/collaborate/comment/:id/delete", DeleteComment)
		auth.GET("/collaborate/notifications", GetNotifications)
		auth.POST("/collaborate/notification/:id/read", MarkNotificationRead)
		auth.POST("/collaborate/notification/read-all", MarkAllNotificationsRead)
		auth.POST("/collaborate/notification/create", CreateSystemNotification)
		auth.POST("/collaborate/invite", InviteCollaborators)

		// Comment (new routes)
		auth.POST("/document/:id/comments", CreateCommentHandler)
		auth.GET("/document/:id/comments", ListCommentsHandler)
		auth.DELETE("/comment/:commentId", DeleteCommentHandler)
		auth.PUT("/comment/:commentId/resolve", ResolveCommentHandler)

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

			// Dashboard
			admin.GET("/admin/dashboard/stats", GetDashboardStats)
			admin.GET("/admin/system/info", GetSystemInfo)

			// Directory listing for path selection
			admin.GET("/admin/directories", AdminListDirectories)

			// LDAP/AD
			admin.POST("/admin/ldap/test", TestLdapConnection)
			admin.POST("/admin/ldap/users", FetchLdapUsers)
			admin.POST("/admin/ldap/import", ImportLdapUsers)

			// Template Management
			admin.GET("/admin/templates", GetAllTemplatesHandler)
			admin.POST("/admin/template/create", CreateTemplateHandler)
			admin.PUT("/admin/template/:id", UpdateTemplateHandler)
			admin.DELETE("/admin/template/:id", DeleteTemplateHandler)

			// Document cleanup (admin only)
			admin.POST("/admin/document/cleanup-expired", CleanupExpired)

			// Follow-up suggestions toggle
			admin.GET("/admin/settings/follow-up-suggestions", GetFollowUpSuggestionsHandler)
			admin.PUT("/admin/settings/follow-up-suggestions", UpdateFollowUpSuggestionsHandler)

			// Operation Logs (admin)
			admin.GET("/admin/operation-logs", ListAllOperationLogs)

			// Feedback Reviews
			feedbackReviews := admin.Group("/admin/feedback/reviews")
			{
				feedbackReviews.GET("", GetPendingReviewsHandler)
				feedbackReviews.POST("/:id/approve", ApproveReviewHandler)
				feedbackReviews.POST("/:id/reject", RejectReviewHandler)
				feedbackReviews.GET("/stats", GetReviewStatsHandler)
			}
		}
	}

	// File content streaming (supports token via query parameter for iframe embedding)
	api.GET("/document/:id/file-content", middleware.AuthQueryToken(), GetDocumentFileContent)

	// OnlyOffice callback (no auth required - OnlyOffice server calls this)
	api.POST("/onlyoffice/callback/:docId", OnlyOfficeCallback)
	api.GET("/onlyoffice/download/:docId", DownloadDocumentForOnlyOffice)

	// Public share link (no auth required)
	api.GET("/knowledge/share/:token", GetSharedConversationHandler)

	// WebSocket
	r.GET("/ws/doc/:id", func(c *gin.Context) {
		ws.HandleWebSocket(hub, c)
	})
	r.GET("/ws/yjs/:id", func(c *gin.Context) {
		ws.HandleYjsWebSocket(hub, c)
	})
}
