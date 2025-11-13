package report

import (
	"org-worker/internal/config"

	"codeberg.org/go-pdf/fpdf"
)

func AddReportHeader(pdf *fpdf.Fpdf, title, subtitle string) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, title)
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 8, subtitle, "", "L", false)
	pdf.Ln(4)
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(6)
}

func AddSectionTitle(pdf *fpdf.Fpdf, title string) {
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, title)
	pdf.Ln(8)
}

func AddReportFooter(pdf *fpdf.Fpdf) {
	pdf.SetY(-20)
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(0, 10, config.GetOrgName())
}
