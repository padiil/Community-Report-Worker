package report

import (
	"bytes"
	"fmt"

	"org-worker/internal/domain"

	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func GenerateCommunityActivityPDF(data domain.CommunityActivityData) (*bytes.Buffer, error) {
	m := GetMarotoInstance(
		"Laporan Aktivitas Komunitas",
		fmt.Sprintf("Komunitas: %s | Periode: %s - %s",
			data.CommunityName,
			data.StartDate.Format("02 Jan 2006"),
			data.EndDate.Format("02 Jan 2006")),
	)

	addSectionTitle(m, "Ringkasan Kinerja")
	renderSummaryCards(m, []summaryCard{
		{Label: "Anggota Baru", Value: fmt.Sprintf("%d Orang", data.NewMemberCount)},
		{Label: "Anggota Aktif", Value: fmt.Sprintf("%d Orang", data.ActiveMemberCount)},
		{Label: "Total Kegiatan", Value: fmt.Sprintf("%d Kegiatan", data.EventsHeldCount)},
	})

	addSectionTitle(m, "Detail Kegiatan & Dokumentasi")
	if len(data.EventDetails) == 0 {
		m.AddRow(8, text.NewCol(12, "Tidak ada kegiatan yang tercatat pada periode ini.", props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
	} else {
		for i, event := range data.EventDetails {
			renderCommunityEvent(m, i, event)
		}
	}

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return marotoDocumentBuffer(document), nil
}

func renderCommunityEvent(m core.Maroto, idx int, event domain.EventDetail) {
	m.AddRow(8, text.NewCol(12, fmt.Sprintf("%d. %s", idx+1, event.Name), props.Text{
		Style: fontstyle.Bold,
		Size:  12,
		Color: ColorTextMain,
	}))
	meta := fmt.Sprintf("Tanggal: %s   |   Fasilitator: %s   |   Peserta: %d",
		event.Date.Format("02 Jan 2006"), event.TutorName, event.ParticipantCount)
	m.AddRow(6, text.NewCol(12, meta, props.Text{Size: 9, Color: ColorTextMute}))

	if len(event.DocumentationURLs) == 0 {
		m.AddRow(6, text.NewCol(12, "(Tidak ada dokumentasi foto)", props.Text{Style: fontstyle.Italic, Color: ColorTextMute}))
	} else {
		cols := make([]core.Col, 0, 4)
		max := len(event.DocumentationURLs)
		if max > 4 {
			max = 4
		}
		for j := 0; j < max; j++ {
			imgBytes, err := downloadImageAsJPG(event.DocumentationURLs[j])
			if err != nil {
				continue
			}
			cols = append(cols, image.NewFromBytesCol(3, imgBytes, "jpg", props.Rect{Percent: 95, Center: true}))
		}
		if len(cols) > 0 {
			m.AddRow(40, cols...)
		}
	}
	m.AddRow(4, line.NewCol(12))
	m.AddRow(3, text.NewCol(12, ""))
}
