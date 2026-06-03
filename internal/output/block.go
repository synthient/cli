package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

type Block struct {
	Name   string
	Values []BlockValue
}

type BlockValue struct {
	Key   string
	Value any
}

func (b Block) Output(out *os.File, styles Styles, indentation int, final bool) {
	maxKeyLength := 0
	for _, v := range b.Values {
		if l := len(v.Key); l > maxKeyLength {
			maxKeyLength = l
		}
	}

	connector := "├"
	rowConnector := "│"
	if final {
		connector = "└"
		rowConnector = " "
	}
	padding := 2 + indentation

	if b.Name != "" {
		WriteLine(
			out,
			fmt.Sprintf(
				"%s  %s",
				styles.Frame.PaddingLeft(indentation).Render(connector),
				styles.Warm.Bold(true).Render(b.Name),
			),
		)
	}

	rows := []string{}
	for _, value := range b.Values {
		valueLines := formatStyledValueLines(styles, value.Value, blockValueWidth(out, maxKeyLength, padding, indentation))
		for i, line := range valueLines {
			key := strings.Repeat(" ", maxKeyLength+padding)
			if i == 0 {
				key = styles.Key.Width(maxKeyLength + padding).Render(value.Key)
			}
			rows = append(
				rows,
				fmt.Sprintf(
					"%s  %s %s",
					styles.Frame.PaddingLeft(indentation).Render(rowConnector),
					key,
					line,
				),
			)
		}
	}

	if len(rows) != 0 {
		WriteLine(out, strings.Join(rows, "\n"))
	}
}

func blockValueWidth(out *os.File, maxKeyLength int, padding int, indentation int) int {
	width, _, err := term.GetSize(out.Fd())
	if err != nil || width <= 0 {
		width = 100
	}
	prefixWidth := indentation + 1 + 2 + maxKeyLength + padding + 1
	return max(24, width-prefixWidth)
}

func formatStyledValueLines(styles Styles, value any, width int) []string {
	text, style := styledValue(styles, value)
	lines := wrapText(text, width)
	for i, line := range lines {
		lines[i] = style.Render(line)
	}
	return lines
}

func styledValue(styles Styles, value any) (string, lipgloss.Style) {
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "None", styles.Empty
		}
		return v, styles.Value
	case []string:
		if len(v) == 0 {
			return "None", styles.Empty
		}
		return strings.Join(v, ", "), styles.Secondary
	case bool:
		if v {
			return "true", styles.Success
		}
		return "false", styles.Muted
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(value), styles.Num
	case nil:
		return "None", styles.Empty
	default:
		return fmt.Sprint(value), styles.Value
	}
}

func wrapText(value string, width int) []string {
	if width <= 0 || lipgloss.Width(value) <= width {
		return []string{value}
	}

	words := strings.Fields(value)
	if len(words) == 0 {
		return []string{value}
	}

	lines := []string{}
	line := ""
	for _, word := range words {
		if lipgloss.Width(word) > width {
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
			lines = append(lines, splitLongWord(word, width)...)
			continue
		}
		if line == "" {
			line = word
			continue
		}
		candidate := line + " " + word
		if lipgloss.Width(candidate) <= width {
			line = candidate
			continue
		}
		lines = append(lines, line)
		line = word
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func splitLongWord(word string, width int) []string {
	lines := []string{}
	var line strings.Builder
	lineWidth := 0
	for _, r := range word {
		next := string(r)
		nextWidth := lipgloss.Width(next)
		if lineWidth+nextWidth > width && line.Len() > 0 {
			lines = append(lines, line.String())
			line.Reset()
			lineWidth = 0
		}
		line.WriteRune(r)
		lineWidth += nextWidth
	}
	if line.Len() > 0 {
		lines = append(lines, line.String())
	}
	return lines
}
