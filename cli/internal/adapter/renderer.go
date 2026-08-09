package adapter

import (
	"fmt"
	"strings"
)

const (
	categoryStatusSupported   = "supported"
	categoryStatusUnsupported = "unsupported"
	categoryStatusMapped      = "mapped"
	categoryStatusPartial     = "partial"
)

type renderer interface {
	target() string
	manifestDirectory() string
	unsupportedCategories() []string
	render(base string) (map[string][]byte, error)
}

type TargetInfo struct {
	ID                    string            `json:"id"`
	ManifestDirectory     string            `json:"manifestDirectory"`
	UnsupportedCategories []string          `json:"unsupportedCategories"`
	CategoryStatuses      map[string]string `json:"categoryStatuses"`
}

var categoryOrder = []string{
	"instructions", "rules", "agents", "guardrails", "hooks", "memories",
	"permissions", "plugins", "profiles", "prompts", "settings", "skills", "tools",
}

var statusesByTarget = map[string]map[string]string{
	"copilot": {
		"instructions": categoryStatusSupported, "rules": categoryStatusSupported,
		"agents": categoryStatusMapped, "guardrails": categoryStatusUnsupported,
		"hooks": categoryStatusSupported, "memories": categoryStatusUnsupported,
		"permissions": categoryStatusMapped, "plugins": categoryStatusMapped,
		"profiles": categoryStatusUnsupported, "prompts": categoryStatusMapped,
		"settings": categoryStatusMapped, "skills": categoryStatusMapped,
		"tools": categoryStatusMapped,
	},
	"codex": {
		"instructions": categoryStatusSupported, "rules": categoryStatusUnsupported,
		"agents": categoryStatusMapped, "guardrails": categoryStatusMapped,
		"hooks": categoryStatusMapped, "memories": categoryStatusMapped,
		"permissions": categoryStatusMapped, "plugins": categoryStatusMapped,
		"profiles": categoryStatusMapped, "prompts": categoryStatusUnsupported,
		"settings": categoryStatusMapped, "skills": categoryStatusSupported,
		"tools": categoryStatusMapped,
	},
	"claude": {
		"instructions": categoryStatusSupported, "rules": categoryStatusSupported,
		"agents": categoryStatusMapped, "guardrails": categoryStatusUnsupported,
		"hooks": categoryStatusSupported, "memories": categoryStatusUnsupported,
		"permissions": categoryStatusUnsupported, "plugins": categoryStatusUnsupported,
		"profiles": categoryStatusUnsupported, "prompts": categoryStatusUnsupported,
		"settings": categoryStatusUnsupported, "skills": categoryStatusMapped,
		"tools": categoryStatusMapped,
	},
}

var rendererRegistry = map[string]renderer{}
var rendererOrder []string

func init() {
	registerRenderer(copilotRenderer{})
	registerRenderer(codexRenderer{})
	registerRenderer(claudeRenderer{})
}

func registerRenderer(r renderer) {
	id := r.target()
	if _, exists := rendererRegistry[id]; exists {
		panic(fmt.Sprintf("duplicate target %q registration", id))
	}
	rendererRegistry[id] = r
	rendererOrder = append(rendererOrder, id)
}

func rendererFor(target string) (renderer, error) {
	renderer, exists := rendererRegistry[target]
	if exists {
		return renderer, nil
	}
	names := RegisteredTargets()
	return nil, fmt.Errorf("unknown target %q (want %s)", target, joinTargetNames(names))
}

func joinTargetNames(targets []string) string {
	if len(targets) == 0 {
		return ""
	}
	if len(targets) == 1 {
		return targets[0]
	}
	if len(targets) == 2 {
		return targets[0] + " or " + targets[1]
	}
	return strings.Join(targets[:len(targets)-1], ", ") + ", or " + targets[len(targets)-1]
}

func RegisteredTargets() []string {
	targets := make([]string, 0, len(rendererOrder))
	targets = append(targets, rendererOrder...)
	return targets
}

func TargetInfos() []TargetInfo {
	var infos []TargetInfo
	for _, id := range RegisteredTargets() {
		r := rendererRegistry[id]
		infos = append(infos, TargetInfo{
			ID:                    id,
			ManifestDirectory:     r.manifestDirectory(),
			UnsupportedCategories: append([]string{}, r.unsupportedCategories()...),
			CategoryStatuses:      categoryStatuses(id),
		})
	}
	return infos
}

func categoryStatuses(target string) map[string]string {
	configured := statusesByTarget[target]
	statuses := make(map[string]string, len(categoryOrder))
	for _, category := range categoryOrder {
		statuses[category] = configured[category]
	}
	return statuses
}
