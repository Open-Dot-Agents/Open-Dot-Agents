package adapter

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

const manifestName = ".open-dot-agents.json"

const importManifestPrefix = ".open-dot-agents-import-"

var commonUnsupportedCategories = []string{
	"guardrails",
	"memories",
	"permissions",
	"plugins",
	"profiles",
	"prompts",
	"settings",
}

type Adapter struct {
	root             string
	renderer         renderer
	allowUnsupported bool
}

type GenerationAction string

const (
	GenerationCreate GenerationAction = "create"
	GenerationUpdate GenerationAction = "update"
	GenerationDelete GenerationAction = "delete"
)

type GenerationPlan struct {
	Create  []string          `json:"create"`
	Update  []string          `json:"update"`
	Delete  []string          `json:"delete"`
	Outputs map[string][]byte `json:"-"`
}

func (p GenerationPlan) AllActions() map[GenerationAction][]string {
	return map[GenerationAction][]string{
		GenerationCreate: p.Create,
		GenerationUpdate: p.Update,
		GenerationDelete: p.Delete,
	}
}

type manifest struct {
	Version int               `json:"version"`
	Target  string            `json:"target,omitempty"`
	Files   map[string]string `json:"files"`
}

func New(root string, allowUnsupported bool) *Adapter {
	a, _ := NewForTarget(root, "copilot", allowUnsupported)
	return a
}

func NewForTarget(root, target string, allowUnsupported bool) (*Adapter, error) {
	renderer, err := rendererFor(target)
	if err != nil {
		return nil, err
	}
	return &Adapter{root: root, renderer: renderer, allowUnsupported: allowUnsupported}, nil
}

func (a *Adapter) Target() string {
	return a.renderer.target()
}

func (a *Adapter) Validate() error {
	if err := validateContracts(a.root, a.Target(), true); err != nil {
		return err
	}
	_, err := a.outputs()
	return err
}

func (a *Adapter) Generate(force bool) error {
	plan, err := a.Plan(force)
	if err != nil {
		return err
	}
	return a.applyGenerationPlan(plan)
}

func (a *Adapter) Import(force bool) error {
	plan, err := a.ImportPlan(force)
	if err != nil {
		return err
	}
	return a.applyImportPlan(plan)
}

func (a *Adapter) ImportPlan(force bool) (*GenerationPlan, error) {
	importer, err := importerFor(a.Target())
	if err != nil {
		return nil, err
	}
	outputs, err := importer(a.root)
	if err != nil {
		return nil, err
	}
	if err := validateOutputPaths(outputs); err != nil {
		return nil, err
	}
	current, _, err := a.readImportManifest()
	if err != nil {
		return nil, err
	}
	if !force {
		for path, hash := range current.Files {
			data, readErr := os.ReadFile(filepath.Join(a.root, path))
			if readErr != nil || digest(data) != hash {
				return nil, fmt.Errorf("imported canonical file %q was modified; rerun with --force after review", path)
			}
		}
	}
	plan := &GenerationPlan{
		Outputs: outputs,
		Create:  make([]string, 0),
		Update:  make([]string, 0),
		Delete:  make([]string, 0),
	}
	for path, data := range outputs {
		full := filepath.Join(a.root, path)
		existing, err := os.ReadFile(full)
		if errors.Is(err, fs.ErrNotExist) {
			if _, owned := current.Files[path]; owned {
				plan.Update = append(plan.Update, path)
			} else {
				plan.Create = append(plan.Create, path)
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		if bytes.Equal(existing, data) {
			if _, owned := current.Files[path]; !owned && !force {
				return nil, fmt.Errorf("import output %q exists but is not adapter-owned; rerun with --force after review", path)
			}
			continue
		}
		if _, owned := current.Files[path]; !owned && !force {
			return nil, fmt.Errorf("import output %q exists but is not adapter-owned; rerun with --force after review", path)
		}
		plan.Update = append(plan.Update, path)
	}
	for path := range current.Files {
		if _, desired := outputs[path]; !desired {
			plan.Delete = append(plan.Delete, path)
		}
	}
	sort.Strings(plan.Create)
	sort.Strings(plan.Update)
	sort.Strings(plan.Delete)
	return plan, nil
}

func (a *Adapter) applyImportPlan(plan *GenerationPlan) error {
	current, _, err := a.readImportManifest()
	if err != nil {
		return err
	}
	next := manifest{Version: 1, Target: a.Target(), Files: make(map[string]string, len(plan.Outputs))}
	for path, data := range plan.Outputs {
		full := filepath.Join(a.root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := writeFileAtomically(full, data, 0o644); err != nil {
			return err
		}
		next.Files[path] = digest(data)
	}
	for path := range current.Files {
		if _, retained := next.Files[path]; !retained {
			if err := os.Remove(filepath.Join(a.root, path)); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	manifestPath := a.importManifestPath()
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		return err
	}
	return writeFileAtomically(manifestPath, append(data, '\n'), 0o644)
}

func (a *Adapter) Plan(force bool) (*GenerationPlan, error) {
	outputs, err := a.outputs()
	if err != nil {
		return nil, err
	}
	current, _, err := a.readManifest()
	if err != nil {
		return nil, err
	}
	if !force {
		for path, hash := range current.Files {
			data, readErr := os.ReadFile(filepath.Join(a.root, path))
			if readErr != nil || digest(data) != hash {
				return nil, fmt.Errorf("generated file %q was modified; rerun with --force after review", path)
			}
		}
	}
	plan := &GenerationPlan{
		Outputs: outputs,
		Create:  make([]string, 0),
		Update:  make([]string, 0),
		Delete:  make([]string, 0),
	}
	for path := range outputs {
		if _, owned := current.Files[path]; !owned {
			full := filepath.Join(a.root, path)
			_, existingErr := os.Stat(full)
			if existingErr == nil {
				if !force {
					return nil, fmt.Errorf("output %q exists but is not adapter-owned; rerun with --force after review", path)
				}
			} else if !errors.Is(existingErr, fs.ErrNotExist) {
				return nil, existingErr
			}
			if !force && errors.Is(existingErr, fs.ErrNotExist) {
				plan.Create = append(plan.Create, path)
			} else {
				plan.Update = append(plan.Update, path)
			}
			continue
		}
		data, err := os.ReadFile(filepath.Join(a.root, path))
		if errors.Is(err, fs.ErrNotExist) {
			plan.Update = append(plan.Update, path)
			continue
		}
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(data, outputs[path]) {
			plan.Update = append(plan.Update, path)
		}
	}
	for path := range current.Files {
		if _, desired := outputs[path]; !desired {
			plan.Delete = append(plan.Delete, path)
		}
	}
	sort.Strings(plan.Create)
	sort.Strings(plan.Update)
	sort.Strings(plan.Delete)
	return plan, nil
}

func (a *Adapter) applyGenerationPlan(plan *GenerationPlan) error {
	current, _, err := a.readManifest()
	if err != nil {
		return err
	}
	next := manifest{Version: 1, Files: make(map[string]string, len(plan.Outputs))}
	for path, data := range plan.Outputs {
		full := filepath.Join(a.root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := writeFileAtomically(full, data, 0o644); err != nil {
			return err
		}
		next.Files[path] = digest(data)
	}
	for path := range current.Files {
		if _, retained := next.Files[path]; !retained {
			if err := os.Remove(filepath.Join(a.root, path)); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	manifestPath := a.manifestPath()
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		return err
	}
	return writeFileAtomically(manifestPath, append(data, '\n'), 0o644)
}

func (a *Adapter) Check() error {
	outputs, err := a.outputs()
	if err != nil {
		return err
	}
	current, hasManifest, err := a.readManifest()
	if err != nil {
		return err
	}
	if !hasManifest {
		return errors.New("no generated compatibility manifest found; run export")
	}
	if len(outputs) != len(current.Files) {
		return errors.New("generated output is stale; run export")
	}
	for path, want := range outputs {
		data, readErr := os.ReadFile(filepath.Join(a.root, path))
		if readErr != nil || !bytes.Equal(data, want) || current.Files[path] != digest(want) {
			return fmt.Errorf("generated output %q is stale or modified; run export", path)
		}
	}
	return nil
}

func (a *Adapter) Clean() error {
	m, exists, err := a.readManifest()
	if err != nil || !exists {
		return err
	}
	for path, hash := range m.Files {
		full := filepath.Join(a.root, path)
		data, readErr := os.ReadFile(full)
		if readErr == nil && digest(data) != hash {
			return fmt.Errorf("refusing to remove modified generated file %q", path)
		}
		if readErr == nil {
			if err := os.Remove(full); err != nil {
				return err
			}
		}
	}
	if err := os.Remove(a.manifestPath()); err != nil {
		return err
	}
	return removeEmptyDirs(filepath.Dir(a.manifestPath()))
}

func (a *Adapter) outputs() (map[string][]byte, error) {
	base := filepath.Join(a.root, ".agents")
	if stat, err := os.Stat(base); err != nil || !stat.IsDir() {
		return nil, errors.New(".agents directory is required")
	}
	if err := a.unsupported(base); err != nil {
		return nil, err
	}
	outputs, err := a.renderer.render(base)
	if err != nil {
		return nil, err
	}
	if err := validateOutputPaths(outputs); err != nil {
		return nil, err
	}
	return outputs, nil
}

func (a *Adapter) unsupported(base string) error {
	statusByCategory, err := targetCategoryStatus(a.root, a.Target(), false)
	if err != nil {
		return err
	}
	if len(statusByCategory) > 0 {
		if err := validateCategoryStatusAlignment(a.Target(), statusByCategory, a.renderer.unsupportedCategories()); err != nil {
			return err
		}
		var populated []string
		for category, status := range statusByCategory {
			if status != categoryStatusUnsupported {
				continue
			}
			dir := filepath.Join(base, category)
			entries, err := os.ReadDir(dir)
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if entry.Name() != "README.md" {
					populated = append(populated, category)
					break
				}
			}
		}
		sort.Strings(populated)
		if len(populated) > 0 && !a.allowUnsupported {
			return fmt.Errorf("unsupported populated categories: %s (pass --allow-unsupported to acknowledge)", strings.Join(populated, ", "))
		}
		return nil
	}

	var populated []string
	for _, name := range a.renderer.unsupportedCategories() {
		dir := filepath.Join(base, name)
		entries, err := os.ReadDir(dir)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.Name() != "README.md" {
				populated = append(populated, name)
				break
			}
		}
	}
	if len(populated) > 0 && !a.allowUnsupported {
		return fmt.Errorf("unsupported populated categories: %s (pass --allow-unsupported to acknowledge)", strings.Join(populated, ", "))
	}
	return nil
}

func (a *Adapter) readManifest() (manifest, bool, error) {
	return readManifestFile(a.manifestPath(), "adapter", "")
}

func (a *Adapter) readImportManifest() (manifest, bool, error) {
	return readManifestFile(a.importManifestPath(), "import", a.Target())
}

func readManifestFile(path, kind, target string) (manifest, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return manifest{}, false, nil
	}
	if err != nil {
		return manifest{}, false, err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil || m.Version != 1 {
		return manifest{}, false, fmt.Errorf("invalid %s manifest", kind)
	}
	if target != "" && m.Target != target {
		return manifest{}, false, fmt.Errorf("invalid %s manifest target %q", kind, m.Target)
	}
	for path := range m.Files {
		clean := filepath.Clean(path)
		if path == "" || filepath.IsAbs(path) || clean == "." || clean == ".." ||
			strings.HasPrefix(clean, ".."+string(filepath.Separator)) ||
			filepath.ToSlash(clean) != path {
			return manifest{}, false, fmt.Errorf("invalid %s manifest path %q", kind, path)
		}
	}
	return m, true, nil
}

func (a *Adapter) manifestPath() string {
	return filepath.Join(a.root, a.renderer.manifestDirectory(), manifestName)
}

func (a *Adapter) importManifestPath() string {
	return filepath.Join(a.root, ".agents", importManifestPrefix+a.Target()+".json")
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func readSourceFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("refusing symlinked source file %s", path)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("source file %s is not regular", path)
	}
	return os.ReadFile(path)
}

func validateOutputPaths(outputs map[string][]byte) error {
	for path := range outputs {
		clean := filepath.Clean(path)
		if path == "" || filepath.IsAbs(path) || clean == "." || clean == ".." ||
			strings.HasPrefix(clean, ".."+string(filepath.Separator)) ||
			filepath.ToSlash(clean) != path {
			return fmt.Errorf("unsafe generated output path %q", path)
		}
	}
	return nil
}

func writeFileAtomically(path string, data []byte, mode fs.FileMode) (err error) {
	file, err := os.CreateTemp(filepath.Dir(path), ".oda-*")
	if err != nil {
		return err
	}
	closed := false
	defer func() {
		if !closed {
			if closeErr := file.Close(); err == nil && closeErr != nil {
				err = closeErr
			}
		}
		if removeErr := os.Remove(file.Name()); err == nil && removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
			err = removeErr
		}
	}()
	if err := file.Chmod(mode); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	closed = true
	return os.Rename(file.Name(), path)
}

func removeEmptyDirs(root string) error {
	var directories []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			directories = append(directories, path)
		}
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	sort.Slice(directories, func(i, j int) bool { return len(directories[i]) > len(directories[j]) })
	for _, dir := range directories {
		if err := os.Remove(dir); err != nil && !errors.Is(err, syscall.ENOTEMPTY) && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}
