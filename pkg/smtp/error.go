package smtp

type ServerError struct {
	Message string
}
type SyntaxError struct {
	Message string
}
type OutOfOrderCmdError struct {
	Message string
}
type InvalidCredentialError struct {
	Message string
}

type AuthRequiredError struct {
	Message string
}

func NewSyntaxError(m string) SyntaxError {
	err := SyntaxError{}
	err.Message = m
	return err
}
func (e SyntaxError) Error() string {
	return e.Message
}
func NewServerError(m string) ServerError {
	err := ServerError{}
	err.Message = m
	return err
}
func (e ServerError) Error() string {
	return e.Message
}
func NewOutOfOrderCmdError(m string) OutOfOrderCmdError {
	err := OutOfOrderCmdError{}
	err.Message = m
	return err
}
func (e OutOfOrderCmdError) Error() string {
	return e.Message
}

func NewInvalidCredentialError(m string) InvalidCredentialError {
	err := InvalidCredentialError{}
	err.Message = m
	return err
}
func (e InvalidCredentialError) Error() string {
	return e.Message
}
func NewAuthRequiredError(m string) AuthRequiredError {
	err := AuthRequiredError{}
	err.Message = m
	return err
}
func (e AuthRequiredError) Error() string {
	return e.Message
}
