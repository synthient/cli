package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/synthient/cli/internal/access"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/feed"
	"github.com/synthient/go-synthient/v2"
	"go.mattglei.ch/timber"
)

func Fatal(err error, msg string, attrs ...timber.Attr) {
	if err == nil {
		return
	}
	if text := Explain(err); text != "" {
		timber.FatalMsg(text, attrs...)
	}
	timber.Fatal(err, msg, attrs...)
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

	var statusErr feed.StatusError
	if errors.As(err, &statusErr) {
		return explainStatus(statusErr)
	}
	return ""
}

func explainStatus(err feed.StatusError) string {
	detail := statusDetail(err.Body)
	switch err.StatusCode {
	case 400:
		return withDetail("Bad request. Check the input values before retrying.", detail)
	case 401:
		return withDetail("Missing or invalid API key. Run `synthient auth` or check SYNTHIENT_API_KEY.", detail)
	case 402:
		return withDetail("Lookup quota exhausted. Run `synthient account` to inspect remaining credits.", detail)
	case 403:
		return withDetail("This API key is authenticated but does not have permission for this resource. Run `synthient scopes` to inspect required feed scopes.", detail)
	case 404:
		return withDetail("Resource not found. Check the stream name and snapshot id.", detail)
	case 429:
		msg := "Rate limited. Retry later using exponential backoff."
		if err.RetryAfter != "" {
			msg = fmt.Sprintf("%s Retry after %s seconds.", msg, err.RetryAfter)
		}
		return withDetail(msg, detail)
	case 500, 502, 503, 504:
		return withDetail("Synthient returned a transient server error. Retry with backoff.", detail)
	default:
		return withDetail(err.Status, detail)
	}
}

func statusDetail(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	var data struct {
		Detail string              `json:"detail"`
		Title  string              `json:"title"`
		Errors map[string][]string `json:"errors"`
	}
	if json.Unmarshal([]byte(body), &data) != nil {
		return body
	}
	if data.Detail != "" {
		return data.Detail
	}
	if data.Title != "" {
		if len(data.Errors) == 0 {
			return data.Title
		}
		parts := []string{data.Title}
		for field, values := range data.Errors {
			parts = append(parts, fmt.Sprintf("%s: %s", field, strings.Join(values, ", ")))
		}
		return strings.Join(parts, "; ")
	}
	return body
}

func withDetail(msg string, detail string) string {
	if strings.TrimSpace(detail) == "" {
		return msg
	}
	return fmt.Sprintf("%s %s", msg, detail)
}
