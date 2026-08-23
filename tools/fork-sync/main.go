package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const version = "1.0.0"

func printUsage() {
	fmt.Printf(`fork-sync v%s - Upstream Synchronization and Invariant Verification Tool

Usage:
  fork-sync <command> [options]

Commands:
  check, status    Inspect upstream remote, incoming commits, and protected invariant files
  start            Prepare a new sync branch (sync/upstream-YYYY-MM-DD) and initiate merge
  verify           Assert all fork invariants across the repository
  help             Show this help message

Options:
  --repo-root <path>        Path to repository root (defaults to auto-detected git root)
  --upstream-remote <name>  Name of upstream remote (default: "upstream")
  --upstream-url <url>      URL for upstream remote (default: "https://github.com/PrimeIntellect-ai/prime-agent.git")
  --upstream-branch <name>  Upstream branch name (default: "main")
  --base <branch>           Local base branch for diff/merge (default: "main" or "HEAD")
  --branch <name>           Custom sync branch name for 'start' command
  --fetch                   Fetch upstream references (default: true for check/status and start)
  --no-fetch                Skip fetching upstream references
  --allow-dirty             Allow starting sync even if working tree has uncommitted changes
  --run-checks              Run 'npm run check' as part of verify command
  --json                    Output results in JSON format
  -v, --verbose             Enable verbose output
  -h, --help                Show help

Examples:
  # Check incoming changes and invariant overlap
  fork-sync status

  # Start a synchronization branch and initiate upstream merge
  fork-sync start

  # Verify all fork invariants and run static checks
  fork-sync verify --run-checks
`, version)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	if cmd == "-h" || cmd == "--help" || cmd == "help" {
		printUsage()
		os.Exit(0)
	}
	if cmd == "-v" || cmd == "--version" || cmd == "version" {
		fmt.Printf("fork-sync version %s\n", version)
		os.Exit(0)
	}

	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	var (
		repoRoot       string
		upstreamRemote string
		upstreamURL    string
		upstreamBranch string
		baseBranch     string
		syncBranch     string
		noFetch        bool
		fetch          bool
		allowDirty     bool
		runChecks      bool
		jsonOutput     bool
		verbose        bool
	)

	fs.StringVar(&repoRoot, "repo-root", "", "Path to repository root")
	fs.StringVar(&upstreamRemote, "upstream-remote", "upstream", "Name of upstream remote")
	fs.StringVar(&upstreamURL, "upstream-url", "https://github.com/PrimeIntellect-ai/prime-agent.git", "URL for upstream remote")
	fs.StringVar(&upstreamBranch, "upstream-branch", "main", "Upstream branch name")
	fs.StringVar(&baseBranch, "base", "", "Local base branch")
	fs.StringVar(&syncBranch, "branch", "", "Custom sync branch name")
	fs.BoolVar(&noFetch, "no-fetch", false, "Skip fetching upstream references")
	fs.BoolVar(&fetch, "fetch", true, "Fetch upstream references")
	fs.BoolVar(&allowDirty, "allow-dirty", false, "Allow dirty working tree")
	fs.BoolVar(&runChecks, "run-checks", false, "Run npm run check during verify")
	fs.BoolVar(&jsonOutput, "json", false, "Output JSON")
	fs.BoolVar(&verbose, "verbose", false, "Verbose output")
	fs.BoolVar(&verbose, "v", false, "Verbose output")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	if repoRoot == "" {
		detectedRoot, err := FindRepoRoot(".")
		if err != nil {
			// Fallback to working directory
			cwd, _ := os.Getwd()
			repoRoot = cwd
		} else {
			repoRoot = detectedRoot
		}
	} else {
		abs, err := filepath.Abs(repoRoot)
		if err == nil {
			repoRoot = abs
		}
	}

	shouldFetch := fetch
	if noFetch {
		shouldFetch = false
	}

	cfg := Config{
		RepoRoot:        repoRoot,
		UpstreamRemote:  upstreamRemote,
		UpstreamURL:     upstreamURL,
		UpstreamBranch:  upstreamBranch,
		LocalBaseBranch: baseBranch,
		SyncBranch:      syncBranch,
		Fetch:           shouldFetch,
		AllowDirty:      allowDirty,
		RunChecks:       runChecks,
		JSON:            jsonOutput,
		Verbose:         verbose,
	}

	var err error
	switch cmd {
	case "check", "status":
		err = RunStatus(cfg)
	case "start":
		err = RunStart(cfg)
	case "verify":
		err = RunVerify(cfg)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		}
		os.Exit(1)
	}
}
