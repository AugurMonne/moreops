package engine

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"moreops/internal/core/config"
	"moreops/internal/core/diff"
	"moreops/internal/core/plugin"
)

// Engine orchestrates modules, plans, diffs, and execution.
type Engine struct {
	in  io.Reader
	out io.Writer
}

// NewEngine creates an Engine instance with customizable IO streams.
func NewEngine(in io.Reader, out io.Writer) *Engine {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	return &Engine{
		in:  in,
		out: out,
	}
}

// PlanResult holds the aggregated plan across all active modules.
type CompositePlan struct {
	Environment string
	ModulePlans []*plugin.PlanResult
	Changes     []diff.FileChange
	ToCreate    int
	ToUpdate    int
	ToDelete    int
	Unchanged   int
}

// Plan generates a preview of all actions without modifying state.
func (e *Engine) Plan(ctx context.Context, cfg *config.Config, envName string) (*CompositePlan, error) {
	if envName == "" {
		envName = "dev"
	}

	plan := &CompositePlan{
		Environment: envName,
		ModulePlans: make([]*plugin.PlanResult, 0),
		Changes:     make([]diff.FileChange, 0),
	}

	modules := plugin.All()
	for _, mod := range modules {
		res, err := mod.Plan(ctx, cfg, envName)
		if err != nil {
			return nil, fmt.Errorf("module '%s' plan failed: %w", mod.Name(), err)
		}
		if res != nil {
			plan.ModulePlans = append(plan.ModulePlans, res)
			for _, fc := range res.FileChanges {
				plan.Changes = append(plan.Changes, fc)
				switch fc.Action {
				case diff.ActionCreate:
					plan.ToCreate++
				case diff.ActionUpdate:
					plan.ToUpdate++
				case diff.ActionDelete:
					plan.ToDelete++
				case diff.ActionNone:
					plan.Unchanged++
				}
			}
		}
	}

	return plan, nil
}

// PrintDiff renders formatted diff for the entire plan.
func (e *Engine) PrintDiff(plan *CompositePlan, useColor bool) {
	if len(plan.Changes) == 0 {
		fmt.Fprintln(e.out, "No infrastructure changes planned.")
		return
	}

	fmt.Fprintf(e.out, "\n--- Planned Changes for Environment [%s] ---\n\n", plan.Environment)
	for _, fc := range plan.Changes {
		fmt.Fprint(e.out, diff.UnifiedDiff(fc, useColor))
	}

	fmt.Fprintf(e.out, "\nPlan summary: %d to create, %d to update, %d to delete, %d unchanged.\n",
		plan.ToCreate, plan.ToUpdate, plan.ToDelete, plan.Unchanged)
}

// Apply executes the plan after preview and explicit confirmation.
func (e *Engine) Apply(ctx context.Context, plan *CompositePlan, autoApprove bool) error {
	hasMutations := plan.ToCreate > 0 || plan.ToUpdate > 0 || plan.ToDelete > 0
	if !hasMutations {
		fmt.Fprintln(e.out, "Infrastructure is up to date. No changes to apply.")
		return nil
	}

	e.PrintDiff(plan, true)

	if !autoApprove {
		confirmed := e.askConfirmation("Do you want to perform these actions? (yes/no)")
		if !confirmed {
			fmt.Fprintln(e.out, "Apply cancelled by user.")
			return nil
		}
	}

	for _, mp := range plan.ModulePlans {
		if err := plugin.ExecutePlan(mp); err != nil {
			return fmt.Errorf("failed applying changes for module '%s': %w", mp.ModuleName, err)
		}
	}

	fmt.Fprintln(e.out, "\n✓ Apply completed successfully!")
	return nil
}

func (e *Engine) askConfirmation(prompt string) bool {
	fmt.Fprintf(e.out, "\n%s: ", prompt)
	reader := bufio.NewReader(e.in)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	ans := strings.ToLower(strings.TrimSpace(line))
	return ans == "yes" || ans == "y"
}
