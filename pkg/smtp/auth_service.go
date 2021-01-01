package smtp

type AuthenticationService interface {
	Authenticate(username string, password []byte) error
	ValidateHMAC(username string, msg []byte, code []byte) error
}
