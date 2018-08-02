package middlewares

type ResState int

const (
	ResStateOk                  = 0
	ResStateForbidden           = 403
	ResStateNotFound            = 404
	ResStateRefererNotFound     = 409
	ResStateIPBlocked           = 410
	ResStateUserIdBlocked       = 411
	ResStateReqExceeded         = 429
	ResStateLoginFailed         = 501
	ResStateBadGateway          = 502
	ResStateBackendUnavailable  = 503
	ResStateBackendTimeout      = 504
	ResStateLoginFailedExceed   = 505
	ResStateCacheUserFaild      = 506
	ResStateLoginFailedPassword = 507
	ResStateInternalError       = 508
	ResStateUnauthorized        = 5
)

func (rs ResState) String() string {
	switch rs {
	case ResStateOk:
		return "OK"
	case ResStateUnauthorized:
		return "尚未登录"
	case ResStateLoginFailed:
		return "登录失败"
	case ResStateLoginFailedExceed:
		return "登录失败次数过多"
	case ResStateLoginFailedPassword:
		return "密码错误"
	case ResStateCacheUserFaild:
		return "验证用户失败"
	case ResStateForbidden:
		return "禁止访问"
	case ResStateNotFound:
		return "不存在的资源"
	case ResStateBackendTimeout:
		return "服务超时"
	case ResStateBackendUnavailable:
		return "服务不可用"
	case ResStateBadGateway:
		return "网关异常"
	case ResStateRefererNotFound:
		return "Referer丢失"
	case ResStateIPBlocked:
		return "您被禁止访问该资源"
	case ResStateUserIdBlocked:
		return "您被禁止访问该资源"
	case ResStateReqExceeded:
		return "访问频率过高"
	case ResStateInternalError:
		return "内部错误"
	}
	return ""
}

type ResponseState struct {
	Code    ResState `json:"code"`
	Message string   `json:"msg"`
}

type CommResponse struct {
	Status ResponseState `json:"status"`
	Data   interface{}   `json:"data"`
}

func makeResponse(code ResState, data interface{}) CommResponse {
	res := CommResponse{
		Status: ResponseState{
			Code:    code,
			Message: code.String(),
		},
	}

	if data != nil {
		res.Data = data
	}
	return res
}
