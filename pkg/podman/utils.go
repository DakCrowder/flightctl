package podman

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flightctl/flightctl/pkg/log"
	"github.com/flightctl/flightctl/pkg/poll"
)

// SanitizePodmanLabel sanitizes a string to be used as a label in Podman.
// Podman labels must be lowercase and can only contain alpha numeric
// characters, hyphens, and underscores. Any other characters are replaced with
// an underscore.
func SanitizePodmanLabel(name string) string {
	var result strings.Builder
	result.Grow(len(name))

	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		// lower case alpha numeric characters, hyphen, and underscore are allowed
		case (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_':
			result.WriteByte(c)
		// upper case alpha characters are converted to lower case
		case c >= 'A' && c <= 'Z':
			// add 32 to ascii value convert to lower case
			result.WriteByte(c + 32)
		// any special characters are replaced with an underscore
		default:
			result.WriteByte('_')
		}
	}

	return result.String()
}

func applyFilters(args, labels, filters []string) []string {
	for _, label := range labels {
		args = append(args, "--filter", fmt.Sprintf("label=%s", label))
	}

	for _, filter := range filters {
		args = append(args, "--filter", filter)
	}
	return args
}

func retryWithBackoff(ctx context.Context, log *log.PrefixLogger, backoff poll.Config, operation func(context.Context) (string, error)) (string, error) {
	var result string
	var retriableErr error
	err := poll.BackoffWithContext(ctx, backoff, func(ctx context.Context) (bool, error) {
		var err error
		retriableErr = nil
		result, err = operation(ctx)
		if err != nil {
			if !IsRetryable(err) {
				log.Error(err)
				return false, err
			}
			retriableErr = err
			log.Warnf("A retriable error occurred: %s", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		if retriableErr != nil {
			err = fmt.Errorf("%w: %w", retriableErr, err)
		}
		return "", err
	}
	return result, nil
}

func logProgress(ctx context.Context, log *log.PrefixLogger, msg string, fn func(ctx context.Context) (string, error)) (string, error) {
	doneCh := make(chan struct{})
	defer close(doneCh)

	startTime := time.Now()
	go func() {
		ticker := time.NewTicker(defaultPullLogInterval)
		defer ticker.Stop()

		for {
			select {
			case <-doneCh:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)
				log.Infof("%s (elapsed: %v)", msg, elapsed)
			}
		}
	}()

	return fn(ctx)
}
