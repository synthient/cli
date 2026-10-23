package app

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/synthient/cli/internal/access"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/synthient-go"
	"go.mattglei.ch/timber"
)

func Fatal(err error, msg string) {
	if err == nil {
		return
	}
	if text := Explain(err); text != "" {
		timber.FatalMsg(text)
	}
	timber.Fatal(err, msg)
}

func Explain(err error) string {
	var missing access.MissingScopeError
	if errors.As(err, &missing) {
		return fmt.Sprintf("Missing scope: %s. This key can authenticate, but it cannot access this resource. Run `synthient scopes` to inspect required scopes.", missing.Scope)
	}
	if errors.Is(err, auth.ErrNoCredentials) || errors.Is(err, synthient.ErrMissingAPIKey) {
		return "Missing API key. Run `synthient auth`, set SYNTHIENT_API_KEY, or add SYNTHIENT_API_KEY to .env."
	}
	apiErr, ok := errors.AsType[*synthient.APIError](err)
	if !ok {
		return ""
	}
	switch apiErr.StatusCode {
	case http.StatusUnauthorized:
		return "Invalid API key. Run `synthient auth` to replace the stored key or check SYNTHIENT_API_KEY."
	case http.StatusPaymentRequired:
		return "Lookup quota exhausted. Run `synthient account` to inspect remaining credits and reset timing."
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return "Bad request. Check the input values and try again."
	case http.StatusForbidden:
		return "This API key is authenticated but does not have permission for this resource. Run `synthient scopes` to inspect required scopes."
	case http.StatusNotFound:
		return "Resource not found. Check the stream name, snapshot id, IP, or domain."
	case http.StatusTooManyRequests:
		return "Rate limited. Retry later using exponential backoff."
	}
	if apiErr.StatusCode >= http.StatusInternalServerError {
		return "Synthient returned a server error. Retry with backoff, and contact support if it persists."
	}
	return fmt.Sprintf("Synthient returned an unexpected response (HTTP %d).", apiErr.StatusCode)
}
