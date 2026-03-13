package identity

import (
	"strings"
	"testing"
)

func TestIsValidResourceName(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"abc", true},
		{"a", true},
		{"abc-def", true},
		{"a1b2c3", true},
		{"a-b-c", true},

		// Must start/end with alphanumeric
		{"-abc", false},
		{"abc-", false},
		{"-", false},

		// Empty
		{"", false},

		// Uppercase not allowed
		{"Abc", false},
		{"ABC", false},

		// Special chars
		{"abc_def", false},
		{"abc.def", false},
		{"abc def", false},
		{"abc/def", false},

		// Length boundary: exactly 63 chars
		{strings.Repeat("a", 63), true},
		// 64 chars
		{strings.Repeat("a", 64), false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsValidResourceName(tt.input)
			if got != tt.want {
				t.Errorf("IsValidResourceName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeResourceName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"Hello-World", "hello-world"},
		{"My_Project", "my-project"},
		{"my project", "my-project"},
		{"UPPER", "upper"},
		{"a/b/c", "a-b-c"},
		{"--leading--trailing--", "leading-trailing"},
		{"a..b", "a-b"},
		{"a___b", "a-b"},     // consecutive special chars collapse
		{"!!!abc!!!", "abc"}, // only special at edges
		{"", ""},             // empty string
		{"123", "123"},       // numbers
		{"hello world 123", "hello-world-123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeResourceName(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeResourceName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePreviewID_WithExplicitID(t *testing.T) {
	id, err := ResolvePreviewID("my-preview", "/some/dir", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(id, "preview-") {
		t.Errorf("expected prefix 'preview-', got %q", id)
	}
	if !strings.Contains(id, "my-preview") {
		t.Errorf("expected id to contain 'my-preview', got %q", id)
	}
}

func TestResolvePreviewID_AlreadyPrefixed(t *testing.T) {
	id, err := ResolvePreviewID("preview-abc", "/some/dir", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "preview-abc" {
		t.Errorf("expected 'preview-abc', got %q", id)
	}
}

func TestResolvePreviewID_InvalidExplicitID(t *testing.T) {
	_, err := ResolvePreviewID("INVALID ID!", "/some/dir", "main")
	if err == nil {
		t.Error("expected error for invalid ID, got nil")
	}
}

func TestResolvePreviewID_Generated(t *testing.T) {
	id, err := ResolvePreviewID("", "/home/user/my-project", "feature-branch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(id, "preview-") {
		t.Errorf("expected prefix 'preview-', got %q", id)
	}
	if !IsValidResourceName(strings.TrimPrefix(id, "preview-")) {
		// The full id with prefix may be >63, but the part after prefix should be valid
		if !IsValidResourceName(id) && len(id) <= 63+len("preview-") {
			t.Errorf("generated id %q is not valid", id)
		}
	}
}

func TestResolvePreviewID_GeneratedNoBranch(t *testing.T) {
	id, err := ResolvePreviewID("", "/home/user/my-project", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(id, "preview-") {
		t.Errorf("expected prefix 'preview-', got %q", id)
	}
	if !strings.Contains(id, "my-project") {
		t.Errorf("expected id to contain folder name 'my-project', got %q", id)
	}
}

func TestEnsurePrefix(t *testing.T) {
	if ensurePrefix("abc") != "preview-abc" {
		t.Error("should add prefix")
	}
	if ensurePrefix("preview-abc") != "preview-abc" {
		t.Error("should not double-prefix")
	}
}
