package internal

import (
	"time"

	"github.com/lestrrat-go/backoff"
	"google.golang.org/api/googleapi"
)

var defaultRetryPolicy = backoff.NewExponential(
	backoff.WithInterval(3*time.Second),
	backoff.WithMaxRetries(5),
)

// IsRetryableError returns true if the error is retryable,
// such as status code is 5xx or network error occurs.
// Otherwise returns false.
// See https://developers.google.com/photos/library/guides/best-practices#retrying-failed-requests
func IsRetryableError(err error) bool {
	if apiErr, ok := err.(*googleapi.Error); ok {
		return IsRetryableStatusCode(apiErr.Code)
	}
	return true
}

// IsRetryableStatusCode returns true if the status code is retryable,
// such as status code is 5xx, 429 (rate limit), or network error occurs.
// Otherwise returns false.
// See https://developers.google.com/photos/library/guides/best-practices#retrying-failed-requests
func IsRetryableStatusCode(code int) bool {
	return code == 429 || (code >= 500 && code <= 599)
}
