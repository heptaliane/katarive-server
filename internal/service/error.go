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
