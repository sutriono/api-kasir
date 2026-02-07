package routes

// import (
// 	"database/sql"
// 	"kasir-api/handlers"
// 	"kasir-api/repositories"

// 	"github.com/gin-gonic/gin"
// )

// func SetupReportRoutes(router *gin.Engine, db *sql.DB) {
// 	reportRepo := repositories.NewReportRepository(db)
// 	reportHandler := handlers.NewReportHandler(reportRepo)

// 	report := router.Group("/api/report")
// 	{
// 		report.GET("/hari-ini", reportHandler.GetTodaysSalesSummary)
// 		report.GET("", reportHandler.GetSalesReport)
// 	}
// }
