// Package parser provides utilities for parsing shell alias definition files.
// It extracts alias definitions from shell scripts while ignoring other code.
package parser

import (
	"bufio"
	"os"
	"strings"
)

// AliasDef represents a single shell alias definition.
type AliasDef struct {
	// Name is the alias name (e.g., "ll" in "alias ll='ls -la'").
	Name string
	// Command is the aliased command (e.g., "ls -la").
	Command string
	// Source is the file path where this alias was defined.
	Source string
}

// ParseAliases extracts alias definitions from a shell script file.
// It strictly parses "alias name='command'" or 'alias name="command"' lines,
// ignoring comments, functions, and other shell constructs.
func ParseAliases(filePath string) ([]AliasDef, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var aliases []AliasDef
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 1. Filter: Must start with "alias "
		if !strings.HasPrefix(line, "alias ") {
			continue
		}

		// 2. Split by first "="
		// alias foo='bar baz'
		// left: "alias foo", right: "'bar baz'"
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		// 3. Extract Name
		// "alias foo" -> "foo"
		namePart := strings.TrimSpace(parts[0])
		name := strings.TrimPrefix(namePart, "alias ")
		name = strings.TrimSpace(name) // "foo"

		// 4. Extract Value
		value := strings.TrimSpace(parts[1])

		// 5. Unquote (Basic)
		// We want the raw command to re-quote it safely later.
		if len(value) >= 2 {
			first := value[0]
			last := value[len(value)-1]
			if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if name != "" && value != "" {
			aliases = append(aliases, AliasDef{
				Name:    name,
				Command: value,
				Source:  filePath,
			})
		}
	}

	return aliases, scanner.Err()
}
