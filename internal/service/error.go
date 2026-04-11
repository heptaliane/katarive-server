package service

import (
	"errors"
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

type JobNotFoundError struct {
	JobId string
}

func (e *JobNotFoundError) Error() string {
	return fmt.Sprintf("No job is found for %s", e.JobId)
}

var _ error = new(JobNotFoundError)

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

var UnspecifiedNarratorError = errors.New("No narrator is set")
