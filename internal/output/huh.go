package output

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var HuhTheme = newHuhTheme()

func newHuhTheme() *huh.Theme {
	theme := huh.ThemeBase()
	accent := lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A277FF"}
	secondary := lipgloss.AdaptiveColor{Light: "#047857", Dark: "#61FFCA"}
	warm := lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#FFCA85"}
	muted := lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9A94A8"}

	theme.Focused.Base = lipgloss.NewStyle().PaddingLeft(1).BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).BorderForeground(accent)
	theme.Focused.Title = lipgloss.NewStyle().Foreground(accent).Bold(true)
	theme.Focused.Description = lipgloss.NewStyle().Foreground(muted)
	theme.Focused.ErrorIndicator = lipgloss.NewStyle().SetString("▲ ").Foreground(warm)
	theme.Focused.ErrorMessage = lipgloss.NewStyle().Foreground(warm)
	theme.Focused.SelectSelector = lipgloss.NewStyle().SetString("◆ ").Foreground(accent)
	theme.Focused.MultiSelectSelector = lipgloss.NewStyle().SetString("◆ ").Foreground(accent)
	theme.Focused.SelectedPrefix = lipgloss.NewStyle().SetString("◼ ").Foreground(secondary)
	theme.Focused.UnselectedPrefix = lipgloss.NewStyle().SetString("◻ ").Foreground(muted)
	theme.Focused.FocusedButton = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("#15141B")).Background(secondary)
	theme.Focused.BlurredButton = lipgloss.NewStyle().Padding(0, 2).Foreground(muted)
	theme.Focused.TextInput.Cursor = lipgloss.NewStyle().Foreground(accent)
	theme.Focused.TextInput.Placeholder = lipgloss.NewStyle().Foreground(muted)
	theme.Focused.TextInput.Prompt = lipgloss.NewStyle().Foreground(secondary)

	theme.Blurred = theme.Focused
	theme.Blurred.Base = lipgloss.NewStyle().PaddingLeft(1).BorderStyle(lipgloss.HiddenBorder()).BorderLeft(true)
	theme.Blurred.SelectSelector = lipgloss.NewStyle().SetString("  ")
	theme.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("  ")
	theme.Group.Title = theme.Focused.Title
	theme.Group.Description = theme.Focused.Description

	return theme
}
