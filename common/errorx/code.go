package errorx

import "errors"

//200	OK	无错误。
//400	INVALID_ARGUMENT	客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。
//400	FAILED_PRECONDITION	请求无法在当前系统状态下执行，例如删除非空目录。
//400	OUT_OF_RANGE	客户端指定了无效范围。
//401	UNAUTHENTICATED	由于 OAuth 令牌丢失、无效或过期，请求未通过身份验证。
//403	PERMISSION_DENIED	客户端权限不足。这可能是因为 OAuth 令牌没有正确的范围、客户端没有权限或者 API 尚未启用。
//404	NOT_FOUND	未找到指定的资源。
//409	ABORTED	并发冲突，例如读取/修改/写入冲突。
//409	ALREADY_EXISTS	客户端尝试创建的资源已存在。
//429	RESOURCE_EXHAUSTED	资源配额不足或达到速率限制。如需了解详情，客户端应该查找 google.rpc.QuotaFailure 错误详细信息。
//499	CANCELLED	请求被客户端取消。
//500	DATA_LOSS	出现不可恢复的数据丢失或数据损坏。客户端应该向用户报告错误。
//500	UNKNOWN	出现未知的服务器错误。通常是服务器错误。
//500	INTERNAL	出现内部服务器错误。通常是服务器错误。
//501	NOT_IMPLEMENTED	API 方法未通过服务器实现。
//502	不适用	到达服务器前发生网络错误。通常是网络中断或配置错误。
//503	UNAVAILABLE	服务不可用。通常是服务器已关闭。
//504	DEADLINE_EXCEEDED	超出请求时限。仅当调用者设置的时限比方法的默认时限短（即请求的时限不足以让服务器处理请求）并且请求未在时限范围内完成时，才会发生这种情况。
var (
	InvalidArgument     = NewCodeError(400, "INVALID_ARGUMENT", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	FAILED_PRECONDITION = NewCodeError(400, "FAILED_PRECONDITION", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	OUT_OF_RANGE        = NewCodeError(400, "OUT_OF_RANGE", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	UNAUTHENTICATED     = NewCodeError(400, "UNAUTHENTICATED", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	PERMISSION_DENIED   = NewCodeError(400, "PERMISSION_DENIED", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	NOT_FOUND           = NewCodeError(400, "NOT_FOUND", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	ABORTED             = NewCodeError(400, "ABORTED", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	ALREADY_EXISTS      = NewCodeError(400, "ALREADY_EXISTS", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	RESOURCE_EXHAUSTED  = NewCodeError(400, "RESOURCE_EXHAUSTED", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	CANCELLED           = NewCodeError(400, "CANCELLED", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	DATA_LOSS           = NewCodeError(400, "DATA_LOSS", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	UNKNOWN             = NewCodeError(400, "UNKNOWN", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	INTERNAL            = NewCodeError(400, "INTERNAL", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	NOT_IMPLEMENTED     = NewCodeError(400, "NOT_IMPLEMENTED", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	UNAVAILABLE         = NewCodeError(400, "UNAVAILABLE", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
	DEADLINE_EXCEEDED   = NewCodeError(400, "DEADLINE_EXCEEDED", errors.New("客户端指定了无效参数。如需了解详情，请查看错误消息和错误详细信息。"))
)
