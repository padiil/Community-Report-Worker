package report

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

var (
	ColorPrimary   = &props.Color{Red: 30, Green: 58, Blue: 138}
	ColorSecondary = &props.Color{Red: 59, Green: 130, Blue: 246}
	ColorBgLight   = &props.Color{Red: 243, Green: 244, Blue: 246}
	ColorTextMain  = &props.Color{Red: 31, Green: 41, Blue: 55}
	ColorTextMute  = &props.Color{Red: 107, Green: 114, Blue: 128}
)

// GetMarotoInstance configures a Maroto PDF with consistent header/footer branding.
func GetMarotoInstance(title, subtitle string) core.Maroto {
	cfg := config.NewBuilder().
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(18).
		Build()

	m := maroto.New(cfg)

	headerRows := []core.Row{
		text.NewRow(16, title, props.Text{
			Style: fontstyle.Bold,
			Size:  16,
			Color: ColorPrimary,
		}),
		text.NewRow(10, subtitle, props.Text{
			Size:  10,
			Color: ColorTextMute,
		}),
		line.NewRow(4),
	}
	if err := m.RegisterHeader(headerRows...); err != nil {
		panic(fmt.Sprintf("failed to register header: %v", err))
	}

	footerRows := []core.Row{
		line.NewRow(4),
		text.NewRow(8, fmt.Sprintf("Generated %s", time.Now().Format("02 Jan 2006")), props.Text{
			Size:  8,
			Style: fontstyle.Italic,
			Color: ColorTextMute,
			Align: align.Right,
		}),
	}
	if err := m.RegisterFooter(footerRows...); err != nil {
		panic(fmt.Sprintf("failed to register footer: %v", err))
	}

	return m
}

// addSectionTitle draws a consistent section heading row across reports.
func addSectionTitle(m core.Maroto, title string) {
	row := m.AddRow(10, text.NewCol(12, title, props.Text{
		Style: fontstyle.Bold,
		Size:  12,
		Color: ColorSecondary,
	}))
	row.WithStyle(&props.Cell{BackgroundColor: ColorBgLight})
	m.AddRow(4, text.NewCol(12, ""))
}

func downloadImageAsJPG(url string) ([]byte, error) {
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
	return buf.Bytes(), nil
}

func marotoDocumentBuffer(doc core.Document) *bytes.Buffer {
	return bytes.NewBuffer(doc.GetBytes())
}
