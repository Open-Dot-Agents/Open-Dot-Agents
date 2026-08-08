package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	agentsSchemaFile   = ".agents/schema/v1/agents.schema.json"
	mappingsSchemaFile = ".agents/schema/v1/mappings.schema.json"
	mappingsFile       = ".agents/mappings.yaml"
	agentsManifestFile = ".agents/manifest.json"
)

const (
	CategoryStatusSupported   = "supported"
	CategoryStatusUnsupported = "unsupported"
	CategoryStatusMapped      = "mapped"
	CategoryStatusPartial     = "partial"
)

var (
	categoryStatusValues = map[string]struct{}{
		CategoryStatusSupported:   {},
		CategoryStatusUnsupported: {},
		CategoryStatusMapped:      {},
		CategoryStatusPartial:     {},
	}
)

var targetIDPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// SchemaMetadata is the schema-aware metadata extracted from agents.schema.json and mappings schema.
type SchemaMetadata struct {
	Categories             []string
	FilePatternsByCategory map[string][]string
}

type schemaUnavailableError struct {
	Path string
}

func (e schemaUnavailableError) Error() string {
	return fmt.Sprintf("schema not available: %s", e.Path)
}

func ValidateContracts(root, target string, strict bool) error {
	metadata, err := LoadMetadata(root, strict)
	if err != nil {
		return err
	}
	if metadata == nil {
		return nil
	}
	if strict {
		if err := ValidateAgentsManifestCompatibility(root); err != nil {
			return err
		}
		if _, err := TargetCategoryStatus(root, target, true); err != nil {
			return err
		}
	}
	return ValidateCategoryFilePatterns(root, metadata)
}

func TargetCategoryStatus(root, target string, strict bool) (map[string]string, error) {
	metadata, err := LoadMetadata(root, strict)
	if err != nil {
		return nil, err
	}
	if metadata == nil {
		return nil, nil
	}
	targetsStatus, err := parseMappingsStatus(root, metadata, strict)
	if err != nil {
		return nil, err
	}
	if targetsStatus == nil {
		return nil, nil
	}
	status, ok := targetsStatus[target]
	if !ok {
		return nil, fmt.Errorf("mappings has no status block for target %q", target)
	}
	return status, nil
}

func LoadMetadata(root string, strict bool) (*SchemaMetadata, error) {
	agentsData, err := os.ReadFile(filepath.Join(root, agentsSchemaFile))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if strict {
				return nil, schemaUnavailableError{Path: agentsSchemaFile}
			}
			return nil, nil
		}
		return nil, err
	}
	mappingsData, err := os.ReadFile(filepath.Join(root, mappingsSchemaFile))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if strict {
				return nil, schemaUnavailableError{Path: mappingsSchemaFile}
			}
			return nil, nil
		}
		return nil, err
	}

	var agentsDoc map[string]any
	if err := json.Unmarshal(agentsData, &agentsDoc); err != nil {
		return nil, fmt.Errorf("invalid agents schema file %s: %w", agentsSchemaFile, err)
	}
	var mappingsDoc map[string]any
	if err := json.Unmarshal(mappingsData, &mappingsDoc); err != nil {
		return nil, fmt.Errorf("invalid mappings schema file %s: %w", mappingsSchemaFile, err)
	}

	requiredTop, err := requiredStringSliceFromObject(agentsDoc, "required")
	if err != nil {
		return nil, err
	}
	if !containsAll(requiredTop, []string{"format_version", "canonical_root", "categories"}) {
		return nil, fmt.Errorf("agents schema missing required top-level keys")
	}

	properties, ok := asObject(agentsDoc["properties"])
	if !ok {
		return nil, fmt.Errorf("agents schema missing properties block")
	}
	formatVersionRaw, ok := asObject(properties["format_version"])
	if !ok {
		return nil, fmt.Errorf("agents schema missing format_version definition")
	}
	constantVersion, ok := formatVersionRaw["const"]
	if !ok || asInt(constantVersion) != 1 {
		return nil, fmt.Errorf("agents schema requires format_version const == 1")
	}
	canonicalRootRaw, ok := asObject(properties["canonical_root"])
	if !ok {
		return nil, fmt.Errorf("agents schema missing canonical_root definition")
	}
	constantRoot, ok := canonicalRootRaw["const"]
	if !ok || !asStringMatches(constantRoot, ".agents") {
		return nil, fmt.Errorf("agents schema requires canonical_root const == \".agents\"")
	}

	definitions, ok := asObject(agentsDoc["$defs"])
	if !ok {
		return nil, fmt.Errorf("agents schema missing $defs block")
	}
	categoryRegistry, ok := asObject(definitions["category_registry"])
	if !ok {
		return nil, fmt.Errorf("agents schema missing category_registry definition")
	}
	categories, err := requiredStringSliceFromObject(categoryRegistry, "required")
	if err != nil || len(categories) == 0 {
		return nil, fmt.Errorf("agents schema missing required category list")
	}

	mappingsDefs, ok := asObject(mappingsDoc["$defs"])
	if !ok {
		return nil, fmt.Errorf("mappings schema missing $defs block")
	}
	statusMatrix, ok := asObject(mappingsDefs["status_matrix"])
	if !ok {
		return nil, fmt.Errorf("mappings schema missing status_matrix definition")
	}
	mappingCategories, err := requiredStringSliceFromObject(statusMatrix, "required")
	if err != nil || len(mappingCategories) == 0 {
		return nil, fmt.Errorf("mappings schema missing required status categories")
	}
	if !sameStringSet(categories, mappingCategories) {
		return nil, fmt.Errorf("agents schema and mappings schema category sets differ")
	}

	categoryDefs, ok := asObject(properties["categories"])
	if !ok {
		return nil, fmt.Errorf("agents schema missing categories property")
	}
	if _, ok := asString(categoryDefs["$ref"]); !ok {
		return nil, fmt.Errorf("agents schema categories property missing $ref")
	}

	categoryRegistryProperties, ok := asObject(categoryRegistry["properties"])
	if !ok {
		return nil, fmt.Errorf("agents schema missing category_registry properties")
	}
	filePatternsByCategory := map[string][]string{}
	for _, category := range categories {
		rawCatSpec, ok := categoryRegistryProperties[category]
		if !ok {
			return nil, fmt.Errorf("agents schema missing property for category %q", category)
		}
		categorySpec, ok := asObject(rawCatSpec)
		if !ok {
			return nil, fmt.Errorf("agents schema category %q is not an object", category)
		}
		filePatternsByCategory[category] = extractFilePatterns(categorySpec, definitions)
	}

	return &SchemaMetadata{Categories: categories, FilePatternsByCategory: filePatternsByCategory}, nil
}

func parseMappingsStatus(root string, metadata *SchemaMetadata, strict bool) (map[string]map[string]string, error) {
	mappingsData, err := os.ReadFile(filepath.Join(root, mappingsFile))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if strict {
				return nil, schemaUnavailableError{Path: mappingsFile}
			}
			return nil, nil
		}
		return nil, err
	}
	var mappings map[string]any
	if err := yaml.Unmarshal(mappingsData, &mappings); err != nil {
		return nil, fmt.Errorf("invalid mappings file %s: %w", mappingsFile, err)
	}

	agentsBlock, ok := asObject(mappings[".agents"])
	if !ok {
		return nil, fmt.Errorf("mappings missing .agents root block")
	}
	manifestRaw, ok := asObject(agentsBlock["manifest"])
	if !ok {
		return nil, fmt.Errorf("mappings missing agents.manifest block")
	}
	if asInt(manifestRaw["format_version"]) != 1 {
		return nil, fmt.Errorf("mappings manifest format_version must be 1")
	}
	if !asStringMatches(manifestRaw["canonical_root"], ".agents") {
		return nil, fmt.Errorf("mappings manifest canonical_root must be \".agents\"")
	}

	targetsRaw, ok := asObject(agentsBlock["targets"])
	if !ok {
		return nil, fmt.Errorf("mappings missing agents.targets block")
	}
	targets := map[string]struct{}{}
	for key, value := range targetsRaw {
		targetSpec, ok := asObject(value)
		if !ok {
			return nil, fmt.Errorf("mappings target %q must be an object", key)
		}
		targetID, ok := asString(targetSpec["id"])
		if !ok || targetID == "" {
			return nil, fmt.Errorf("mappings target %q missing id", key)
		}
		if targetID != key {
			return nil, fmt.Errorf("mappings target key %q does not match id %q", key, targetID)
		}
		if !targetIDPattern.MatchString(targetID) {
			return nil, fmt.Errorf("mappings target %q has invalid id; expected [a-z0-9-]+", targetID)
		}
		name, ok := asString(targetSpec["name"])
		if !ok || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("mappings target %q missing name", key)
		}
		docs, err := asStringSlice(targetSpec["docs"])
		if err != nil || len(docs) == 0 {
			return nil, fmt.Errorf("mappings target %q docs must be a non-empty list", key)
		}
		targets[key] = struct{}{}
		_ = docs
		_ = name
	}

	statusRaw, ok := asObject(agentsBlock["status"])
	if !ok {
		return nil, fmt.Errorf("mappings missing agents.status block")
	}
	targetStatusByCategory := map[string]map[string]string{}
	for _, category := range metadata.Categories {
		rawCategoryStatus, ok := statusRaw[category]
		if !ok {
			return nil, fmt.Errorf("mappings missing status for category %q", category)
		}
		categoryStatus, ok := asObject(rawCategoryStatus)
		if !ok {
			return nil, fmt.Errorf("mappings category %q status block must be an object", category)
		}
		for target, rawValue := range categoryStatus {
			if _, ok := targets[target]; !ok {
				return nil, fmt.Errorf("mappings references unknown target %q for category %q", target, category)
			}
			value, ok := asString(rawValue)
			if !ok {
				return nil, fmt.Errorf("mappings category %q target %q status must be a string", category, target)
			}
			if _, ok := categoryStatusValues[value]; !ok {
				return nil, fmt.Errorf("mappings category %q target %q status %q is invalid", category, target, value)
			}
			targetStatus, ok := targetStatusByCategory[target]
			if !ok {
				targetStatus = map[string]string{}
				targetStatusByCategory[target] = targetStatus
			}
			targetStatus[category] = value
		}
	}

	for category := range statusRaw {
		found := false
		for _, declared := range metadata.Categories {
			if category == declared {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("mappings contains unknown category %q in status block", category)
		}
	}

	for target := range targets {
		targetStatus, ok := targetStatusByCategory[target]
		if !ok || len(targetStatus) != len(metadata.Categories) {
			return nil, fmt.Errorf("mappings target %q is missing required category status entries", target)
		}
	}

	return targetStatusByCategory, nil
}

func ValidateCategoryFilePatterns(root string, metadata *SchemaMetadata) error {
	base := filepath.Join(root, ".agents")
	if _, err := os.Stat(base); err != nil {
		return nil
	}

	return filepath.WalkDir(base, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(base, current)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if relative == "." {
			return nil
		}

		parts := strings.Split(relative, "/")
		if len(parts) > 0 && parts[0] == "schema" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Base(relative) == "README.md" {
			return nil
		}
		if len(parts) == 1 {
			return nil
		}
		category := parts[0]
		patterns := metadata.FilePatternsByCategory[category]
		if len(patterns) == 0 {
			return nil
		}
		matched := false
		for _, pattern := range patterns {
			ok, err := path.Match(pattern, relative)
			if err != nil {
				return fmt.Errorf("invalid file pattern %q for category %q", pattern, category)
			}
			if ok {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("entry %q does not match allowed file patterns for \"%s\" category", relative, category)
		}
		return nil
	})
}

func ValidateCategoryStatusAlignment(target string, statusByCategory map[string]string, unsupportedCategories []string) error {
	unsupportedByRenderer := map[string]struct{}{}
	for _, category := range unsupportedCategories {
		unsupportedByRenderer[category] = struct{}{}
	}
	var mismatches []string
	categories := make([]string, 0, len(statusByCategory))
	for category := range statusByCategory {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	for _, category := range categories {
		isUnsupportedByMapping := statusByCategory[category] == CategoryStatusUnsupported
		_, rendererUnsupported := unsupportedByRenderer[category]
		if isUnsupportedByMapping == rendererUnsupported {
			continue
		}
		if isUnsupportedByMapping {
			mismatches = append(mismatches, fmt.Sprintf("mappings marks %s as unsupported but %s renderer still emits this category", category, target))
		} else {
			mismatches = append(mismatches, fmt.Sprintf("%s is unsupported by %s renderer but mappings marks it %s", category, target, statusByCategory[category]))
		}
	}
	if len(mismatches) == 0 {
		return nil
	}
	return fmt.Errorf("target %s status mismatch: %s", target, strings.Join(mismatches, "; "))
}

func ValidateAgentsManifestCompatibility(root string) error {
	data, err := os.ReadFile(filepath.Join(root, agentsManifestFile))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("invalid .agents/manifest.json: %w", err)
	}
	if asInt(manifest["format_version"]) != 1 {
		return fmt.Errorf(".agents/manifest.json requires format_version == \"1\"")
	}
	if !asStringMatches(manifest["canonical_root"], ".agents") {
		return fmt.Errorf(".agents/manifest.json requires canonical_root == \".agents\"")
	}
	return nil
}

func extractFilePatterns(spec map[string]any, definitions map[string]any) []string {
	patterns := map[string]struct{}{}
	visitFilePatterns(spec, definitions, map[string]struct{}{}, patterns)
	var keys []string
	for pattern := range patterns {
		keys = append(keys, pattern)
	}
	sort.Strings(keys)
	return keys
}

func visitFilePatterns(spec map[string]any, definitions map[string]any, visitedRefs map[string]struct{}, patterns map[string]struct{}) {
	refRaw, ok := spec["$ref"]
	if ok {
		ref, ok := asString(refRaw)
		if ok && strings.HasPrefix(ref, "#/$defs/") {
			name := strings.TrimPrefix(ref, "#/$defs/")
			if _, seen := visitedRefs[name]; !seen {
				visitedRefs[name] = struct{}{}
				if target, ok := asObject(definitions[name]); ok {
					for _, pattern := range filePatternsFromNode(target["file_patterns"]) {
						patterns[pattern] = struct{}{}
					}
					visitFilePatterns(target, definitions, visitedRefs, patterns)
				}
			}
		}
	}
	if allOfRaw, ok := spec["allOf"]; ok {
		if allOf, ok := allOfRaw.([]any); ok {
			for _, item := range allOf {
				asMap, ok := item.(map[string]any)
				if !ok {
					continue
				}
				for _, pattern := range filePatternsFromNode(asMap["file_patterns"]) {
					patterns[pattern] = struct{}{}
				}
				visitFilePatterns(asMap, definitions, visitedRefs, patterns)
			}
		}
	}
	if propertiesRaw, ok := spec["properties"]; ok {
		properties, ok := asObject(propertiesRaw)
		if ok {
			if filePatternsRaw, ok := properties["file_patterns"]; ok {
				for _, pattern := range filePatternsFromNode(filePatternsRaw) {
					patterns[pattern] = struct{}{}
				}
			}
		}
	}
	for _, pattern := range filePatternsFromNode(spec["file_patterns"]) {
		patterns[pattern] = struct{}{}
	}
}

func filePatternsFromNode(node any) []string {
	out := make([]string, 0)
	for _, pattern := range filePatternsFromNodeRaw(node) {
		if strings.TrimSpace(pattern) != "" {
			out = append(out, pattern)
		}
	}
	sort.Strings(out)
	return dedupeStrings(out)
}

func filePatternsFromNodeRaw(node any) []string {
	switch typed := node.(type) {
	case string:
		return []string{typed}
	case []any:
		patterns := make([]string, 0, len(typed))
		for _, item := range typed {
			patterns = append(patterns, filePatternsFromNodeRaw(item)...)
		}
		return patterns
	case map[string]any:
		if constant, ok := typed["const"]; ok {
			patterns := filePatternsFromNodeRaw(constant)
			if len(patterns) > 0 {
				return patterns
			}
		}
		if enums, ok := typed["enum"]; ok {
			patterns := filePatternsFromNodeRaw(enums)
			if len(patterns) > 0 {
				return patterns
			}
		}
		if items, ok := typed["items"]; ok {
			patterns := filePatternsFromNodeRaw(items)
			if len(patterns) > 0 {
				return patterns
			}
		}
	}
	return nil
}

func requiredStringSliceFromObject(value map[string]any, key string) ([]string, error) {
	raw, ok := value[key]
	if !ok {
		return nil, fmt.Errorf("missing required key %q", key)
	}
	values, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("expected array for %q, got %T", key, raw)
	}
	out := make([]string, 0, len(values))
	for i, value := range values {
		item, ok := asString(value)
		if !ok {
			return nil, fmt.Errorf("expected string at index %d for %q", i, key)
		}
		out = append(out, item)
	}
	return out, nil
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func asObject(value any) (map[string]any, bool) {
	result, ok := value.(map[string]any)
	return result, ok
}

func asString(value any) (string, bool) {
	result, ok := value.(string)
	return result, ok
}

func asStringMatches(value any, expected string) bool {
	got, ok := asString(value)
	return ok && got == expected
}

func asInt(value any) int {
	switch cast := value.(type) {
	case int:
		return cast
	case int64:
		return int(cast)
	case uint:
		return int(cast)
	case uint64:
		return int(cast)
	case float64:
		if cast != float64(int(cast)) {
			return 0
		}
		return int(cast)
	default:
		return 0
	}
}

func asStringSlice(value any) ([]string, error) {
	raw, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("expected []any, got %T", value)
	}
	out := make([]string, 0, len(raw))
	for i, value := range raw {
		item, ok := asString(value)
		if !ok {
			return nil, fmt.Errorf("expected string at index %d", i)
		}
		out = append(out, item)
	}
	return out, nil
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	left := append([]string(nil), a...)
	right := append([]string(nil), b...)
	sort.Strings(left)
	sort.Strings(right)
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func containsAll(values, required []string) bool {
	set := map[string]struct{}{}
	for _, value := range values {
		set[value] = struct{}{}
	}
	for _, value := range required {
		if _, ok := set[value]; !ok {
			return false
		}
	}
	return true
}
