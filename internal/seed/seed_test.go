package seed

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/servusdei2018/wt/internal/config"
)

func TestComputePort(t *testing.T) {
	p1 := ComputePort("feature/login")
	p2 := ComputePort("feature/login")
	p3 := ComputePort("feature/dashboard")

	if p1 != p2 {
		t.Errorf("expected deterministic port for same branch, got %d and %d", p1, p2)
	}
	if p1 < 3000 || p1 >= 9000 {
		t.Errorf("port %d out of expected 3000-8999 range", p1)
	}
	if p3 < 3000 || p3 >= 9000 {
		t.Errorf("port %d out of expected 3000-8999 range", p3)
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"feature/login-ui", "feature-login-ui"},
		{"user_profile@v2", "user-profile-v2"},
		{"---main---", "main"},
	}

	for _, tt := range tests {
		got := Slugify(tt.input)
		if got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSeedFiles(t *testing.T) {
	repoRoot := t.TempDir()
	targetWt := t.TempDir()

	// Write repo root seed files
	envContent := "BRANCH={{ .Branch }}\nSLUG={{ .BranchSlug }}\nPORT={{ .Port }}\n"
	if err := os.WriteFile(filepath.Join(repoRoot, ".env"), []byte(envContent), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(repoRoot, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	jsonContent := `{"branch": "{{ .Branch }}", "port": {{ .Port }}}`
	if err := os.WriteFile(filepath.Join(repoRoot, "config", "local.json"), []byte(jsonContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Binary file
	binData := []byte{0x00, 0x01, 0x02, 0x03, 'T', 'E', 'S', 'T'}
	if err := os.WriteFile(filepath.Join(repoRoot, "secret.bin"), binData, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.SeedConfig{
		Include: []string{".env", "config/local.json", "secret.bin", "missing.env"},
	}

	branch := "feat/auth-v1"
	report, err := SeedFiles(repoRoot, targetWt, branch, cfg)
	if err != nil {
		t.Fatalf("SeedFiles unexpected error: %v", err)
	}

	if report.Count() != 3 {
		t.Errorf("expected 3 files copied, got %d", report.Count())
	}

	// Check .env content
	envData, err := os.ReadFile(filepath.Join(targetWt, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	expectedPort := ComputePort(branch)
	expectedEnv := fmt.Sprintf("BRANCH=feat/auth-v1\nSLUG=feat-auth-v1\nPORT=%d\n", expectedPort)
	if string(envData) != expectedEnv {
		t.Errorf("interpolated .env = %q, want %q", string(envData), expectedEnv)
	}

	// Check permissions preserved
	info, err := os.Stat(filepath.Join(targetWt, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf(".env permissions = %v, want %v", info.Mode().Perm(), os.FileMode(0o600))
	}

	// Check config/local.json
	jsonData, err := os.ReadFile(filepath.Join(targetWt, "config", "local.json"))
	if err != nil {
		t.Fatal(err)
	}
	expectedJson := fmt.Sprintf(`{"branch": "feat/auth-v1", "port": %d}`, expectedPort)
	if string(jsonData) != expectedJson {
		t.Errorf("interpolated config/local.json = %q, want %q", string(jsonData), expectedJson)
	}

	// Check binary file copied accurately
	gotBin, err := os.ReadFile(filepath.Join(targetWt, "secret.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotBin, binData) {
		t.Errorf("binary content mismatch")
	}
}

func TestSeedFilesDoNotOverwriteExisting(t *testing.T) {
	repoRoot := t.TempDir()
	targetWt := t.TempDir()

	if err := os.WriteFile(filepath.Join(repoRoot, ".env"), []byte("ROOT=1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetWt, ".env"), []byte("EXISTING=1"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.SeedConfig{Include: []string{".env"}}
	report, err := SeedFiles(repoRoot, targetWt, "feat/test", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if report.Count() != 0 {
		t.Errorf("expected 0 files copied when target already exists, got %d", report.Count())
	}

	content, _ := os.ReadFile(filepath.Join(targetWt, ".env"))
	if string(content) != "EXISTING=1" {
		t.Errorf("target file was overwritten: %q", string(content))
	}
}

func TestSeedFilesPathTraversal(t *testing.T) {
	repoRoot := t.TempDir()
	targetWt := t.TempDir()

	// Filenames starting with leading dots like ..env should be allowed
	if err := os.WriteFile(filepath.Join(repoRoot, "..env"), []byte("SECRET=1"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgValid := config.SeedConfig{Include: []string{"..env"}}
	report, err := SeedFiles(repoRoot, targetWt, "feat/test", cfgValid)
	if err != nil {
		t.Fatalf("expected valid file starting with .. to be allowed, got error: %v", err)
	}
	if report.Count() != 1 {
		t.Errorf("expected 1 file copied for ..env, got %d", report.Count())
	}

	// Escaping path traversal attempts should be rejected
	invalidPaths := []string{"../outside.txt", "foo/../../outside.txt"}
	for _, p := range invalidPaths {
		cfg := config.SeedConfig{Include: []string{p}}
		_, err := SeedFiles(repoRoot, targetWt, "feat/test", cfg)
		if err == nil {
			t.Errorf("expected error for path traversal attempt %q, got nil", p)
		}
	}
}

func TestSeedFilesInterpolationError(t *testing.T) {
	repoRoot := t.TempDir()
	targetWt := t.TempDir()

	// Write an invalid template with unclosed braces or invalid syntax
	invalidContent := "BRANCH={{ .Branch }\n"
	if err := os.WriteFile(filepath.Join(repoRoot, ".env"), []byte(invalidContent), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := config.SeedConfig{
		Include: []string{".env"},
	}

	_, err := SeedFiles(repoRoot, targetWt, "feat/test", cfg)
	if err == nil {
		t.Error("expected error for invalid template syntax, got nil")
	}

	// Write a template referencing a non-existent variable
	missingVarContent := "BRANCH={{ .NonExistent }}\n"
	if err := os.WriteFile(filepath.Join(repoRoot, ".env.missing"), []byte(missingVarContent), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg2 := config.SeedConfig{
		Include: []string{".env.missing"},
	}

	_, err = SeedFiles(repoRoot, targetWt, "feat/test", cfg2)
	if err == nil {
		t.Error("expected error for missing template key, got nil")
	}
}
