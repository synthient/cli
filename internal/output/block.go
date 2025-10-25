package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true)
	dataStyle   = lipgloss.NewStyle().PaddingLeft(3)
)

type Block struct {
	Name   string
	Values []BlockValue
}

type BlockValue struct {
	Key   string
	Value any
}

func (b Block) Output(indentation int) {
	maxKeyLength := 0
	for _, v := range b.Values {
		if l := len(v.Key); l > maxKeyLength {
			maxKeyLength = l
		}
	}

	var (
		rows       = []string{}
		maxRowSize = 0
		padding    = 1 + indentation
	)
	for _, value := range b.Values {
		row := fmt.Sprintf(
			"%s %v",
			SYNTHIENT_COLOR.PaddingLeft(indentation).Width(maxKeyLength+padding).Render(value.Key),
			value.Value,
		)
		rows = append(rows, row)
		width, _ := lipgloss.Size(row)
		if width > maxRowSize {
			maxRowSize = width
		}
	}

	if b.Name != "" {
		fmt.Println(
			headerStyle.PaddingLeft(indentation).Width(maxRowSize + 6).
				Render(b.Name),
		)
	}

	fmt.Println(dataStyle.Render(strings.Join(rows, "\n")))
}
