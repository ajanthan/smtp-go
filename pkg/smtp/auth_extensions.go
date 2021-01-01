package smtp

import (
	"bytes"
	"encoding/base64"
)

var NullByte = []byte("\x00")

func HandlePlainAuth(cred string, authService AuthenticationService) error {

	credential, err := base64Decode(cred)
	if err != nil {
		return NewServerError(err.Error())
	}
	parts := bytes.Split(credential, NullByte)
	//ignoring parts[0]=> identity
	return authService.Authenticate(string(parts[1]), parts[2])
}
func HandleLoginAuth(username, password string, authService AuthenticationService) error {
	decodedUsername, err := base64Decode(username)
	if err != nil {
		return NewServerError(err.Error())
	}
	decodedPassword, err := base64Decode(password)
	if err != nil {
		return NewServerError(err.Error())
	}

	return authService.Authenticate(string(decodedUsername), decodedPassword)
}
func HandleMD5CRAMAuth(cred string, challenge []byte, authService AuthenticationService) error {
	decodedCred, err := base64Decode(cred)
	if err != nil {
		return NewServerError(err.Error())
	}
	creds := bytes.Split(decodedCred, []byte{' '})
	if len(creds) != 2 {
		return NewServerError("invalid MD5-CRAM format")
	}
	return authService.ValidateHMAC(string(creds[0]), []byte(challenge), creds[1])
}

func base64Decode(in string) ([]byte, error) {
	data := []byte(in)
	decodedData := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	l, err := base64.StdEncoding.Decode(decodedData, data)
	if err != nil {
		return []byte{}, NewServerError(err.Error())
	}
	return decodedData[:l], err
}
func base64Encode(in string) []byte {
	data := []byte(in)
	encodedData := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encodedData, data)
	return encodedData
}
