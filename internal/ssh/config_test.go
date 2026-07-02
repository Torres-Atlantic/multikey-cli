package ssh

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateConfigIncludesIdentitiesOnly(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{
		sshConfigPath:      filepath.Join(dir, "config"),
		multikeyConfigPath: filepath.Join(dir, "multikey.conf"),
	}

	profiles := []Profile{
		{SSHHost: "github-work", IdentityFile: "/home/me/.ssh/id_ed25519_work"},
		{SSHHost: "github-personal", IdentityFile: "/home/me/.ssh/id_ed25519_personal"},
	}

	if err := m.GenerateConfig(profiles); err != nil {
		t.Fatalf("GenerateConfig: %v", err)
	}

	data, err := os.ReadFile(m.multikeyConfigPath)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	content := string(data)

	// Every host block must carry IdentitiesOnly yes — this is the core promise:
	// SSH must present only the specified key, never fall through to another
	// agent identity.
	if got := strings.Count(content, "IdentitiesOnly yes"); got != len(profiles) {
		t.Errorf("expected IdentitiesOnly yes for each of %d host blocks, found %d\n%s",
			len(profiles), got, content)
	}

	for _, p := range profiles {
		if !strings.Contains(content, "Host "+p.SSHHost) {
			t.Errorf("missing Host block for %s\n%s", p.SSHHost, content)
		}
	}

	// The generated file holds SSH routing config and should be owner-only.
	info, err := os.Stat(m.multikeyConfigPath)
	if err != nil {
		t.Fatalf("stat generated config: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected 0600 permissions, got %o", perm)
	}
}

func TestEnsureIncludePrependsAndBacksUp(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")
	existing := "Host *\n  AddKeysToAgent yes\n"
	if err := os.WriteFile(configPath, []byte(existing), 0600); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	m := &Manager{
		sshConfigPath:      configPath,
		multikeyConfigPath: filepath.Join(dir, "multikey.conf"),
	}

	if err := m.EnsureInclude(); err != nil {
		t.Fatalf("EnsureInclude: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	content := string(data)

	// The include must come before the pre-existing block (first-match-wins).
	includeIdx := strings.Index(content, IncludeDirective)
	hostIdx := strings.Index(content, "Host *")
	if includeIdx == -1 || hostIdx == -1 || includeIdx > hostIdx {
		t.Errorf("include directive should be prepended before existing content:\n%s", content)
	}

	// The original config must be backed up before mutation.
	bak, err := os.ReadFile(configPath + ".bak")
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(bak) != existing {
		t.Errorf("backup does not match original: got %q want %q", string(bak), existing)
	}

	// Running again must be idempotent (no duplicate include).
	if err := m.EnsureInclude(); err != nil {
		t.Fatalf("EnsureInclude (2nd): %v", err)
	}
	data, _ = os.ReadFile(configPath)
	if got := strings.Count(string(data), IncludeDirective); got != 1 {
		t.Errorf("expected include directive exactly once, found %d", got)
	}
}
