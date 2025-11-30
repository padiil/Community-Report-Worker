package report

import (
	"bytes"
	"fmt"
	"org-worker/internal/domain"

	"codeberg.org/go-pdf/fpdf"
)

func GenerateDemographicsPDF(data domain.ParticipantDemographicsData) (*bytes.Buffer, error) {
	pdf := NewReportPDF(
		"Participant Demographics Report",
		fmt.Sprintf("Community: %s\nTotal Participants: %d", data.CommunityName, data.TotalParticipants),
	)
	pageWidth, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	usableWidth := pageWidth - left - right

	pdf.Ln(5)
	AddSectionTitle(pdf, "Report Snapshot")
	renderSummaryCards(pdf, []summaryCard{
		{Label: "Total Participants", Value: fmt.Sprintf("%d People", data.TotalParticipants)},
		{Label: "Statuses Tracked", Value: fmt.Sprintf("%d Segments", len(data.ByStatus))},
		{Label: "Locations Tracked", Value: fmt.Sprintf("%d Regions", len(data.ByLocation))},
	}, left, usableWidth)

	addStatSection(pdf, "By Employment Status", data.ByStatus, data.TotalParticipants)
	addStatSection(pdf, "By Age Group", data.ByAge, data.TotalParticipants)
	addStatSection(pdf, "By Location (Top 10)", data.ByLocation, data.TotalParticipants)

	// --- Pie Chart: Status Pekerjaan ---
	if data.TotalParticipants > 0 {
		pieImg, err := createPieChartImage(data.ByStatus, data.TotalParticipants)
		if err == nil {
			pdf.RegisterImageOptionsReader("pieStatus", fpdf.ImageOptions{ImageType: "PNG"}, pieImg)
			pdf.Image("pieStatus", 10, pdf.GetY(), 90, 0, false, "", 0, "")
		}
		pdf.Ln(95)
		// Pie Chart: Kategori Usia
		pieImg, err = createPieChartImage(data.ByAge, data.TotalParticipants)
		if err == nil {
			pdf.RegisterImageOptionsReader("pieAge", fpdf.ImageOptions{ImageType: "PNG"}, pieImg)
			pdf.Image("pieAge", 110, pdf.GetY()-95, 90, 0, false, "", 0, "")
		}
		pdf.Ln(100)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

func addStatSection(pdf *fpdf.Fpdf, title string, stats []domain.DemographicStat, total int64) {
	AddSectionTitle(pdf, title)
	pdf.SetFont("Arial", "", 12)
	for _, stat := range stats {
		label := stat.ID
		if label == "" {
			label = "Not Specified"
		}
		percentage := float64(stat.Count) / float64(total) * 100
		pdf.Cell(0, 8, fmt.Sprintf("  - %s: %d (%.1f%%)", label, stat.Count, percentage))
		pdf.Ln(6)
	}
	pdf.Ln(6)
}
