package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SupportedAgent represents a supported AI coding agent CLI.
type SupportedAgent string

const (
	AgentCodex       SupportedAgent = "codex"
	AgentAntigravity SupportedAgent = "antigravity"
	AgentAgy         SupportedAgent = "agy"
	AgentClaude      SupportedAgent = "claude"
)

// AgentRunner invokes an external AI agent CLI.
type AgentRunner interface {
	RunPrompt(agent SupportedAgent, prompt string, repoRoot string) (string, error)
	IsAvailable(agent SupportedAgent) bool
}

// DefaultAgentRunner implements AgentRunner using os/exec.
type DefaultAgentRunner struct{}

// IsAvailable checks if the CLI tool for the given agent exists in PATH.
func (d *DefaultAgentRunner) IsAvailable(agent SupportedAgent) bool {
	binary := getAgentBinary(agent)
	if binary == "" {
		return false
	}
	_, err := exec.LookPath(binary)
	return err == nil
}

func getAgentBinary(agent SupportedAgent) string {
	switch strings.ToLower(string(agent)) {
	case "codex":
		return "codex"
	case "antigravity":
		if _, err := exec.LookPath("antigravity"); err == nil {
			return "antigravity"
		}
		return "agy"
	case "agy":
		return "agy"
	case "claude":
		return "claude"
	default:
		return string(agent)
	}
}

// RunPrompt executes the agent CLI non-interactively with the provided prompt.
func (d *DefaultAgentRunner) RunPrompt(agent SupportedAgent, prompt string, repoRoot string) (string, error) {
	binary := getAgentBinary(agent)
	if binary == "" {
		return "", fmt.Errorf("unsupported agent: %s", agent)
	}

	if _, err := exec.LookPath(binary); err != nil {
		return "", fmt.Errorf("agent executable '%s' not found in PATH", binary)
	}

	var cmd *exec.Cmd

	switch strings.ToLower(string(agent)) {
	case "codex":
		// codex exec [prompt]
		cmd = exec.Command(binary, "exec", prompt)
	case "antigravity", "agy":
		// agy -p <prompt> --dangerously-skip-permissions
		cmd = exec.Command(binary, "-p", prompt, "--dangerously-skip-permissions")
	case "claude":
		// claude -p <prompt> --allow-dangerously-skip-permissions
		cmd = exec.Command(binary, "-p", prompt, "--allow-dangerously-skip-permissions")
	default:
		// Generic fallback: <binary> -p <prompt>
		cmd = exec.Command(binary, "-p", prompt)
	}

	cmd.Dir = repoRoot
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("agent %s failed (exit %v): %s\n%s", agent, err, stderr.String(), stdout.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" && stderr.Len() > 0 {
		output = strings.TrimSpace(stderr.String())
	}
	return output, nil
}

// BuildSummaryPrompt constructs the instruction prompt for the agent to review incoming upstream diffs.
func BuildSummaryPrompt(report StatusReport, diffText string, outputPath string) string {
	var sb strings.Builder

	sb.WriteString("You are analyzing incoming upstream changes for the ubunatic/prime-agent repository fork.\n\n")
	sb.WriteString(fmt.Sprintf("Upstream Target: %s (%s)\n", report.TargetRef, report.UpstreamURL))
	sb.WriteString(fmt.Sprintf("Incoming Commits: %d\n", len(report.IncomingCommits)))
	sb.WriteString(fmt.Sprintf("Total Modified Files: %d (%d overlap with protected fork invariants)\n\n", report.TotalFilesCount, report.InvariantCount))

	if report.InvariantCount > 0 {
		sb.WriteString("Protected Fork Invariant Overlap:\n")
		for _, f := range report.ModifiedFiles {
			if f.IsInvariant {
				sb.WriteString(fmt.Sprintf("- %s (Category: %s, Detail: %s)\n", f.Path, f.Category, f.Description))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Incoming Commits:\n")
	limit := 20
	if len(report.IncomingCommits) < limit {
		limit = len(report.IncomingCommits)
	}
	for i := 0; i < limit; i++ {
		c := report.IncomingCommits[i]
		sb.WriteString(fmt.Sprintf("- [%s] %s (%s, %s)\n", c.Hash, c.Subject, c.Author, c.Date))
	}
	if len(report.IncomingCommits) > limit {
		sb.WriteString(fmt.Sprintf("- ... and %d more commits\n", len(report.IncomingCommits)-limit))
	}
	sb.WriteString("\n")

	sb.WriteString("TASK:\n")
	sb.WriteString(fmt.Sprintf("1. Analyze the upstream changes against this fork's architectural invariants (docs/ForkArchitecture.md).\n"))
	sb.WriteString(fmt.Sprintf("2. Highlight potential merge issues, breaking protocol/API changes, or conflicts that might arise during merge.\n"))
	sb.WriteString(fmt.Sprintf("3. Write a comprehensive markdown report directly to `%s`.\n", outputPath))
	sb.WriteString("Format the markdown with:\n")
	sb.WriteString("- # Upstream Sync Pre-Merge Analysis (<date>)\n")
	sb.WriteString("- ## Executive Summary\n")
	sb.WriteString("- ## High-Risk & Fork Invariant Impact (telemetry, installer, onboarding, releases, etc.)\n")
	sb.WriteString("- ## Notable Upstream Features & Fixes\n")
	sb.WriteString("- ## Recommended Merge Strategy & Action Items\n")

	if diffText != "" {
		sb.WriteString("\n--- SUMMARY OF FILE STATS ---\n")
		sb.WriteString(diffText)
	}

	return sb.String()
}

// GeneratePreMergeSummary triggers an agent review of incoming changes and saves the summary markdown.
func GeneratePreMergeSummary(cfg Config, report StatusReport, runner AgentRunner) (string, error) {
	if cfg.Agent == "" {
		return "", nil
	}

	agent := SupportedAgent(cfg.Agent)
	if !runner.IsAvailable(agent) {
		return "", fmt.Errorf("agent '%s' requested but binary is not available in PATH", cfg.Agent)
	}

	// Determine output summary path
	outputPath := cfg.SummaryOutput
	if outputPath == "" {
		dateStr := time.Now().Format("2006-01-02")
		outputPath = filepath.Join(cfg.RepoRoot, "issues", fmt.Sprintf("upstream-sync-summary-%s.md", dateStr))
	}

	targetRef := report.TargetRef
	baseRef := cfg.LocalBaseBranch
	if baseRef == "" {
		baseRef = "HEAD"
	}

	// Query diff stat
	diffStat, _ := GitRunnerDefault.GetDiffStat(cfg.RepoRoot, baseRef, targetRef)

	prompt := BuildSummaryPrompt(report, diffStat, outputPath)

	if !cfg.JSON {
		fmt.Printf("\n[*] Spawning agent '%s' to analyze incoming changes and write summary to:\n    %s\n", agent, outputPath)
	}

	output, err := runner.RunPrompt(agent, prompt, cfg.RepoRoot)
	if err != nil {
		return "", fmt.Errorf("agent %s failed to generate summary: %w", agent, err)
	}

	// Check if file was written by agent; if not, write agent stdout to the target path
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		if strings.TrimSpace(output) != "" {
			if writeErr := os.WriteFile(outputPath, []byte(output), 0644); writeErr != nil {
				return "", fmt.Errorf("failed to write agent output to %s: %w", outputPath, writeErr)
			}
		}
	}

	return outputPath, nil
}
