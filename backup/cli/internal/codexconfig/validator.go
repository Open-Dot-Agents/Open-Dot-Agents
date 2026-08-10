package codexconfig

import (
	"fmt"
	"io"

	"github.com/pelletier/go-toml/v2"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Validate checks a Codex TOML document against the supplied JSON Schema.
func Validate(config io.Reader, schema io.Reader, schemaLocation string) error {
	var document map[string]any
	if err := toml.NewDecoder(config).Decode(&document); err != nil {
		return fmt.Errorf("parse Codex TOML: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft7
	if err := compiler.AddResource(schemaLocation, schema); err != nil {
		return fmt.Errorf("load Codex config schema: %w", err)
	}
	compiled, err := compiler.Compile(schemaLocation)
	if err != nil {
		return fmt.Errorf("compile Codex config schema: %w", err)
	}
	if err := compiled.Validate(document); err != nil {
		return fmt.Errorf("Codex config does not match schema: %w", err)
	}
	return nil
}
