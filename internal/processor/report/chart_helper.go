package report

import (
	"bytes"
	"fmt"
	"org-worker/internal/domain"

	"github.com/wcharczuk/go-chart/v2"
)

func createPieChartImage(stats []domain.DemographicStat, total int64) ([]byte, error) {
	if total == 0 {
		return nil, nil
	}
	var values []chart.Value
	for _, stat := range stats {
		percentage := (float64(stat.Count) / float64(total)) * 100
		label := fmt.Sprintf("%s (%.1f%%)", stat.ID, percentage)
		values = append(values, chart.Value{Value: float64(stat.Count), Label: label})
	}
	pie := chart.PieChart{
		Width:  512,
		Height: 512,
		Values: values,
	}
	buf := new(bytes.Buffer)
	if err := pie.Render(chart.PNG, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func createBarChartImage(stats []domain.MilestoneStat, title string) ([]byte, error) {
	var values []chart.Value
	for _, stat := range stats {
		values = append(values, chart.Value{Value: float64(stat.Count), Label: stat.ID})
	}
	bar := chart.BarChart{
		Width:  512,
		Height: 512,
		Bars:   values,
	}
	if title != "" {
		bar.Title = title
	}
	buf := new(bytes.Buffer)
	if err := bar.Render(chart.PNG, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func convertFinancialStatToDemographic(stats []domain.FinancialStat) []domain.DemographicStat {
	var result []domain.DemographicStat
	for _, s := range stats {
		result = append(result, domain.DemographicStat{ID: s.ID, Count: int(s.Total)})
	}
	return result
}
