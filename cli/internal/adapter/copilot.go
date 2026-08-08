package adapter

type copilotRenderer struct{}

func (copilotRenderer) target() string {
	return "copilot"
}

func (copilotRenderer) manifestDirectory() string {
	return ".github"
}

func (copilotRenderer) unsupportedCategories() []string {
	return commonUnsupportedCategories
}

func (copilotRenderer) render(base string) (map[string][]byte, error) {
	out := make(map[string][]byte)
	if err := instructions(base, ".github/copilot-instructions.md", out); err != nil {
		return nil, err
	}
	if err := rules(base, ".github/instructions", ".instructions.md", out); err != nil {
		return nil, err
	}
	if err := agents(base, ".github/agents", ".agent.md", out); err != nil {
		return nil, err
	}
	if err := copyJSON(base, "hooks", ".github/hooks", out); err != nil {
		return nil, err
	}
	if err := mcpJSON(base, ".github/mcp.json", out); err != nil {
		return nil, err
	}
	if err := copySkills(base, ".github/skills", out); err != nil {
		return nil, err
	}
	return out, nil
}
