package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Block struct {
	Name   string
	Values []BlockValue
}

type BlockValue struct {
	Key   string
	Value any
}

func (b Block) Output(out *os.File, styles Styles, indentation int) {
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
			styles.SynthientColor.PaddingLeft(indentation).
				Width(maxKeyLength+padding).
				Render(value.Key),
			value.Value,
		)
		rows = append(rows, row)
		width, _ := lipgloss.Size(row)
		if width > maxRowSize {
			maxRowSize = width
		}
	}

	if b.Name != "" {
		WriteLine(
			out,
			styles.Bold.PaddingLeft(indentation).Width(maxRowSize+6).
				Render(b.Name),
		)
	}

	WriteLine(out, styles.BlockData.Render(strings.Join(rows, "\n")))
}
