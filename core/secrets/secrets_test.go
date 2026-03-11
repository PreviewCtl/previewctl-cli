package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseKeyValues(t *testing.T) {
	tests := []struct {
		name    string
		entries []string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "simple values",
			entries: []string{"KEY=value", "FOO=bar"},
			want:    map[string]string{"KEY": "value", "FOO": "bar"},
		},
		{
			name:    "quoted values",
			entries: []string{`KEY="hello world"`, `FOO='bar baz'`},
			want:    map[string]string{"KEY": "hello world", "FOO": "bar baz"},
		},
		{
			name:    "empty slice",
			entries: []string{},
			want:    map[string]string{},
		},
		{
			name:    "missing equals",
			entries: []string{"NOEQUALS"},
			wantErr: true,
		},
		{
			name:    "empty key",
			entries: []string{"=value"},
			wantErr: true,
		},
		{
			name:    "inline comment stripped",
			entries: []string{"KEY=value #comment"},
			want:    map[string]string{"KEY": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeyValues(tt.entries)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseKeyValues() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d entries, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("key %q: got %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestParseEnvFile(t *testing.T) {
	content := `# Database config
DB_HOST=localhost
DB_PORT=5432
export DB_NAME=mydb

# Quoted value
SECRET="my secret value"
SINGLE='quoted'

# Empty line above

INLINE=value # this is a comment
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ParseEnvFile(path)
	if err != nil {
		t.Fatalf("ParseEnvFile() error: %v", err)
	}

	expected := map[string]string{
		"DB_HOST": "localhost",
		"DB_PORT": "5432",
		"DB_NAME": "mydb",
		"SECRET":  "my secret value",
		"SINGLE":  "quoted",
		"INLINE":  "value",
	}

	for k, want := range expected {
		if got[k] != want {
			t.Errorf("key %q: got %q, want %q", k, got[k], want)
		}
	}
}

func TestParseEnvFile_NotExist(t *testing.T) {
	got, err := ParseEnvFile("/nonexistent/path/.env")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestParseEnvFile_InvalidLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("NOEQUALS\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ParseEnvFile(path)
	if err == nil {
		t.Error("expected error for invalid line, got nil")
	}
}

func TestMerge(t *testing.T) {
	a := map[string]string{"A": "1", "B": "2"}
	b := map[string]string{"B": "overridden", "C": "3"}

	got := Merge(a, b)

	if got["A"] != "1" {
		t.Errorf("A: got %q, want %q", got["A"], "1")
	}
	if got["B"] != "overridden" {
		t.Errorf("B: got %q, want %q", got["B"], "overridden")
	}
	if got["C"] != "3" {
		t.Errorf("C: got %q, want %q", got["C"], "3")
	}
}

func TestMerge_Empty(t *testing.T) {
	got := Merge()
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestParseOSEnv(t *testing.T) {
	got := ParseOSEnv()
	// PATH is practically always set on any OS
	if _, ok := got["PATH"]; !ok {
		// On Windows it might be "Path"
		if _, ok := got["Path"]; !ok {
			t.Error("expected PATH or Path in OS env")
		}
	}
}
