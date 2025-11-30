package report

import (
	"bytes"
	"fmt"
	"strings"

	"org-worker/internal/domain"

	"codeberg.org/go-pdf/fpdf"
)

func GenerateImpactPDF(data domain.ProgramImpactData) (*bytes.Buffer, error) {
	pdf := NewReportPDF(
		"Program Impact Report",
		fmt.Sprintf("Community: %s\nPeriod: %s to %s",
			data.CommunityName,
			data.StartDate.Format("02 Jan 2006"),
			data.EndDate.Format("02 Jan 2006")),
	)

	pageWidth, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	usableWidth := pageWidth - left - right

	AddSectionTitle(pdf, "Achievement Statistics")
	pdf.SetFont("Arial", "", 12)
	if len(data.Stats) == 0 {
		pdf.Cell(0, 8, "No achievements recorded for this period.")
	} else {
		cards := make([]summaryCard, 0, len(data.Stats))
		for _, stat := range data.Stats {
			cards = append(cards, summaryCard{
				Label: translateMilestoneType(stat.ID),
				Value: fmt.Sprintf("%d Participants", stat.Count),
			})
		}
		renderSummaryCards(pdf, cards, left, usableWidth)
		for _, stat := range data.Stats {
			label := translateMilestoneType(stat.ID)
			pdf.Cell(80, 8, fmt.Sprintf("  - %s:", label))
			pdf.SetFont("Arial", "B", 12)
			pdf.Cell(0, 8, fmt.Sprintf("%d", stat.Count))
			pdf.SetFont("Arial", "", 12)
			pdf.Ln(7)
		}
		pdf.Ln(5)
		barImg, err := createBarChartImage(data.Stats, "Achievement Distribution")
		if err == nil {
			pdf.RegisterImageOptionsReader("barImpact", fpdf.ImageOptions{ImageType: "PNG"}, barImg)
			const chartHeight = 80.0
			xCenter := left + (usableWidth-chartHeight)/2
			pdf.Image("barImpact", xCenter, pdf.GetY(), 0, chartHeight, false, "", 0, "")
			pdf.Ln(chartHeight + 10)
		}
	}

	if len(data.Highlights) > 0 {
		ensureVerticalSpace(pdf, 70)
		AddSectionTitle(pdf, "Participant Project Highlights")
		for idx, highlight := range data.Highlights {
			renderImpactHighlight(pdf, idx, highlight, left, usableWidth)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

func translateMilestoneType(milestoneType string) string {
	switch milestoneType {
	case "project_submitted":
		return "Final Projects Submitted"
	case "level_up":
		return "Participants Leveling Up"
	case "job_placement":
		return "Participants Hired"
	default:
		return milestoneType
	}
}

func renderImpactHighlight(pdf *fpdf.Fpdf, idx int, highlight domain.ImpactHighlight, left float64, usableWidth float64) {
	ensureVerticalSpace(pdf, 60)
	title := fmt.Sprintf("  %d. %s", idx+1, highlight.Title)
	drawCardHeader(pdf, title, usableWidth)
	pdf.SetFont("Arial", "I", 10)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 6, fmt.Sprintf("    Submitted by: %s", highlight.OwnerName))
	pdf.Ln(8)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 10)
	if summary := strings.TrimSpace(highlight.Summary); summary != "" {
		pdf.MultiCell(0, 5, fmt.Sprintf("    %s", summary), "", "L", false)
		pdf.Ln(2)
	}

	if len(highlight.DocumentationURLs) == 0 {
		pdf.SetFont("Arial", "I", 9)
		pdf.Cell(0, 6, "    (No visual documentation yet)")
		pdf.Ln(8)
		return
	}
	thumbnailWidth := 100.0
	drawHighlightThumbnail(pdf, highlight.DocumentationURLs[0], left+10, thumbnailWidth, idx)
	if extra := len(highlight.DocumentationURLs) - 1; extra > 0 {
		pdf.SetFont("Arial", "I", 8)
		pdf.Cell(0, 5, fmt.Sprintf("    (+%d additional documentation items)", extra))
		pdf.Ln(7)
	}
	pdf.Ln(4)
}

func drawHighlightThumbnail(pdf *fpdf.Fpdf, url string, x float64, width float64, idx int) {
	imgBuf, imgW, imgH, err := downloadImageAsJPG(url)
	if err != nil || imgW == 0 {
		pdf.SetFont("Arial", "I", 9)
		pdf.Cell(0, 6, "    (Image could not be loaded)")
		pdf.Ln(8)
		return
	}
	imgName := fmt.Sprintf("highlight_img_%d_%d", pdf.PageNo(), idx)
	pdf.RegisterImageOptionsReader(imgName, fpdf.ImageOptions{ImageType: "JPG"}, imgBuf)
	ratio := imgH / imgW
	imgHeight := width * ratio
	currentY := pdf.GetY()
	pdf.Image(imgName, x, currentY, width, 0, false, "", 0, "")
	pdf.SetY(currentY + imgHeight + 5)
}
