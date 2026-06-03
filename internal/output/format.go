package output

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func PadRight(value string, width int) string {
	gap := width - lipgloss.Width(value)
	if gap <= 0 {
		return value
	}
	return value + strings.Repeat(" ", gap)
}

func PadLeft(value string, width int) string {
	gap := width - lipgloss.Width(value)
	if gap <= 0 {
		return value
	}
	return strings.Repeat(" ", gap) + value
}
