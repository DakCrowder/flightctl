package podman

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"syscall"

	"github.com/flightctl/flightctl/pkg/poll"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Error constants for podman operations.
var (
	ErrNotFound      = errors.New("not found")
	ErrAuthFailed    = errors.New("authentication failed")
	ErrNetwork       = errors.New("network error")
	ErrNoRetry       = errors.New("non-retryable error")
	ErrImageNotFound = errors.New("image not found")
	ErrImageShortName = errors.New("failed to resolve image short name")
)

// Error substring constants for stderr parsing.
const (
	noSuchImageErrorSubstring      = "no such image"
	noSuchArtifactErrorSubstring   = "no such artifact"
	imageNotKnownErrorSubstring    = "image not known"
	artifactNotKnownErrorSubstring = "artifact not known"
)

// stderrError wraps an error with stderr output details.
type stderrError struct {
	wrapped error
	reason  string
	code    int
	stderr  string
}

func (e *stderrError) Error() string {
	return fmt.Sprintf("%s: code: %d: %s", e.wrapped.Error(), e.code, e.stderr)
}

func (e *stderrError) Unwrap() error {
	return e.wrapped
}

func (e *stderrError) Reason() string {
	return e.reason
}

// FromStderr converts stderr output from a command into an error type.
func FromStderr(stderr string, exitCode int) error {
	errMap := map[string]error{
		// authentication
		"authentication required": ErrAuthFailed,
		"unauthorized":            ErrAuthFailed,
		"access denied":           ErrAuthFailed,
		// not found
		"not found":        ErrNotFound,
		"manifest unknown": ErrImageNotFound,
		// networking
		"no such host":           ErrNetwork,
		"connection refused":     ErrNetwork,
		"unable to resolve host": ErrNetwork,
		"network is unreachable": ErrNetwork,
		"i/o timeout":            ErrNetwork,
		"unexpected EOF":         ErrNetwork,
		// context
		"context canceled":          context.Canceled,
		"context deadline exceeded": context.DeadlineExceeded,
		// container image resolution
		"short-name resolution enforced": ErrImageShortName,
		// no such object
		"no such object": ErrNotFound,
	}
	for check, err := range errMap {
		if strings.Contains(stderr, check) {
			return &stderrError{
				wrapped: err,
				reason:  check,
				code:    exitCode,
				stderr:  stderr,
			}
		}
	}
	return fmt.Errorf("code: %d: %s", exitCode, stderr)
}

// IsRetryable determines if an error should be retried.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var dnsErr *net.DNSError
	switch {
	case errors.As(err, &dnsErr):
		return dnsErr.Temporary()
	case isTimeoutError(err):
		return true
	case errors.Is(err, ErrNetwork):
		return true
	case errors.Is(err, poll.ErrMaxSteps):
		return true
	case errors.Is(err, syscall.ECONNRESET):
		return true
	case errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF):
		return true
	case strings.Contains(err.Error(), "unexpected EOF"):
		return true
	case errors.Is(err, ErrNoRetry):
		return false
	case errors.Is(err, ErrAuthFailed):
		return false
	default:
		return false
	}
}

func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if wait.Interrupted(err) {
		return true
	}
	if errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return false
}
