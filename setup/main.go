package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/Poomon001/day-trading-package/identification"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Stock struct {
	StockName string `json:"stock_name"`
}

const (
    user_host = "user_database"
    stock_host = "stock_database"
    tx_host = "tx_database"
    // host     = "localhost" // for local testing
    user_port     = 5432
    stock_port    = 5431
    tx_port      = 5430
    user     = "nt_user"
    password = "db123"
    dbname   = "nt_db"
)

type AddStockRequest struct {
	StockID  string  `json:"stock_id"`
	Quantity float64 `json:"quantity"`
}

type ErrorResponse struct {
	Success bool              `json:"success"`
	Data    map[string]string `json:"data"`
}

type PostResponse struct {
	Success bool    `json:"success"`
	Data    *string `json:"data"`
}

func handleError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := ErrorResponse{
		Success: false,
		Data:    map[string]string{"error": message},
	}
	c.IndentedJSON(statusCode, errorResponse)
}

func createStock(c *gin.Context) {
	user_name, exists := c.Get("user_name")
	if !exists || user_name == nil {
		handleError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var json Stock

	if err := c.BindJSON(&json); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Generate UUID as string for the new stock
	stockID := uuid.New().String()

	// Save stock to database with generated stockID
	err := saveStockToDatabase(json, stockID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to save stock to database", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"stock_id": stockID,
		},
	})
}

func saveStockToDatabase(stock Stock, stockID string) error {
	// Define formatted string for database connection
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", stock_host, stock_port, user, password, dbname)

	// Attempt to connect to database
	db, err := sql.Open("postgres", postgresqlDbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// Insert stock into the stocks table with provided stockID
    _, err = db.Exec("INSERT INTO stocks (stock_id, stock_name, time_added) VALUES ($1, $2, $3)", stockID, stock.StockName, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func addStockToUser(c *gin.Context) {
	// Get user name from identification middleware
	userName, _ := c.Get("user_name")
	if userName == nil {
		handleError(c, http.StatusBadRequest, "Failed to obtain the user name", nil)
		return
	}

	var req AddStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Connect to the PostgreSQL database
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", stock_host, stock_port, user, password, dbname))
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to connect to the database", err)
		return
	}
	defer db.Close()

	fmt.Println("userName:", userName)
	fmt.Println("ID:", req.StockID)
	fmt.Println("quantity:", req.Quantity)
	// Insert stock into user_stocks table
	_, err = db.Exec(`
		INSERT INTO user_stocks (user_name, stock_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_name, stock_id)
		DO UPDATE SET quantity = user_stocks.quantity + EXCLUDED.quantity;
	`, userName, req.StockID, req.Quantity)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to add stock to user", err)
		return
	}

	// If everything succeeded, return success response
	response := PostResponse{
		Success: true,
		Data:    nil,
	}
	c.IndentedJSON(http.StatusOK, response)
}

func wipeDatabaseTables(c *gin.Context) {
	/*
		This function is needed when running the postman collection tests, as not doing so
		will cause certain tests to fail
	*/

	stock_db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", stock_host, stock_port, user, password, dbname))
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to connect to the database", err)
		return
	}
	defer stock_db.Close()

	// Define a list of tables to truncate
	stock_tables := []string{"stocks", "user_stocks"}

	// Truncate each table. This will delete all rows in the table
	for _, stock_table := range stock_tables {
		_, err = stock_db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", stock_table))
		if err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to truncate table", err)
			return
		}
	}

	// Reset the stock_id sequence to start at  1 after truncating the stocks table. This is necessary because
	// when we test the endpoint in the postman collection, the stock_id will be auto incremented by 1, causing
	// the test to fail
	_, err = stock_db.Exec("ALTER SEQUENCE stocks_stock_id_seq RESTART WITH  1")
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to reset stock_id sequence", err)
		return
	}

	user_db, user_err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", user_host, user_port, user, password, dbname))
	if user_err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to connect to the database", err)
		return
	}
	defer user_db.Close()

	// Define a list of tables to truncate
	user_tables := []string{"users"}

	// Truncate each table. This will delete all rows in the table
	for _, user_table := range user_tables {
		_, err = user_db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", user_table))
		if err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to truncate table", err)
			return
		}
	}

	tx_db, tx_err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", tx_host, tx_port, user, password, dbname))
	if tx_err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to connect to the database", err)
		return
	}
	defer tx_db.Close()

	// Define a list of tables to truncate
	tx_tables := []string{"stock_transactions", "wallet_transactions"}

	// Truncate each table. This will delete all rows in the table
	for _, tx_table := range tx_tables {
		_, err = tx_db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tx_table))
		if err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to truncate table", err)
			return
		}
	}

	response := PostResponse{
		Success: true,
		Data:    nil,
	}
	c.IndentedJSON(http.StatusOK, response)
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	identification.Test()
	router.POST("/createStock", identification.Identification, createStock)
	router.POST("/addStockToUser", identification.Identification, addStockToUser)

	// For testring purposes: all database tables are wiped before running postman-collection tests
	router.DELETE("/wipeDatabaseTables", wipeDatabaseTables)

	router.Run(":8080")
}
