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

type UnexpectedTypeError struct {
	Value    any
	Expected any
}

func (e *UnexpectedTypeError) Error() string {
	return fmt.Sprintf("Unexpected type is detected. Expected %T but got %T",
		e.Value,
		e.Expected,
	)
}

var _ error = new(UnexpectedTypeError)

type JobNotFoundError struct {
	Id string
}

func (e *JobNotFoundError) Error() string {
	return fmt.Sprintf("No job is found for %s", e.Id)
}

var _ error = new(JobNotFoundError)
