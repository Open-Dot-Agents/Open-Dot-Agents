package codexconfig

import (
	"strings"
	"testing"
)

const testSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "model": {"type": "string"}
  }
}`

func TestValidateAcceptsMatchingTOML(t *testing.T) {
	err := Validate(strings.NewReader("model = \"gpt-5\"\n"), strings.NewReader(testSchema), "https://example.test/config-schema.json")
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsUnknownField(t *testing.T) {
	err := Validate(strings.NewReader("unknown = true\n"), strings.NewReader(testSchema), "https://example.test/config-schema.json")
	if err == nil || !strings.Contains(err.Error(), "additionalProperties") {
		t.Fatalf("Validate() error = %v, want additionalProperties failure", err)
	}
}

func TestValidateRejectsInvalidTOML(t *testing.T) {
	err := Validate(strings.NewReader("model = [\n"), strings.NewReader(testSchema), "https://example.test/config-schema.json")
	if err == nil || !strings.Contains(err.Error(), "parse Codex TOML") {
		t.Fatalf("Validate() error = %v, want TOML parse failure", err)
	}
}
