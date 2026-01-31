package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAliases_ValidFile(t *testing.T) {
	// Create temp file with valid aliases
	content := `# This is a comment
alias ll='ls -la'
alias gs="git status"
alias gp='git push'

# Another comment
function not_an_alias() {
    echo "ignored"
}

alias complex='echo "hello=world"'
`
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	aliases, err := ParseAliases(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(aliases) != 4 {
		t.Errorf("expected 4 aliases, got %d", len(aliases))
	}

	// Check first alias
	if aliases[0].Name != "ll" {
		t.Errorf("expected name 'll', got '%s'", aliases[0].Name)
	}
	if aliases[0].Command != "ls -la" {
		t.Errorf("expected command 'ls -la', got '%s'", aliases[0].Command)
	}

	// Check complex alias with = in value
	if aliases[3].Name != "complex" {
		t.Errorf("expected name 'complex', got '%s'", aliases[3].Name)
	}
	if aliases[3].Command != `echo "hello=world"` {
		t.Errorf("expected command with =, got '%s'", aliases[3].Command)
	}
}

func TestParseAliases_EmptyFile(t *testing.T) {
	tmpFile := createTempFile(t, "")
	defer os.Remove(tmpFile)

	aliases, err := ParseAliases(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(aliases) != 0 {
		t.Errorf("expected 0 aliases, got %d", len(aliases))
	}
}

func TestParseAliases_OnlyComments(t *testing.T) {
	content := `# Comment 1
# Comment 2
# alias fake='not parsed'
`
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	aliases, err := ParseAliases(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(aliases) != 0 {
		t.Errorf("expected 0 aliases, got %d", len(aliases))
	}
}

func TestParseAliases_MalformedLines(t *testing.T) {
	content := `alias noequals
alias valid='works'
`
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	aliases, err := ParseAliases(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only parse the valid one (noequals has no = so skipped)
	if len(aliases) != 1 {
		t.Errorf("expected 1 valid alias, got %d", len(aliases))
	}
	if len(aliases) > 0 && aliases[0].Name != "valid" {
		t.Errorf("expected 'valid', got '%s'", aliases[0].Name)
	}
}

func TestParseAliases_NonexistentFile(t *testing.T) {
	_, err := ParseAliases("/nonexistent/path/file.sh")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseAliases_SingleQuotes(t *testing.T) {
	content := `alias single='echo hello'`
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	aliases, err := ParseAliases(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(aliases) != 1 {
		t.Fatalf("expected 1 alias, got %d", len(aliases))
	}
	if aliases[0].Command != "echo hello" {
		t.Errorf("expected 'echo hello', got '%s'", aliases[0].Command)
	}
}

func TestParseAliases_DoubleQuotes(t *testing.T) {
	content := `alias double="echo world"`
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	aliases, err := ParseAliases(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(aliases) != 1 {
		t.Fatalf("expected 1 alias, got %d", len(aliases))
	}
	if aliases[0].Command != "echo world" {
		t.Errorf("expected 'echo world', got '%s'", aliases[0].Command)
	}
}

func TestParseAliases_SourceField(t *testing.T) {
	content := `alias test='echo test'`
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	aliases, err := ParseAliases(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if aliases[0].Source != tmpFile {
		t.Errorf("expected source '%s', got '%s'", tmpFile, aliases[0].Source)
	}
}

// Helper function to create temp files for testing
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_aliases.sh")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}
