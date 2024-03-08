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
	// host = "database"
	host = "localhost" // for local testing
	port     = 5432
	user     = "nt_user"
	password = "db123"
	dbname   = "nt_db"
)

type ErrorResponse struct {
	Success bool              `json:"success"`
	Data    map[string]string `json:"data"`
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

type StockResponse struct {
	Success bool        `json:"success"`
	Data    []StockData `json:"data"`
}

type StockData struct {
	StockID      int     `json:"stock_id"`
	StockName    string  `json:"stock_name"`
	CurrentPrice float64 `json:"current_price"`
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

type WalletTransactionItem struct {
	WalletTxID string  `json:"wallet_tx_id"`
	StockTxID  string  `json:"stock_tx_id"`
	IsDebit    bool    `json:"is_debit"`
	Amount     float64 `json:"amount"`
	TimeStamp  string  `json:"time_stamp"`
}

type WalletTransactionResponse struct {
	Success bool                    `json:"success"`
	Data    []WalletTransactionItem `json:"data"`
}

type StockTransactionItem struct {
	StockTxID   string  `json:"stock_tx_id"`
	StockID     int     `json:"stock_id"`
	WalletTxID  string  `json:"wallet_tx_id"`
	OrderStatus string  `json:"order_status"`
	ParentTxID  *string  `json:"parent_tx_id"`
	IsBuy       bool    `json:"is_buy"`
	OrderType   string  `json:"order_type"`
	StockPrice  float64 `json:"stock_price"`
	Quantity    int     `json:"quantity"`
	TimeStamp   string  `json:"time_stamp"`
}

type StockTransactionResponse struct {
	Success bool                   `json:"success"`
	Data    []StockTransactionItem `json:"data"`
}

func handleError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := ErrorResponse{
		Success: false,
		Data:    map[string]string{"error": message},
	}
	c.IndentedJSON(statusCode, errorResponse)
}

// Helper function to establish a database connection
func openConnection() (*sql.DB, error) {
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	return sql.Open("postgres", postgresqlDbInfo)
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

	if addMoney.Amount <= 0 {
		handleError(c, http.StatusBadRequest, "Invalid amount", nil)
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

func getStockPrices(c *gin.Context) {
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

	rows, err := db.Query(`
        SELECT stock_id, stock_name, current_price
        FROM stocks`)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to query stock prices", err)
		return
	}
	defer rows.Close()

	var stocks []StockData
	for rows.Next() {
		var item StockData
		if err := rows.Scan(&item.StockID, &item.StockName, &item.CurrentPrice); err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to scan row", err)
			return
		}
		stocks = append(stocks, item)
	}

	response := StockResponse{
		Success: true,
		Data:    stocks,
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

	// If the portfolio is empty, return an empty array
	if len(portfolio) == 0 {
		portfolio = []StockPortfolioItem{}
	}

	response := StockPortfolioResponse{
		Success: true,
		Data:    portfolio,
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getWalletTransactions(c *gin.Context) {
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

	rows, err := db.Query(`
        SELECT wt.wallet_tx_id, st.stock_tx_id, wt.is_debit, wt.amount, wt.time_stamp
        FROM wallet_transactions wt
        JOIN stock_transactions st ON st.wallet_tx_id = wt.wallet_tx_id
        WHERE wt.user_name = $1
		ORDER BY wt.time_stamp ASC`, userName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to query wallet transactions", err)
		return
	}
	defer rows.Close()

	var wallet_transactions []WalletTransactionItem
	for rows.Next() {
		var item WalletTransactionItem
		if err := rows.Scan(&item.WalletTxID, &item.StockTxID, &item.IsDebit, &item.Amount, &item.TimeStamp); err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to scan row", err)
			return
		}
		wallet_transactions = append(wallet_transactions, item)
	}

	response := WalletTransactionResponse{
		Success: true,
		Data:    wallet_transactions,
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getStockTransactions(c *gin.Context) {
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

	rows, err := db.Query(`
        SELECT stock_tx_id, stock_id, wallet_tx_id, order_status, parent_tx_id, is_buy, order_type, stock_price, quantity, time_stamp
        FROM stock_transactions
        WHERE user_name = $1
		ORDER BY time_stamp ASC`, userName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to query stock transactions", err)
		return
	}
	defer rows.Close()

	var stock_transactions []StockTransactionItem
	for rows.Next() {
		var item StockTransactionItem
		if err := rows.Scan(&item.StockTxID, &item.StockID, &item.WalletTxID, &item.OrderStatus, &item.ParentTxID, &item.IsBuy, &item.OrderType, &item.StockPrice, &item.Quantity, &item.TimeStamp); err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to scan row", err)
			return
		}
		fmt.Println(item)
		stock_transactions = append(stock_transactions, item)
	}

	response := StockTransactionResponse{
		Success: true,
		Data:    stock_transactions,
	}
	c.IndentedJSON(http.StatusOK, response)
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
	router.GET("/getWalletTransactions", identification.Identification, getWalletTransactions)
	router.GET("/getStockTransactions", identification.Identification, getStockTransactions)
	router.GET("/getStockPrices", identification.Identification, getStockPrices)
	router.Run(":5433")
}
