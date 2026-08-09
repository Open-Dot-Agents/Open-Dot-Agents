package dotacore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLocalAdapterIsChecksumPinnedAndRejectedInCI(t *testing.T) {
	root := t.TempDir()
	executable := filepath.Join(t.TempDir(), "adapter")
	if err := os.WriteFile(executable, []byte("adapter-v1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := AddLocalAdapter(root, "org.example.local", "1.0.0", executable, []string{"validate"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ResolveAdapter(root, "org.example.local", true); err == nil || !strings.Contains(err.Error(), "DOTA4004") {
		t.Fatalf("CI error = %v", err)
	}
	if err := os.WriteFile(executable, []byte("adapter-v2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ResolveAdapter(root, "org.example.local", false); err == nil || !strings.Contains(err.Error(), "DOTA4005") {
		t.Fatalf("checksum error = %v", err)
	}
}

func TestReleaseAdapterDownloadVerifiesChecksum(t *testing.T) {
	artifactBytes := []byte("released adapter")
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write(artifactBytes)
	}))
	defer server.Close()
	previousClient := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() { http.DefaultClient = previousClient })

	for _, test := range []struct {
		name    string
		sha256  string
		wantErr bool
	}{
		{name: "valid", sha256: hashBytes(artifactBytes)},
		{name: "mismatch", sha256: strings.Repeat("0", 64), wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			cache := t.TempDir()
			t.Setenv("XDG_CACHE_HOME", cache)
			t.Setenv("LocalAppData", cache)
			root := t.TempDir()
			publisher := PublisherManifest{
				ID: "org.example.release", Name: "Release adapter", Version: "1.0.0", ProtocolVersion: "1.0",
				Capabilities: []string{"validate"},
				Artifacts:    []Artifact{{OS: runtime.GOOS, Arch: runtime.GOARCH, URL: server.URL + "/adapter", SHA256: test.sha256}},
			}
			manifestPath := filepath.Join(t.TempDir(), "publisher.json")
			data, err := json.Marshal(publisher)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := AddPublishedAdapter(root, manifestPath); err != nil {
				t.Fatal(err)
			}
			err = InstallAdapters(root)
			if test.wantErr {
				if err == nil || !strings.Contains(err.Error(), "DOTA4005") {
					t.Fatalf("install error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			_, executable, err := ResolveAdapter(root, publisher.ID, true)
			if err != nil {
				t.Fatal(err)
			}
			if got, err := os.ReadFile(executable); err != nil || string(got) != string(artifactBytes) {
				t.Fatalf("installed bytes = %q, %v", got, err)
			}
		})
	}
}

func TestPublishedAdapterRequiresHTTPSArtifacts(t *testing.T) {
	publisher := PublisherManifest{
		ID: "org.example.insecure", Name: "Insecure adapter", Version: "1.0.0", ProtocolVersion: "1.0",
		Capabilities: []string{"validate"},
		Artifacts:    []Artifact{{OS: runtime.GOOS, Arch: runtime.GOARCH, URL: "http://example.com/adapter", SHA256: strings.Repeat("0", 64)}},
	}
	data, err := json.Marshal(publisher)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(t.TempDir(), "publisher.json")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := AddPublishedAdapter(t.TempDir(), manifestPath); err == nil || !strings.Contains(err.Error(), "DOTA4007") {
		t.Fatalf("HTTPS error = %v", err)
	}
}

func TestLoadLockRejectsDuplicateAdapterIDs(t *testing.T) {
	root := t.TempDir()
	entry := `{"id":"org.example.duplicate","version":"1","protocolVersion":"1.0","source":{"type":"path","path":"/adapter","sha256":"` + strings.Repeat("0", 64) + `"},"capabilities":["validate"]}`
	data := []byte(`{"lockVersion":1,"adapters":[` + entry + `,` + entry + `]}`)
	path := filepath.Join(root, filepath.FromSlash(LockPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadLock(root); err == nil || !strings.Contains(err.Error(), "DOTA4001") {
		t.Fatalf("duplicate error = %v", err)
	}
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
