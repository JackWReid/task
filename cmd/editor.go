package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jackreid/task/internal/model"
)

type frontmatter struct {
	Title     string
	Type      string
	Status    string
	Labels    []string
	HasTitle  bool
	HasType   bool
	HasStatus bool
	HasLabels bool
}

func renderTaskTemplate(title string, taskType model.TaskType, status model.Status, labels []string, description string) string {
	var builder strings.Builder
	builder.WriteString("---\n")
	builder.WriteString("title: ")
	builder.WriteString(formatYAMLString(title))
	builder.WriteString("\n")
	builder.WriteString("type: ")
	builder.WriteString(taskType.String())
	builder.WriteString("\n")
	builder.WriteString("status: ")
	builder.WriteString(status.String())
	builder.WriteString("\n")
	if len(labels) == 0 {
		builder.WriteString("labels: []\n")
	} else {
		builder.WriteString("labels:\n")
		for _, label := range labels {
			builder.WriteString("  - ")
			builder.WriteString(formatYAMLString(label))
			builder.WriteString("\n")
		}
	}
	builder.WriteString("---\n")
	if description != "" {
		builder.WriteString(description)
		if !strings.HasSuffix(description, "\n") {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

func openEditorWithTemplate(template string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		return "", errors.New("EDITOR is not set")
	}

	tempFile, err := os.CreateTemp("", "task-edit-*.md")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.WriteString(template); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("writing temp file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("closing temp file: %w", err)
	}

	cmd := editorCommand(editor, tempPath)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("running editor: %w", err)
	}

	edited, err := os.ReadFile(tempPath)
	if err != nil {
		return "", fmt.Errorf("reading temp file: %w", err)
	}
	return string(edited), nil
}

func editorCommand(editor, path string) *exec.Cmd {
	return exec.Command("sh", "-c", fmt.Sprintf("%s %s", editor, shellEscape(path)))
}

func shellEscape(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func parseFrontmatter(content string) (frontmatter, string, error) {
	var fm frontmatter
	lines := strings.Split(content, "\n")

	start := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			start = i
			break
		}
		if trimmed != "" {
			return fm, "", errors.New("frontmatter must start with ---")
		}
	}
	if start == -1 {
		return fm, "", errors.New("frontmatter must start with ---")
	}

	end := -1
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return fm, "", errors.New("frontmatter must end with ---")
	}

	frontLines := lines[start+1 : end]
	body := strings.Join(lines[end+1:], "\n")

	parsed, err := parseFrontmatterLines(frontLines)
	if err != nil {
		return fm, "", err
	}
	return parsed, body, nil
}

func parseFrontmatterLines(lines []string) (frontmatter, error) {
	fm := frontmatter{}
	for i := 0; i < len(lines); {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			i++
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return fm, fmt.Errorf("invalid frontmatter line: %s", line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "title":
			fm.Title = unquoteIfQuoted(value)
			fm.HasTitle = true
			i++
		case "type":
			fm.Type = unquoteIfQuoted(value)
			fm.HasType = true
			i++
		case "status":
			fm.Status = unquoteIfQuoted(value)
			fm.HasStatus = true
			i++
		case "labels":
			fm.HasLabels = true
			labels, nextIndex, err := parseLabels(value, lines, i+1)
			if err != nil {
				return fm, err
			}
			fm.Labels = labels
			i = nextIndex
		default:
			i++
		}
	}

	return fm, nil
}

func parseLabels(value string, lines []string, start int) ([]string, int, error) {
	if value == "" {
		var labels []string
		i := start
		for i < len(lines) {
			trimmed := strings.TrimSpace(lines[i])
			if trimmed == "" {
				i++
				continue
			}
			if strings.HasPrefix(trimmed, "-") {
				label := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
				label = unquoteIfQuoted(label)
				if label != "" {
					labels = append(labels, label)
				}
				i++
				continue
			}
			break
		}
		return labels, i, nil
	}

	if value == "[]" {
		return []string{}, start, nil
	}

	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
		if inner == "" {
			return []string{}, start, nil
		}
		parts := strings.Split(inner, ",")
		labels := make([]string, 0, len(parts))
		for _, part := range parts {
			label := strings.TrimSpace(part)
			label = unquoteIfQuoted(label)
			if label != "" {
				labels = append(labels, label)
			}
		}
		return labels, start, nil
	}

	return []string{unquoteIfQuoted(value)}, start, nil
}

func unquoteIfQuoted(value string) string {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return value
	}
	if value[0] == '"' && value[len(value)-1] == '"' {
		unquoted, err := strconv.Unquote(value)
		if err == nil {
			return unquoted
		}
		return value[1 : len(value)-1]
	}
	if value[0] == '\'' && value[len(value)-1] == '\'' {
		inner := value[1 : len(value)-1]
		return strings.ReplaceAll(inner, "''", "'")
	}
	return value
}

func formatYAMLString(value string) string {
	if value == "" {
		return `""`
	}
	if strings.ContainsAny(value, ":\n") || strings.HasPrefix(value, " ") || strings.HasSuffix(value, " ") {
		return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return value
}

func normalizeLabels(labels []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(labels))
	for _, label := range labels {
		trimmed := strings.TrimSpace(label)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		result = append(result, trimmed)
	}
	return result
}

func normalizeDescription(body string) *string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return nil
	}
	clean := strings.TrimRight(body, "\n")
	return &clean
}
