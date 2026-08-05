package output

import (
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

var HuhTheme = huh.ThemeFunc(newHuhStyles)

func newHuhStyles(isDark bool) *huh.Styles {
	styles := huh.ThemeBase(isDark)
	p := newPalette(isDark, detectProfile(os.Stdout))

	styles.Focused.Base = lipgloss.NewStyle().PaddingLeft(1).BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).BorderForeground(p.accent)
	styles.Focused.Title = lipgloss.NewStyle().Foreground(p.accent).Bold(true)
	styles.Focused.Description = lipgloss.NewStyle().Foreground(p.muted)
	styles.Focused.ErrorIndicator = lipgloss.NewStyle().SetString("▲ ").Foreground(p.warm)
	styles.Focused.ErrorMessage = lipgloss.NewStyle().Foreground(p.warm)
	styles.Focused.SelectSelector = lipgloss.NewStyle().SetString("◆ ").Foreground(p.accent)
	styles.Focused.MultiSelectSelector = lipgloss.NewStyle().SetString("◆ ").Foreground(p.accent)
	styles.Focused.SelectedPrefix = lipgloss.NewStyle().SetString("◼ ").Foreground(p.secondary)
	styles.Focused.UnselectedPrefix = lipgloss.NewStyle().SetString("◻ ").Foreground(p.muted)
	styles.Focused.FocusedButton = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("#15141B")).Background(p.secondary)
	styles.Focused.BlurredButton = lipgloss.NewStyle().Padding(0, 2).Foreground(p.muted)
	styles.Focused.TextInput.Cursor = lipgloss.NewStyle().Foreground(p.accent)
	styles.Focused.TextInput.Placeholder = lipgloss.NewStyle().Foreground(p.muted)
	styles.Focused.TextInput.Prompt = lipgloss.NewStyle().Foreground(p.secondary)

	styles.Blurred = styles.Focused
	styles.Blurred.Base = lipgloss.NewStyle().PaddingLeft(1).BorderStyle(lipgloss.HiddenBorder()).BorderLeft(true)
	styles.Blurred.SelectSelector = lipgloss.NewStyle().SetString("  ")
	styles.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("  ")
	styles.Group.Title = styles.Focused.Title
	styles.Group.Description = styles.Focused.Description

	return styles
}
