package report

import (
	"bytes"
	"fmt"
	"org-worker/internal/domain"

	"codeberg.org/go-pdf/fpdf"
)

func GenerateDemographicsPDF(data domain.ParticipantDemographicsData) (*bytes.Buffer, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	AddReportHeader(pdf, "Laporan Demografi Peserta",
		fmt.Sprintf("Komunitas: %s\nTotal Peserta: %d", data.CommunityName, data.TotalParticipants))

	addStatSection(pdf, "Berdasarkan Status Pekerjaan", data.ByStatus, data.TotalParticipants)
	addStatSection(pdf, "Berdasarkan Kategori Usia", data.ByAge, data.TotalParticipants)
	addStatSection(pdf, "Berdasarkan Domisili (Top 10)", data.ByLocation, data.TotalParticipants)

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
	AddReportFooter(pdf)
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
			label = "Tidak Diisi"
		}
		percentage := float64(stat.Count) / float64(total) * 100
		pdf.Cell(0, 8, fmt.Sprintf("  - %s: %d (%.1f%%)", label, stat.Count, percentage))
		pdf.Ln(6)
	}
	pdf.Ln(6)
}
