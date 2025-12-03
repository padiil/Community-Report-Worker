package report

import (
	"fmt"

	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

type summaryCard struct {
	Label string
	Value string
}

func renderSummaryCards(m core.Maroto, cards []summaryCard) {
	if len(cards) == 0 {
		return
	}
	columns := len(cards)
	if columns > 4 {
		columns = 4
	}
	width := 12 / columns
	cols := make([]core.Col, 0, columns)
	for i := 0; i < columns; i++ {
		card := cards[i]
		cols = append(cols, text.NewCol(width, fmt.Sprintf("%s\n%s", card.Label, card.Value), props.Text{
			Align: align.Center,
			Top:   4,
			Size:  11,
			Style: fontstyle.Bold,
		}))
	}
	row := m.AddRow(25, cols...)
	row.WithStyle(&props.Cell{BackgroundColor: ColorBgLight})
	m.AddRow(4, text.NewCol(12, ""))

	if len(cards) > columns {
		more := fmt.Sprintf("+%d more metrics", len(cards)-columns)
		m.AddRow(6, text.NewCol(12, more, props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
	}
}
