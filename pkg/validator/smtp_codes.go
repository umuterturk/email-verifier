package validator

// SMTP response code constants
const (
	SMTPCodeOK                = 250 // Requested action completed
	SMTPCodeUserNotLocal      = 251 // User not local, will forward
	SMTPCodeStartMail         = 354 // Start mail input
	SMTPCodeServiceNotAvail   = 421 // Service not available (greylisting)
	SMTPCodeMailboxBusy       = 450 // Mailbox busy (greylisting)
	SMTPCodeLocalError        = 451 // Local error (greylisting)
	SMTPCodeInsufficientSpace = 452 // Insufficient storage
	SMTPCodeCommandError      = 500 // Syntax error
	SMTPCodeParamError        = 501 // Syntax error in parameters
	SMTPCodeNotImplemented    = 502 // Command not implemented
	SMTPCodeBadSequence       = 503 // Bad sequence of commands
	SMTPCodeParamNotImpl      = 504 // Parameter not implemented
	SMTPCodeMailboxNotFound   = 550 // Mailbox not found
	SMTPCodeUserNotLocal550   = 551 // User not local
	SMTPCodeExceededStorage   = 552 // Exceeded storage allocation
	SMTPCodeMailboxNameErr    = 553 // Mailbox name not allowed
	SMTPCodeTransactionFail   = 554 // Transaction failed
)

// SMTPResponseCategory categorizes SMTP responses
type SMTPResponseCategory int

const (
	SMTPCategorySuccess   SMTPResponseCategory = iota // 2xx responses
	SMTPCategoryTemporary                             // 4xx responses (greylisting, rate limiting)
	SMTPCategoryPermanent                             // 5xx responses (mailbox doesn't exist)
	SMTPCategoryUnknown                               // Other responses
)

// CategorizeResponse determines the category of an SMTP response code
func CategorizeResponse(code int) SMTPResponseCategory {
	switch {
	case code >= 200 && code < 300:
		return SMTPCategorySuccess
	case code >= 400 && code < 500:
		return SMTPCategoryTemporary
	case code >= 500 && code < 600:
		return SMTPCategoryPermanent
	default:
		return SMTPCategoryUnknown
	}
}

// IsGreylisting checks if response indicates greylisting or temporary rejection
func IsGreylisting(code int) bool {
	return code == SMTPCodeServiceNotAvail ||
		code == SMTPCodeMailboxBusy ||
		code == SMTPCodeLocalError
}

// IsMailboxNotFound checks if response indicates mailbox doesn't exist
func IsMailboxNotFound(code int) bool {
	return code == SMTPCodeMailboxNotFound ||
		code == SMTPCodeMailboxNameErr ||
		code == SMTPCodeUserNotLocal550
}
