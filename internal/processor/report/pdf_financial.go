package report

import (
	"bytes"
	"fmt"
	"org-worker/internal/domain"

	"codeberg.org/go-pdf/fpdf"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func GenerateFinancialPDF(data domain.FinancialReportData) (*bytes.Buffer, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	p := message.NewPrinter(language.Indonesian)
	formatRp := func(val float64) string {
		return p.Sprintf("Rp %.0f", val)
	}

	pdf.AddPage()
	AddReportHeader(pdf, "Laporan Transparansi Keuangan",
		fmt.Sprintf("Periode: %s s/d %s", data.StartDate.Format("02 Jan 2006"), data.EndDate.Format("02 Jan 2006")))

	AddSectionTitle(pdf, "Ringkasan Keuangan")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(60, 8, "Total Pemasukan:")
	pdf.Cell(0, 8, formatRp(data.TotalIncome))
	pdf.Ln(6)
	if data.TotalInKindValue > 0 {
		pdf.Cell(60, 8, "Nilai Donasi In-Kind:")
		pdf.Cell(0, 8, formatRp(data.TotalInKindValue))
		pdf.Ln(6)
	}
	pdf.Cell(60, 8, "Total Pengeluaran:")
	pdf.Cell(0, 8, formatRp(data.TotalExpenses))
	pdf.Ln(6)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(60, 8, "Pemasukan Bersih (Net):")
	pdf.Cell(0, 8, formatRp(data.NetIncome))
	pdf.Ln(12)

	AddSectionTitle(pdf, "Alokasi Pengeluaran")
	pdf.SetFont("Arial", "", 12)
	if data.TotalExpenses > 0 {
		// --- Pie Chart: Alokasi Pengeluaran ---
		pieImg, err := createPieChartImage(convertFinancialStatToDemographic(data.ExpensesByCategory), int64(data.TotalExpenses))
		if err == nil {
			pdf.RegisterImageOptionsReader("pieExpense", fpdf.ImageOptions{ImageType: "PNG"}, pieImg)
			pdf.Image("pieExpense", 10, pdf.GetY(), 190, 0, false, "", 0, "")
			pdf.Ln(70)
		}
		for _, stat := range data.ExpensesByCategory {
			percentage := (stat.Total / data.TotalExpenses) * 100
			label := fmt.Sprintf("  - %s:", stat.ID)
			value := fmt.Sprintf("%s (%.1f%%)", formatRp(stat.Total), percentage)
			pdf.Cell(70, 8, label)
			pdf.Cell(0, 8, value)
			pdf.Ln(6)
		}
	} else {
		pdf.Cell(0, 8, "Tidak ada pengeluaran tercatat pada periode ini.")
	}
	pdf.Ln(10)

	AddSectionTitle(pdf, "5 Donasi Terbesar")
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(80, 7, "Sumber", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 7, "Tanggal", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 7, "Jumlah", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 10)
	for _, donation := range data.TopDonations {
		pdf.CellFormat(80, 7, donation.Source, "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 7, donation.Date.Format("02 Jan 2006"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(60, 7, formatRp(donation.Amount), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}
	AddReportFooter(pdf)
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}
