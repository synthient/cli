package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/synthient/cli/internal/access"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/go-synthient/v2"
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
	if errors.Is(err, auth.ErrNoCredentials) || errors.Is(err, synthient.ErrNoToken) {
		return "Missing API key. Run `synthient auth`, set SYNTHIENT_API_KEY, or add SYNTHIENT_API_KEY to .env."
	}
	if errors.Is(err, synthient.ErrUnauthorized) {
		return "Invalid API key. Run `synthient auth` to replace the stored key or check SYNTHIENT_API_KEY."
	}
	if errors.Is(err, synthient.ErrPaymentRequired) {
		return "Lookup quota exhausted. Run `synthient account` to inspect remaining credits and reset timing."
	}
	if errors.Is(err, synthient.ErrBadRequest) {
		return "Bad request. Check the input values and try again."
	}
	if errors.Is(err, synthient.ErrInternalServerError) {
		return "Synthient returned a server error. Retry with backoff, and contact support if it persists."
	}
	if errors.Is(err, synthient.ErrUnexpectedStatusCode) {
		text := err.Error()
		if strings.Contains(text, "403") {
			return "This API key is authenticated but does not have permission for this resource. Run `synthient scopes` to inspect required scopes."
		}
		if strings.Contains(text, "404") {
			return "Resource not found. Check the stream name, snapshot id, IP, or domain."
		}
		if strings.Contains(text, "429") {
			return "Rate limited. Retry later using exponential backoff."
		}
		return "Synthient returned an unexpected response. Re-run with SYNTHIENT_DEBUG=true for details."
	}
	return ""
}
