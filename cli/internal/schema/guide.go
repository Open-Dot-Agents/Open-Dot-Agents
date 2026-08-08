package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type GuideSchemaContract struct {
	CanonicalRoot  string `json:"canonicalRoot"`
	AgentsSchema   string `json:"agentsSchema"`
	MappingsSchema string `json:"mappingsSchema"`
	AgentsManifest string `json:"agentsManifest"`
}

type VendorGuide struct {
	Contract           GuideSchemaContract `json:"contract"`
	Categories         []string            `json:"categories"`
	Targets            []VendorGuideTarget `json:"targets"`
	CompatibilityNotes []string            `json:"compatibilityNotes"`
}

type VendorGuideTarget struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Docs        []string          `json:"docs"`
	Status      map[string]string `json:"status"`
	StatusNotes map[string]string `json:"statusNotes"`
	Notes       []string          `json:"notes"`
}

func GenerateVendorGuide(root string, targetFilter []string) (*VendorGuide, error) {
	metadata, err := LoadMetadata(root, true)
	if err != nil {
		return nil, err
	}
	if metadata == nil {
		return nil, fmt.Errorf("schema metadata unavailable")
	}

	statusByTarget, err := parseMappingsStatus(root, metadata, true)
	if err != nil {
		return nil, err
	}

	mappingsData, err := os.ReadFile(filepath.Join(root, mappingsFile))
	if err != nil {
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

	targets, err := parseGuideTargets(agentsBlock)
	if err != nil {
		return nil, err
	}
	statusNotes := parseStatusNotes(agentsBlock)
	compatNotes, err := parseManifestNotes(root)
	if err != nil {
		return nil, err
	}

	normalizedTargets := make([]string, 0, len(targetFilter))
	seen := make(map[string]struct{})
	if len(targetFilter) == 0 {
		for target := range statusByTarget {
			targetFilter = append(targetFilter, target)
		}
		sort.Strings(targetFilter)
	}
	for _, target := range targetFilter {
		if _, ok := statusByTarget[target]; !ok {
			return nil, fmt.Errorf("mappings has no status block for target %q", target)
		}
		if _, ok := targets[target]; !ok {
			return nil, fmt.Errorf("mappings has no target block for %q", target)
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		normalizedTargets = append(normalizedTargets, target)
	}

	categories := append([]string(nil), metadata.Categories...)
	sort.Strings(categories)

	guideTargets := make([]VendorGuideTarget, 0, len(normalizedTargets))
	for _, target := range normalizedTargets {
		spec := targets[target]
		docs := toStringSlice(spec["docs"])
		name := strings.TrimSpace(asStringOrBlank(spec["name"]))
		if name == "" {
			name = target
		}

		statusNotesForTarget := make(map[string]string)
		for category, notes := range statusNotes {
			if note, ok := notes[target]; ok {
				statusNotesForTarget[category] = note
			}
		}
		status := mapFromMap(statusByTarget[target])

		guideTargets = append(guideTargets, VendorGuideTarget{
			ID:          target,
			Name:        name,
			Docs:        docs,
			Status:      status,
			StatusNotes: statusNotesForTarget,
			Notes:       compatibilityForTarget(target, compatNotes),
		})
	}

	return &VendorGuide{
		Contract: GuideSchemaContract{
			CanonicalRoot:  ".agents",
			AgentsSchema:   agentsSchemaFile,
			MappingsSchema: mappingsSchemaFile,
			AgentsManifest: agentsManifestFile,
		},
		Categories:         categories,
		Targets:            guideTargets,
		CompatibilityNotes: compatNotes.general,
	}, nil
}

func RenderVendorGuideMarkdown(guide *VendorGuide) string {
	if guide == nil {
		return ""
	}

	var out strings.Builder
	out.WriteString("# Vendor implementation guide\n\n")
	out.WriteString("This guide is generated from repository schema and mapping metadata.\n\n")

	out.WriteString("## Contract metadata\n\n")
	out.WriteString("- Canonical root: " + toMarkdownInlineCode(guide.Contract.CanonicalRoot) + "\n")
	out.WriteString("- Agents schema: " + guide.Contract.AgentsSchema + "\n")
	out.WriteString("- Mappings schema: " + guide.Contract.MappingsSchema + "\n")
	if guide.Contract.AgentsManifest != "" {
		out.WriteString("- Manifest compatibility file: " + guide.Contract.AgentsManifest + "\n")
	}
	if len(guide.Categories) == 0 {
		out.WriteString("- Categories: none\n")
	} else {
		out.WriteString("- Categories: " + strings.Join(quoteItems(guide.Categories), ", ") + "\n")
	}
	out.WriteString("\n")

	if len(guide.CompatibilityNotes) > 0 {
		out.WriteString("## Compatibility notes\n\n")
		for _, note := range guide.CompatibilityNotes {
			out.WriteString("- " + note + "\n")
		}
		out.WriteString("\n")
	}

	for _, target := range guide.Targets {
		name := target.ID
		if strings.TrimSpace(target.Name) != "" && target.Name != target.ID {
			name = target.Name + " (" + target.ID + ")"
		}
		out.WriteString("## Target: " + name + "\n\n")
		if len(target.Docs) > 0 {
			out.WriteString("### Target documentation\n")
			for _, doc := range target.Docs {
				out.WriteString("- " + toMarkdownLink(doc) + "\n")
			}
			out.WriteString("\n")
		}
		if len(target.Notes) > 0 {
			out.WriteString("### Vendor notes\n")
			for _, note := range target.Notes {
				out.WriteString("- " + note + "\n")
			}
			out.WriteString("\n")
		}

		out.WriteString("### Category status\n\n")
		out.WriteString("| Category | Status | Notes |\n")
		out.WriteString("| --- | --- | --- |\n")
		for _, category := range guide.Categories {
			status := target.Status[category]
			if status == "" {
				status = "missing"
			}
			note := target.StatusNotes[category]
			out.WriteString("| " + escapeTableCell(category) + " | " + escapeTableCell(status) + " | " + escapeTableCell(note) + " |\n")
		}
		out.WriteString("\n")
	}
	return out.String()
}

type targetManifestNotes struct {
	general []string
}

func parseManifestNotes(root string) (targetManifestNotes, error) {
	manifestPath := filepath.Join(root, agentsManifestFile)
	notes := targetManifestNotes{}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return notes, nil
		}
		return notes, err
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return notes, fmt.Errorf("invalid .agents/manifest.json: %w", err)
	}
	notes.general = append(notes.general, toStringSlice(payload["compatibility_notes"])...)
	notes.general = append(notes.general, toStringSlice(payload["deprecations"])...)
	return notes, nil
}

func parseGuideTargets(root map[string]any) (map[string]map[string]any, error) {
	rawTargets, ok := root["targets"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("mappings missing agents.targets block")
	}
	targets := make(map[string]map[string]any)
	for key, rawTarget := range rawTargets {
		targetSpec, ok := asObject(rawTarget)
		if !ok {
			return nil, fmt.Errorf("mapping target %q must be an object", key)
		}
		targetID, ok := asString(targetSpec["id"])
		if !ok || strings.TrimSpace(targetID) == "" {
			return nil, fmt.Errorf("mapping target %q missing id", key)
		}
		if targetID != key {
			return nil, fmt.Errorf("mappings target key %q does not match id %q", key, targetID)
		}
		if _, ok := targetSpec["docs"]; !ok {
			return nil, fmt.Errorf("mappings target %q missing docs", key)
		}
		if _, err := asStringSlice(targetSpec["docs"]); err != nil {
			return nil, fmt.Errorf("mappings target %q docs must be a non-empty list", key)
		}
		targets[key] = targetSpec
	}
	return targets, nil
}

func parseStatusNotes(root map[string]any) map[string]map[string]string {
	notes := map[string]map[string]string{}
	rawNotes, ok := root["status_notes"].(map[string]any)
	if !ok {
		return notes
	}
	for category, rawValues := range rawNotes {
		values, ok := rawValues.(map[string]any)
		if !ok {
			continue
		}
		noteForTarget := map[string]string{}
		for target, rawNote := range values {
			note, ok := rawNote.(string)
			if ok && strings.TrimSpace(note) != "" {
				noteForTarget[target] = strings.TrimSpace(note)
			}
		}
		if len(noteForTarget) > 0 {
			notes[category] = noteForTarget
		}
	}
	return notes
}

func compatibilityForTarget(_ string, notes targetManifestNotes) []string {
	if len(notes.general) == 0 {
		return nil
	}
	return append([]string{}, notes.general...)
}

func mapFromMap(source map[string]string) map[string]string {
	copied := make(map[string]string, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
}

func toStringSlice(value any) []string {
	parts, err := asStringSlice(value)
	if err != nil {
		return nil
	}
	return parts
}

func asStringOrBlank(value any) string {
	raw, ok := value.(string)
	if !ok {
		return ""
	}
	return raw
}

func quoteItems(values []string) []string {
	quoted := make([]string, 0, len(values))
	for _, item := range values {
		quoted = append(quoted, toMarkdownInlineCode(item))
	}
	return quoted
}

func toMarkdownLink(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return "[" + trimmed + "](" + trimmed + ")"
}

func toMarkdownInlineCode(value string) string {
	return "`" + strings.TrimSpace(value) + "`"
}

func escapeTableCell(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.ReplaceAll(trimmed, "|", "\\|")
	trimmed = strings.ReplaceAll(trimmed, "\n", "<br>")
	return trimmed
}
