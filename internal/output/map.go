package output

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerColor = lipgloss.NewStyle().Bold(true)
	keyColor    = lipgloss.NewStyle().Foreground(lipgloss.Color("#5d3fd3"))
)

type Block struct {
	Name   string
	Values map[string]any
}

func (b Block) Output() {
	keys := make([]string, 0, len(b.Values))
	for k := range b.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	maxKeyLength := 0
	for _, k := range keys {
		if l := len(k); l > maxKeyLength {
			maxKeyLength = l
		}
	}

	var (
		rows       = []string{}
		maxRowSize = 0
		padding    = 3
	)
	for _, key := range keys {
		value := b.Values[key]
		row := fmt.Sprintf(
			"%s %v",
			keyColor.Width(maxKeyLength+padding+1).Render(key),
			value,
		)
		rows = append(rows, row)
		width, _ := lipgloss.Size(row)
		if width > maxRowSize {
			maxRowSize = width
		}
	}

	fmt.Println(
		headerColor.Width(maxRowSize).
			Render(b.Name),
	)

	for _, row := range rows {
		fmt.Println(row)
	}
}
