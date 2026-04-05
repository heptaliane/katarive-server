package service

import (
	"fmt"
)

type UnsupportedSourceURLError struct {
	URL string
}

func (e *UnsupportedSourceURLError) Error() string {
	return fmt.Sprintf("No source for '%s' is available.", e.URL)
}

var _ error = new(UnsupportedSourceURLError)

type NarrateError struct {
	Reason string
}

func (e *NarrateError) Error() string {
	return fmt.Sprintf("Narrate failed with error: %s", e.Reason)
}

var _ error = new(NarrateError)
