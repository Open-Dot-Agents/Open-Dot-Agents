package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/codexconfig"
)

const (
	defaultSchemaURL = "https://developers.openai.com/codex/config-schema.json"
	maxSchemaBytes   = 4 << 20
)

func main() {
	configPath := flag.String("config", "", "path to generated Codex config.toml")
	schemaURL := flag.String("schema", defaultSchemaURL, "official Codex config JSON Schema URL")
	flag.Parse()
	if *configPath == "" || flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: codex-config-validator --config PATH [--schema URL]")
		os.Exit(2)
	}

	config, err := os.Open(*configPath)
	if err != nil {
		fatalf("open config: %v", err)
	}
	defer config.Close()

	client := &http.Client{Timeout: 20 * time.Second}
	response, err := client.Get(*schemaURL)
	if err != nil {
		fatalf("download official Codex config schema: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fatalf("download official Codex config schema: HTTP %s", response.Status)
	}

	schema, err := io.ReadAll(io.LimitReader(response.Body, maxSchemaBytes+1))
	if err != nil {
		fatalf("read official Codex config schema: %v", err)
	}
	if len(schema) > maxSchemaBytes {
		fatalf("official Codex config schema exceeds %d bytes", maxSchemaBytes)
	}
	if err := codexconfig.Validate(config, bytes.NewReader(schema), *schemaURL); err != nil {
		fatalf("%v", err)
	}
	fmt.Printf("Codex config matches %s\n", *schemaURL)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "codex-config-validator: "+format+"\n", args...)
	os.Exit(1)
}
