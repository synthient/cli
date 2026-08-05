package output

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
	"github.com/synthient/cli/internal/options"
)

var StdoutStyles = NewStyles(os.Stdout)

type Styles struct {
	Bold      lipgloss.Style
	Muted     lipgloss.Style
	Subtle    lipgloss.Style
	Accent    lipgloss.Style
	Secondary lipgloss.Style
	Warm      lipgloss.Style

	Frame lipgloss.Style
	Title lipgloss.Style
	Key   lipgloss.Style
	Value lipgloss.Style
	Empty lipgloss.Style
	Bool  lipgloss.Style
	Num   lipgloss.Style

	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style

	BlockData lipgloss.Style

	SynthientColor lipgloss.Style
}

type palette struct {
	accent     color.Color
	secondary  color.Color
	warm       color.Color
	muted      color.Color
	subtle     color.Color
	foreground color.Color
	danger     color.Color
}

func newPalette(isDark bool, profile colorprofile.Profile) palette {
	lightDark := lipgloss.LightDark(isDark)
	pick := func(light, dark string) color.Color {
		return profile.Convert(lightDark(lipgloss.Color(light), lipgloss.Color(dark)))
	}
	return palette{
		accent:     pick("#7C3AED", "#A277FF"),
		secondary:  pick("#047857", "#61FFCA"),
		warm:       pick("#B45309", "#FFCA85"),
		muted:      pick("#6B7280", "#9A94A8"),
		subtle:     pick("#9CA3AF", "#4D4D4D"),
		foreground: pick("#111827", "#EDECEE"),
		danger:     profile.Convert(lipgloss.Color("#FF6767")),
	}
}

func detectProfile(out *os.File) colorprofile.Profile {
	if options.NoColor {
		return colorprofile.NoTTY
	}
	return colorprofile.Detect(out, os.Environ())
}

func hasDarkBackground(out *os.File) bool {
	if !term.IsTerminal(os.Stdin.Fd()) || !term.IsTerminal(out.Fd()) {
		return true
	}
	return lipgloss.HasDarkBackground(os.Stdin, out)
}

func NewStyles(out *os.File) Styles {
	p := newPalette(hasDarkBackground(out), detectProfile(out))
	return Styles{
		Bold:           lipgloss.NewStyle().Bold(true).Foreground(p.foreground),
		Muted:          lipgloss.NewStyle().Foreground(p.muted),
		Subtle:         lipgloss.NewStyle().Foreground(p.subtle),
		Accent:         lipgloss.NewStyle().Foreground(p.accent),
		Secondary:      lipgloss.NewStyle().Foreground(p.secondary),
		Warm:           lipgloss.NewStyle().Foreground(p.warm),
		Frame:          lipgloss.NewStyle().Foreground(p.accent),
		Title:          lipgloss.NewStyle().Bold(true).Foreground(p.accent),
		Key:            lipgloss.NewStyle().Foreground(p.secondary),
		Value:          lipgloss.NewStyle().Foreground(p.foreground),
		Empty:          lipgloss.NewStyle().Foreground(p.muted).Italic(true),
		Bool:           lipgloss.NewStyle().Foreground(p.warm),
		Num:            lipgloss.NewStyle().Foreground(p.warm),
		Success:        lipgloss.NewStyle().Foreground(p.secondary),
		Warning:        lipgloss.NewStyle().Foreground(p.warm),
		Error:          lipgloss.NewStyle().Foreground(p.danger),
		BlockData:      lipgloss.NewStyle().PaddingLeft(3),
		SynthientColor: lipgloss.NewStyle().Foreground(p.accent),
	}
}
