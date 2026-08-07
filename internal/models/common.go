package models

type (
	// Message use for notification responses
	Message struct {
		Message string `json:"message"` // Notification message
	}

	// HTTPError use for error responses
	HTTPError struct {
		Error   string `json:"error"`   // Shorthand error
		Message string `json:"message"` // Pretty error message
	}

	// ErrorDetail points at the exact field that failed. Without it a 422 on a
	// tour with twenty hints tells the admin only that something, somewhere, is
	// wrong.
	ErrorDetail struct {
		Path    string `json:"path" example:"hints[1].selector"`
		Message string `json:"message" example:"required unless placement is center"`
	}

	// ErrorBody is the single error shape every endpoint answers with. The code
	// is machine readable so a client can branch on it without parsing prose.
	ErrorBody struct {
		Code    string        `json:"code" example:"VALIDATION_FAILED"`
		Message string        `json:"message" example:"Tour cannot be published"`
		Details []ErrorDetail `json:"details,omitempty"`
	}

	// ErrorEnvelope wraps ErrorBody so that a successful payload and an error
	// payload can never be confused for one another.
	ErrorEnvelope struct {
		Error ErrorBody `json:"error"`
	}
)

// Error codes. Chosen once here so that handlers cannot invent variants.
const (
	CodeBadRequest   = "BAD_REQUEST"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeNotFound     = "NOT_FOUND"
	CodeConflict     = "CONFLICT"
	CodeValidation   = "VALIDATION_FAILED"
	CodeRateLimited  = "RATE_LIMITED"
	CodeInternal     = "INTERNAL"
)

// CodeForStatus is the fallback for errors that do not name their own code.
func CodeForStatus(status int) string {
	switch status {
	case 400:
		return CodeBadRequest
	case 401:
		return CodeUnauthorized
	case 403:
		return CodeForbidden
	case 404:
		return CodeNotFound
	case 409:
		return CodeConflict
	case 422:
		return CodeValidation
	case 429:
		return CodeRateLimited
	default:
		return CodeInternal
	}
}
