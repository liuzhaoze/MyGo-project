package consts

const (
	ErrnoSuccess      = 0
	ErrnoUnknownError = 1

	ErrnoBindRequestError     = 1000
	ErrorRequestValidateError = 1001
)

var ErrMsg = map[int]string{
	ErrnoSuccess:      "success",
	ErrnoUnknownError: "unknown error",

	ErrnoBindRequestError:     "binding request error",
	ErrorRequestValidateError: "validate request error",
}
