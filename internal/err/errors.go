package err

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

var (
	ErrEmailAlreadyExists = &Error{
		Code:    CodeEmailAlreadyExists,
		Message: "email already exists",
	}

	ErrUnauthorized = &Error{
		Code:    CodeUnauthorized,
		Message: "unauthorized",
	}
)
