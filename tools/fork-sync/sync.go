package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Config configures the fork-sync execution.
type Config struct {
	RepoRoot        string `json:"repo_root"`
	UpstreamRemote  string `json:"upstream_remote"`
	UpstreamURL     string `json:"upstream_url"`
	UpstreamBranch  string `json:"upstream_branch"`
	LocalBaseBranch string `json:"local_base_branch"`
	SyncBranch      string `json:"sync_branch"`
	Fetch           bool   `json:"fetch"`
	AllowDirty      bool   `json:"allow_dirty"`
	RunChecks       bool   `json:"run_checks"`
	Agent           string `json:"agent"`
	SummaryOutput   string `json:"summary_output"`
	JSON            bool   `json:"json"`
	Verbose         bool   `json:"verbose"`
}

// StatusReport holds data for check/status commands.
type StatusReport struct {
	RepoRoot         string           `json:"repo_root"`
	CurrentBranch    string           `json:"current_branch"`
	UpstreamRemote   string           `json:"upstream_remote"`
	UpstreamURL      string           `json:"upstream_url"`
	UpstreamBranch   string           `json:"upstream_branch"`
	TargetRef        string           `json:"target_ref"`
	WorkingTreeClean bool             `json:"working_tree_clean"`
	IncomingCommits  []CommitInfo     `json:"incoming_commits"`
	ModifiedFiles    []InvariantMatch `json:"modified_files"`
	InvariantCount   int              `json:"invariant_count"`
	TotalFilesCount  int              `json:"total_files_count"`
	HasSyncPending   bool             `json:"has_sync_pending"`
	SummaryFile      string           `json:"summary_file,omitempty"`
}

// RunStatus executes the status / check inspection workflow.
func RunStatus(cfg Config) error {
	report := StatusReport{
		RepoRoot:       cfg.RepoRoot,
		UpstreamRemote: cfg.UpstreamRemote,
		UpstreamURL:    cfg.UpstreamURL,
		UpstreamBranch: cfg.UpstreamBranch,
	}

	// 1. Check current branch
	currBranch, err := GetCurrentBranch(cfg.RepoRoot)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	report.CurrentBranch = currBranch

	// 2. Check clean working tree
	clean, _, err := IsWorkingTreeClean(cfg.RepoRoot)
	if err != nil {
		return fmt.Errorf("failed to check working tree status: %w", err)
	}
	report.WorkingTreeClean = clean

	// 3. Remote configuration
	url, added, err := EnsureUpstreamRemote(cfg.RepoRoot, cfg.UpstreamRemote, cfg.UpstreamURL)
	if err != nil {
		return fmt.Errorf("failed to configure upstream remote: %w", err)
	}
	report.UpstreamURL = url
	if added && !cfg.JSON {
		fmt.Printf("[+] Added upstream remote '%s' (%s)\n", cfg.UpstreamRemote, url)
	}

	// 4. Fetch upstream if requested
	if cfg.Fetch {
		if !cfg.JSON {
			fmt.Printf("[*] Fetching refs from '%s'...\n", cfg.UpstreamRemote)
		}
		if err := FetchRemote(cfg.RepoRoot, cfg.UpstreamRemote); err != nil {
			return fmt.Errorf("fetch error: %w", err)
		}
	}

	// 5. Inspect incoming commits and diff
	targetRef := fmt.Sprintf("%s/%s", cfg.UpstreamRemote, cfg.UpstreamBranch)
	report.TargetRef = targetRef
	baseRef := cfg.LocalBaseBranch
	if baseRef == "" {
		baseRef = "HEAD"
	}

	if !RefExists(cfg.RepoRoot, targetRef) {
		if !cfg.Fetch {
			if !cfg.JSON {
				fmt.Printf("[*] Upstream ref '%s' not found locally. Fetching from '%s'...\n", targetRef, cfg.UpstreamRemote)
			}
			if err := FetchRemote(cfg.RepoRoot, cfg.UpstreamRemote); err != nil {
				return fmt.Errorf("upstream ref '%s' not found and fetch failed: %w", targetRef, err)
			}
		} else {
			return fmt.Errorf("upstream ref '%s' not found after fetch; check upstream branch name (%s)", targetRef, cfg.UpstreamBranch)
		}
	}

	commits, err := GetIncomingCommits(cfg.RepoRoot, baseRef, targetRef)
	if err != nil {
		return fmt.Errorf("failed to query incoming commits between %s and %s: %w", baseRef, targetRef, err)
	}
	report.IncomingCommits = commits

	diffFiles, err := GetDiffFiles(cfg.RepoRoot, baseRef, targetRef)
	if err != nil {
		return fmt.Errorf("failed to query file diff between %s and %s: %w", baseRef, targetRef, err)
	}

	report.TotalFilesCount = len(diffFiles)
	for _, f := range diffFiles {
		match := ClassifyFile(f)
		if match.IsInvariant {
			report.InvariantCount++
		}
		report.ModifiedFiles = append(report.ModifiedFiles, match)
	}

	report.HasSyncPending = len(commits) > 0

	// 6. Generate Agent Summary if requested
	if cfg.Agent != "" && report.HasSyncPending {
		summaryFile, err := GeneratePreMergeSummary(cfg, report, &DefaultAgentRunner{})
		if err != nil {
			return err
		}
		report.SummaryFile = summaryFile
	}

	if cfg.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	// Render human-readable status
	fmt.Printf("\n=== Fork Sync Status ===\n\n")
	fmt.Printf("  Repository Root: %s\n", cfg.RepoRoot)
	fmt.Printf("  Current Branch:  %s\n", report.CurrentBranch)
	fmt.Printf("  Upstream Remote: %s (%s)\n", report.UpstreamRemote, report.UpstreamURL)
	fmt.Printf("  Target Ref:      %s\n", targetRef)
	fmt.Printf("  Working Tree:    %s\n\n", map[bool]string{true: "Clean", false: "Modified (uncommitted changes exist)"}[report.WorkingTreeClean])

	if len(commits) == 0 {
		fmt.Println("  Status: Already up to date with upstream/main. No incoming commits.")
		return nil
	}

	fmt.Printf("  Incoming Commits: %d\n", len(commits))
	fmt.Println("  ------------------------------------------------------------")
	displayLimit := 10
	if len(commits) < displayLimit || cfg.Verbose {
		displayLimit = len(commits)
	}
	for i := 0; i < displayLimit; i++ {
		c := commits[i]
		fmt.Printf("  * [%s] %s (%s, %s)\n", c.Hash, c.Subject, c.Author, c.Date)
	}
	if len(commits) > displayLimit {
		fmt.Printf("  ... and %d more commits (use --verbose to view all)\n", len(commits)-displayLimit)
	}
	fmt.Println()

	fmt.Printf("  Modified Files: %d total (%d touching fork invariants)\n", report.TotalFilesCount, report.InvariantCount)
	if report.InvariantCount > 0 {
		fmt.Println("\n  [!] CAUTION: The following incoming modified files overlap with Fork Invariants:")
		for _, f := range report.ModifiedFiles {
			if f.IsInvariant {
				fmt.Printf("    - %s\n      Category: %s\n      Detail:   %s\n", f.Path, f.Category, f.Description)
			}
		}
		fmt.Println("\n  When merging, preserve fork invariants (telemetry, installer, onboarding, releases).")
	}

	if cfg.Verbose {
		fmt.Println("\n  General upstream changed files:")
		for _, f := range report.ModifiedFiles {
			if !f.IsInvariant {
				fmt.Printf("    - %s\n", f.Path)
			}
		}
	}

	if report.SummaryFile != "" {
		fmt.Printf("\n  [✓] Pre-merge analysis markdown written to:\n      %s\n", report.SummaryFile)
	}

	fmt.Printf("\n  To begin sync, run:\n    tools/fork-sync start\n\n")
	return nil
}

// RunStart prepares a sync branch and initiates merge.
func RunStart(cfg Config) error {
	// 1. Ensure clean working tree
	if !cfg.AllowDirty {
		clean, status, err := IsWorkingTreeClean(cfg.RepoRoot)
		if err != nil {
			return fmt.Errorf("failed to check working tree: %w", err)
		}
		if !clean {
			return fmt.Errorf("cannot start sync with uncommitted changes in working tree:\n%s\n\nCommit, stash, or run with --allow-dirty", status)
		}
	}

	// 2. Ensure remote and fetch
	url, _, err := EnsureUpstreamRemote(cfg.RepoRoot, cfg.UpstreamRemote, cfg.UpstreamURL)
	if err != nil {
		return fmt.Errorf("remote error: %w", err)
	}
	if !cfg.JSON {
		fmt.Printf("[*] Fetching upstream refs from %s (%s)...\n", cfg.UpstreamRemote, url)
	}
	if err := FetchRemote(cfg.RepoRoot, cfg.UpstreamRemote); err != nil {
		return fmt.Errorf("fetch error: %w", err)
	}

	// 3. Determine sync branch name
	syncBranch := cfg.SyncBranch
	if syncBranch == "" {
		dateStr := time.Now().Format("2006-01-02")
		syncBranch = fmt.Sprintf("sync/upstream-%s", dateStr)
	}

	// 4. Create and checkout sync branch from current base
	baseRef := cfg.LocalBaseBranch
	if baseRef == "" {
		baseRef = "main"
	}
	if !cfg.JSON {
		fmt.Printf("[*] Creating and checking out branch '%s' from '%s'...\n", syncBranch, baseRef)
	}
	if err := CreateAndCheckoutBranch(cfg.RepoRoot, syncBranch, baseRef); err != nil {
		return fmt.Errorf("failed to create sync branch %s: %w", syncBranch, err)
	}

	// 5. Initiate merge
	targetRef := fmt.Sprintf("%s/%s", cfg.UpstreamRemote, cfg.UpstreamBranch)
	if !cfg.JSON {
		fmt.Printf("[*] Merging %s into %s...\n", targetRef, syncBranch)
	}

	mergeRes, err := MergeBranch(cfg.RepoRoot, targetRef)
	if err != nil && len(mergeRes.ConflictedFiles) == 0 {
		return fmt.Errorf("merge failed: %w", err)
	}

	if mergeRes.Clean {
		if !cfg.JSON {
			fmt.Printf("\n[✓] Clean merge completed successfully!\n\n")
			fmt.Println("Running fork invariant validation...")
		}
		res, ok := VerifyAll(cfg.RepoRoot)
		if !ok {
			if !cfg.JSON {
				fmt.Println("\n[!] Invariant validation warning: some fork invariants may have been modified.")
			}
		} else if !cfg.JSON {
			fmt.Println("[✓] All fork invariants intact.")
		}

		if cfg.JSON {
			output := map[string]interface{}{
				"status":      "merged_clean",
				"branch":      syncBranch,
				"invariants":  res,
				"all_passed":  ok,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(output)
		}

		fmt.Println("\nNext steps:")
		fmt.Println("  1. Run static checks and tests: npm run check")
		fmt.Println("  2. Verify invariants anytime:   tools/fork-sync verify --run-checks")
		fmt.Println("  3. Commit, push branch, and open PR.")
		return nil
	}

	// Merge has conflicts
	var invariantConflicts []InvariantMatch
	var generalConflicts []string
	for _, f := range mergeRes.ConflictedFiles {
		match := ClassifyFile(f)
		if match.IsInvariant {
			invariantConflicts = append(invariantConflicts, match)
		} else {
			generalConflicts = append(generalConflicts, f)
		}
	}

	if cfg.JSON {
		output := map[string]interface{}{
			"status":              "conflict",
			"branch":              syncBranch,
			"conflicted_files":    mergeRes.ConflictedFiles,
			"invariant_conflicts": invariantConflicts,
			"general_conflicts":   generalConflicts,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	fmt.Printf("\n[!] Merge conflicts detected across %d file(s).\n\n", len(mergeRes.ConflictedFiles))

	if len(invariantConflicts) > 0 {
		fmt.Printf("=== CRITICAL: %d Fork Invariant File(s) in Conflict ===\n", len(invariantConflicts))
		fmt.Println("You MUST resolve these preserving fork behavior (telemetry opt-in, non-sudo user-local install, etc.):")
		for _, ic := range invariantConflicts {
			fmt.Printf("  * %s\n    Invariant:   %s\n    Description: %s\n", ic.Path, ic.Category, ic.Description)
		}
		fmt.Println()
	}

	if len(generalConflicts) > 0 {
		fmt.Printf("=== General Upstream Conflicts (%d files) ===\n", len(generalConflicts))
		for _, gc := range generalConflicts {
			fmt.Printf("  * %s\n", gc)
		}
		fmt.Println()
	}

	fmt.Println("Resolution Guide:")
	fmt.Println("  1. Inspect and edit conflicted files.")
	fmt.Println("  2. Stage resolved files: git add <resolved-files>")
	fmt.Println("  3. Run invariant verifier: tools/fork-sync verify --run-checks")
	fmt.Println("  4. Complete merge commit: git commit")
	return nil
}

// RunVerify executes fork invariant validation.
func RunVerify(cfg Config) error {
	results, allPassed := VerifyAll(cfg.RepoRoot)

	npmCheckPassed := true
	interactiveTestPassed := true
	var npmOutput string
	var interactiveOutput string

	if cfg.RunChecks {
		if !cfg.JSON {
			fmt.Println("[*] Executing repository static checks (npm run check)...")
		}
		cmd := exec.Command("npm", "run", "check")
		cmd.Dir = cfg.RepoRoot
		out, err := cmd.CombinedOutput()
		npmOutput = string(out)
		if err != nil {
			npmCheckPassed = false
			allPassed = false
		}

		// Run make interactive-test if Makefile exists and tmux is present
		makefilePath := filepath.Join(cfg.RepoRoot, "Makefile")
		if _, statErr := os.Stat(makefilePath); statErr == nil {
			if _, lookErr := exec.LookPath("tmux"); lookErr == nil {
				if !cfg.JSON {
					fmt.Println("[*] Executing interactive TUI smoke test (make interactive-test)...")
				}
				makeCmd := exec.Command("make", "interactive-test")
				makeCmd.Dir = cfg.RepoRoot
				iOut, iErr := makeCmd.CombinedOutput()
				interactiveOutput = string(iOut)
				if iErr != nil {
					interactiveTestPassed = false
					allPassed = false
				}
			}
		}
	}

	if cfg.JSON {
		output := map[string]interface{}{
			"all_passed":              allPassed,
			"invariant_results":       results,
			"run_checks":              cfg.RunChecks,
			"npm_check_passed":        npmCheckPassed,
			"npm_output":              npmOutput,
			"interactive_test_passed": interactiveTestPassed,
			"interactive_output":      interactiveOutput,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(output); err != nil {
			return err
		}
		if !allPassed {
			return fmt.Errorf("verification failed: one or more fork invariants or checks violated")
		}
		return nil
	}

	fmt.Printf("\n=== Fork Invariant Verification ===\n\n")
	for _, r := range results {
		statusStr := "[PASS]"
		if !r.Passed {
			statusStr = "[FAIL]"
		}
		fmt.Printf("  %-6s %s (%s)\n", statusStr, r.Name, r.Category)
		fmt.Printf("         %s\n", r.Description)
		if r.Details != "" {
			fmt.Printf("         %s\n", r.Details)
		}
		fmt.Println()
	}

	if cfg.RunChecks {
		fmt.Printf("  Repository static checks (npm run check): ")
		if npmCheckPassed {
			fmt.Println("[PASS]")
		} else {
			fmt.Println("[FAIL]")
			fmt.Printf("\n%s\n", npmOutput)
		}

		if interactiveOutput != "" {
			fmt.Printf("  Interactive TUI smoke test (make interactive-test): ")
			if interactiveTestPassed {
				fmt.Println("[PASS]")
			} else {
				fmt.Println("[FAIL]")
				fmt.Printf("\n%s\n", interactiveOutput)
			}
		}
		fmt.Println()
	}

	if !allPassed {
		return fmt.Errorf("verification failed: one or more fork invariants or checks violated")
	}

	fmt.Println("All fork invariant assertions and checks PASSED.")
	return nil
}
