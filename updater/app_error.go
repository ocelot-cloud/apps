package main

// AppError is returned by helpers to provide a message for the health report
// and the original underlying error.
type AppError struct {
	message string
	err     error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.err == nil {
		return e.message
	}
	return e.message + ": " + e.err.Error()
}
