package handlers

import (
	"encoding/json"
	"kasir-api/repositories"
	"net/http"
	"time"
)

type ReportHandler struct {
	transactionRepo *repositories.TransactionRepository
}

func NewReportHandler(transactionRepo *repositories.TransactionRepository) *ReportHandler {
	return &ReportHandler{
		transactionRepo: transactionRepo,
	}
}

// GetDailySalesReport - GET /api/report/hari-ini
func (h *ReportHandler) GetDailySalesReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	report, err := h.transactionRepo.GetDailySalesReport()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

// GetSalesReportByDate - GET /api/report?start_date=2026-01-01&end_date=2026-02-01
func (h *ReportHandler) GetSalesReportByDate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Validasi parameter
	if startDate == "" || endDate == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "start_date dan end_date harus disediakan (format: YYYY-MM-DD)",
		})
		return
	}

	// Validasi format tanggal
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "start_date format tidak valid (harus YYYY-MM-DD)",
		})
		return
	}

	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "end_date format tidak valid (harus YYYY-MM-DD)",
		})
		return
	}

	// Validasi range tanggal
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	if start.After(end) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "start_date tidak boleh lebih besar dari end_date",
		})
		return
	}

	report, err := h.transactionRepo.GetSalesReportByDateRange(startDate, endDate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}
