package report

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"org-worker/internal/domain"
	"org-worker/internal/repository"
	"org-worker/internal/storage"
)

type ReportHandler struct {
	repo    *repository.ReportRepository
	storage storage.StorageProvider
}

func NewReportHandler(repo *repository.ReportRepository, storage storage.StorageProvider) *ReportHandler {
	return &ReportHandler{repo: repo, storage: storage}
}

func (h *ReportHandler) HandleReportGeneration(ctx context.Context, logger *slog.Logger, reportDoc domain.ReportDoc) error {
	logger.Info("Mulai memproses laporan")
	pdfBuffer, err := h.generatePDF(ctx, reportDoc)
	if err != nil {
		logger.Error("Gagal membuat buffer PDF", "err", err)
		_ = h.repo.UpdateReportStatus(ctx, reportDoc.ID, "failed", "", err.Error())
		return err
	}
	filename := fmt.Sprintf("%s-%s.pdf", reportDoc.Type, reportDoc.ID)
	fileURL, err := h.storage.Save(ctx, reportDoc.Type, filename, pdfBuffer)
	if err != nil {
		logger.Error("Gagal menyimpan file ke storage", "err", err)
		_ = h.repo.UpdateReportStatus(ctx, reportDoc.ID, "failed", "", err.Error())
		return err
	}
	if err := h.repo.UpdateReportStatus(ctx, reportDoc.ID, "completed", fileURL, ""); err != nil {
		logger.Error("Gagal memperbarui status laporan", "err", err)
		return err
	}
	logger.Info("Laporan berhasil dibuat dan disimpan", "fileURL", fileURL)
	return nil
}

func (h *ReportHandler) generatePDF(ctx context.Context, reportDoc domain.ReportDoc) (*bytes.Buffer, error) {
	switch reportDoc.Type {
	case "community_activity":
		data, err := h.repo.GetCommunityActivityData(ctx, reportDoc.Filters)
		if err != nil {
			return nil, err
		}
		return GenerateCommunityActivityPDF(data)
	case "participant_demographics":
		data, err := h.repo.GetParticipantDemographicsData(ctx, reportDoc.Filters)
		if err != nil {
			return nil, err
		}
		return GenerateDemographicsPDF(data)
	case "program_impact":
		data, err := h.repo.GetProgramImpactData(ctx, reportDoc.Filters)
		if err != nil {
			return nil, err
		}
		return GenerateImpactPDF(data)
	case "financial_summary":
		data, err := h.repo.GetFinancialSummaryData(ctx, reportDoc.Filters)
		if err != nil {
			return nil, err
		}
		return GenerateFinancialPDF(data)
	}
	return nil, fmt.Errorf("tipe laporan tidak dikenal: %s", reportDoc.Type)
}
