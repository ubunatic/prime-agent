package main

import (
	"os"
	"path/filepath"
	"testing"
)

type mockGitRunner struct {
	responses map[string]string
	errors    map[string]error
	calls     [][]string
}

func (m *mockGitRunner) Run(dir string, args ...string) (string, error) {
	cmdKey := ""
	if len(args) > 0 {
		cmdKey = args[0]
	}
	m.calls = append(m.calls, args)
	if err, ok := m.errors[cmdKey]; ok {
		return m.responses[cmdKey], err
	}
	return m.responses[cmdKey], nil
}

func TestGitFindRepoRoot(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	root, err := FindRepoRoot(cwd)
	if err != nil {
		t.Fatalf("FindRepoRoot failed: %v", err)
	}
	if root == "" {
		t.Errorf("FindRepoRoot returned empty string")
	}

	expectedRoot := filepath.Clean(filepath.Join(cwd, "..", ".."))
	if root != expectedRoot {
		t.Errorf("FindRepoRoot = %q, want %q", root, expectedRoot)
	}
}

func TestGetRemotesMock(t *testing.T) {
	oldRunner := defaultGitRunner
	defer func() { defaultGitRunner = oldRunner }()

	mock := &mockGitRunner{
		responses: map[string]string{
			"remote": "origin\tgit@github.com:ubunatic/prime-agent.git (fetch)\norigin\tgit@github.com:ubunatic/prime-agent.git (push)\nupstream\thttps://github.com/PrimeIntellect-ai/prime-agent.git (fetch)\n",
		},
	}
	defaultGitRunner = mock

	remotes, err := GetRemotes("/mock/root")
	if err != nil {
		t.Fatalf("GetRemotes error: %v", err)
	}
	if len(remotes) != 2 {
		t.Errorf("Expected 2 remotes, got %d", len(remotes))
	}
	if remotes["origin"] != "git@github.com:ubunatic/prime-agent.git" {
		t.Errorf("origin = %q", remotes["origin"])
	}
	if remotes["upstream"] != "https://github.com/PrimeIntellect-ai/prime-agent.git" {
		t.Errorf("upstream = %q", remotes["upstream"])
	}
}

func TestEnsureUpstreamRemoteMock(t *testing.T) {
	oldRunner := defaultGitRunner
	defer func() { defaultGitRunner = oldRunner }()

	// Case 1: remote already exists
	mock := &mockGitRunner{
		responses: map[string]string{
			"remote": "upstream\thttps://github.com/PrimeIntellect-ai/prime-agent.git (fetch)\n",
		},
	}
	defaultGitRunner = mock

	url, added, err := EnsureUpstreamRemote("/mock/root", "upstream", "https://github.com/PrimeIntellect-ai/prime-agent.git")
	if err != nil {
		t.Fatalf("EnsureUpstreamRemote error: %v", err)
	}
	if added {
		t.Errorf("Expected added=false for existing remote")
	}
	if url != "https://github.com/PrimeIntellect-ai/prime-agent.git" {
		t.Errorf("url = %q", url)
	}

	// Case 2: remote is added
	mockAdded := &mockGitRunner{
		responses: map[string]string{
			"remote": "origin\tgit@github.com:ubunatic/prime-agent.git (fetch)\n",
		},
	}
	defaultGitRunner = mockAdded

	url, added, err = EnsureUpstreamRemote("/mock/root", "upstream", "https://github.com/PrimeIntellect-ai/prime-agent.git")
	if err != nil {
		t.Fatalf("EnsureUpstreamRemote error: %v", err)
	}
	if !added {
		t.Errorf("Expected added=true for new remote")
	}
	if url != "https://github.com/PrimeIntellect-ai/prime-agent.git" {
		t.Errorf("url = %q", url)
	}
}

func TestGetIncomingCommitsAndDiffMock(t *testing.T) {
	oldRunner := defaultGitRunner
	defer func() { defaultGitRunner = oldRunner }()

	mock := &mockGitRunner{
		responses: map[string]string{
			"log":  "abc1234\tJane Doe\t2026-08-20\tfeat: upstream update\ndef5678\tJohn Smith\t2026-08-21\tfix: resolve bug\n",
			"diff": "packages/ai/src/index.ts\ninstall.sh\n",
		},
	}
	defaultGitRunner = mock

	commits, err := GetIncomingCommits("/mock/root", "main", "upstream/main")
	if err != nil {
		t.Fatalf("GetIncomingCommits error: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("Expected 2 commits, got %d", len(commits))
	}
	if commits[0].Hash != "abc1234" || commits[0].Author != "Jane Doe" || commits[0].Subject != "feat: upstream update" {
		t.Errorf("Commit[0] mismatch: %+v", commits[0])
	}

	files, err := GetDiffFiles("/mock/root", "main", "upstream/main")
	if err != nil {
		t.Fatalf("GetDiffFiles error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(files))
	}
	if files[0] != "packages/ai/src/index.ts" || files[1] != "install.sh" {
		t.Errorf("Files mismatch: %+v", files)
	}
}
