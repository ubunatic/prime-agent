# Proposal: Disable Forced Login / Make Prime Intellect Authentication Optional

This document analyzes where Prime Intellect authentication (`Prime Inference` / `/login`) is triggered, where forced login prompts or default fallbacks occur, and how to make local or alternative provider configurations the default experience without requiring a Prime account.

**Status:** Resolved (Forced Prime Intellect login splash & defaults disabled)

---

## 1. Context & Motivation

Currently, when launching `prime-agent` for the first time without pre-configured credentials:
1. **Onboarding Splash**: Checks for Prime CLI credentials (`~/.prime/config.json`) and prompts the user to log in or select Prime Inference models.
2. **Default Fallback Model**: If no model or provider API key (e.g. `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`) is found in environment or settings, the resolver falls back to Prime Inference or prompts `/login`.
3. **Trace Uploads**: Automatically checks for Prime credentials to upload agent session traces.

Users should be able to use `prime-agent` fully offline or with their own API keys (Anthropic, OpenAI, Ollama, OpenRouter, etc.) without ever encountering forced login prompts or browser auth redirects for Prime Intellect.

---

## 2. Technical Findings: Key Intervention Points

### A. Startup Onboarding (`packages/coding-agent/src/modes/interactive/onboarding.ts`)
- **Location:** [`packages/coding-agent/src/modes/interactive/onboarding.ts`](../packages/coding-agent/src/modes/interactive/onboarding.ts)
- **Current Behavior:** `shouldRunOnboarding()` checks if any model is configured with valid auth. If not, it runs the interactive onboarding flow, which presents Prime CLI / Prime Inference splash components (`PrimeOnboardingSplashComponent`).
- **Proposed Modification:**
  - Skip forced login onboarding splash when `onboardingShown` is false.
  - Default the initial model selection view directly to provider/API key entry or local models (Ollama/Custom OpenAI proxy) rather than defaulting to Prime Inference auth flow.

### B. Initial Model Fallback Resolution (`packages/coding-agent/src/core/model-resolver.ts`)
- **Location:** `packages/coding-agent/src/core/model-resolver.ts`
- **Current Behavior:** If no default model is specified, model resolution attempts to discover available provider keys. If none exist, it offers `prime-inference` as the default provider option.
- **Proposed Modification:**
  - Fall back to standard environment key detection or local endpoint providers (e.g. Ollama `http://localhost:11434` or standard `OPENAI_API_KEY`/`ANTHROPIC_API_KEY`).
  - Do not automatically prompt or force Prime Intellect login on missing model errors.

### C. Agent Traces Logging (`packages/coding-agent/src/core/agent-traces.ts`)
- **Location:** `packages/coding-agent/src/core/agent-traces.ts`
- **Current Behavior:** Tracing checks for Prime CLI credentials to upload telemetry/traces (`PRIME_AGENT_TRACES_PROVIDER_ID`).
- **Proposed Modification:**
  - Keep traces local only by default. Skip background checks for `~/.prime/config.json` unless `/traces login` is explicitly run by the user.

---

## 3. Recommended Code Changes

1. **Update `onboarding.ts`**:
   ```ts
   // Make Prime CLI splash opt-in and ensure onboarding allows selecting key-based providers directly
   export function shouldRunPrimeCliOnboardingSplash(_state: OnboardingStartupState): boolean {
       return false; // Disable forced Prime CLI onboarding splash
   }
   ```

2. **Update Auth Guidance Messages (`packages/coding-agent/src/core/auth-guidance.ts`)**:
   - Change generic "Run /login to update credentials" messages to guide the user to set environment variables (`export ANTHROPIC_API_KEY=...`) or use key-based setup.

---

## 4. Benefits
- **Privacy & Autonomy**: Users can run `prime-agent` completely standalone with standard LLM provider keys or local models (Ollama/vLLM).
- **No Third-Party Credentials Required**: Completely removes mandatory account creation/login for Prime Intellect.
