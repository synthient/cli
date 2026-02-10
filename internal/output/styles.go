package output

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

var StdoutStyles = NewStyles(os.Stdout)

type Styles struct {
	renderer *lipgloss.Renderer

	Bold      lipgloss.Style
	BlockData lipgloss.Style

	SynthientColor lipgloss.Style
}

func NewStyles(out *os.File) Styles {
	renderer := lipgloss.NewRenderer(out)
	return Styles{
		renderer:       renderer,
		Bold:           renderer.NewStyle().Bold(true),
		BlockData:      renderer.NewStyle().PaddingLeft(3),
		SynthientColor: renderer.NewStyle().Foreground(lipgloss.Color("#5d3fd3")),
	}
}
