package smtp

const (
	StatusReady                  = 220
	StatusAuthSuccess            = 235
	StatusOk                     = 250
	StatusClose                  = 221
	StatusAuthChallenge          = 334
	StatusContinue               = 354
	StatusTempAuthError          = 454
	StatusSyntaxError            = 501
	StatusCommandNotImplemented  = 502
	StatusOutOfSequenceCmdError  = 503
	StatusAuthRequired           = 503
	StatusInvalidCredentialError = 535
	StatusTLSRequired            = 538
	StatusUnknownError           = 554
)
