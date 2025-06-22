package main

import (
	"log"
	"os"
	"student-money-manager/database"
	"student-money-manager/handlers"
	"student-money-manager/middleware"

	"github.com/gin-gonic/gin"
	// "github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	// if err := godotenv.Load(); err != nil {
	// 	log.Println("No .env file found")
	// }

	// Initialize database
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize handlers
	handler := handlers.NewHandler(db)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.JWTAuthMiddleware())
		{
			// Account routes
			protected.GET("/account", handler.GetAccount)
			protected.PUT("/account", handler.UpdateAccount)
			protected.POST("/account/auto-allowance", handler.ProcessAutoAllowance)

			// Transaction routes
			transactions := protected.Group("/transactions")
			{
				transactions.GET("", handler.GetTransactions)
				transactions.POST("", handler.CreateTransaction)
				transactions.GET("/:id", handler.GetTransaction)
				transactions.PUT("/:id", handler.UpdateTransaction)
				transactions.PATCH("/:id", handler.PatchTransaction)
				transactions.DELETE("/:id", handler.DeleteTransaction)
			}

			// Analytics routes
			analytics := protected.Group("/analytics")
			{
				analytics.GET("/summary", handler.GetSummary)
				analytics.GET("/categories", handler.GetCategoryAnalytics)
			}

			// Savings routes
			savings := protected.Group("/savings")
			{
				savings.GET("/goals", handler.GetSavingsGoals)
				savings.POST("/goals", handler.CreateSavingsGoal)
				savings.PUT("/goals/:id", handler.UpdateSavingsGoal)
				savings.DELETE("/goals/:id", handler.DeleteSavingsGoal)
				savings.GET("/transactions", handler.GetSavingsTransactions)
				savings.POST("/transfer", handler.TransferToSavings)
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}
