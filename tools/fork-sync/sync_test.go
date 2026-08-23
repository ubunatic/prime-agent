package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunVerify(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(cwd, "..", ".."))

	cfg := Config{
		RepoRoot:  repoRoot,
		JSON:      true,
		RunChecks: false,
	}

	err = RunVerify(cfg)
	if err != nil {
		t.Errorf("RunVerify failed on repo root: %v", err)
	}
}

func TestRunVerifyFailure(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		RepoRoot: tmpDir,
		JSON:     true,
	}

	err := RunVerify(cfg)
	if err == nil {
		t.Errorf("Expected RunVerify to fail on empty directory")
	}
}

func TestRunStatusMock(t *testing.T) {
	oldRunner := defaultGitRunner
	defer func() { defaultGitRunner = oldRunner }()

	mock := &mockGitRunner{
		responses: map[string]string{
			"rev-parse": "main\n",
			"status":    "",
			"remote":    "upstream\thttps://github.com/PrimeIntellect-ai/prime-agent.git (fetch)\n",
			"log":       "abc1234\tDeveloper\t2026-08-22\tfeat: new feature\n",
			"diff":      "packages/coding-agent/src/core/telemetry.ts\nREADME.md\n",
		},
	}
	defaultGitRunner = mock

	cfg := Config{
		RepoRoot:        "/mock/root",
		UpstreamRemote:  "upstream",
		UpstreamURL:     "https://github.com/PrimeIntellect-ai/prime-agent.git",
		UpstreamBranch:  "main",
		LocalBaseBranch: "main",
		Fetch:           false,
		JSON:            true,
	}

	err := RunStatus(cfg)
	if err != nil {
		t.Fatalf("RunStatus error: %v", err)
	}
}

func TestRunStartMockClean(t *testing.T) {
	oldRunner := defaultGitRunner
	defer func() { defaultGitRunner = oldRunner }()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(cwd, "..", ".."))

	mock := &mockGitRunner{
		responses: map[string]string{
			"status":   "",
			"remote":   "upstream\thttps://github.com/PrimeIntellect-ai/prime-agent.git (fetch)\n",
			"fetch":    "",
			"checkout": "",
			"merge":    "Updating abc1234..def5678\nFast-forward\n",
		},
	}
	defaultGitRunner = mock

	cfg := Config{
		RepoRoot:        repoRoot,
		UpstreamRemote:  "upstream",
		UpstreamURL:     "https://github.com/PrimeIntellect-ai/prime-agent.git",
		UpstreamBranch:  "main",
		LocalBaseBranch: "main",
		SyncBranch:      "sync/test-branch",
		Fetch:           false,
		JSON:            true,
	}

	err = RunStart(cfg)
	if err != nil {
		t.Fatalf("RunStart error: %v", err)
	}
}
