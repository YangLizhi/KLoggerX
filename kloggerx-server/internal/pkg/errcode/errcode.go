package errcode

const (
	Success        = 0
	ErrBadRequest  = 400
	ErrUnauthorized = 401
	ErrForbidden   = 403
	ErrNotFound    = 404
	ErrTooManyReqs = 429
	ErrInternal    = 500

	// 业务错误码
	ErrDocNotFound      = 10001
	ErrDocPermission    = 10002
	ErrKbNotFound       = 10003
	ErrKbPermission     = 10004
	ErrUserNotFound     = 10005
	ErrInvalidPassword  = 10006
	ErrStorageConnect   = 10007
	ErrFileTypeNotAllow = 10008
	ErrFileTooLarge     = 10009
)

var messages = map[int]string{
	Success:             "成功",
	ErrBadRequest:       "请求参数错误",
	ErrUnauthorized:     "未授权，请先登录",
	ErrForbidden:        "无权限执行此操作",
	ErrNotFound:         "资源不存在",
	ErrTooManyReqs:      "请求过于频繁，请稍后再试",
	ErrInternal:         "服务器内部错误",
	ErrDocNotFound:      "文档不存在",
	ErrDocPermission:    "无文档操作权限",
	ErrKbNotFound:       "知识库不存在",
	ErrKbPermission:     "无知识库操作权限",
	ErrUserNotFound:     "用户不存在",
	ErrInvalidPassword:  "密码错误",
	ErrStorageConnect:   "存储连接失败",
	ErrFileTypeNotAllow: "文件类型不允许",
	ErrFileTooLarge:     "文件大小超出限制",
}

// Message returns the message corresponding to the given error code.
func Message(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "未知错误"
}
