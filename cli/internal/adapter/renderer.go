package adapter

import (
	"fmt"
	"strings"
)

type renderer interface {
	target() string
	manifestDirectory() string
	unsupportedCategories() []string
	render(base string) (map[string][]byte, error)
}

type TargetCapabilities struct {
	Instructions bool `json:"instructions"`
	Rules        bool `json:"rules"`
	Agents       bool `json:"agents"`
	Hooks        bool `json:"hooks"`
	Skills       bool `json:"skills"`
	MCP          bool `json:"mcp"`
}

type TargetInfo struct {
	ID                    string             `json:"id"`
	ManifestDirectory     string             `json:"manifestDirectory"`
	UnsupportedCategories []string           `json:"unsupportedCategories"`
	Capabilities          TargetCapabilities `json:"capabilities"`
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
			Capabilities:          capabilitiesFrom(rendererUnsupportedToMap(r.unsupportedCategories())),
		})
	}
	return infos
}

func rendererUnsupportedToMap(unsupported []string) map[string]bool {
	supported := map[string]bool{
		"instructions": true,
		"rules":        true,
		"agents":       true,
		"hooks":        true,
		"skills":       true,
		"tools":        true,
	}
	for _, category := range unsupported {
		supported[category] = false
	}
	return supported
}

func capabilitiesFrom(supported map[string]bool) TargetCapabilities {
	return TargetCapabilities{
		Instructions: supported["instructions"],
		Rules:        supported["rules"],
		Agents:       supported["agents"],
		Hooks:        supported["hooks"],
		Skills:       supported["skills"],
		MCP:          supported["tools"],
	}
}
