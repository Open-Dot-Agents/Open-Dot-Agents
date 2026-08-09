package dotacore

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/pkg/adapterprotocol"
)

const maxSnapshotBytes = 16 << 20

type ChangeSet struct {
	Create []string `json:"create"`
	Update []string `json:"update"`
	Delete []string `json:"delete"`
}

type ApplyOptions struct {
	Force  bool
	DryRun bool
	Backup bool
}

type ownershipManifest struct {
	Version int               `json:"version"`
	Adapter string            `json:"adapter"`
	Files   map[string]string `json:"files"`
}

func Snapshot(root string, patterns []string, adapterLimit int64) (adapterprotocol.SnapshotParams, error) {
	var files []adapterprotocol.File
	var total int64
	limit := int64(maxSnapshotBytes)
	if adapterLimit > 0 && adapterLimit < limit {
		limit = adapterLimit
	}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if relative == "." {
			return nil
		}
		if entry.IsDir() {
			if relative == ".git" || strings.HasPrefix(relative, ".git/") || relative == ".dota-backups" {
				return filepath.SkipDir
			}
			return nil
		}
		if !matchesAny(patterns, relative) {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return diagnosticError("DOTA3000", relative, "adapter input may not be a symlink")
		}
		data, err := os.ReadFile(current)
		if err != nil {
			return err
		}
		total += int64(len(data))
		if total > limit {
			return diagnosticError("DOTA3001", relative, "adapter snapshot exceeds declared byte limit")
		}
		files = append(files, adapterprotocol.File{Path: relative, Encoding: "base64", Content: base64.StdEncoding.EncodeToString(data)})
		return nil
	})
	if err != nil {
		return adapterprotocol.SnapshotParams{}, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return adapterprotocol.SnapshotParams{Files: files}, nil
}

func ApplyPlan(root, adapterID, operation string, plan *adapterprotocol.Plan, options ApplyOptions) (ChangeSet, error) {
	if plan == nil {
		return ChangeSet{}, diagnosticError("DOTA3002", "", "adapter returned no plan")
	}
	desired := map[string][]byte{}
	canonicalPaths := map[string]string{}
	for _, file := range plan.Files {
		if err := validateOutputPath(file.Path, operation); err != nil {
			return ChangeSet{}, err
		}
		if _, exists := desired[file.Path]; exists {
			return ChangeSet{}, diagnosticError("DOTA3003", file.Path, "adapter returned duplicate output")
		}
		key := strings.ToLower(file.Path)
		if previous, exists := canonicalPaths[key]; exists {
			return ChangeSet{}, diagnosticError("DOTA3003", file.Path, "adapter output collides with "+previous)
		}
		canonicalPaths[key] = file.Path
		data, err := decodeProtocolFile(file)
		if err != nil {
			return ChangeSet{}, err
		}
		desired[file.Path] = data
	}
	manifestPath := ownershipPath(root, adapterID, operation)
	current, exists, err := readOwnership(manifestPath, adapterID)
	if err != nil {
		return ChangeSet{}, err
	}
	if !exists {
		current = ownershipManifest{Version: 1, Adapter: adapterID, Files: map[string]string{}}
	}
	changes := ChangeSet{}
	for output, data := range desired {
		full := filepath.Join(root, filepath.FromSlash(output))
		existing, readErr := readOutput(full)
		_, owned := current.Files[output]
		switch {
		case errors.Is(readErr, fs.ErrNotExist):
			changes.Create = append(changes.Create, output)
		case readErr != nil:
			return ChangeSet{}, readErr
		case !owned && !options.Force && !bytesEqual(existing, data):
			return ChangeSet{}, diagnosticError("DOTA3004", output, "output exists but is not adapter-owned; use --force after review")
		case !owned && !options.Force:
			return ChangeSet{}, diagnosticError("DOTA3004", output, "matching output exists but is not adapter-owned; use --force to adopt it")
		case owned && digest(existing) != current.Files[output] && !options.Force:
			return ChangeSet{}, diagnosticError("DOTA3005", output, "owned output was modified; use --force after review")
		case !bytesEqual(existing, data):
			changes.Update = append(changes.Update, output)
		}
	}
	for output := range current.Files {
		if _, retained := desired[output]; retained {
			continue
		}
		full := filepath.Join(root, filepath.FromSlash(output))
		data, readErr := readOutput(full)
		if readErr == nil && digest(data) != current.Files[output] && !options.Force {
			return ChangeSet{}, diagnosticError("DOTA3005", output, "refusing to delete modified owned output")
		}
		if readErr == nil {
			changes.Delete = append(changes.Delete, output)
		} else if !errors.Is(readErr, fs.ErrNotExist) {
			return ChangeSet{}, readErr
		}
	}
	sort.Strings(changes.Create)
	sort.Strings(changes.Update)
	sort.Strings(changes.Delete)
	if options.DryRun {
		return changes, nil
	}
	if options.Backup {
		if err := backupChanged(root, adapterID, append(append([]string{}, changes.Update...), changes.Delete...)); err != nil {
			return ChangeSet{}, err
		}
	}
	next := ownershipManifest{Version: 1, Adapter: adapterID, Files: map[string]string{}}
	paths := make([]string, 0, len(desired))
	for output := range desired {
		paths = append(paths, output)
	}
	sort.Strings(paths)
	for _, output := range paths {
		full := filepath.Join(root, filepath.FromSlash(output))
		if err := ensureNoSymlinkParents(root, full); err != nil {
			return ChangeSet{}, err
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return ChangeSet{}, err
		}
		if err := atomicWrite(full, desired[output], 0o644); err != nil {
			return ChangeSet{}, err
		}
		next.Files[output] = digest(desired[output])
	}
	for _, output := range changes.Delete {
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(output))); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return ChangeSet{}, err
		}
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return ChangeSet{}, err
	}
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		return ChangeSet{}, err
	}
	if err := atomicWrite(manifestPath, append(data, '\n'), 0o644); err != nil {
		return ChangeSet{}, err
	}
	return changes, nil
}

func CheckPlan(root, adapterID, operation string, plan *adapterprotocol.Plan) error {
	if plan == nil {
		return diagnosticError("DOTA3002", "", "adapter returned no plan")
	}
	manifest, exists, err := readOwnership(ownershipPath(root, adapterID, operation), adapterID)
	if err != nil {
		return err
	}
	if !exists {
		return diagnosticError("DOTA3006", "", "no dota ownership manifest; run export or import")
	}
	if len(manifest.Files) != len(plan.Files) {
		return diagnosticError("DOTA3007", "", "generated output is stale")
	}
	seen := map[string]string{}
	for _, file := range plan.Files {
		if err := validateOutputPath(file.Path, operation); err != nil {
			return err
		}
		key := strings.ToLower(file.Path)
		if previous, exists := seen[key]; exists {
			return diagnosticError("DOTA3003", file.Path, "adapter output collides with "+previous)
		}
		seen[key] = file.Path
		data, err := decodeProtocolFile(file)
		if err != nil {
			return err
		}
		current, err := readOutput(filepath.Join(root, filepath.FromSlash(file.Path)))
		if err != nil || !bytesEqual(current, data) || manifest.Files[file.Path] != digest(data) {
			return diagnosticError("DOTA3007", file.Path, "generated output is stale or modified")
		}
	}
	return nil
}

func Clean(root, adapterID, operation string) error {
	manifestPath := ownershipPath(root, adapterID, operation)
	manifest, exists, err := readOwnership(manifestPath, adapterID)
	if err != nil || !exists {
		return err
	}
	for output, want := range manifest.Files {
		full := filepath.Join(root, filepath.FromSlash(output))
		data, readErr := readOutput(full)
		if readErr == nil && digest(data) != want {
			return diagnosticError("DOTA3005", output, "refusing to remove modified owned output")
		}
		if readErr == nil {
			if err := os.Remove(full); err != nil {
				return err
			}
		} else if !errors.Is(readErr, fs.ErrNotExist) {
			return readErr
		}
	}
	return os.Remove(manifestPath)
}

func validateOutputPath(output, operation string) error {
	clean := path.Clean(output)
	if output == "" || clean != output || strings.HasPrefix(clean, "/") || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return diagnosticError("DOTA3008", output, "unsafe adapter output path")
	}
	if operation == "export" && (output == ".agents" || strings.HasPrefix(output, ".agents/")) {
		return diagnosticError("DOTA3008", output, "export adapters may not write the canonical tree")
	}
	if operation == "import" {
		if !strings.HasPrefix(output, ".agents/") || output == ManifestPath || output == LockPath || strings.HasPrefix(output, ".agents/.dota/") {
			return diagnosticError("DOTA3008", output, "import adapters may write only canonical categories or extensions")
		}
		parts := strings.Split(output, "/")
		if len(parts) < 3 || (!categoryNames[parts[1]] && parts[1] != "extensions") {
			return diagnosticError("DOTA3008", output, "import output is not a category or extension path")
		}
		if parts[1] == "extensions" && (len(parts) < 4 || !reverseDNSPattern.MatchString(parts[2])) {
			return diagnosticError("DOTA3008", output, "import extension output requires a reverse-DNS namespace")
		}
	}
	return nil
}

func decodeProtocolFile(file adapterprotocol.File) ([]byte, error) {
	switch file.Encoding {
	case "utf-8":
		return []byte(file.Content), nil
	case "base64":
		data, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			return nil, diagnosticError("DOTA3009", file.Path, "invalid base64 output")
		}
		return data, nil
	default:
		return nil, diagnosticError("DOTA3009", file.Path, "unsupported output encoding")
	}
}

func ownershipPath(root, adapterID, operation string) string {
	return filepath.Join(root, ".agents", ".dota", operation, adapterID+".json")
}

func readOwnership(path, adapterID string) (ownershipManifest, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return ownershipManifest{}, false, nil
	}
	if err != nil {
		return ownershipManifest{}, false, err
	}
	var manifest ownershipManifest
	if err := decodeStrictJSON(data, &manifest); err != nil || manifest.Version != 1 || manifest.Adapter != adapterID || manifest.Files == nil {
		return ownershipManifest{}, false, diagnosticError("DOTA3010", filepath.ToSlash(path), "invalid ownership manifest")
	}
	for output := range manifest.Files {
		if err := validateOutputPath(output, filepath.Base(filepath.Dir(path))); err != nil {
			return ownershipManifest{}, false, err
		}
	}
	return manifest, true, nil
}

func matchesAny(patterns []string, name string) bool {
	for _, pattern := range patterns {
		if matchPattern(pattern, name) {
			return true
		}
	}
	return false
}

func matchPattern(pattern, name string) bool {
	if pattern == "**" {
		return true
	}
	if !strings.Contains(pattern, "**") {
		matched, _ := path.Match(pattern, name)
		return matched
	}
	before, after, _ := strings.Cut(pattern, "**")
	if !strings.HasPrefix(name, before) {
		return false
	}
	remainder := strings.TrimPrefix(name, before)
	if after == "" {
		return true
	}
	after = strings.TrimPrefix(after, "/")
	for {
		matched, _ := path.Match(after, remainder)
		if matched {
			return true
		}
		index := strings.IndexByte(remainder, '/')
		if index < 0 {
			return false
		}
		remainder = remainder[index+1:]
	}
}

func ensureNoSymlinkParents(root, destination string) error {
	parent := filepath.Dir(destination)
	for parent != root && parent != filepath.Dir(parent) {
		info, err := os.Lstat(parent)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return diagnosticError("DOTA3011", filepath.ToSlash(parent), "refusing symlinked output parent")
		}
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		parent = filepath.Dir(parent)
	}
	return nil
}

func readOutput(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, diagnosticError("DOTA3012", filepath.ToSlash(path), "output is not a regular file")
	}
	return os.ReadFile(path)
}

func atomicWrite(destination string, data []byte, mode fs.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(destination), ".dota-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryName, destination)
}

func backupChanged(root, adapterID string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	backupRoot := filepath.Join(root, ".dota-backups", adapterID, time.Now().UTC().Format("20060102-150405.000000000"))
	for _, output := range paths {
		data, err := readOutput(filepath.Join(root, filepath.FromSlash(output)))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		destination := filepath.Join(backupRoot, filepath.FromSlash(output))
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		if err := atomicWrite(destination, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func bytesEqual(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
