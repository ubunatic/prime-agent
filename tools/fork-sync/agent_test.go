package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type MockAgentRunner struct {
	AvailableAgents map[SupportedAgent]bool
	LastPrompt      string
	LastAgent       SupportedAgent
	OutputToReturn  string
	ShouldError     bool
}

func (m *MockAgentRunner) IsAvailable(agent SupportedAgent) bool {
	if m.AvailableAgents == nil {
		return true
	}
	return m.AvailableAgents[agent]
}

func (m *MockAgentRunner) RunPrompt(agent SupportedAgent, prompt string, repoRoot string) (string, error) {
	m.LastAgent = agent
	m.LastPrompt = prompt
	if m.ShouldError {
		return "", os.ErrPermission
	}
	return m.OutputToReturn, nil
}

func TestBuildSummaryPrompt(t *testing.T) {
	report := StatusReport{
		TargetRef:       "upstream/main",
		UpstreamURL:     "https://github.com/PrimeIntellect-ai/prime-agent.git",
		TotalFilesCount: 50,
		InvariantCount:  2,
		ModifiedFiles: []InvariantMatch{
			{Path: "packages/coding-agent/src/core/telemetry.ts", Category: "Telemetry", Description: "Opt-in telemetry", IsInvariant: true},
			{Path: "install.sh", Category: "Installer", Description: "User-local install", IsInvariant: true},
		},
		IncomingCommits: []CommitInfo{
			{Hash: "abc1234", Subject: "feat: new cool feature", Author: "alice", Date: "2026-08-20"},
		},
	}

	prompt := BuildSummaryPrompt(report, "50 files changed", "issues/summary.md")

	if !strings.Contains(prompt, "upstream/main") {
		t.Errorf("prompt missing target ref")
	}
	if !strings.Contains(prompt, "telemetry.ts") {
		t.Errorf("prompt missing invariant file")
	}
	if !strings.Contains(prompt, "issues/summary.md") {
		t.Errorf("prompt missing output path")
	}
}

func TestGeneratePreMergeSummary(t *testing.T) {
	tmpDir := t.TempDir()
	summaryFile := filepath.Join(tmpDir, "summary.md")

	mockRunner := &MockAgentRunner{
		OutputToReturn: "# Upstream Sync Pre-Merge Analysis\n\nLooks good!",
	}

	cfg := Config{
		RepoRoot:      tmpDir,
		Agent:         "codex",
		SummaryOutput: summaryFile,
	}

	report := StatusReport{
		TargetRef:      "upstream/main",
		HasSyncPending: true,
	}

	outPath, err := GeneratePreMergeSummary(cfg, report, mockRunner)
	if err != nil {
		t.Fatalf("GeneratePreMergeSummary failed: %v", err)
	}

	if outPath != summaryFile {
		t.Errorf("expected outPath %s, got %s", summaryFile, outPath)
	}

	data, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatalf("failed to read written summary file: %v", err)
	}

	if !strings.Contains(string(data), "Upstream Sync Pre-Merge Analysis") {
		t.Errorf("written summary file content mismatch: %s", string(data))
	}
}

func TestAgentRunnerUnavailable(t *testing.T) {
	mockRunner := &MockAgentRunner{
		AvailableAgents: map[SupportedAgent]bool{
			AgentCodex: false,
		},
	}

	cfg := Config{
		RepoRoot: ".",
		Agent:    "codex",
	}

	report := StatusReport{
		HasSyncPending: true,
	}

	_, err := GeneratePreMergeSummary(cfg, report, mockRunner)
	if err == nil {
		t.Errorf("expected error for unavailable agent, got nil")
	}
}
