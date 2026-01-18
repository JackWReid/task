package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jackreid/task/internal/id"
)

func runNote(args []string) error {
	fs := flag.NewFlagSet("note", flag.ContinueOnError)
	fs.SetOutput(stderr)

	fs.Usage = func() {
		fmt.Fprintln(stderr, `Add a note to a task.

Usage:
  task note <id> <content>
  echo "content" | task note <id>

The note content can be provided as the second positional argument,
or via stdin for piping longer content.

Examples:
  task note abc "This is a note"
  echo "Multi-line note" | task note abc
  cat notes.txt | task note abc`)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		errorf("Error: task ID is required")
		fs.Usage()
		return fmt.Errorf("task ID is required")
	}

	taskID := fs.Arg(0)

	// Get note content from argument or stdin
	var content string
	if fs.NArg() >= 2 {
		// Content from positional argument
		content = fs.Arg(1)
	} else {
		// Try to read from stdin
		stdinContent, err := readStdin()
		if err != nil {
			errorf("Error reading stdin: %v", err)
			return err
		}
		content = stdinContent
	}

	content = strings.TrimSpace(content)
	if content == "" {
		errorf("Error: note content is required")
		fs.Usage()
		return fmt.Errorf("note content is required")
	}

	s := getStore()

	task, err := s.FindByID(taskID)
	if err != nil {
		errorf("Error: %v", err)
		return err
	}

	if task == nil {
		errorf("Error: task not found: %s", taskID)
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Generate note ID
	noteID, err := id.GenerateNoteID(taskID)
	if err != nil {
		errorf("Error generating note ID: %v", err)
		return err
	}

	task.AddNote(noteID, content)

	if err := s.Update(task); err != nil {
		errorf("Error: %v", err)
		return err
	}

	fmt.Fprintf(stdout, "Added note to task %s\n", taskID)
	return nil
}

// readStdin reads content from stdin if available
func readStdin() (string, error) {
	// Check if stdin has data (not a terminal)
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", nil // Can't stat, assume no stdin
	}

	// Check if stdin is a pipe or has data
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", nil // stdin is a terminal, no pipe data
	}

	// Read all stdin content
	reader := bufio.NewReader(stdin)
	var builder strings.Builder

	for {
		line, err := reader.ReadString('\n')
		builder.WriteString(line)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return builder.String(), nil
}
