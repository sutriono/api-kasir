package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
	"time"
)

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// GetTodaysSalesSummary menampilkan summary penjualan hari ini
func (repo *ReportRepository) GetTodaysSalesSummary() (*models.SalesSummary, error) {
	today := time.Now().Format("2006-01-02")

	summary := &models.SalesSummary{}

	// Query total revenue dan transaksi hari ini
	err := repo.db.QueryRow(`
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COUNT(id) as total_transaksi
		FROM transactions
		WHERE DATE(created_at) = $1
	`, today).Scan(&summary.TotalRevenue, &summary.TotalTransaksi)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Query produk terlaris hari ini
	err = repo.db.QueryRow(`
		SELECT 
			p.name,
			SUM(td.quantity) as qty_terjual
		FROM transaction_details td
		JOIN products p ON td.product_id = p.id
		JOIN transactions t ON td.transaction_id = t.id
		WHERE DATE(t.created_at) = $1
		GROUP BY p.id, p.name
		ORDER BY qty_terjual DESC
		LIMIT 1
	`, today).Scan(&summary.ProdukTerlaris.Nama, &summary.ProdukTerlaris.QtyTerjual)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return summary, nil
}

// GetSalesReportByDateRange menampilkan report penjualan berdasarkan tanggal range
func (repo *ReportRepository) GetSalesReportByDateRange(startDate, endDate string) (*models.SalesReport, error) {
	// Validasi format tanggal
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return nil, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
	}

	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		return nil, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD")
	}

	report := &models.SalesReport{
		DateRange: models.DateRange{
			StartDate: startDate,
			EndDate:   endDate,
		},
	}

	// Query total revenue dan transaksi dalam range
	err := repo.db.QueryRow(`
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COUNT(id) as total_transaksi
		FROM transactions
		WHERE DATE(created_at) BETWEEN $1 AND $2
	`, startDate, endDate).Scan(&report.TotalRevenue, &report.TotalTransaksi)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Query produk terlaris dalam range
	err = repo.db.QueryRow(`
		SELECT 
			p.name,
			SUM(td.quantity) as qty_terjual
		FROM transaction_details td
		JOIN products p ON td.product_id = p.id
		JOIN transactions t ON td.transaction_id = t.id
		WHERE DATE(t.created_at) BETWEEN $1 AND $2
		GROUP BY p.id, p.name
		ORDER BY qty_terjual DESC
		LIMIT 1
	`, startDate, endDate).Scan(&report.ProdukTerlaris.Nama, &report.ProdukTerlaris.QtyTerjual)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Query daily breakdown (optional)
	rows, err := repo.db.Query(`
		SELECT 
			DATE(t.created_at) as date,
			COALESCE(SUM(t.total_amount), 0) as revenue,
			COUNT(DISTINCT t.id) as transaksi
		FROM transactions t
		WHERE DATE(t.created_at) BETWEEN $1 AND $2
		GROUP BY DATE(t.created_at)
		ORDER BY DATE(t.created_at) ASC
	`, startDate, endDate)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report.DailyBreakdown = make([]models.DailySalesData, 0)

	for rows.Next() {
		var daily models.DailySalesData
		err := rows.Scan(&daily.Date, &daily.Revenue, &daily.Transaksi)
		if err != nil {
			return nil, err
		}

		// Get top product untuk hari tersebut
		err = repo.db.QueryRow(`
			SELECT 
				p.name,
				SUM(td.quantity) as qty_terjual
			FROM transaction_details td
			JOIN products p ON td.product_id = p.id
			JOIN transactions t ON td.transaction_id = t.id
			WHERE DATE(t.created_at) = $1
			GROUP BY p.id, p.name
			ORDER BY qty_terjual DESC
			LIMIT 1
		`, daily.Date).Scan(&daily.TopProduct.Nama, &daily.TopProduct.QtyTerjual)

		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		report.DailyBreakdown = append(report.DailyBreakdown, daily)
	}

	return report, nil
}
