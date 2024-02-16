package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Poomon001/day-trading-package/identification"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	host     = "database"
	port     = 5432
	user     = "nt_user"
	password = "db123"
	dbname   = "nt_db"
)

type Error struct {
	Success bool    `json:"success"`
	Data    *string `json:"data"`
	Message string  `json:"message"`
}

type AddMoney struct {
	Amount int `json:"amount"`
}

type PostResponse struct {
	Success bool    `json:"success"`
	Data    *string `json:"data"`
}

type WalletBalanceResponse struct {
	Success bool       `json:"success"`
	Data    WalletData `json:"data"`
}

type WalletData struct {
	Balance float64 `json:"balance"`
}

type StockPortfolioItem struct {
	StockID       int    `json:"stock_id"`
	StockName     string `json:"stock_name"`
	QuantityOwned int    `json:"quantity_owned"`
}

type StockPortfolioResponse struct {
	Success bool                 `json:"success"`
	Data    []StockPortfolioItem `json:"data"`
}

// Helper function to establish a database connection
func openConnection() (*sql.DB, error) {
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	return sql.Open("postgres", postgresqlDbInfo)
}

func handleError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := Error{
		Success: false,
		Data:    nil,
		Message: fmt.Sprintf("%s: %v", message, err),
	}
	c.IndentedJSON(statusCode, errorResponse)
}

func addMoneyToWallet(c *gin.Context) {
	userName, _ := c.Get("user_name")

	if userName == nil {
		handleError(c, http.StatusBadRequest, "Failed to obtain the user name", nil)
		return
	}

	var addMoney AddMoney
	if err := c.ShouldBindJSON(&addMoney); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	db, err := openConnection()
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to connect to the database", err)
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE users SET wallet = wallet + $1 WHERE user_name = $2", addMoney.Amount, userName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to update wallet", err)
		return
	}

	response := PostResponse{
		Success: true,
		Data:    nil,
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getWalletBalance(c *gin.Context) {
	userName, _ := c.Get("user_name")

	if userName == nil {
		handleError(c, http.StatusBadRequest, "Failed to obtain the user name", nil)
		return
	}

	db, err := openConnection()
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to connect to the database", err)
		return
	}
	defer db.Close()

	var balance float64
	err = db.QueryRow("SELECT wallet FROM users WHERE user_name = $1", userName).Scan(&balance)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to query wallet balance", err)
		return
	}

	response := WalletBalanceResponse{
		Success: true,
		Data: WalletData{
			Balance: balance,
		},
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getStockPortfolio(c *gin.Context) {
	userName, _ := c.Get("user_name")

	if userName == nil {
		handleError(c, http.StatusBadRequest, "Failed to obtain the user name", nil)
		return
	}

	db, err := openConnection()
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to connect to the database", err)
		return
	}
	defer db.Close()
	// Retrieves the stock ID, stock name, and quantity owned for all stocks
	// associated with a particular user. Performs a join operation between the 'user_stocks'
	// and 'stocks' tables
	rows, err := db.Query(`
        SELECT s.stock_id, s.stock_name, us.quantity
        FROM user_stocks us
        JOIN stocks s ON s.stock_id = us.stock_id
        WHERE us.user_name = $1`, userName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to query stock portfolio", err)
		return
	}
	defer rows.Close()

	var portfolio []StockPortfolioItem
	for rows.Next() {
		var item StockPortfolioItem
		if err := rows.Scan(&item.StockID, &item.StockName, &item.QuantityOwned); err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to scan row", err)
			return
		}
		portfolio = append(portfolio, item)
	}

	response := StockPortfolioResponse{
		Success: true,
		Data:    portfolio,
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getCookies(c *gin.Context) {
	cookie := c.GetHeader("Authorization")

	if cookie == "" {
		handleError(c, http.StatusBadRequest, "Authorization token missing", nil)
		return
	}

	c.String(http.StatusOK, "Authorization token: "+cookie)
}

func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	identification.Test()
	router.POST("/addMoneyToWallet", identification.Identification, addMoneyToWallet)
	router.GET("/getWalletBalance", identification.Identification, getWalletBalance)
	router.GET("/getStockPortfolio", identification.Identification, getStockPortfolio)
	router.GET("/eatCookies", getCookies)
	router.Run(":5433")
}
