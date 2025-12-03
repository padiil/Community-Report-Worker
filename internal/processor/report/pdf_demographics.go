package report

import (
	"bytes"
	"fmt"

	"org-worker/internal/domain"

	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func GenerateDemographicsPDF(data domain.ParticipantDemographicsData) (*bytes.Buffer, error) {
	m := GetMarotoInstance(
		"Laporan Demografi Peserta",
		fmt.Sprintf("Komunitas: %s | Total Peserta: %d", data.CommunityName, data.TotalParticipants),
	)

	addSectionTitle(m, "Ringkasan Laporan")
	renderSummaryCards(m, []summaryCard{
		{Label: "Total Peserta", Value: fmt.Sprintf("%d Orang", data.TotalParticipants)},
		{Label: "Status yang Dipantau", Value: fmt.Sprintf("%d Segmen", len(data.ByStatus))},
		{Label: "Lokasi yang Dipantau", Value: fmt.Sprintf("%d Wilayah", len(data.ByLocation))},
	})

	renderDemographicSection(m, "Berdasarkan Status Pekerjaan", data.ByStatus, data.TotalParticipants)
	renderDemographicSection(m, "Berdasarkan Kelompok Usia", data.ByAge, data.TotalParticipants)
	renderDemographicSection(m, "Berdasarkan Lokasi (Top 10)", data.ByLocation, data.TotalParticipants)

	addSectionTitle(m, "Grafik Distribusi")
	if data.TotalParticipants <= 0 {
		m.AddRow(8, text.NewCol(12, "Tidak dapat menampilkan grafik tanpa total peserta.", props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
	} else {
		var chartCols []core.Col
		if statusChart, err := createPieChartImage(data.ByStatus, data.TotalParticipants); err == nil && statusChart != nil {
			chartCols = append(chartCols, image.NewFromBytesCol(6, statusChart, "png", props.Rect{Percent: 90, Center: true}))
		}
		if ageChart, err := createPieChartImage(data.ByAge, data.TotalParticipants); err == nil && ageChart != nil {
			chartCols = append(chartCols, image.NewFromBytesCol(6, ageChart, "png", props.Rect{Percent: 90, Center: true}))
		}
		if len(chartCols) == 0 {
			m.AddRow(8, text.NewCol(12, "Grafik tidak dapat dibuat.", props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
		} else {
			m.AddRow(80, chartCols...)
		}
	}

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return marotoDocumentBuffer(document), nil
}

func renderDemographicSection(m core.Maroto, title string, stats []domain.DemographicStat, total int64) {
	addSectionTitle(m, title)
	if len(stats) == 0 || total == 0 {
		m.AddRow(6, text.NewCol(12, "Tidak ada data untuk kategori ini.", props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
		return
	}
	for _, stat := range stats {
		label := stat.ID
		if label == "" {
			label = "Tidak Ditentukan"
		}
		percentage := (float64(stat.Count) / float64(total)) * 100
		rowText := fmt.Sprintf("- %s: %d (%.1f%%)", label, stat.Count, percentage)
		m.AddRow(6, text.NewCol(12, rowText, props.Text{Size: 10}))
	}
	m.AddRow(4, text.NewCol(12, ""))
}
