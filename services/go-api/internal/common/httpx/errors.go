package httpx

const (
	CodeOK              = 0
	CodeUnauthorized    = 40101
	CodeTokenExpired    = 40102
	CodeForbidden       = 40301
	CodeInviteRequired  = 40321
	CodeNotFound        = 40400
	CodeConflict        = 40900
	CodeValidationError = 42200
	CodeSystemError     = 50000
	CodeInvalidRequest  = CodeValidationError
	CodeInternalError   = CodeSystemError
)
