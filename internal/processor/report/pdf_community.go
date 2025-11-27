package report

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"org-worker/internal/domain"

	"codeberg.org/go-pdf/fpdf"
	_ "github.com/chai2010/webp"
)

var (
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

func downloadImageAsJPG(url string) (*bytes.Buffer, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}
	return buf, nil
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
	cardGap := 4.0
	cardWidth := (usableWidth - (cardGap * 2)) / 3
	cardHeight := 22.0
	startY := pdf.GetY()
	drawSummaryCard(pdf, "New Members", fmt.Sprintf("%d People", data.NewMemberCount), left, startY, cardWidth, cardHeight)
	drawSummaryCard(pdf, "Active Members", fmt.Sprintf("%d People", data.ActiveMemberCount), left+cardWidth+cardGap, startY, cardWidth, cardHeight)
	drawSummaryCard(pdf, "Total Events", fmt.Sprintf("%d Events", data.EventsHeldCount), left+(cardWidth+cardGap)*2, startY, cardWidth, cardHeight)
	pdf.SetY(startY + cardHeight + 12)

	AddSectionTitle(pdf, "Event Details & Documentation")
	for i, event := range data.EventDetails {
		if pdf.GetY() > 230 {
			pdf.AddPage()
		}

		r, g, b := hexToRGB(ColorBgLight)
		pdf.SetFillColor(r, g, b)
		pdf.SetDrawColor(r, g, b)
		pdf.SetFont("Arial", "B", 11)
		r, g, b = hexToRGB(ColorTextMain)
		pdf.SetTextColor(r, g, b)
		eventTitle := fmt.Sprintf("  %d. %s", i+1, event.Name)
		pdf.CellFormat(usableWidth, 9, eventTitle, "0", 1, "L", true, 0, "")

		pdf.SetFont("Arial", "", 10)
		r, g, b = hexToRGB(ColorTextMute)
		pdf.SetTextColor(r, g, b)
		meta := fmt.Sprintf("Date: %s   |   Tutor: %s   |   Participants: %d",
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
			imgBuf, err := downloadImageAsJPG(imgURL)
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
