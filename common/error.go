package common

import "fmt"

// OodlesError is an error that should be shown to the user directly
type OodlesError string

// Error returns an OodlesError
func Error(tmpl string, args ...interface{}) OodlesError {
	return OodlesError(fmt.Sprintf(tmpl, args...))
}

func (e OodlesError) Error() string {
	return string(e)
}

// IsOodlesError returns true if the given error is an OodlesError (user error)
func IsOodlesError(err error) bool {
	_, ok := err.(*OodlesError)
	if ok {
		return true
	}

	_, ok = err.(OodlesError)
	return ok
}
