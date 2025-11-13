package report

import (
	"bytes"
	"fmt"
	"org-worker/internal/domain"

	"codeberg.org/go-pdf/fpdf"
)

func GenerateCommunityActivityPDF(data domain.CommunityActivityData) (*bytes.Buffer, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	AddReportHeader(pdf, "Laporan Aktivitas Komunitas",
		fmt.Sprintf("Komunitas: %s\nPeriode: %s s/d %s", data.CommunityName, data.StartDate.Format("02 Jan 2006"), data.EndDate.Format("02 Jan 2006")))

	AddSectionTitle(pdf, "Ringkasan Utama")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(60, 8, "Anggota Baru:")
	pdf.Cell(0, 8, fmt.Sprintf("%d orang", data.NewMemberCount))
	pdf.Ln(6)
	pdf.Cell(60, 8, "Anggota Aktif (hadir min. 1x):")
	pdf.Cell(0, 8, fmt.Sprintf("%d orang", data.ActiveMemberCount))
	pdf.Ln(6)
	pdf.Cell(60, 8, "Total Event Diadakan:")
	pdf.Cell(0, 8, fmt.Sprintf("%d acara", data.EventsHeldCount))
	pdf.Ln(12)

	AddSectionTitle(pdf, "Detail Event")
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(80, 7, "Nama Event", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Tanggal", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Tutor", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Peserta", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 10)
	for _, event := range data.EventDetails {
		pdf.CellFormat(80, 7, event.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 7, event.Date.Format("02 Jan 2006"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, event.TutorName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("%d", event.ParticipantCount), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}
	AddReportFooter(pdf)
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}
