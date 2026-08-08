package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/adapter"
	"github.com/Open-Dot-Agents/Open-Dot-Agents/cli/internal/schema"
)

var version = "dev"

func currentVersion() string {
	buildVersion := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		buildVersion = info.Main.Version
	}
	return resolveVersion(version, buildVersion)
}

func resolveVersion(linkerVersion, buildVersion string) string {
	if linkerVersion != "" && linkerVersion != "dev" {
		return linkerVersion
	}
	if buildVersion != "" && buildVersion != "(devel)" {
		return buildVersion
	}
	return "dev"
}

type commandResult struct {
	Command string                  `json:"command"`
	Target  string                  `json:"target"`
	Status  string                  `json:"status"`
	Error   string                  `json:"error,omitempty"`
	Backup  string                  `json:"backup,omitempty"`
	Guide   *schema.VendorGuide     `json:"guide,omitempty"`
	Plan    *adapter.GenerationPlan `json:"plan,omitempty"`
	Diff    []string                `json:"diff,omitempty"`
}

type commandOutput struct {
	Command string          `json:"command"`
	Targets []commandResult `json:"targets"`
}

type runContext struct {
	root      string
	force     bool
	dryRun    bool
	diff      bool
	jsonOut   bool
	ciMode    bool
	backup    bool
	backupDir string
}

type parsedCLI struct {
	root             string
	target           string
	force            bool
	backup           bool
	backupDir        string
	allowUnsupported bool
	dryRun           bool
	diff             bool
	format           string
	ci               bool
	help             bool
	shortHelp        bool
	version          bool
	command          string
	commandArgs      []string
}

func main() {
	commands := strings.Join([]string{"validate", "generate", "export", "import", "check", "clean", "guide", "completion"}, ", ")
	targetList := strings.Join(append(adapter.RegisteredTargets(), "all"), ", ")
	flag.Usage = func() {
		fmt.Fprintln(os.Stdout, "Open-Dot-Agents CLI (oda)")
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Usage:")
		fmt.Fprintln(os.Stdout, "  oda <command> [options]")
		fmt.Fprintln(os.Stdout, "  oda [options] <command>")
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Purpose:")
		fmt.Fprintln(os.Stdout, "  Keep `.agents/` canonical and import/export repository-scoped harness configuration.")
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Common workflow:")
		fmt.Fprintln(os.Stdout, "  1. validate   - verify current .agents tree and mappings/manifest shape")
		fmt.Fprintln(os.Stdout, "  2. export     - materialize runtime-specific files (or compare with --dry-run)")
		fmt.Fprintln(os.Stdout, "  3. check      - detect drift between tracked source and generated output")
		fmt.Fprintln(os.Stdout, "  4. clean      - remove generated files for the selected target(s)")
		fmt.Fprintln(os.Stdout, "  5. guide      - generate a vendor implementation guide from schema + mappings")
		fmt.Fprintln(os.Stdout, "  6. completion - completion script and CLI shell helpers")
		fmt.Fprintln(os.Stdout)
		fmt.Fprintf(os.Stdout, "Targets: %s (or all)\n\n", targetList)
		fmt.Fprintln(os.Stdout, "Commands:")
		fmt.Fprintln(os.Stdout, "  validate    Validate `.agents/schema/v0.0.1/*.json`, mappings, manifest compatibility and source shape")
		fmt.Fprintln(os.Stdout, "  export      Export `.agents` into the selected target(s)")
		fmt.Fprintln(os.Stdout, "  generate    Backward-compatible alias for export")
		fmt.Fprintln(os.Stdout, "  import      Reverse-project from generated target files into `.agents`")
		fmt.Fprintln(os.Stdout, "  check       Verify generated output matches the target projection plan")
		fmt.Fprintln(os.Stdout, "  clean       Remove generated files tracked by the target manifest")
		fmt.Fprintln(os.Stdout, "  guide       Print a machine-readable vendor implementation guide")
		fmt.Fprintln(os.Stdout, "  completion  Print shell completion script (bash/zsh)")
		fmt.Fprintln(os.Stdout, "Validation details:")
		fmt.Fprintln(os.Stdout, "  - `validate` checks schema files, mapping status matrix, front-matter requirements, and unsupported categories.")
		fmt.Fprintln(os.Stdout, "  - If `.agents/schema/v0.0.1/*` or `.agents/mappings.yaml` are missing, validation reports a clear `schema not available` error.")
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Options:")
		fmt.Fprintln(os.Stdout, "  --root PATH             Repository root to operate on (default: \".\")")
		fmt.Fprintln(os.Stdout, "  --target TARGET         Adapter target (copilot/codex/claude/all)")
		fmt.Fprintln(os.Stdout, "  --force                 Replace conflicting import/export files after review")
		fmt.Fprintln(os.Stdout, "  --backup                Back up the destination before a forced import/export")
		fmt.Fprintln(os.Stdout, "  --backup-dir PATH       Backup root (default: <root>/.oda-backups)")
		fmt.Fprintln(os.Stdout, "  --allow-unsupported     Permit adapters that would otherwise fail on unsupported categories")
		fmt.Fprintln(os.Stdout, "  --dry-run               Show what write commands would do without touching files")
		fmt.Fprintln(os.Stdout, "  --diff                  Include file diff summary lines for dry-run import/export")
		fmt.Fprintln(os.Stdout, "  --ci                    Enable CI mode. import/export/check report drift with non-zero exit status")
		fmt.Fprintln(os.Stdout, "  --format text|json      Output format (default: text)")
		fmt.Fprintln(os.Stdout, "  --help, -h              Show this help")
		fmt.Fprintln(os.Stdout, "  --version               Print oda version and exit")
		fmt.Fprintln(os.Stdout, "Examples:")
		fmt.Fprintln(os.Stdout, "  oda validate")
		fmt.Fprintln(os.Stdout, "  oda --target all validate")
		fmt.Fprintln(os.Stdout, "  oda validate --target all")
		fmt.Fprintln(os.Stdout, "  oda --target copilot --force --backup generate")
		fmt.Fprintln(os.Stdout, "  oda --target codex --force --backup export")
		fmt.Fprintln(os.Stdout, "  oda --target all --force --backup --backup-dir /tmp/oda-backups generate")
		fmt.Fprintln(os.Stdout, "  oda --target copilot --force --dry-run import")
		fmt.Fprintln(os.Stdout, "  oda --target copilot --force import")
		fmt.Fprintln(os.Stdout, "  oda guide")
		fmt.Fprintln(os.Stdout, "  oda completion")
		fmt.Fprintln(os.Stdout, "  eval \"$(oda completion bash)\"")
		fmt.Fprintln(os.Stdout, "  source <(oda completion zsh)")
		fmt.Fprintln(os.Stdout, "  oda --target codex --ci check")
		fmt.Fprintln(os.Stdout, "  oda --target copilot --dry-run --diff export")
		fmt.Fprintf(os.Stdout, "Known commands: %s\n", strings.TrimSpace(commands))
	}

	parsed, err := parseCLIArgs(os.Args[1:], flag.Usage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "oda:", err)
		flag.Usage()
		os.Exit(2)
	}
	if parsed.help || parsed.shortHelp {
		flag.Usage()
		return
	}
	if parsed.version {
		fmt.Fprintf(os.Stdout, "oda %s\n", currentVersion())
		return
	}
	if parsed.command == "" {
		flag.Usage()
		os.Exit(2)
	}
	if parsed.format != "text" && parsed.format != "json" {
		fmt.Fprintln(os.Stderr, "oda: --format must be one of: text, json")
		os.Exit(2)
	}
	command := parsed.command
	if err := validateCommandTarget(command, parsed.target); err != nil {
		fmt.Fprintln(os.Stderr, "oda:", err)
		os.Exit(2)
	}
	if command == "completion" {
		if len(parsed.commandArgs) > 1 {
			fmt.Fprintln(os.Stderr, "oda: completion accepts at most one optional shell argument")
			os.Exit(2)
		}
		shell := "bash"
		if len(parsed.commandArgs) == 1 {
			shell = parsed.commandArgs[0]
		}
		if err := printCompletionScript(shell); err != nil {
			fmt.Fprintln(os.Stderr, "oda:", err)
			os.Exit(1)
		}
		return
	}
	if len(parsed.commandArgs) != 0 {
		flag.Usage()
		os.Exit(2)
	}

	targetsToRun, err := resolveTargets(parsed.target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "oda:", err)
		os.Exit(2)
	}
	ctx := runContext{
		root:      parsed.root,
		force:     parsed.force,
		dryRun:    parsed.dryRun || parsed.diff,
		diff:      parsed.diff,
		jsonOut:   parsed.format == "json",
		ciMode:    parsed.ci,
		backup:    parsed.backup,
		backupDir: parsed.backupDir,
	}

	results := make([]commandResult, 0, len(targetsToRun))
	failed := false
	for _, selected := range targetsToRun {
		a, err := adapter.NewForTarget(parsed.root, selected, parsed.allowUnsupported)
		if err != nil {
			if ctx.jsonOut {
				results = append(results, commandResult{
					Command: command,
					Target:  selected,
					Status:  "error",
					Error:   err.Error(),
				})
			} else {
				fmt.Fprintln(os.Stderr, "oda:", err)
			}
			failed = true
			continue
		}
		if command == "guide" {
			guide, err := schema.GenerateVendorGuide(parsed.root, []string{selected})
			if err != nil {
				result := commandResult{
					Command: command,
					Target:  selected,
					Status:  "error",
					Error:   err.Error(),
				}
				results = append(results, result)
				failed = true
				continue
			}
			if ctx.jsonOut {
				result := commandResult{
					Command: command,
					Target:  selected,
					Status:  "ok",
					Guide:   guide,
				}
				results = append(results, result)
			} else {
				fmt.Println(schema.RenderVendorGuideMarkdown(guide))
				results = append(results, commandResult{
					Command: command,
					Target:  selected,
					Status:  "ok",
				})
			}
			continue
		}

		result := run(ctx, command, a)
		results = append(results, result)
		if result.Status != "ok" {
			failed = true
		}
	}
	if ctx.jsonOut {
		if err := writeJSONOutput(os.Stdout, command, results); err != nil {
			fmt.Fprintln(os.Stderr, "oda:", err)
			os.Exit(1)
		}
		if failed {
			os.Exit(1)
		}
		return
	}
	if command == "guide" {
		if failed {
			os.Exit(1)
		}
		return
	} else {
		for _, result := range results {
			if result.Status == "ok" {
				fmt.Printf("oda: %s %s: ok\n", result.Command, result.Target)
			} else {
				fmt.Fprintf(os.Stderr, "oda: %s %s: %s\n", result.Command, result.Target, result.Error)
			}
			if result.Backup != "" {
				fmt.Printf("oda: backup for %s: %s\n", result.Target, result.Backup)
			}
			if len(result.Diff) > 0 {
				for _, line := range result.Diff {
					fmt.Fprintln(os.Stdout, line)
				}
			}
		}
	}
	if failed {
		os.Exit(1)
	}
}

func parseCLIArgs(args []string, usage func()) (parsedCLI, error) {
	parsed := parsedCLI{
		root:   ".",
		target: "copilot",
		format: "text",
	}

	fs := flag.NewFlagSet("oda", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() { usage() }

	fs.StringVar(&parsed.root, "root", parsed.root, "repository root")
	fs.StringVar(&parsed.target, "target", parsed.target, "target adapter or 'all'")
	fs.BoolVar(&parsed.force, "force", false, "replace a generated compatibility tree")
	fs.BoolVar(&parsed.backup, "backup", false, "backup the destination before a forced generate/import")
	fs.StringVar(&parsed.backupDir, "backup-dir", "", "directory where backups are stored (default: <root>/.oda-backups)")
	fs.BoolVar(&parsed.allowUnsupported, "allow-unsupported", false, "allow populated unsupported categories")
	fs.BoolVar(&parsed.dryRun, "dry-run", false, "compute changes without applying them")
	fs.BoolVar(&parsed.diff, "diff", false, "show file-level import/export diff summary")
	fs.StringVar(&parsed.format, "format", parsed.format, "output format: text or json")
	fs.BoolVar(&parsed.ci, "ci", false, "non-zero if import, export, or check would change output")
	fs.BoolVar(&parsed.help, "help", false, "show help")
	fs.BoolVar(&parsed.shortHelp, "h", false, "show help")
	fs.BoolVar(&parsed.version, "version", false, "show version")

	commandIndex := findCommandArg(args)
	parseArgs := args
	if commandIndex >= 0 {
		parseArgs = append([]string{}, args...)
		parseArgs = append(parseArgs[:commandIndex], parseArgs[commandIndex+1:]...)
	}
	if err := fs.Parse(parseArgs); err != nil {
		return parsed, err
	}

	if commandIndex >= 0 {
		parsed.command = args[commandIndex]
		parsed.commandArgs = fs.Args()
		return parsed, nil
	}
	parsed.command = ""
	parsed.commandArgs = fs.Args()
	return parsed, nil
}

func findCommandArg(args []string) int {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if isKnownCommand(arg) {
			return i
		}
		if strings.HasPrefix(arg, "--") {
			if consumesValue(arg) {
				i++
			}
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
	}
	return -1
}

func isKnownCommand(arg string) bool {
	return arg == "validate" || arg == "generate" || arg == "export" || arg == "import" || arg == "check" || arg == "clean" || arg == "guide" || arg == "completion"
}

func consumesValue(flagName string) bool {
	switch flagName {
	case "--root", "--target", "--backup-dir", "--format":
		return true
	default:
		return false
	}
}

func resolveTargets(target string) ([]string, error) {
	if target == "all" {
		return adapter.RegisteredTargets(), nil
	}
	return []string{target}, nil
}

func validateCommandTarget(command, target string) error {
	if command == "import" && target == "all" {
		return fmt.Errorf("import requires one explicit source target; --target all is not supported")
	}
	return nil
}

func writeJSONOutput(w io.Writer, command string, results []commandResult) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(commandOutput{Command: command, Targets: results})
}

func run(ctx runContext, command string, a *adapter.Adapter) commandResult {
	switch command {
	case "validate":
		if err := a.Validate(); err != nil {
			return commandResult{Command: command, Target: a.Target(), Status: "error", Error: err.Error()}
		}
		return commandResult{Command: command, Target: a.Target(), Status: "ok"}
	case "generate", "export":
		if !ctx.dryRun && ctx.force && ctx.backup {
			manifestDir, err := backupDirectory(command, a.Target())
			if err != nil {
				return commandResult{Command: command, Target: a.Target(), Status: "error", Error: err.Error()}
			}
			backupPath, err := backupGeneratedOutput(ctx.root, manifestDir, ctx.backupDir)
			if err != nil {
				return commandResult{Command: command, Target: a.Target(), Status: "error", Error: err.Error()}
			}
			// Plan after backup so backup snapshots pre-generation state.
			if err := a.Generate(ctx.force); err != nil {
				return commandResult{Command: command, Target: a.Target(), Status: "error", Backup: backupPath, Error: err.Error()}
			}
			return commandResult{Command: command, Target: a.Target(), Status: "ok", Backup: backupPath}
		}
		if ctx.dryRun {
			plan, err := a.Plan(ctx.force)
			if err != nil {
				return commandResult{Command: command, Target: a.Target(), Status: "error", Error: err.Error()}
			}
			status := "ok"
			var errMsg string
			if ctx.ciMode && hasChanges(plan) {
				status = "drift"
				errMsg = "generated output is out of sync; rerun export"
			}
			result := commandResult{Command: command, Target: a.Target(), Status: status, Error: errMsg, Plan: plan}
			if ctx.diff {
				result.Diff = planDiff(plan)
			}
			return result
		}
		if err := a.Generate(ctx.force); err != nil {
			return commandResult{Command: command, Target: a.Target(), Status: "error", Error: err.Error()}
		}
		return commandResult{Command: command, Target: a.Target(), Status: "ok"}
	case "import":
		if !ctx.dryRun && ctx.force && ctx.backup {
			manifestDir, err := backupDirectory(command, a.Target())
			if err != nil {
				return commandResult{Command: command, Target: a.Target(), Status: "error", Error: err.Error()}
			}
			backupPath, err := backupGeneratedOutput(ctx.root, manifestDir, ctx.backupDir)
			if err != nil {
				return commandResult{Command: command, Target: a.Target(), Status: "error", Error: err.Error()}
			}
			if err := a.Import(ctx.force); err != nil {
				return commandResult{Command: command, Target: a.Target(), Status: "error", Backup: backupPath, Error: err.Error()}
			}
			return commandResult{Command: command, Target: a.Target(), Status: "ok", Backup: backupPath}
		}
		if ctx.dryRun {
			plan, err := a.ImportPlan(ctx.force)
			if err != nil {
				return commandResult{Command: command, Target: a.Target(), Status: "error", Error: err.Error()}
			}
			status := "ok"
			var errMsg string
			if ctx.ciMode && hasChanges(plan) {
				status = "drift"
				errMsg = "generated .agents output is out of sync; rerun import"
			}
			result := commandResult{Command: command, Target: a.Target(), Status: status, Error: errMsg, Plan: plan}
			if ctx.diff {
				result.Diff = planDiff(plan)
			}
			return result
		}
		if err := a.Import(ctx.force); err != nil {
			return commandResult{Command: command, Target: a.Target(), Status: "error", Error: err.Error()}
		}
		return commandResult{Command: command, Target: a.Target(), Status: "ok"}
	case "check":
		if ctx.ciMode {
			if err := a.Check(); err != nil {
				return commandResult{
					Command: command,
					Target:  a.Target(),
					Status:  "drift",
					Error:   err.Error(),
				}
			}
			return commandResult{Command: command, Target: a.Target(), Status: "ok"}
		}
		if err := a.Check(); err != nil {
			return commandResult{Command: command, Target: a.Target(), Status: "error", Error: err.Error()}
		}
		return commandResult{Command: command, Target: a.Target(), Status: "ok"}
	case "clean":
		if err := a.Clean(); err != nil {
			return commandResult{Command: command, Target: a.Target(), Status: "error", Error: err.Error()}
		}
		return commandResult{Command: command, Target: a.Target(), Status: "ok"}
	default:
		return commandResult{Command: command, Target: a.Target(), Status: "error", Error: "unsupported command"}
	}
}

func planDiff(plan *adapter.GenerationPlan) []string {
	if plan == nil {
		return nil
	}
	var lines []string
	for _, path := range plan.Create {
		lines = append(lines, "A "+path)
	}
	for _, path := range plan.Update {
		lines = append(lines, "M "+path)
	}
	for _, path := range plan.Delete {
		lines = append(lines, "D "+path)
	}
	return lines
}

func hasChanges(plan *adapter.GenerationPlan) bool {
	if plan == nil {
		return false
	}
	return len(plan.Create) > 0 || len(plan.Update) > 0 || len(plan.Delete) > 0
}

func targetManifestDirectory(target string) (string, error) {
	for _, info := range adapter.TargetInfos() {
		if info.ID == target {
			return info.ManifestDirectory, nil
		}
	}
	return "", fmt.Errorf("unknown target %q", target)
}

func backupDirectory(command, target string) (string, error) {
	if command == "import" {
		return ".agents", nil
	}
	return targetManifestDirectory(target)
}

func printCompletionScript(shell string) error {
	shell = strings.ToLower(shell)
	if shell == "" {
		shell = filepath.Base(os.Getenv("SHELL"))
	}
	switch shell {
	case "", "bash":
		fmt.Print(bashCompletionScript())
		return nil
	case "zsh":
		fmt.Print(zshCompletionScript())
		return nil
	default:
		return fmt.Errorf("unsupported shell %q", shell)
	}
}

func bashCompletionScript() string {
	commands := strings.Join([]string{"validate", "generate", "export", "import", "check", "clean", "guide", "completion"}, " ")
	targets := strings.Join(append(adapter.RegisteredTargets(), "all"), " ")
	flags := strings.Join([]string{
		"--help",
		"-h",
		"--version",
		"--root",
		"--target",
		"--force",
		"--backup",
		"--backup-dir",
		"--allow-unsupported",
		"--dry-run",
		"--diff",
		"--format",
		"--ci",
	}, " ")
	return fmt.Sprintf(`_oda_completion() {
  local cur
  local prev
  local command
  local command_pos
  local i
  local token
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  command=""
  command_pos=-1
  i=1
  while [[ $i -le $COMP_CWORD ]]; do
    token="${COMP_WORDS[i]}"
    case "${token}" in
      --root|--target|--backup-dir|--format)
        ((i += 2))
        continue
        ;;
      --help|-h|--version|--force|--backup|--allow-unsupported|--dry-run|--diff|--ci)
        ((i += 1))
        continue
        ;;
    esac
    if [[ "${token}" != --* && "${token}" != "" ]]; then
      command="${token}"
      command_pos=$i
      break
    fi
    ((i += 1))
  done

  if [[ $command_pos -lt 0 || $COMP_CWORD -eq $command_pos ]]; then
    COMPREPLY=( $(compgen -W "%s %s" -- "${cur}") )
    return 0
  fi
  if [[ "${prev}" == "--target" ]]; then
    COMPREPLY=( $(compgen -W "%s" -- "${cur}") )
    return 0
  fi
  if [[ "${prev}" == "--format" ]]; then
    COMPREPLY=( $(compgen -W "text json" -- "${cur}") )
    return 0
  fi
  if [[ "${prev}" == "--root" || "${prev}" == "--backup-dir" ]]; then
    COMPREPLY=( $(compgen -f -- "${cur}") )
    return 0
  fi

  case "${command}" in
    validate|generate|export|import|check|clean|guide|completion)
      if [[ "${command}" == "completion" ]]; then
        COMPREPLY=( $(compgen -W "bash zsh" -- "${cur}") )
      else
        COMPREPLY=( $(compgen -W "%s" -- "${cur}") )
      fi
      return 0
      ;;
  esac
		}

complete -F _oda_completion -o default oda
`, commands, flags, targets, flags)
}

func zshCompletionScript() string {
	commands := strings.Join([]string{"validate", "generate", "export", "import", "check", "clean", "guide", "completion"}, " ")
	targets := strings.Join(append(adapter.RegisteredTargets(), "all"), " ")
	return fmt.Sprintf(`#compdef oda
_oda() {
  local -a oda_commands
  local -a oda_options
  local i
  local command
  local current
  local previous
  local token
  oda_commands=(%s)
  oda_options=(
    '--help:show help'
    '-h:show help'
    '--version:show version'
    '--root:repository root:_files'
    '--target:target:(%s)'
    '--force:replace existing files'
    '--backup:backup existing target directory'
    '--backup-dir:backup root:_files'
    '--allow-unsupported:allow unsupported categories'
    '--dry-run:plan changes'
    '--diff:show diff summary'
    '--format:output format:(text json)'
    '--ci:ci mode'
  )
  command=""
  i=2
  while (( i < CURRENT )); do
    token=$words[i]
    case "$token" in
      --root|--target|--backup-dir|--format)
        ((i += 2))
        continue
        ;;
      --help|-h|--version|--force|--backup|--allow-unsupported|--dry-run|--diff|--ci)
        ((i += 1))
        continue
        ;;
      -*)
        ((i += 1))
        continue
        ;;
      *)
        command=$token
        break
        ;;
    esac
  done

  current=$words[$CURRENT]
  previous=$words[$CURRENT-1]

  if [[ -z "$command" || $i -eq $CURRENT ]]; then
    compadd $oda_commands[@]
    if [[ $i -eq $CURRENT ]]; then
      return 0
    fi
    for option in "${oda_options[@]}"; do
      if [[ "$option" == *=* ]]; then
        continue
      fi
      compadd "${option%%:*}"
    done
    return 0
  fi

  if [[ "$previous" == "--target" ]]; then
    compadd %s
    return 0
  fi
  if [[ "$previous" == "--root" || "$previous" == "--backup-dir" ]]; then
    _files
    return 0
  fi
  if [[ "$previous" == "--format" ]]; then
    compadd text json
    return 0
  fi

  if [[ "$command" == "completion" ]]; then
    compadd bash zsh
    return 0
  fi
  for option in "${oda_options[@]}"; do
    compadd "${option%%:*}"
  done
}

compdef _oda oda
`, shellQuotedList(commands), targets, targets)
}

func shellQuotedList(values string) string {
	items := strings.Fields(values)
	var quoted []string
	for _, item := range items {
		quoted = append(quoted, "'"+item+"'")
	}
	return strings.Join(quoted, " ")
}

func backupGeneratedOutput(root, manifestDir, backupRoot string) (string, error) {
	source := filepath.Join(root, manifestDir)
	if _, err := os.Stat(source); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	backupBase := strings.TrimSpace(backupRoot)
	if backupBase == "" {
		backupBase = filepath.Join(root, ".oda-backups")
	}
	backupBase = filepath.Clean(backupBase)
	if backupBase == "." || backupBase == string(filepath.Separator) {
		return "", fmt.Errorf("invalid backup directory %q", backupBase)
	}
	if err := os.MkdirAll(backupBase, 0o755); err != nil {
		return "", err
	}

	destination := filepath.Join(backupBase, manifestDir, time.Now().Format("20060102-150405.000000000"))
	if err := copyDir(source, destination); err != nil {
		return "", err
	}
	return destination, nil
}

func copyDir(source, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(destination, info.Mode().Perm()); err != nil {
		return err
	}

	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if rel == "." {
			return nil
		}
		entryInfo, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(target, entryInfo.Mode().Perm())
		}
		if entryInfo.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, entryInfo.Mode().Perm())
	})
}
