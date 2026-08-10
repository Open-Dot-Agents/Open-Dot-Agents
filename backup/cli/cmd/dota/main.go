package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/dotacore"
	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/specdata"
	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/pkg/adapterprotocol"
)

var version = "dev"

type outputEnvelope struct {
	OutputVersion int                                 `json:"outputVersion"`
	Command       string                              `json:"command"`
	Status        string                              `json:"status"`
	Adapter       *adapterprotocol.AdapterDescription `json:"adapter,omitempty"`
	Changes       *dotacore.ChangeSet                 `json:"changes,omitempty"`
	Data          any                                 `json:"data,omitempty"`
}

func main() {
	code, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "dota:", err)
	}
	os.Exit(code)
}

func run(args []string) (int, error) {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printHelp()
		return 0, nil
	}
	if args[0] == "--version" || args[0] == "version" {
		fmt.Printf("dota %s\n", currentVersion())
		return 0, nil
	}
	command := args[0]
	var err error
	switch command {
	case "init":
		err = runInit(args[1:])
	case "validate":
		err = runValidate(args[1:])
	case "inspect":
		err = runInspect(args[1:])
	case "adapter":
		err = runAdapter(args[1:])
	case "import", "export", "check", "clean":
		err = runOperation(command, args[1:])
	case "conformance":
		err = runConformance(args[1:])
	default:
		printHelp()
		return 2, fmt.Errorf("unknown command %q", command)
	}
	if err != nil {
		return errorExitCode(err), err
	}
	return 0, nil
}

func printHelp() {
	fmt.Println("Dot Agents CLI (dota)")
	fmt.Println()
	fmt.Println("Usage: dota <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init                         Create a v1 .agents manifest")
	fmt.Println("  validate                     Validate the portable v1 tree")
	fmt.Println("  inspect                      Inspect categories and locked adapters")
	fmt.Println("  adapter add|install|list|doctor")
	fmt.Println("  import|export|check|clean    Run one explicitly locked adapter")
	fmt.Println("  conformance tree|adapter     Run normative conformance checks")
	fmt.Println()
	fmt.Println("Adapters are never discovered from PATH. Commit .agents/adapters.lock.json.")
}

func runInit(args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	root := flags.String("root", ".", "repository root")
	force := flags.Bool("force", false, "replace an existing manifest")
	if err := flags.Parse(args); err != nil {
		return usageError(err)
	}
	path := filepath.Join(*root, filepath.FromSlash(dotacore.ManifestPath))
	if _, err := os.Stat(path); err == nil && !*force {
		return fmt.Errorf("%s already exists", dotacore.ManifestPath)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	manifest := dotacore.Manifest{Schema: "https://open-dot-agents.github.io/Open-Dot-Agents/spec/v1/schema/manifest.schema.json", SpecVersion: specdata.SpecVersion, Conformance: []string{"core"}}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Println("dota: initialized", filepath.ToSlash(path))
	return nil
}

func runValidate(args []string) error {
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	root := flags.String("root", ".", "repository root")
	adapterID := flags.String("adapter", "", "locked adapter id")
	jsonOutput := flags.Bool("json", false, "machine-readable output")
	ci := flags.Bool("ci", false, "CI trust policy")
	if err := flags.Parse(args); err != nil {
		return usageError(err)
	}
	result, err := dotacore.Validate(*root)
	if err != nil {
		return err
	}
	var description *adapterprotocol.AdapterDescription
	if *adapterID != "" {
		desc, _, operationResult, err := callAdapter(*root, *adapterID, *ci, "validate")
		if err != nil {
			return err
		}
		if err := diagnosticFailure(operationResult.Diagnostics); err != nil {
			return err
		}
		description = &desc
	}
	if *jsonOutput {
		return writeJSON(outputEnvelope{OutputVersion: 1, Command: "validate", Status: "ok", Adapter: description, Data: result.Manifest})
	}
	fmt.Printf("dota: validate: ok (%s)\n", result.Manifest.SpecVersion)
	return nil
}

func runInspect(args []string) error {
	flags := flag.NewFlagSet("inspect", flag.ContinueOnError)
	root := flags.String("root", ".", "repository root")
	jsonOutput := flags.Bool("json", false, "machine-readable output")
	if err := flags.Parse(args); err != nil {
		return usageError(err)
	}
	result, err := dotacore.Validate(*root)
	if err != nil {
		return err
	}
	categories, err := dotacore.CategoryList(*root)
	if err != nil {
		return err
	}
	lock, lockErr := dotacore.LoadLock(*root)
	if errors.Is(lockErr, os.ErrNotExist) {
		lock = dotacore.AdapterLock{LockVersion: 1}
	} else if lockErr != nil {
		return lockErr
	}
	data := map[string]any{"manifest": result.Manifest, "categories": categories, "adapters": lock.Adapters}
	if *jsonOutput {
		return writeJSON(outputEnvelope{OutputVersion: 1, Command: "inspect", Status: "ok", Data: data})
	}
	fmt.Printf("spec: %s\ncategories: %s\n", result.Manifest.SpecVersion, strings.Join(categories, ", "))
	for _, adapter := range lock.Adapters {
		fmt.Printf("adapter: %s@%s (%s)\n", adapter.ID, adapter.Version, adapter.Source.Type)
	}
	return nil
}

func runAdapter(args []string) error {
	if len(args) == 0 {
		return usageError(errors.New("adapter subcommand is required"))
	}
	subcommand := args[0]
	flags := flag.NewFlagSet("adapter "+subcommand, flag.ContinueOnError)
	root := flags.String("root", ".", "repository root")
	switch subcommand {
	case "add":
		id := flags.String("id", "", "reverse-DNS adapter id")
		adapterVersion := flags.String("version", "dev", "adapter version")
		executable := flags.String("path", "", "local adapter executable")
		publisherManifest := flags.String("manifest", "", "publisher manifest path or https URL")
		capabilities := flags.String("capabilities", "validate,import,export", "comma-separated capabilities")
		if err := flags.Parse(args[1:]); err != nil {
			return usageError(err)
		}
		if *publisherManifest != "" {
			if *id != "" || *executable != "" {
				return usageError(errors.New("--manifest cannot be combined with --id or --path"))
			}
			return dotacore.AddPublishedAdapter(*root, *publisherManifest)
		}
		if *id == "" || *executable == "" {
			return usageError(errors.New("use --manifest, or provide --id and --path"))
		}
		return dotacore.AddLocalAdapter(*root, *id, *adapterVersion, *executable, splitCSV(*capabilities))
	case "install":
		if err := flags.Parse(args[1:]); err != nil {
			return usageError(err)
		}
		return dotacore.InstallAdapters(*root)
	case "list":
		jsonOutput := flags.Bool("json", false, "machine-readable output")
		if err := flags.Parse(args[1:]); err != nil {
			return usageError(err)
		}
		lock, err := dotacore.LoadLock(*root)
		if err != nil {
			return err
		}
		if *jsonOutput {
			return writeJSON(outputEnvelope{OutputVersion: 1, Command: "adapter list", Status: "ok", Data: lock.Adapters})
		}
		for _, adapter := range lock.Adapters {
			fmt.Printf("%s\t%s\t%s\n", adapter.ID, adapter.Version, adapter.Source.Type)
		}
		return nil
	case "doctor":
		id := flags.String("adapter", "", "locked adapter id")
		ci := flags.Bool("ci", false, "CI trust policy")
		if err := flags.Parse(args[1:]); err != nil {
			return usageError(err)
		}
		if *id == "" {
			return usageError(errors.New("--adapter is required"))
		}
		description, _, _, err := callAdapter(*root, *id, *ci, "describe")
		if err != nil {
			return err
		}
		fmt.Printf("dota: adapter doctor: ok (%s %s, protocol %s)\n", description.ID, description.Version, description.ProtocolVersion)
		return nil
	default:
		return usageError(fmt.Errorf("unknown adapter subcommand %q", subcommand))
	}
}

func runOperation(command string, args []string) error {
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	root := flags.String("root", ".", "repository root")
	adapterID := flags.String("adapter", "", "locked adapter id")
	ci := flags.Bool("ci", false, "CI trust policy")
	force := flags.Bool("force", false, "adopt or replace conflicting files")
	dryRun := flags.Bool("dry-run", false, "calculate changes without writing")
	backup := flags.Bool("backup", false, "back up replaced files")
	jsonOutput := flags.Bool("json", false, "machine-readable output")
	mode := flags.String("operation", "export", "ownership operation for check/clean")
	if err := flags.Parse(args); err != nil {
		return usageError(err)
	}
	if *adapterID == "" {
		return usageError(errors.New("--adapter is required"))
	}
	if _, err := dotacore.Validate(*root); err != nil {
		return err
	}
	operation := command
	if command == "check" || command == "clean" {
		operation = *mode
	}
	if operation != "import" && operation != "export" {
		return usageError(errors.New("--operation must be import or export"))
	}
	if command == "clean" {
		return dotacore.Clean(*root, *adapterID, operation)
	}
	method := operation + "Plan"
	description, _, result, err := callAdapter(*root, *adapterID, *ci, method)
	if err != nil {
		return err
	}
	if err := diagnosticFailure(result.Diagnostics); err != nil {
		return err
	}
	if command == "check" {
		return dotacore.CheckPlan(*root, *adapterID, operation, result.Plan)
	}
	changes, err := dotacore.ApplyPlan(*root, *adapterID, operation, result.Plan, dotacore.ApplyOptions{Force: *force, DryRun: *dryRun, Backup: *backup})
	if err != nil {
		return err
	}
	if *jsonOutput {
		return writeJSON(outputEnvelope{OutputVersion: 1, Command: command, Status: "ok", Adapter: &description, Changes: &changes, Data: map[string]any{"losses": result.Losses, "diagnostics": result.Diagnostics}})
	}
	fmt.Printf("dota: %s %s: ok (create=%d update=%d delete=%d)\n", command, description.Target, len(changes.Create), len(changes.Update), len(changes.Delete))
	for _, loss := range result.Losses {
		fmt.Printf("loss: %s: %s\n", loss.Path, loss.Reason)
	}
	return nil
}

func runConformance(args []string) error {
	if len(args) == 0 {
		return usageError(errors.New("conformance subcommand is required"))
	}
	switch args[0] {
	case "tree":
		return runValidate(args[1:])
	case "adapter":
		return runAdapterConformance(args[1:])
	default:
		return usageError(fmt.Errorf("unknown conformance subcommand %q", args[0]))
	}
}

func runAdapterConformance(args []string) error {
	flags := flag.NewFlagSet("conformance adapter", flag.ContinueOnError)
	root := flags.String("root", ".", "repository root")
	adapterID := flags.String("adapter", "", "locked adapter id")
	ci := flags.Bool("ci", false, "CI trust policy")
	if err := flags.Parse(args); err != nil {
		return usageError(err)
	}
	if *adapterID == "" {
		return usageError(errors.New("--adapter is required"))
	}
	if _, err := dotacore.Validate(*root); err != nil {
		return err
	}
	description, _, _, err := callAdapter(*root, *adapterID, *ci, "describe")
	if err != nil {
		return err
	}
	methods := []struct{ capability, method string }{{"validate", "validate"}, {"export", "exportPlan"}, {"import", "importPlan"}}
	for _, operation := range methods {
		if !contains(description.Capabilities, operation.capability) {
			continue
		}
		_, _, first, err := callAdapter(*root, *adapterID, *ci, operation.method)
		if err != nil {
			return err
		}
		_, _, second, err := callAdapter(*root, *adapterID, *ci, operation.method)
		if err != nil {
			return err
		}
		firstJSON, _ := json.Marshal(first)
		secondJSON, _ := json.Marshal(second)
		if !bytes.Equal(firstJSON, secondJSON) {
			return fmt.Errorf("DOTA4008 adapter %s is nondeterministic", operation.method)
		}
		if first.Diagnostics == nil || first.Losses == nil {
			return fmt.Errorf("DOTA4008 adapter %s must return diagnostics and losses arrays", operation.method)
		}
		for _, diagnostic := range first.Diagnostics {
			if !validDiagnosticCode(diagnostic.Code) || !validSeverity(diagnostic.Severity) || strings.TrimSpace(diagnostic.Message) == "" {
				return fmt.Errorf("DOTA4008 adapter %s returned an invalid diagnostic", operation.method)
			}
		}
		for _, loss := range first.Losses {
			if strings.TrimSpace(loss.Path) == "" || strings.TrimSpace(loss.Reason) == "" || !validSeverity(loss.Severity) {
				return fmt.Errorf("DOTA4008 adapter %s returned an invalid loss", operation.method)
			}
		}
		if operation.method != "validate" && first.Plan == nil {
			return fmt.Errorf("DOTA4008 adapter %s returned no plan", operation.method)
		}
	}
	fmt.Printf("dota: conformance adapter: ok (%s %s)\n", description.ID, description.Version)
	return nil
}

func validDiagnosticCode(code string) bool {
	if len(code) != 8 || !strings.HasPrefix(code, "DOTA") {
		return false
	}
	for _, character := range code[4:] {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func validSeverity(severity string) bool {
	return severity == "info" || severity == "warning" || severity == "error"
}

func callAdapter(root, id string, ci bool, method string) (adapterprotocol.AdapterDescription, dotacore.LockedAdapter, adapterprotocol.OperationResult, error) {
	locked, executable, err := dotacore.ResolveAdapter(root, id, ci)
	if err != nil {
		return adapterprotocol.AdapterDescription{}, dotacore.LockedAdapter{}, adapterprotocol.OperationResult{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := adapterprotocol.Start(ctx, executable)
	if err != nil {
		return adapterprotocol.AdapterDescription{}, locked, adapterprotocol.OperationResult{}, err
	}
	defer client.Close()
	if err := client.Initialize(currentVersion()); err != nil {
		return adapterprotocol.AdapterDescription{}, locked, adapterprotocol.OperationResult{}, err
	}
	description, err := client.Describe()
	if err != nil {
		return adapterprotocol.AdapterDescription{}, locked, adapterprotocol.OperationResult{}, err
	}
	if err := validateAdapterDescription(locked, description, method); err != nil {
		return description, locked, adapterprotocol.OperationResult{}, err
	}
	if method == "describe" {
		return description, locked, adapterprotocol.OperationResult{}, nil
	}
	snapshot, err := dotacore.Snapshot(root, description.InputPatterns, description.MaxSnapshotBytes)
	if err != nil {
		return description, locked, adapterprotocol.OperationResult{}, err
	}
	result, err := client.Operation(method, snapshot)
	return description, locked, result, err
}

func validateAdapterDescription(locked dotacore.LockedAdapter, description adapterprotocol.AdapterDescription, method string) error {
	if description.ID != locked.ID || description.Version != locked.Version || description.ProtocolVersion != locked.ProtocolVersion {
		return fmt.Errorf("DOTA4006 adapter description does not match lock entry")
	}
	lockedCapabilities := append([]string(nil), locked.Capabilities...)
	describedCapabilities := append([]string(nil), description.Capabilities...)
	sort.Strings(lockedCapabilities)
	sort.Strings(describedCapabilities)
	if strings.Join(lockedCapabilities, "\x00") != strings.Join(describedCapabilities, "\x00") {
		return fmt.Errorf("DOTA4006 adapter capabilities do not match lock entry")
	}
	required := map[string]string{"validate": "validate", "importPlan": "import", "exportPlan": "export"}[method]
	if required != "" && !contains(description.Capabilities, required) {
		return fmt.Errorf("DOTA4006 adapter does not declare the %s capability", required)
	}
	if description.Name == "" || description.Target == "" || len(description.CategoryStatuses) == 0 || len(description.InputPatterns) == 0 {
		return fmt.Errorf("DOTA4006 adapter description is incomplete")
	}
	categories := map[string]bool{
		"agents": true, "guardrails": true, "hooks": true, "instructions": true, "memories": true,
		"permissions": true, "plugins": true, "profiles": true, "prompts": true, "rules": true,
		"settings": true, "skills": true, "tools": true,
	}
	for category, status := range description.CategoryStatuses {
		if !categories[category] {
			return fmt.Errorf("DOTA4006 adapter description contains an unknown category")
		}
		if status != "supported" && status != "mapped" && status != "partial" && status != "unsupported" {
			return fmt.Errorf("DOTA4006 adapter description contains an invalid category status")
		}
	}
	for _, pattern := range description.InputPatterns {
		clean := path.Clean(pattern)
		if pattern == "" || strings.HasPrefix(pattern, "/") || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(pattern, "\\") {
			return fmt.Errorf("DOTA4006 adapter description contains an unsafe input pattern")
		}
	}
	if description.MaxSnapshotBytes < 0 {
		return fmt.Errorf("DOTA4006 adapter description contains an invalid snapshot limit")
	}
	return nil
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func diagnosticFailure(diagnostics []adapterprotocol.Diagnostic) error {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == "error" {
			return fmt.Errorf("%s %s: %s", diagnostic.Code, diagnostic.Path, diagnostic.Message)
		}
	}
	return nil
}

func currentVersion() string {
	if version != "" && version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func writeJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func splitCSV(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	sort.Strings(result)
	return result
}

func usageError(err error) error { return fmt.Errorf("DOTA0002 %w", err) }

func errorExitCode(err error) int {
	message := err.Error()
	switch {
	case strings.Contains(message, "DOTA0002"):
		return 2
	case strings.Contains(message, "DOTA3006"), strings.Contains(message, "DOTA3007"):
		return 3
	case strings.Contains(message, "DOTA4"):
		return 4
	default:
		return 1
	}
}
