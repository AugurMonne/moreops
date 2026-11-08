package wizard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Choice represents a selectable item in a TUI menu.
type Choice struct {
	ID          int
	Key         string
	Label       string
	Description string
}

// fdGetter is an interface for types that have a file descriptor (like *os.File).
type fdGetter interface {
	Fd() uintptr
}

// IsInteractive checks if both input and output are interactive terminals.
func IsInteractive(in io.Reader, out io.Writer) bool {
	inFile, inOk := in.(fdGetter)
	outFile, outOk := out.(fdGetter)
	if !inOk || !outOk {
		return false
	}
	return isTerminal(inFile.Fd()) && isTerminal(outFile.Fd())
}

// SelectOne presents a single-choice interactive menu with arrow key navigation.
// Falls back to numeric text input if the terminal is not interactive.
func SelectOne(reader *bufio.Reader, in io.Reader, out io.Writer, prompt string, choices []Choice, defaultIndex int) (Choice, error) {
	if len(choices) == 0 {
		return Choice{}, fmt.Errorf("no choices provided")
	}
	if defaultIndex < 0 || defaultIndex >= len(choices) {
		defaultIndex = 0
	}

	if !IsInteractive(in, out) {
		return fallbackSelectOne(reader, out, prompt, choices, defaultIndex)
	}

	inFile := in.(fdGetter)
	fd := int(inFile.Fd())

	oldState, err := makeRaw(fd)
	if err != nil {
		return fallbackSelectOne(reader, out, prompt, choices, defaultIndex)
	}
	defer restoreTerminal(fd, oldState)

	current := defaultIndex
	buf := make([]byte, 3)

	// Helper to render the interactive list
	render := func(firstRender bool) {
		if !firstRender {
			// Clear choices lines + 1 footer line
			linesToClear := len(choices) + 1
			for i := 0; i < linesToClear; i++ {
				fmt.Fprintf(out, "\033[A\033[2K\r")
			}
		}

		for i, c := range choices {
			prefix := "    "
			labelColor := ansiWhite
			descColor := ansiDim
			if i == current {
				prefix = "  " + ansiCyan + ansiBold + "❯ " + ansiReset
				labelColor = ansiCyan + ansiBold
				descColor = ansiYellow
			}

			if c.Description != "" {
				fmt.Fprintf(out, "%s%s[%d] %-28s%s %s%s%s\n",
					prefix, labelColor, c.ID, c.Label, ansiReset, descColor, c.Description, ansiReset)
			} else {
				fmt.Fprintf(out, "%s%s[%d] %s%s\n", prefix, labelColor, c.ID, c.Label, ansiReset)
			}
		}

		fmt.Fprintf(out, "  %s(Use ↑/↓ or j/k to navigate, digits 1-%d, Enter to select)%s\n",
			ansiDim, len(choices), ansiReset)
	}

	// Print header prompt
	fmt.Fprintf(out, "%s?%s %s%s%s\n", ansiCyan+ansiBold, ansiReset, ansiBold, prompt, ansiReset)
	render(true)

	for {
		n, err := in.Read(buf)
		if err != nil || n == 0 {
			break
		}

		b := buf[0]
		if b == 3 { // Ctrl+C
			restoreTerminal(fd, oldState)
			fmt.Fprintln(out, "^C")
			os.Exit(130)
		}

		if b == '\r' || b == '\n' { // Enter
			break
		}

		if b == '\x1b' && n >= 3 && buf[1] == '[' {
			switch buf[2] {
			case 'A': // Up
				current = (current - 1 + len(choices)) % len(choices)
				render(false)
			case 'B': // Down
				current = (current + 1) % len(choices)
				render(false)
			}
			continue
		}

		// Vim keys
		if b == 'k' || b == 'K' {
			current = (current - 1 + len(choices)) % len(choices)
			render(false)
			continue
		}
		if b == 'j' || b == 'J' {
			current = (current + 1) % len(choices)
			render(false)
			continue
		}

		// Numeric selection
		if b >= '1' && b <= '9' {
			num := int(b - '1')
			if num < len(choices) {
				current = num
				break
			}
		}
	}

	// Clear choices and footer
	linesToClear := len(choices) + 1
	for i := 0; i < linesToClear; i++ {
		fmt.Fprintf(out, "\033[A\033[2K\r")
	}

	// Clear prompt line and rewrite with selected answer
	fmt.Fprintf(out, "\033[A\033[2K\r")
	fmt.Fprintf(out, "%s✓%s %s%s%s: %s%s%s\n",
		ansiGreen+ansiBold, ansiReset,
		ansiBold, prompt, ansiReset,
		ansiCyan+ansiBold, choices[current].Label, ansiReset)

	return choices[current], nil
}

// SelectMulti presents an interactive multi-select menu with checkboxes.
func SelectMulti(reader *bufio.Reader, in io.Reader, out io.Writer, prompt string, choices []Choice, defaultIndices []int) ([]Choice, error) {
	if len(choices) == 0 {
		return nil, fmt.Errorf("no choices provided")
	}

	if !IsInteractive(in, out) {
		return fallbackSelectMulti(reader, out, prompt, choices, defaultIndices)
	}

	inFile := in.(fdGetter)
	fd := int(inFile.Fd())

	oldState, err := makeRaw(fd)
	if err != nil {
		return fallbackSelectMulti(reader, out, prompt, choices, defaultIndices)
	}
	defer restoreTerminal(fd, oldState)

	selected := make(map[int]bool)
	for _, idx := range defaultIndices {
		if idx >= 0 && idx < len(choices) {
			selected[idx] = true
		}
	}

	current := 0
	buf := make([]byte, 3)

	render := func(firstRender bool) {
		if !firstRender {
			linesToClear := len(choices) + 1
			for i := 0; i < linesToClear; i++ {
				fmt.Fprintf(out, "\033[A\033[2K\r")
			}
		}

		for i, c := range choices {
			pointer := "  "
			labelColor := ansiWhite
			if i == current {
				pointer = ansiCyan + ansiBold + "❯ " + ansiReset
				labelColor = ansiCyan + ansiBold
			}

			box := ansiDim + "[ ]" + ansiReset
			if selected[i] {
				box = ansiGreen + ansiBold + "[✓]" + ansiReset
			}

			fmt.Fprintf(out, " %s%s %s%d. %-32s%s %s%s%s\n",
				pointer, box, labelColor, c.ID, c.Label, ansiReset,
				ansiDim, c.Description, ansiReset)
		}

		fmt.Fprintf(out, "  %s(↑/↓ to navigate, Space to toggle, 'a' for all, Enter to confirm)%s\n",
			ansiDim, ansiReset)
	}

	fmt.Fprintf(out, "%s?%s %s%s%s\n", ansiCyan+ansiBold, ansiReset, ansiBold, prompt, ansiReset)
	render(true)

	for {
		n, err := in.Read(buf)
		if err != nil || n == 0 {
			break
		}

		b := buf[0]
		if b == 3 { // Ctrl+C
			restoreTerminal(fd, oldState)
			fmt.Fprintln(out, "^C")
			os.Exit(130)
		}

		if b == '\r' || b == '\n' { // Enter confirms
			break
		}

		if b == ' ' { // Space toggles
			selected[current] = !selected[current]
			render(false)
			continue
		}

		if b == 'a' || b == 'A' { // Toggle all
			allSelected := true
			for i := range choices {
				if !selected[i] {
					allSelected = false
					break
				}
			}
			for i := range choices {
				selected[i] = !allSelected
			}
			render(false)
			continue
		}

		if b == '\x1b' && n >= 3 && buf[1] == '[' {
			switch buf[2] {
			case 'A': // Up
				current = (current - 1 + len(choices)) % len(choices)
				render(false)
			case 'B': // Down
				current = (current + 1) % len(choices)
				render(false)
			}
			continue
		}

		if b == 'k' || b == 'K' {
			current = (current - 1 + len(choices)) % len(choices)
			render(false)
			continue
		}
		if b == 'j' || b == 'J' {
			current = (current + 1) % len(choices)
			render(false)
			continue
		}

		// Direct digit toggle
		if b >= '1' && b <= '9' {
			idx := int(b - '1')
			if idx < len(choices) {
				selected[idx] = !selected[idx]
				current = idx
				render(false)
			}
		}
	}

	// Clean menu lines
	linesToClear := len(choices) + 1
	for i := 0; i < linesToClear; i++ {
		fmt.Fprintf(out, "\033[A\033[2K\r")
	}

	var results []Choice
	var names []string
	for i, c := range choices {
		if selected[i] {
			results = append(results, c)
			names = append(names, c.Label)
		}
	}

	// Overwrite prompt with clean summary
	fmt.Fprintf(out, "\033[A\033[2K\r")
	summary := strings.Join(names, ", ")
	if summary == "" {
		summary = "none"
	}
	fmt.Fprintf(out, "%s✓%s %s%s%s: %s%s%s\n",
		ansiGreen+ansiBold, ansiReset,
		ansiBold, prompt, ansiReset,
		ansiCyan+ansiBold, summary, ansiReset)

	return results, nil
}

func fallbackSelectOne(reader *bufio.Reader, out io.Writer, prompt string, choices []Choice, defaultIndex int) (Choice, error) {
	fmt.Fprintf(out, "\n%s?%s %s%s%s:\n", ansiCyan+ansiBold, ansiReset, ansiBold, prompt, ansiReset)
	for i, c := range choices {
		rec := ""
		if i == defaultIndex {
			rec = " (Recommended)"
		}
		if c.Description != "" {
			fmt.Fprintf(out, "  [%d] %s%s - %s\n", c.ID, c.Label, rec, c.Description)
		} else {
			fmt.Fprintf(out, "  [%d] %s%s\n", c.ID, c.Label, rec)
		}
	}

	defaultVal := strconv.Itoa(choices[defaultIndex].ID)
	fmt.Fprintf(out, "%s[%s]%s: ", ansiDim, defaultVal, ansiReset)

	line, err := reader.ReadString('\n')
	if err != nil {
		return choices[defaultIndex], nil
	}
	val := strings.TrimSpace(line)
	if val == "" {
		return choices[defaultIndex], nil
	}

	if num, err := strconv.Atoi(val); err == nil {
		for _, c := range choices {
			if c.ID == num {
				return c, nil
			}
		}
	}

	// Match by key or label prefix
	valLower := strings.ToLower(val)
	for _, c := range choices {
		if strings.ToLower(c.Key) == valLower || strings.HasPrefix(strings.ToLower(c.Label), valLower) {
			return c, nil
		}
	}

	return choices[defaultIndex], nil
}

func fallbackSelectMulti(reader *bufio.Reader, out io.Writer, prompt string, choices []Choice, defaultIndices []int) ([]Choice, error) {
	fmt.Fprintf(out, "\n%s?%s %s%s%s:\n", ansiCyan+ansiBold, ansiReset, ansiBold, prompt, ansiReset)
	for _, c := range choices {
		fmt.Fprintf(out, "  [%d] %-30s %s\n", c.ID, c.Label, c.Description)
	}

	var defIDs []string
	for _, idx := range defaultIndices {
		if idx >= 0 && idx < len(choices) {
			defIDs = append(defIDs, strconv.Itoa(choices[idx].ID))
		}
	}
	defaultVal := strings.Join(defIDs, " ")

	fmt.Fprintf(out, "%sChoose components [%s]%s: ", ansiDim, defaultVal, ansiReset)

	line, err := reader.ReadString('\n')
	if err != nil {
		var res []Choice
		for _, idx := range defaultIndices {
			res = append(res, choices[idx])
		}
		return res, nil
	}
	val := strings.TrimSpace(line)
	if val == "" {
		val = defaultVal
	}

	selectedIDs := ParseSelection(val, len(choices))
	idMap := make(map[int]bool)
	for _, id := range selectedIDs {
		idMap[id] = true
	}

	var results []Choice
	for _, c := range choices {
		if idMap[c.ID] {
			results = append(results, c)
		}
	}
	return results, nil
}
