package spidererrors

import "net/http"

type SpiderError struct {
	Message    string
	Code       string
	StatusCode int
}

func (e *SpiderError) Error() string { return e.Message }

func New(message, code string, status int) *SpiderError {
	return &SpiderError{Message: message, Code: code, StatusCode: status}
}

func Authentication(message string) *SpiderError {
	if message == "" {
		message = "Invalid email or password"
	}
	return New(message, "unauthenticated", http.StatusUnauthorized)
}

func Authorization(message string) *SpiderError {
	return New(message, "forbidden", http.StatusForbidden)
}

func NotFound(message string) *SpiderError {
	return New(message, "not_found", http.StatusNotFound)
}

func Validation(message string) *SpiderError {
	return New(message, "validation_error", http.StatusUnprocessableEntity)
}

func WorkerAuth(message string) *SpiderError {
	return New(message, "worker_unauthorized", http.StatusUnauthorized)
}

func Pipeline(message string) *SpiderError {
	return New(message, "security_pipeline_error", http.StatusInternalServerError)
}

func Serving(message string) *SpiderError {
	return New(message, "serving_error", http.StatusBadGateway)
}
