package diff

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileAction describes what operation will happen to a file.
type FileAction string

const (
	ActionCreate FileAction = "CREATE"
	ActionUpdate FileAction = "UPDATE"
	ActionDelete FileAction = "DELETE"
	ActionNone   FileAction = "NOOP"
)

// FileChange represents a planned change to a file.
type FileChange struct {
	Path    string
	OldData string
	NewData string
	Action  FileAction
}

// ComputeFileChange compares existing file content on disk with desired new content.
func ComputeFileChange(path, newContent string) (FileChange, error) {
	fc := FileChange{
		Path:    path,
		NewData: newContent,
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		fc.Action = ActionCreate
		return fc, nil
	} else if err != nil {
		return fc, err
	}

	fc.OldData = string(data)
	if fc.OldData == fc.NewData {
		fc.Action = ActionNone
	} else {
		fc.Action = ActionUpdate
	}
	return fc, nil
}

// UnifiedDiff produces a unified diff string representation.
func UnifiedDiff(fc FileChange, useColor bool) string {
	var sb strings.Builder

	reset := ""
	green := ""
	red := ""
	cyan := ""
	yellow := ""

	if useColor {
		reset = "\033[0m"
		green = "\033[32m"
		red = "\033[31m"
		cyan = "\033[36m"
		yellow = "\033[33m"
	}

	switch fc.Action {
	case ActionCreate:
		sb.WriteString(fmt.Sprintf("%s+++ %s (new file)%s\n", green, fc.Path, reset))
		lines := strings.Split(fc.NewData, "\n")
		for _, line := range lines {
			if line == "" && strings.HasSuffix(fc.NewData, "\n") {
				continue
			}
			sb.WriteString(fmt.Sprintf("%s+%s%s\n", green, line, reset))
		}
	case ActionDelete:
		sb.WriteString(fmt.Sprintf("%s--- %s (deleted)%s\n", red, fc.Path, reset))
		lines := strings.Split(fc.OldData, "\n")
		for _, line := range lines {
			sb.WriteString(fmt.Sprintf("%s-%s%s\n", red, line, reset))
		}
	case ActionUpdate:
		sb.WriteString(fmt.Sprintf("%s--- a/%s%s\n", yellow, fc.Path, reset))
		sb.WriteString(fmt.Sprintf("%s+++ b/%s%s\n", yellow, fc.Path, reset))
		sb.WriteString(simpleLineDiff(fc.OldData, fc.NewData, green, red, reset))
	case ActionNone:
		sb.WriteString(fmt.Sprintf("%s# %s (unchanged)%s\n", cyan, fc.Path, reset))
	}

	return sb.String()
}

func simpleLineDiff(oldStr, newStr, green, red, reset string) string {
	oldLines := strings.Split(oldStr, "\n")
	newLines := strings.Split(newStr, "\n")

	var sb strings.Builder
	max := len(oldLines)
	if len(newLines) > max {
		max = len(newLines)
	}

	for i := 0; i < max; i++ {
		var o, n string
		hasOld := i < len(oldLines)
		hasNew := i < len(newLines)

		if hasOld {
			o = oldLines[i]
		}
		if hasNew {
			n = newLines[i]
		}

		if hasOld && hasNew {
			if o != n {
				sb.WriteString(fmt.Sprintf("%s-%s%s\n", red, o, reset))
				sb.WriteString(fmt.Sprintf("%s+%s%s\n", green, n, reset))
			} else {
				sb.WriteString(fmt.Sprintf(" %s\n", o))
			}
		} else if hasOld {
			sb.WriteString(fmt.Sprintf("%s-%s%s\n", red, o, reset))
		} else if hasNew {
			sb.WriteString(fmt.Sprintf("%s+%s%s\n", green, n, reset))
		}
	}

	return sb.String()
}

// WriteFile writes content ensuring parent directories exist.
func WriteFile(path, content string) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// DeleteFile removes a file if it exists.
func DeleteFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
