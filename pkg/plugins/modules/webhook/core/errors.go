package core

import "errors"

// Common webhook errors
var (
	ErrInvalidPlatform = errors.New("invalid or empty platform")
	ErrInvalidURL      = errors.New("invalid or empty target URL")
	ErrSenderNotFound  = errors.New("sender not found for platform")
	ErrSendFailed      = errors.New("failed to send webhook")
	ErrInvalidConfig   = errors.New("invalid webhook configuration")
)

