package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
	"time"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	var (
		res *models.Transaction
	)

	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// inisialisasi subtotal -> jumlah total transaksi keseluruhan
	totalAmount := 0
	// inisialisasi modeling transactionDetails -> nanti kita insert ke db
	details := make([]models.TransactionDetail, 0, len(items))
	// loop setiap item
	for _, item := range items {
		var productName string
		var productID, price, stock int
		// get product dapet pricing
		err = tx.QueryRow("SELECT id, name, price, stock FROM products WHERE id=$1", item.ProductID).Scan(&productID, &productName, &price, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}

		if err != nil {
			return nil, err
		}

		// validasi stok cukup
		if item.Quantity > stock {
			return nil, fmt.Errorf("insufficient stock for product '%s': requested %d, available %d",
				productName, item.Quantity, stock)
		}

		// hitung current total = quantity * pricing
		// ditambahin ke dalam subtotal
		subtotal := item.Quantity * price
		totalAmount += subtotal

		// kurangi jumlah stok
		_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Quantity, productID)
		if err != nil {
			return nil, err
		}

		// item nya dimasukkin ke transactionDetails
		details = append(details, models.TransactionDetail{
			ProductID:   productID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	// insert transaction
	var transactionID int
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id", totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	// insert transaction details
	for i := range details {
		details[i].TransactionID = transactionID
	}

	// Prepare statement sekali untuk optimasi
	stmt, err := tx.Prepare(
		"INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2, $3, $4)",
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Execute untuk setiap detail
	for _, detail := range details {
		_, err := stmt.Exec(transactionID, detail.ProductID, detail.Quantity, detail.Subtotal)
		if err != nil {
			return nil, fmt.Errorf("failed to insert transaction detail for product %d: %w", detail.ProductID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	res = &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		Details:     details,
	}

	return res, nil
}

// GetDailySalesReport - Get sales summary untuk hari ini
func (repo *TransactionRepository) GetDailySalesReport() (*models.DailySalesReport, error) {
	startDate := time.Now().Format("2006-01-02")
	endDate := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	return repo.GetSalesReportByDateRange(startDate, endDate)
}

// GetSalesReportByDateRange - Get sales summary berdasarkan range tanggal
func (repo *TransactionRepository) GetSalesReportByDateRange(startDate, endDate string) (*models.DailySalesReport, error) {
	report := &models.DailySalesReport{}

	// Query 1: Total revenue dan jumlah transaksi
	err := repo.db.QueryRow(`
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COUNT(DISTINCT id) as total_transaksi
		FROM transactions
		WHERE DATE(created_at) >= $1 AND DATE(created_at) < $2
	`, startDate, endDate).Scan(&report.TotalRevenue, &report.TotalTransaksi)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get revenue summary: %w", err)
	}

	// Query 2: Produk terlaris
	topProduct := &models.TopProduct{}
	err = repo.db.QueryRow(`
		SELECT 
			p.name,
			SUM(td.quantity) as qty_terjual
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		WHERE DATE(t.created_at) >= $1 AND DATE(t.created_at) < $2
		GROUP BY p.id, p.name
		ORDER BY qty_terjual DESC
		LIMIT 1
	`, startDate, endDate).Scan(&topProduct.Nama, &topProduct.QtyTerjual)

	if err == sql.ErrNoRows {
		// Tidak ada data transaksi, tapi report tetap valid dengan nilai 0
		report.ProdukTerlaris = nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get top product: %w", err)
	} else {
		report.ProdukTerlaris = topProduct
	}

	return report, nil
}
