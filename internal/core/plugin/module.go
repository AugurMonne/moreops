package plugin

import (
	"context"
	"fmt"
	"sync"

	"moreops/internal/core/config"
	"moreops/internal/core/diff"
)

// PlanResult encapsulates all planned mutations by a module.
type PlanResult struct {
	ModuleName  string
	FileChanges []diff.FileChange
	Commands    []string
	Summary     string
}

// Module is the uniform interface all MoreOps modules implement.
type Module interface {
	Name() string
	Description() string
	Plan(ctx context.Context, cfg *config.Config, envName string) (*PlanResult, error)
	Apply(ctx context.Context, plan *PlanResult) error
}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Module)
)

// Register adds a module to the global registry.
func Register(m Module) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[m.Name()] = m
}

// Get retrieves a module by name.
func Get(name string) (Module, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	m, ok := registry[name]
	return m, ok
}

// All returns all registered modules.
func All() []Module {
	registryMu.RLock()
	defer registryMu.RUnlock()
	modules := make([]Module, 0, len(registry))
	for _, m := range registry {
		modules = append(modules, m)
	}
	return modules
}

// ClearRegistry resets the module registry (useful for testing).
func ClearRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = make(map[string]Module)
}

// ExecutePlan writes planned file changes to disk.
func ExecutePlan(plan *PlanResult) error {
	for _, change := range plan.FileChanges {
		switch change.Action {
		case diff.ActionCreate, diff.ActionUpdate:
			if err := diff.WriteFile(change.Path, change.NewData); err != nil {
				return fmt.Errorf("failed to write %s: %w", change.Path, err)
			}
		case diff.ActionDelete:
			if err := diff.DeleteFile(change.Path); err != nil {
				return fmt.Errorf("failed to delete %s: %w", change.Path, err)
			}
		case diff.ActionNone:
			// noop
		}
	}
	return nil
}
