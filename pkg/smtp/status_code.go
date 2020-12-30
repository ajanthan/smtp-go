package smtp

const (
	StatusReady                 = 220
	StatusOk                    = 250
	StatusClose                 = 221
	StatusContinue              = 354
	StatusSyntaxError           = 501
	StatusCommandNotImplemented = 502
	StatusOutOfSequenceCmdError = 503
	StatusUnknownError          = 554
)
