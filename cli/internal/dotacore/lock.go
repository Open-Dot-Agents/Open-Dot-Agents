package dotacore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const LockPath = ".agents/adapters.lock.json"

const (
	maxPublisherManifestBytes = 4 << 20
	maxAdapterArtifactBytes   = 128 << 20
)

type AdapterLock struct {
	Schema      string          `json:"$schema,omitempty"`
	LockVersion int             `json:"lockVersion"`
	Adapters    []LockedAdapter `json:"adapters"`
}

type LockedAdapter struct {
	ID              string        `json:"id"`
	Version         string        `json:"version"`
	ProtocolVersion string        `json:"protocolVersion"`
	Source          AdapterSource `json:"source"`
	Capabilities    []string      `json:"capabilities"`
}

type AdapterSource struct {
	Type      string     `json:"type"`
	Path      string     `json:"path,omitempty"`
	SHA256    string     `json:"sha256,omitempty"`
	Manifest  string     `json:"manifest,omitempty"`
	Artifacts []Artifact `json:"artifacts,omitempty"`
}

type Artifact struct {
	OS             string `json:"os"`
	Arch           string `json:"arch"`
	URL            string `json:"url"`
	SHA256         string `json:"sha256"`
	SigstoreBundle string `json:"sigstoreBundle,omitempty"`
}

type PublisherManifest struct {
	Schema          string     `json:"$schema,omitempty"`
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Version         string     `json:"version"`
	ProtocolVersion string     `json:"protocolVersion"`
	Capabilities    []string   `json:"capabilities"`
	Artifacts       []Artifact `json:"artifacts"`
}

func LoadLock(root string) (AdapterLock, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(LockPath)))
	if err != nil {
		return AdapterLock{}, err
	}
	if err := validateEmbeddedSchema("adapters-lock.schema.json", data); err != nil {
		return AdapterLock{}, diagnosticError("DOTA4000", LockPath, err.Error())
	}
	var lock AdapterLock
	if err := decodeStrictJSON(data, &lock); err != nil {
		return AdapterLock{}, diagnosticError("DOTA4000", LockPath, err.Error())
	}
	seen := map[string]bool{}
	for _, adapter := range lock.Adapters {
		if seen[adapter.ID] {
			return AdapterLock{}, diagnosticError("DOTA4001", LockPath, "duplicate adapter id "+adapter.ID)
		}
		seen[adapter.ID] = true
	}
	return lock, nil
}

func SaveLock(root string, lock AdapterLock) error {
	sort.Slice(lock.Adapters, func(i, j int) bool { return lock.Adapters[i].ID < lock.Adapters[j].ID })
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	if err := validateEmbeddedSchema("adapters-lock.schema.json", data); err != nil {
		return diagnosticError("DOTA4000", LockPath, err.Error())
	}
	path := filepath.Join(root, filepath.FromSlash(LockPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return atomicWrite(path, append(data, '\n'), 0o644)
}

func AddLocalAdapter(root, id, version, executable string, capabilities []string) error {
	if !reverseDNSPattern.MatchString(id) {
		return diagnosticError("DOTA4002", LockPath, "adapter id must use reverse-DNS form")
	}
	abs, err := filepath.Abs(executable)
	if err != nil {
		return err
	}
	digest, err := fileDigest(abs)
	if err != nil {
		return err
	}
	lock, err := LoadLock(root)
	if errors.Is(err, os.ErrNotExist) {
		lock = AdapterLock{Schema: "https://open-dot-agents.github.io/Open-Dot-Agents/spec/v1/schema/adapters-lock.schema.json", LockVersion: 1}
	} else if err != nil {
		return err
	}
	entry := LockedAdapter{ID: id, Version: version, ProtocolVersion: "1.0", Source: AdapterSource{Type: "path", Path: abs, SHA256: digest}, Capabilities: capabilities}
	replaced := false
	for index := range lock.Adapters {
		if lock.Adapters[index].ID == id {
			lock.Adapters[index] = entry
			replaced = true
		}
	}
	if !replaced {
		lock.Adapters = append(lock.Adapters, entry)
	}
	return SaveLock(root, lock)
}

func AddPublishedAdapter(root, manifestReference string) error {
	data, err := readPublisherManifest(manifestReference)
	if err != nil {
		return err
	}
	if err := validateEmbeddedSchema("adapter-publisher.schema.json", data); err != nil {
		return diagnosticError("DOTA4007", manifestReference, err.Error())
	}
	var publisher PublisherManifest
	if err := decodeStrictJSON(data, &publisher); err != nil {
		return diagnosticError("DOTA4007", manifestReference, err.Error())
	}
	if !reverseDNSPattern.MatchString(publisher.ID) || publisher.Name == "" || publisher.Version == "" || publisher.ProtocolVersion != "1.0" || len(publisher.Capabilities) == 0 || len(publisher.Artifacts) == 0 {
		return diagnosticError("DOTA4007", manifestReference, "invalid adapter publisher manifest")
	}
	for _, artifact := range publisher.Artifacts {
		parsed, err := url.Parse(artifact.URL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return diagnosticError("DOTA4007", artifact.URL, "adapter artifacts require https URLs")
		}
	}
	lock, err := LoadLock(root)
	if errors.Is(err, os.ErrNotExist) {
		lock = AdapterLock{Schema: "https://open-dot-agents.github.io/Open-Dot-Agents/spec/v1/schema/adapters-lock.schema.json", LockVersion: 1}
	} else if err != nil {
		return err
	}
	lockedReference := manifestReference
	if parsed, parseErr := url.Parse(manifestReference); parseErr == nil && parsed.Scheme == "" {
		absolute, absoluteErr := filepath.Abs(manifestReference)
		if absoluteErr != nil {
			return absoluteErr
		}
		lockedReference = (&url.URL{Scheme: "file", Path: filepath.ToSlash(absolute)}).String()
	}
	entry := LockedAdapter{
		ID: publisher.ID, Version: publisher.Version, ProtocolVersion: publisher.ProtocolVersion,
		Source:       AdapterSource{Type: "release", Manifest: lockedReference, Artifacts: publisher.Artifacts},
		Capabilities: publisher.Capabilities,
	}
	replaced := false
	for index := range lock.Adapters {
		if lock.Adapters[index].ID == entry.ID {
			lock.Adapters[index] = entry
			replaced = true
		}
	}
	if !replaced {
		lock.Adapters = append(lock.Adapters, entry)
	}
	return SaveLock(root, lock)
}

func readPublisherManifest(reference string) ([]byte, error) {
	parsed, err := url.Parse(reference)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" || parsed.Scheme == "file" {
		path := reference
		if parsed.Scheme == "file" {
			path = parsed.Path
		}
		return os.ReadFile(path)
	}
	if parsed.Scheme != "https" {
		return nil, diagnosticError("DOTA4007", reference, "publisher manifests require https or a local file")
	}
	response, err := http.Get(reference)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("publisher manifest returned %s", response.Status)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxPublisherManifestBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxPublisherManifestBytes {
		return nil, diagnosticError("DOTA4007", reference, "publisher manifest exceeds 4 MiB")
	}
	return data, nil
}

func ResolveAdapter(root, id string, ci bool) (LockedAdapter, string, error) {
	lock, err := LoadLock(root)
	if err != nil {
		return LockedAdapter{}, "", err
	}
	for _, adapter := range lock.Adapters {
		if adapter.ID != id {
			continue
		}
		switch adapter.Source.Type {
		case "path":
			if ci {
				return adapter, "", diagnosticError("DOTA4004", LockPath, "local-path adapters are forbidden in CI mode")
			}
			if err := verifyFile(adapter.Source.Path, adapter.Source.SHA256); err != nil {
				return adapter, "", err
			}
			return adapter, adapter.Source.Path, nil
		case "release":
			path, err := cachedAdapterPath(adapter)
			if err != nil {
				return adapter, "", err
			}
			artifact, err := platformArtifact(adapter)
			if err != nil {
				return adapter, "", err
			}
			if err := verifyFile(path, artifact.SHA256); err != nil {
				return adapter, "", diagnosticError("DOTA4005", path, "adapter is not installed or failed integrity verification")
			}
			return adapter, path, nil
		}
	}
	return LockedAdapter{}, "", diagnosticError("DOTA4003", LockPath, "adapter is not locked: "+id)
}

func InstallAdapters(root string) error {
	lock, err := LoadLock(root)
	if err != nil {
		return err
	}
	for _, adapter := range lock.Adapters {
		if adapter.Source.Type != "release" {
			continue
		}
		artifact, err := platformArtifact(adapter)
		if err != nil {
			return err
		}
		destination, err := cachedAdapterPath(adapter)
		if err != nil {
			return err
		}
		if verifyFile(destination, artifact.SHA256) == nil {
			continue
		}
		request, err := http.NewRequest(http.MethodGet, artifact.URL, nil)
		if err != nil {
			return err
		}
		if request.URL.Scheme != "https" || request.URL.Host == "" {
			return diagnosticError("DOTA4007", artifact.URL, "adapter artifacts require https URLs")
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			return fmt.Errorf("adapter download returned %s", response.Status)
		}
		limited := io.LimitReader(response.Body, maxAdapterArtifactBytes+1)
		data, readErr := io.ReadAll(limited)
		response.Body.Close()
		if readErr != nil {
			return readErr
		}
		if len(data) > maxAdapterArtifactBytes {
			return diagnosticError("DOTA4005", artifact.URL, "adapter artifact exceeds 128 MiB")
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != artifact.SHA256 {
			return diagnosticError("DOTA4005", artifact.URL, "download checksum mismatch")
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		if err := atomicWrite(destination, data, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func platformArtifact(adapter LockedAdapter) (Artifact, error) {
	for _, artifact := range adapter.Source.Artifacts {
		if artifact.OS == runtime.GOOS && artifact.Arch == runtime.GOARCH {
			return artifact, nil
		}
	}
	return Artifact{}, fmt.Errorf("adapter %s has no artifact for %s/%s", adapter.ID, runtime.GOOS, runtime.GOARCH)
}

func cachedAdapterPath(adapter LockedAdapter) (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	name := "adapter"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(root, "dota", "adapters", adapter.ID, adapter.Version, runtime.GOOS+"-"+runtime.GOARCH, name), nil
}

func verifyFile(path, expected string) error {
	digest, err := fileDigest(path)
	if err != nil {
		return err
	}
	if !strings.EqualFold(digest, expected) {
		return diagnosticError("DOTA4005", path, "adapter checksum mismatch")
	}
	return nil
}

func fileDigest(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
