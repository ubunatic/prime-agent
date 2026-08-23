package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClassifyFile(t *testing.T) {
	tests := []struct {
		path         string
		wantCategory InvariantCategory
		isInvariant  bool
	}{
		{
			path:         "packages/coding-agent/src/core/telemetry.ts",
			wantCategory: CategoryTelemetry,
			isInvariant:  true,
		},
		{
			path:         "packages/coding-agent/src/core/settings-manager.ts",
			wantCategory: CategoryTelemetry,
			isInvariant:  true,
		},
		{
			path:         "install.sh",
			wantCategory: CategoryInstaller,
			isInvariant:  true,
		},
		{
			path:         "install-beta.sh",
			wantCategory: CategoryInstaller,
			isInvariant:  true,
		},
		{
			path:         "packages/coding-agent/src/modes/interactive/onboarding.ts",
			wantCategory: CategoryOnboarding,
			isInvariant:  true,
		},
		{
			path:         "scripts/release.mjs",
			wantCategory: CategoryRelease,
			isInvariant:  true,
		},
		{
			path:         ".github/workflows/build-binaries.yml",
			wantCategory: CategoryWorkflows,
			isInvariant:  true,
		},
		{
			path:         "packages/tui/src/fullscreen.ts",
			wantCategory: CategoryTUIControls,
			isInvariant:  true,
		},
		{
			path:         "tools/fork-sync/main.go",
			wantCategory: CategoryOtherFork,
			isInvariant:  true,
		},
		{
			path:         "packages/ai/src/index.ts",
			wantCategory: "",
			isInvariant:  false,
		},
		{
			path:         "README.md",
			wantCategory: "",
			isInvariant:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			got := ClassifyFile(tc.path)
			if got.IsInvariant != tc.isInvariant {
				t.Errorf("ClassifyFile(%q).IsInvariant = %v, want %v", tc.path, got.IsInvariant, tc.isInvariant)
			}
			if tc.isInvariant && got.Category != tc.wantCategory {
				t.Errorf("ClassifyFile(%q).Category = %q, want %q", tc.path, got.Category, tc.wantCategory)
			}
		})
	}
}

func TestVerifyRealRepositoryInvariants(t *testing.T) {
	// Locate repository root (two levels up from tools/fork-sync)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(cwd, "..", ".."))

	results, allPassed := VerifyAll(repoRoot)
	if !allPassed {
		t.Errorf("VerifyAll on current repository failed: %+v", results)
	}
	if len(results) != len(Verifiers) {
		t.Errorf("Expected %d verifier results, got %d", len(Verifiers), len(results))
	}
	for _, r := range results {
		if !r.Passed {
			t.Errorf("Invariant check %q failed: %s", r.Name, r.Details)
		}
	}
}

func TestVerifiersSyntheticFailures(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Missing telemetry file
	res := VerifyTelemetryInvariant(tmpDir)
	if res.Passed {
		t.Errorf("Expected VerifyTelemetryInvariant to fail on empty dir")
	}

	// 2. Corrupted telemetry file (missing override checks)
	telemetryDir := filepath.Join(tmpDir, "packages", "coding-agent", "src", "core")
	if err := os.MkdirAll(telemetryDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(telemetryDir, "telemetry.ts"), []byte("const telemetry = true;"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	res = VerifyTelemetryInvariant(tmpDir)
	if res.Passed {
		t.Errorf("Expected VerifyTelemetryInvariant to fail when overrides missing")
	}

	// 3. Corrupted installer (sudo npm install -g)
	if err := os.WriteFile(filepath.Join(tmpDir, "install.sh"), []byte("sudo npm install -g prime-agent"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	res = VerifyInstallerInvariant(tmpDir)
	if res.Passed {
		t.Errorf("Expected VerifyInstallerInvariant to fail when sudo npm install present")
	}

	// 4. Corrupted onboarding (returning true for prime cli splash)
	onboardingDir := filepath.Join(tmpDir, "packages", "coding-agent", "src", "modes", "interactive")
	if err := os.MkdirAll(onboardingDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(onboardingDir, "onboarding.ts"), []byte("export function shouldRunPrimeCliOnboardingSplash() { return true; }"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	res = VerifyOnboardingInvariant(tmpDir)
	if res.Passed {
		t.Errorf("Expected VerifyOnboardingInvariant to fail when shouldRunPrimeCliOnboardingSplash returns true")
	}

	// 5. Corrupted release script (missing PI_SKIP_NPM_PUBLISH)
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "release.mjs"), []byte("console.log('standard release');"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	res = VerifyReleaseScriptInvariant(tmpDir)
	if res.Passed {
		t.Errorf("Expected VerifyReleaseScriptInvariant to fail when PI_SKIP_NPM_PUBLISH missing")
	}

	// 6. Missing build-binaries.yml workflow
	res = VerifyWorkflowsInvariant(tmpDir)
	if res.Passed {
		t.Errorf("Expected VerifyWorkflowsInvariant to fail when workflow file missing")
	}
}
