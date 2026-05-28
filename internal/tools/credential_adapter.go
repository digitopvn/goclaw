package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"sync"

	"github.com/nextlevelbuilder/goclaw/internal/store"
)

// CredentialAdapter transforms a user credential into the shape a specific CLI
// binary expects. Each binary's adapter_name column routes runtime exec to one
// of these. The default "passthrough" adapter is a no-op so unrelated presets
// (gh, aws, kubectl, gcloud, terraform, gws) keep their legacy env-injection
// path bit-for-bit.
//
// Adapters never see or log raw secret bytes — `Prepare` returns an Injection
// describing how to mutate argv/env, plus a Cleanup hook for ephemeral files
// and a ScrubValues slice that the caller registers with the per-request
// scrubber (see scrub.go WithScrubBag).
type CredentialAdapter interface {
	Name() string
	ShouldInject(argv []string) bool
	Prepare(ctx context.Context, bin *store.SecureCLIBinary, cred *store.SecureCLIUserCredential, argv []string) (*Injection, error)
}

// Injection is the adapter's instruction set for one exec call.
//
//   - ArgvPrefix is spliced in BETWEEN binary and user-supplied args (so
//     `git clone X` becomes `git <prefix...> clone X`). The slice is not
//     re-quoted or shell-parsed — it goes straight into exec.Command.
//   - Env is merged on top of the base env after legacy ValidateGrantEnvVars
//     has already vetted the user-controlled credential blob. Keys here come
//     from hard-coded adapter code only.
//   - Cleanup is deferred AFTER the synchronous exec returns; safe to be nil.
//   - ScrubValues lists adapter-derived secrets the per-request scrubber must
//     redact from stdout/stderr/error messages before they reach the LLM.
type Injection struct {
	ArgvPrefix  []string
	Env         map[string]string
	Cleanup     func() error
	ScrubValues []string
}

var (
	adaptersMu sync.RWMutex
	adapters   = map[string]CredentialAdapter{
		"passthrough": passthroughAdapter{},
	}
)

// RegisterAdapter registers a CredentialAdapter under its Name(). Idempotent —
// re-registering the same name overwrites. Intended for init() of each
// adapter file (e.g. credential_adapter_git.go in Phase 3).
func RegisterAdapter(a CredentialAdapter) {
	if a == nil {
		return
	}
	adaptersMu.Lock()
	defer adaptersMu.Unlock()
	adapters[a.Name()] = a
}

// AdapterFor resolves an adapter by name, falling back to passthrough on
// empty/unknown so an operator-set adapter_name typo can never break exec —
// it just degrades to legacy behavior with a clear audit trail.
func AdapterFor(name string) CredentialAdapter {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	if a, ok := adapters[name]; ok {
		return a
	}
	return adapters["passthrough"]
}

// passthroughAdapter is the default. It does nothing — preserves the legacy
// env-injection-only path so every existing preset behaves identically.
type passthroughAdapter struct{}

func (passthroughAdapter) Name() string                { return "passthrough" }
func (passthroughAdapter) ShouldInject([]string) bool  { return false }
func (passthroughAdapter) Prepare(context.Context, *store.SecureCLIBinary, *store.SecureCLIUserCredential, []string) (*Injection, error) {
	return &Injection{}, nil
}

// hashHostScope returns the SHA-256 hex prefix of the host-scope value, or
// "none" when nil. Used in audit logs instead of plaintext hostname so SIEM
// pipelines can correlate without exposing internal infra hostnames.
func hashHostScope(hs *string) string {
	if hs == nil || *hs == "" {
		return "none"
	}
	sum := sha256.Sum256([]byte(*hs))
	return hex.EncodeToString(sum[:4]) // 8 hex chars
}

// sortedKeys returns env map keys in deterministic order for audit logs.
func sortedKeys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
