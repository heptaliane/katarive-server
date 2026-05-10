package model

import (
	"fmt"
)

type UnsupportedSourceURLError struct {
	Url string
}

func (e *UnsupportedSourceURLError) Error() string {
	return fmt.Sprintf("No source for '%s' is available.", e.Url)
}

var _ error = new(UnsupportedSourceURLError)
