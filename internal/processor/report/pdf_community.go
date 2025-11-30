package report

import (
	"bytes"
	"fmt"
	"org-worker/internal/domain"

	"codeberg.org/go-pdf/fpdf"
	_ "github.com/chai2010/webp"
)

func GenerateCommunityActivityPDF(data domain.CommunityActivityData) (*bytes.Buffer, error) {
	pdf := NewReportPDF(
		"Community Activity Report",
		fmt.Sprintf("Community: %s  |  Period: %s - %s",
			data.CommunityName,
			data.StartDate.Format("02 Jan 2006"),
			data.EndDate.Format("02 Jan 2006")),
	)

	pageWidth, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	usableWidth := pageWidth - left - right

	pdf.Ln(5)
	AddSectionTitle(pdf, "Performance Summary")
	renderSummaryCards(pdf, []summaryCard{
		{Label: "New Members", Value: fmt.Sprintf("%d People", data.NewMemberCount)},
		{Label: "Active Members", Value: fmt.Sprintf("%d People", data.ActiveMemberCount)},
		{Label: "Total Events", Value: fmt.Sprintf("%d Events", data.EventsHeldCount)},
	}, left, usableWidth)

	AddSectionTitle(pdf, "Event Details & Documentation")
	for i, event := range data.EventDetails {
		ensureVerticalSpace(pdf, 100)
		eventTitle := fmt.Sprintf("  %d. %s", i+1, event.Name)
		drawCardHeader(pdf, eventTitle, usableWidth)

		pdf.SetFont("Arial", "", 10)
		r, g, b := hexToRGB(ColorTextMute)
		pdf.SetTextColor(r, g, b)
		meta := fmt.Sprintf("Date: %s   |   Facilitator: %s   |   Participants: %d",
			event.Date.Format("Monday, 02 Jan 2006"), event.TutorName, event.ParticipantCount)
		pdf.CellFormat(usableWidth, 7, "      "+meta, "B", 1, "L", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
		pdf.Ln(4)

		if len(event.DocumentationURLs) == 0 {
			pdf.SetFont("Arial", "I", 9)
			r, g, b = hexToRGB(ColorTextMute)
			pdf.SetTextColor(r, g, b)
			pdf.CellFormat(0, 6, "      (No photo documentation)", "", 1, "L", false, 0, "")
			pdf.SetTextColor(0, 0, 0)
			pdf.Ln(4)
			continue
		}

		maxImages := 4
		gap := 3.0
		imgWidth := (usableWidth - gap*float64(maxImages-1)) / float64(maxImages)
		imgHeight := imgWidth * 0.65
		xStart := left
		yStart := pdf.GetY()
		if yStart+imgHeight > 250 {
			pdf.AddPage()
			yStart = pdf.GetY()
		}

		count := len(event.DocumentationURLs)
		if count > maxImages {
			count = maxImages
		}
		for j := 0; j < count; j++ {
			currentX := xStart + float64(j)*(imgWidth+gap)
			imgURL := event.DocumentationURLs[j]
			imgBuf, _, _, err := downloadImageAsJPG(imgURL)
			if err == nil {
				imgName := fmt.Sprintf("evt_%d_%d", i, j)
				pdf.RegisterImageOptionsReader(imgName, fpdf.ImageOptions{ImageType: "JPG"}, imgBuf)
				pdf.SetDrawColor(220, 220, 220)
				pdf.Rect(currentX, yStart, imgWidth, imgHeight, "D")
				pdf.Image(imgName, currentX+0.5, yStart+0.5, imgWidth-1, imgHeight-1, false, "", 0, "")
			} else {
				pdf.SetFillColor(240, 240, 240)
				pdf.Rect(currentX, yStart, imgWidth, imgHeight, "F")
				pdf.SetFont("Arial", "I", 7)
				pdf.SetTextColor(150, 150, 150)
				pdf.SetXY(currentX, yStart+(imgHeight/2)-2)
				pdf.CellFormat(imgWidth, 4, "Img Err", "", 0, "C", false, 0, "")
				pdf.SetTextColor(0, 0, 0)
			}
		}
		pdf.SetY(yStart + imgHeight + 8)

		if len(event.DocumentationURLs) > maxImages {
			pdf.SetFont("Arial", "I", 8)
			r, g, b = hexToRGB(ColorTextMute)
			pdf.SetTextColor(r, g, b)
			pdf.CellFormat(0, 5, fmt.Sprintf("      (+%d more documentation photos...)", len(event.DocumentationURLs)-maxImages), "0", 1, "L", false, 0, "")
			pdf.SetTextColor(0, 0, 0)
			pdf.Ln(2)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}
