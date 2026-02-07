package models

type Transaction struct {
	ID          int                 `json:"id"`
	TotalAmount int                 `json:"total_amount"`
	Details     []TransactionDetail `json:"details"`
}

type TransactionDetail struct {
	ID            int    `json:"id"`
	TransactionID int    `json:"transaction_id"`
	ProductID     int    `json:"product_id"`
	ProductName   string `json:"product_name"`
	Quantity      int    `json:"quantity"`
	Subtotal      int    `json:"subtotal"`
}

type CheckoutRequest struct {
	Items []CheckoutItem `json:"items"`
}

type CheckoutItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type SalesSummary struct {
	TotalRevenue   int                `json:"total_revenue"`
	TotalTransaksi int                `json:"total_transaksi"`
	ProdukTerlaris BestSellingProduct `json:"produk_terlaris"`
}

type BestSellingProduct struct {
	Nama       string `json:"nama"`
	QtyTerjual int    `json:"qty_terjual"`
}

type SalesReport struct {
	TotalRevenue   int                `json:"total_revenue"`
	TotalTransaksi int                `json:"total_transaksi"`
	ProdukTerlaris BestSellingProduct `json:"produk_terlaris"`
	DateRange      DateRange          `json:"date_range"`
	DailyBreakdown []DailySalesData   `json:"daily_breakdown,omitempty"`
}

type DateRange struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type DailySalesData struct {
	Date       string             `json:"date"`
	Revenue    int                `json:"revenue"`
	Transaksi  int                `json:"transaksi"`
	TopProduct BestSellingProduct `json:"top_product"`
}

type DailySalesReport struct {
	TotalRevenue   int         `json:"total_revenue"`
	TotalTransaksi int         `json:"total_transaksi"`
	ProdukTerlaris *TopProduct `json:"produk_terlaris"`
}

type TopProduct struct {
	Nama       string `json:"nama"`
	QtyTerjual int    `json:"qty_terjual"`
}

type SalesReportFilter struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}
