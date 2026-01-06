package fileio

import (
	"encoding/base64"
	"fmt"

	"github.com/flightctl/flightctl/api/v1beta1"
)

// DecodeContent decodes the content based on the encoding type and returns the
// decoded content as a byte slice.
func DecodeContent(content string, encoding *v1beta1.EncodingType) ([]byte, error) {
	if encoding == nil || *encoding == "plain" {
		return []byte(content), nil
	}

	switch *encoding {
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 content: %w", err)
		}
		return decoded, nil
	default:
		return nil, fmt.Errorf("unsupported content encoding: %q", *encoding)
	}
}
