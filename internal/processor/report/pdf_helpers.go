package report

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"net/http"

	"codeberg.org/go-pdf/fpdf"
)

const (
	ColorPrimary   = "#1e3a8a"
	ColorSecondary = "#3b82f6"
	ColorBgLight   = "#f3f4f6"
	ColorTextMain  = "#1f2937"
	ColorTextMute  = "#6b7280"
)

func hexToRGB(hex string) (int, int, int) {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

func downloadImageAsJPG(url string) (*bytes.Buffer, float64, float64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, 0, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, 0, 0, err
	}

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, 0, 0, err
	}

	bounds := img.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	return buf, width, height, nil
}

func drawSummaryCard(pdf *fpdf.Fpdf, label, value string, x, y, w, h float64) {
	r, g, b := hexToRGB(ColorBgLight)
	pdf.SetFillColor(r, g, b)
	pdf.SetDrawColor(230, 230, 230)
	pdf.Rect(x, y, w, h, "FD")

	r, g, b = hexToRGB(ColorSecondary)
	pdf.SetFillColor(r, g, b)
	pdf.Rect(x, y, 1.5, h, "F")

	pdf.SetFont("Arial", "", 9)
	r, g, b = hexToRGB(ColorTextMute)
	pdf.SetTextColor(r, g, b)
	pdf.SetXY(x+5, y+5)
	pdf.Cell(w-10, 5, label)

	pdf.SetFont("Arial", "B", 14)
	r, g, b = hexToRGB(ColorPrimary)
	pdf.SetTextColor(r, g, b)
	pdf.SetXY(x+5, y+11)
	pdf.Cell(w-10, 8, value)
	pdf.SetTextColor(0, 0, 0)
}

type summaryCard struct {
	Label string
	Value string
}

func renderSummaryCards(pdf *fpdf.Fpdf, cards []summaryCard, left float64, usableWidth float64) {
	if len(cards) == 0 {
		return
	}
	maxCards := len(cards)
	if maxCards > 3 {
		maxCards = 3
	}
	gap := 4.0
	cardWidth := (usableWidth - gap*float64(maxCards-1)) / float64(maxCards)
	cardHeight := 22.0
	startY := pdf.GetY()
	for i := 0; i < maxCards; i++ {
		card := cards[i]
		x := left + float64(i)*(cardWidth+gap)
		drawSummaryCard(pdf, card.Label, card.Value, x, startY, cardWidth, cardHeight)
	}
	pdf.SetY(startY + cardHeight + 12)

	if len(cards) > maxCards {
		pdf.SetFont("Arial", "I", 9)
		pdf.SetTextColor(120, 120, 120)
		pdf.Cell(0, 6, fmt.Sprintf("+%d more metrics", len(cards)-maxCards))
		pdf.SetTextColor(0, 0, 0)
		pdf.Ln(6)
	}
}

func drawCardHeader(pdf *fpdf.Fpdf, text string, width float64) {
	r, g, b := hexToRGB(ColorBgLight)
	pdf.SetFillColor(r, g, b)
	pdf.SetDrawColor(r, g, b)
	pdf.SetFont("Arial", "B", 11)
	r, g, b = hexToRGB(ColorTextMain)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(width, 9, text, "0", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

func ensureVerticalSpace(pdf *fpdf.Fpdf, required float64) {
	_, pageHeight := pdf.GetPageSize()
	_, _, _, bottom := pdf.GetMargins()
	currentY := pdf.GetY()
	maxY := pageHeight - bottom
	if currentY+required > maxY {
		pdf.AddPage()
	}
}
