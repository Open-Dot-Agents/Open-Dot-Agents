package adapter

type claudeRenderer struct{}

func (claudeRenderer) target() string {
	return "claude"
}

func (claudeRenderer) manifestDirectory() string {
	return ".claude"
}

func (claudeRenderer) unsupportedCategories() []string {
	return commonUnsupportedCategories
}

func (claudeRenderer) render(base string) (map[string][]byte, error) {
	out := make(map[string][]byte)
	if err := instructions(base, "CLAUDE.md", out); err != nil {
		return nil, err
	}
	if err := claudeRules(base, out); err != nil {
		return nil, err
	}
	if err := agents(base, ".claude/agents", ".md", out); err != nil {
		return nil, err
	}
	if err := copySkills(base, ".claude/skills", out); err != nil {
		return nil, err
	}
	if err := mcpJSON(base, ".mcp.json", out); err != nil {
		return nil, err
	}
	return out, claudeHooks(base, out)
}
