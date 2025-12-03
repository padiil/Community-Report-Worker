package report

import (
	"bytes"
	"fmt"
	"strings"

	"org-worker/internal/domain"

	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func GenerateImpactPDF(data domain.ProgramImpactData) (*bytes.Buffer, error) {
	m := GetMarotoInstance(
		"Laporan Dampak Program",
		fmt.Sprintf("Community: %s | Period: %s - %s",
			data.CommunityName,
			data.StartDate.Format("02 Jan 2006"),
			data.EndDate.Format("02 Jan 2006")),
	)

	statMap := map[string]int{}
	for _, stat := range data.Stats {
		statMap[stat.ID] = stat.Count
	}
	addSectionTitle(m, "Ringkasan Kinerja")
	renderSummaryCards(m, []summaryCard{
		{Label: "Proyek Diajukan", Value: fmt.Sprintf("%d Proyek", statMap["project_submitted"])},
		{Label: "Level Up", Value: fmt.Sprintf("%d Anggota", statMap["level_up"])},
		{Label: "Penempatan Kerja", Value: fmt.Sprintf("%d Penempatan", statMap["job_placement"])},
	})

	addSectionTitle(m, "Sorotan Dampak & Dokumentasi")
	if len(data.Highlights) == 0 {
		m.AddRow(8, text.NewCol(12, "Tidak ada sorotan dampak yang tercatat pada periode ini.", props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
	} else {
		for idx, highlight := range data.Highlights {
			renderImpactHighlight(m, idx, highlight)
		}
	}

	addSectionTitle(m, "Grafik Distribusi Pencapaian")
	if len(data.Stats) == 0 {
		m.AddRow(8, text.NewCol(12, "Tidak ada data milestone untuk divisualisasikan.", props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
	} else if chartBytes, err := createBarChartImage(data.Stats, ""); err == nil && chartBytes != nil {
		m.AddRow(70, image.NewFromBytesCol(12, chartBytes, "png", props.Rect{Percent: 90, Center: true}))
	}

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return marotoDocumentBuffer(document), nil
}

func renderImpactHighlight(m core.Maroto, idx int, highlight domain.ImpactHighlight) {
	m.AddRow(8, text.NewCol(12, fmt.Sprintf("%d. %s", idx+1, highlight.Title), props.Text{
		Style: fontstyle.Bold,
		Size:  12,
		Color: ColorTextMain,
	}))
	m.AddRow(5, text.NewCol(12, fmt.Sprintf("Penanggung Jawab: %s", highlight.OwnerName), props.Text{Size: 9, Color: ColorTextMute}))

	if summary := strings.TrimSpace(highlight.Summary); summary != "" {
		m.AddRow(10, text.NewCol(12, summary, props.Text{Size: 10, Align: align.Left}))
	} else {
		m.AddRow(6, text.NewCol(12, "(Tidak ada deskripsi sorotan)", props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
	}

	if len(highlight.DocumentationURLs) == 0 {
		m.AddRow(6, text.NewCol(12, "(Tidak ada foto dokumentasi)", props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
	} else {
		cols := make([]core.Col, 0, 4)
		max := len(highlight.DocumentationURLs)
		if max > 4 {
			max = 4
		}
		for i := 0; i < max; i++ {
			imgBytes, err := downloadImageAsJPG(highlight.DocumentationURLs[i])
			if err != nil {
				continue
			}
			cols = append(cols, image.NewFromBytesCol(3, imgBytes, "jpg", props.Rect{Percent: 95, Center: true}))
		}
		if len(cols) > 0 {
			m.AddRow(40, cols...)
		}
		if len(highlight.DocumentationURLs) > max {
			m.AddRow(6, text.NewCol(12, fmt.Sprintf("(+%d foto dokumentasi lainnya)", len(highlight.DocumentationURLs)-max), props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
		}
	}

	m.AddRow(4, line.NewCol(12))
	m.AddRow(4, text.NewCol(12, ""))
}
