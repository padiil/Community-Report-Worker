package report

import (
	"bytes"
	"fmt"

	"org-worker/internal/domain"

	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func GenerateFinancialPDF(data domain.FinancialReportData) (*bytes.Buffer, error) {
	m := GetMarotoInstance(
		"Laporan Transparansi Keuangan",
		fmt.Sprintf("Periode: %s s.d. %s",
			data.StartDate.Format("02 Jan 2006"),
			data.EndDate.Format("02 Jan 2006")),
	)
	p := message.NewPrinter(language.Indonesian)
	formatRp := func(val float64) string {
		return p.Sprintf("Rp %.0f", val)
	}

	addSectionTitle(m, "Ringkasan Keuangan")
	renderSummaryCards(m, []summaryCard{
		{Label: "Total Pemasukan", Value: formatRp(data.TotalIncome)},
		{Label: "Total Pengeluaran", Value: formatRp(data.TotalExpenses)},
		{Label: "Saldo Bersih", Value: formatRp(data.NetIncome)},
	})
	if data.TotalInKindValue > 0 {
		m.AddRow(8, text.NewCol(12, fmt.Sprintf("Donasi barang tercatat: %s", formatRp(data.TotalInKindValue)), props.Text{Size: 10, Color: ColorTextMain}))
	}

	renderExpenseAllocation(m, data, formatRp)
	renderDonationTable(m, data.TopDonations, formatRp)

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return marotoDocumentBuffer(document), nil
}

func renderExpenseAllocation(m core.Maroto, data domain.FinancialReportData, formatRp func(float64) string) {
	addSectionTitle(m, "Alokasi Pengeluaran")
	if data.TotalExpenses <= 0 {
		m.AddRow(8, text.NewCol(12, "Tidak ada pengeluaran yang tercatat pada periode ini.", props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
		return
	}
	if chartBytes, err := createPieChartImage(convertFinancialStatToDemographic(data.ExpensesByCategory), int64(data.TotalExpenses)); err == nil && chartBytes != nil {
		m.AddRow(70, image.NewFromBytesCol(12, chartBytes, "png", props.Rect{Percent: 80, Center: true}))
	}
	for _, stat := range data.ExpensesByCategory {
		percentage := 0.0
		if data.TotalExpenses > 0 {
			percentage = (stat.Total / data.TotalExpenses) * 100
		}
		lineText := fmt.Sprintf("- %s: %s (%.1f%%)", stat.ID, formatRp(stat.Total), percentage)
		m.AddRow(6, text.NewCol(12, lineText, props.Text{Size: 10}))
	}
}

func renderDonationTable(m core.Maroto, donations []domain.TopDonation, formatRp func(float64) string) {
	addSectionTitle(m, "5 Donasi Tunai Teratas")
	header := []core.Col{
		text.NewCol(6, "Sumber", props.Text{Align: align.Center, Style: fontstyle.Bold}),
		text.NewCol(3, "Tanggal", props.Text{Align: align.Center, Style: fontstyle.Bold}),
		text.NewCol(3, "Jumlah", props.Text{Align: align.Center, Style: fontstyle.Bold}),
	}
	row := m.AddRow(8, header...)
	row.WithStyle(&props.Cell{BackgroundColor: ColorBgLight})
	if len(donations) == 0 {
		m.AddRow(6, text.NewCol(12, "Tidak ada donasi tunai yang tercatat.", props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
		return
	}
	for _, donation := range donations {
		m.AddRow(8,
			text.NewCol(6, donation.Source, props.Text{Size: 10}),
			text.NewCol(3, donation.Date.Format("02 Jan 2006"), props.Text{Size: 10, Align: align.Center}),
			text.NewCol(3, formatRp(donation.Amount), props.Text{Size: 10, Align: align.Right}),
		)
		m.AddRow(1, line.NewCol(12))
	}
}
