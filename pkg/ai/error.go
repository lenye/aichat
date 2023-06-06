package ai

import "errors"

var (
	ErrInvalidAPIKey  = errors.New("invalid api key")
	ErrInvalidAPIType = errors.New("invalid api type")
	ErrInvalidBaseURL = errors.New("invalid base URL")
	ErrInvalidProxy   = errors.New("invalid proxy")
)
