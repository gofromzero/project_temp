package errorx

const defaultCode = 500

type CodeError struct {
	Code    int     `json:"code"`
	Msg     string  `json:"msg"`
	Details []error `json:"details"`
}

type CodeErrorResponse struct {
	Code    int      `json:"code"`
	Msg     string   `json:"msg"`
	Details []string `json:"details"`
}

func NewCodeError(code int, msg string, details ...error) *CodeError {
	return &CodeError{Code: code, Msg: msg, Details: details}
}

func NewDefaultError(msg string) error {
	return NewCodeError(defaultCode, msg)
}
func (e *CodeError) With(err ...error) *CodeError {
	e.Details = append(e.Details, err...)
	return e
}

func (e *CodeError) WithMsg(msg string) *CodeError {
	e.Msg = msg
	return e
}

func (e *CodeError) Error() string {
	return e.Msg
}

func (e *CodeError) Data() *CodeErrorResponse {
	details := make([]string, 0, len(e.Details))
	for _, err := range e.Details {
		details = append(details, err.Error())
	}
	return &CodeErrorResponse{
		Code:    e.Code,
		Msg:     e.Msg,
		Details: details,
	}
}
