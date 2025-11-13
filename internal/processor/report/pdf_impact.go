package report

import (
	"bytes"
	"fmt"
	"org-worker/internal/domain"

	"codeberg.org/go-pdf/fpdf"
)

func GenerateImpactPDF(data domain.ProgramImpactData) (*bytes.Buffer, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	AddReportHeader(pdf, "Laporan Dampak Program",
		fmt.Sprintf("Komunitas: %s\nPeriode: %s s/d %s", data.CommunityName, data.StartDate.Format("02 Jan 2006"), data.EndDate.Format("02 Jan 2006")))

	AddSectionTitle(pdf, "Pencapaian Peserta")
	pdf.SetFont("Arial", "", 12)
	if len(data.Stats) == 0 {
		pdf.Cell(0, 8, "Belum ada pencapaian yang tercatat pada periode ini.")
	} else {
		for _, stat := range data.Stats {
			label := translateMilestoneType(stat.ID)
			pdf.Cell(80, 8, fmt.Sprintf("  - %s:", label))
			pdf.Cell(0, 8, fmt.Sprintf("%d", stat.Count))
			pdf.Ln(7)
		}
		// --- Bar Chart: Pencapaian Peserta ---
		barImg, err := createBarChartImage(data.Stats)
		if err == nil {
			pdf.RegisterImageOptionsReader("barImpact", fpdf.ImageOptions{ImageType: "PNG"}, barImg)
			pdf.Image("barImpact", 10, pdf.GetY(), 190, 0, false, "", 0, "")
			pdf.Ln(70)
		}
	}
	AddReportFooter(pdf)
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

func translateMilestoneType(milestoneType string) string {
	switch milestoneType {
	case "project_submitted":
		return "Proyek Akhir Disubmit"
	case "level_up":
		return "Peserta Naik Level"
	case "job_placement":
		return "Peserta Mendapat Pekerjaan"
	default:
		return milestoneType
	}
}
