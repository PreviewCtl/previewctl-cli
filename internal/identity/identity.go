package identity

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"path/filepath"
	"strings"
)

// ResolvePreviewID returns a valid preview ID. If inputID is non-empty it is
// validated and returned. Otherwise a new ID is generated from the workspace
// directory name with a random numeric suffix.
func ResolvePreviewID(inputID string, workingDir string, branch string) (id string, err error) {
	trimmed := strings.TrimSpace(inputID)
	if trimmed != "" {
		if !IsValidResourceName(trimmed) {
			return "", fmt.Errorf("invalid preview id %q: must be lowercase alphanumeric or '-', 1-63 chars, start/end with alphanumeric", trimmed)
		}
		return ensurePrefix(trimmed), nil
	}

	folderName := filepath.Base(workingDir)
	base := SanitizeResourceName(folderName)
	if base == "" {
		base = "preview"
	}

	branchPart := SanitizeResourceName(branch)
	if branchPart != "" {
		base = base + "-" + branchPart
	}

	suffix, err := randomNumericSuffix(8)
	if err != nil {
		return "", fmt.Errorf("failed to generate preview id: %w", err)
	}

	maxBaseLen := 63 - 1 - len(suffix)
	if maxBaseLen < 1 {
		maxBaseLen = 1
	}
	if len(base) > maxBaseLen {
		base = strings.Trim(base[:maxBaseLen], "-")
	}
	if base == "" {
		base = "preview"
	}

	id = base + "-" + suffix
	if !IsValidResourceName(id) {
		return "", fmt.Errorf("failed to generate valid preview id")
	}

	return ensurePrefix(id), nil
}

// IsValidResourceName checks whether value is valid for a Docker network name
// or Kubernetes namespace (RFC 1123 label: lowercase alphanumeric + '-',
// 1-63 chars, must start and end with alphanumeric).
func IsValidResourceName(value string) bool {
	if len(value) < 1 || len(value) > 63 {
		return false
	}
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' {
			continue
		}
		if ch == '-' && i > 0 && i < len(value)-1 {
			continue
		}
		return false
	}
	return true
}

// SanitizeResourceName converts an arbitrary string into a valid resource name
// base (lowercase alphanumeric with single dashes, no leading/trailing dash).
func SanitizeResourceName(value string) string {
	v := strings.ToLower(value)
	var b strings.Builder
	b.Grow(len(v))
	lastDash := false

	for i := 0; i < len(v); i++ {
		ch := v[i]
		if ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' {
			b.WriteByte(ch)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(b.String(), "-")
}

const previewPrefix = "preview-"

func ensurePrefix(id string) string {
	if strings.HasPrefix(id, previewPrefix) {
		return id
	}
	return previewPrefix + id
}

func randomNumericSuffix(width int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(width)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", width, n.Int64()), nil
}
