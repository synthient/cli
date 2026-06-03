package output

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/synthient/cli/internal/options"
)

var StdoutStyles = NewStyles(os.Stdout)

type Styles struct {
	renderer *lipgloss.Renderer

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

func NewStyles(out *os.File) Styles {
	renderer := lipgloss.NewRenderer(out)
	if options.NoColor {
		renderer.SetColorProfile(termenv.Ascii)
	}
	accent := lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A277FF"}
	secondary := lipgloss.AdaptiveColor{Light: "#047857", Dark: "#61FFCA"}
	warm := lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#FFCA85"}
	muted := lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9A94A8"}
	subtle := lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#4D4D4D"}
	foreground := lipgloss.AdaptiveColor{Light: "#111827", Dark: "#EDECEE"}
	return Styles{
		renderer:       renderer,
		Bold:           renderer.NewStyle().Bold(true).Foreground(foreground),
		Muted:          renderer.NewStyle().Foreground(muted),
		Subtle:         renderer.NewStyle().Foreground(subtle),
		Accent:         renderer.NewStyle().Foreground(accent),
		Secondary:      renderer.NewStyle().Foreground(secondary),
		Warm:           renderer.NewStyle().Foreground(warm),
		Frame:          renderer.NewStyle().Foreground(accent),
		Title:          renderer.NewStyle().Bold(true).Foreground(accent),
		Key:            renderer.NewStyle().Foreground(secondary),
		Value:          renderer.NewStyle().Foreground(foreground),
		Empty:          renderer.NewStyle().Foreground(muted).Italic(true),
		Bool:           renderer.NewStyle().Foreground(warm),
		Num:            renderer.NewStyle().Foreground(warm),
		Success:        renderer.NewStyle().Foreground(secondary),
		Warning:        renderer.NewStyle().Foreground(warm),
		Error:          renderer.NewStyle().Foreground(lipgloss.Color("#FF6767")),
		BlockData:      renderer.NewStyle().PaddingLeft(3),
		SynthientColor: renderer.NewStyle().Foreground(accent),
	}
}
