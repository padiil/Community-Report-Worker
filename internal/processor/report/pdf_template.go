package report

import (
	"fmt"
	"org-worker/internal/config"

	"codeberg.org/go-pdf/fpdf"
)

type ReportTheme struct {
	FontFamily string
	Primary    struct{ R, G, B int }
	Accent     struct{ R, G, B int }
	LogoPath   string
}

var DefaultTheme = ReportTheme{
	FontFamily: "Arial",
	Primary:    struct{ R, G, B int }{33, 37, 41},
	Accent:     struct{ R, G, B int }{220, 230, 241},
	LogoPath:   "",
}

func NewReportPDF(title, subtitle string) *fpdf.Fpdf {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 25, 15)
	pdf.SetAutoPageBreak(true, 20)

	pdf.SetHeaderFuncMode(func() {
		if DefaultTheme.LogoPath != "" {
			pdf.ImageOptions(DefaultTheme.LogoPath, 15, 10, 25, 0, false, fpdf.ImageOptions{ImageType: ""}, 0, "")
		}
		pdf.SetY(12)
		pdf.SetFont(DefaultTheme.FontFamily, "B", 16)
		pdf.SetTextColor(DefaultTheme.Primary.R, DefaultTheme.Primary.G, DefaultTheme.Primary.B)
		pdf.Cell(0, 8, title)
		pdf.Ln(9)
		pdf.SetFont(DefaultTheme.FontFamily, "", 11)
		pdf.MultiCell(0, 6, subtitle, "", "L", false)
		pdf.Ln(2)
		pdf.SetDrawColor(200, 200, 200)
		pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
		pdf.Ln(4)
	}, true)

	pdf.SetFooterFunc(func() {
		pdf.SetY(-18)
		pdf.SetDrawColor(DefaultTheme.Accent.R, DefaultTheme.Accent.G, DefaultTheme.Accent.B)
		pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
		pdf.Ln(4)
		pdf.SetFont(DefaultTheme.FontFamily, "I", 10)
		pdf.CellFormat(0, 6, config.GetOrgName(), "", 0, "L", false, 0, "")
		pageInfo := fmt.Sprintf("%d/{nb}", pdf.PageNo())
		pdf.CellFormat(0, 6, pageInfo, "", 0, "R", false, 0, "")
	})

	pdf.AliasNbPages("{nb}")
	pdf.AddPage()
	pdf.SetFont(DefaultTheme.FontFamily, "", 12)
	pdf.SetTextColor(0, 0, 0)
	return pdf
}

func AddSectionTitle(pdf *fpdf.Fpdf, title string) {
	pdf.SetFillColor(DefaultTheme.Accent.R, DefaultTheme.Accent.G, DefaultTheme.Accent.B)
	pdf.SetFont(DefaultTheme.FontFamily, "B", 13)
	pdf.CellFormat(0, 9, title, "", 1, "L", true, 0, "")
	pdf.Ln(3)
}
