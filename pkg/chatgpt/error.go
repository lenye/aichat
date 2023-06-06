package chatgpt

import "errors"

var (
	ErrInvalidAPIKey  = errors.New("invalid api_key")
	ErrInvalidAPIType = errors.New("invalid api_type")
	ErrInvalidBaseURL = errors.New("invalid base URL")
	ErrInvalidProxy   = errors.New("invalid proxy")
)
