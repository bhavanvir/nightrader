package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/Poomon001/day-trading-package/identification"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var user_db *sql.DB
var stock_db *sql.DB
var tx_db *sql.DB

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

var (
	stmtAddMoney *sql.Stmt
	stmtWalletBalance *sql.Stmt
	stmtStockPortfolio *sql.Stmt
	stmtWalletTransactions *sql.Stmt
	stmtStockTransactions *sql.Stmt
	stmtStockPrices *sql.Stmt
)

type ErrorResponse struct {
	Success bool              `json:"success"`
	Data    map[string]string `json:"data"`
}

type AddMoney struct {
	Amount float64 `json:"amount"`
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
	StockID      string  `json:"stock_id"`
	StockName    string  `json:"stock_name"`
	CurrentPrice float64 `json:"current_price"`
}

type StockPortfolioItem struct {
	StockID       string  `json:"stock_id"`
	StockName     string  `json:"stock_name"`
	QuantityOwned float64 `json:"quantity_owned"`
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
	StockID     string  `json:"stock_id"`
	WalletTxID  *string `json:"wallet_tx_id"`
	OrderStatus string  `json:"order_status"`
	ParentTxID  *string `json:"parent_stock_tx_id"`
	IsBuy       bool    `json:"is_buy"`
	OrderType   string  `json:"order_type"`
	StockPrice  float64 `json:"stock_price"`
	Quantity    float64 `json:"quantity"`
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
		handleError(c, http.StatusOK, "Invalid amount", nil)
		return
	}

	_, err := stmtAddMoney.Exec(addMoney.Amount, userName)
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

	var balance float64
	err := stmtWalletBalance.QueryRow(userName).Scan(&balance)
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
	rows, err := stmtStockPrices.Query()
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

	rows, err := stmtStockPortfolio.Query(userName)
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

	rows, err := stmtWalletTransactions.Query(userName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to query wallet transactions", err)
		return
	}
	defer rows.Close()

	var walletTransactions []WalletTransactionItem
	for rows.Next() {
		var item WalletTransactionItem
		if err := rows.Scan(&item.WalletTxID, &item.StockTxID, &item.IsDebit, &item.Amount, &item.TimeStamp); err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to scan row", err)
			return
		}
		walletTransactions = append(walletTransactions, item)
	}

	response := WalletTransactionResponse{
		Success: true,
		Data:    walletTransactions,
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getStockTransactions(c *gin.Context) {
	userName, _ := c.Get("user_name")

	if userName == nil {
		handleError(c, http.StatusBadRequest, "Failed to obtain the user name", nil)
		return
	}

	rows, err := stmtStockTransactions.Query(userName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to query stock transactions", err)
		return
	}
	defer rows.Close()

	var stockTransactions []StockTransactionItem
	for rows.Next() {
		var item StockTransactionItem
		if err := rows.Scan(&item.StockTxID, &item.StockID, &item.WalletTxID, &item.OrderStatus, &item.ParentTxID, &item.IsBuy, &item.OrderType, &item.StockPrice, &item.Quantity, &item.TimeStamp); err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to scan row", err)
			return
		}
		stockTransactions = append(stockTransactions, item)
	}

	response := StockTransactionResponse{
		Success: true,
		Data:    stockTransactions,
	}
	c.IndentedJSON(http.StatusOK, response)
}

func initializeDB() error {
	var err error
    postgresqlUserDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", user_host, user_port, user, password, dbname)
    user_db, err = sql.Open("postgres", postgresqlUserDbInfo)
    if err != nil {
        return fmt.Errorf("failed to connect to the user database: %v", err)
    }

    postgresqlStockDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", stock_host, stock_port, user, password, dbname)
    stock_db, err = sql.Open("postgres", postgresqlStockDbInfo)
    if err != nil {
        return fmt.Errorf("failed to connect to the stock database: %v", err)
    }

    postgresqlTxDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", tx_host, tx_port, user, password, dbname)
    tx_db, err = sql.Open("postgres", postgresqlTxDbInfo)
    if err != nil {
        return fmt.Errorf("failed to connect to the transaction database: %v", err)
    }

	// Ensure the database connection is fully established
    for {
        err = user_db.Ping()
        if err == nil {
            break
        }
        fmt.Println("Waiting for the user database connection to be established...")
        time.Sleep(1 * time.Second)
    }

    for {
        err = stock_db.Ping()
        if err == nil {
            break
        }
        fmt.Println("Waiting for the stock database connection to be established...")
        time.Sleep(1 * time.Second)
    }

    for {
        err = tx_db.Ping()
        if err == nil {
            break
        }
        fmt.Println("Waiting for the transaction database connection to be established...")
        time.Sleep(1 * time.Second)
    }

	return nil
}

func prepareStatements() error {
	var err error

	stmtAddMoney, err = user_db.Prepare("UPDATE users SET wallet = wallet + $1 WHERE user_name = $2")
	if err != nil {
		return fmt.Errorf("failed to prepare addMoney statement: %v", err)
	}

	stmtWalletBalance, err = user_db.Prepare("SELECT wallet FROM users WHERE user_name = $1")
	if err != nil {
		return fmt.Errorf("failed to prepare walletBalance statement: %v", err)
	}

	stmtStockPortfolio, err = stock_db.Prepare(`
        SELECT s.stock_id, s.stock_name, us.quantity
        FROM user_stocks us
        JOIN stocks s ON s.stock_id = us.stock_id
        WHERE us.user_name = $1
		ORDER BY us.time_added ASC`)
	if err != nil {
		return fmt.Errorf("failed to prepare stockPortfolio statement: %v", err)
	}

	stmtWalletTransactions, err = tx_db.Prepare(`
        SELECT wt.wallet_tx_id, st.stock_tx_id, wt.is_debit, wt.amount, wt.time_stamp
        FROM wallet_transactions wt
        JOIN stock_transactions st ON st.wallet_tx_id = wt.wallet_tx_id
        WHERE wt.user_name = $1
		ORDER BY wt.time_stamp ASC`)
	if err != nil {
		return fmt.Errorf("failed to prepare walletTransactions statement: %v", err)
	}

	stmtStockTransactions, err = tx_db.Prepare(`
        SELECT stock_tx_id, stock_id, wallet_tx_id, order_status, parent_stock_tx_id, is_buy, order_type, stock_price, quantity, time_stamp
        FROM stock_transactions
        WHERE user_name = $1
		ORDER BY time_stamp ASC`)
	if err != nil {
		return fmt.Errorf("failed to prepare stockTransactions statement: %v", err)
	}

	stmtStockPrices, err = stock_db.Prepare(`
		SELECT stock_id, stock_name, current_price
		FROM stocks
		ORDER BY time_added ASC`)
	if err != nil {
		return fmt.Errorf("failed to prepare stockPrices statement: %v", err)
	}

	return nil
}

func main() {
	err := initializeDB()
	if err != nil {
		fmt.Printf("Failed to initialize the database: %v\n", err)
		return
	}
    defer user_db.Close()
    defer stock_db.Close()
    defer tx_db.Close()

	err = prepareStatements()
	if err != nil {
		fmt.Printf("Failed to prepare SQL statements: %v\n", err)
		return
	}

	defer stmtAddMoney.Close()
	defer stmtWalletBalance.Close()
	defer stmtStockPortfolio.Close()
	defer stmtWalletTransactions.Close()
	defer stmtStockTransactions.Close()
	defer stmtStockPrices.Close()

    user_db.SetMaxOpenConns(10) // Set maximum number of open connections
    user_db.SetMaxIdleConns(5) // Set maximum number of idle connections

    stock_db.SetMaxOpenConns(10) // Set maximum number of open connections
    stock_db.SetMaxIdleConns(5) // Set maximum number of idle connections

    tx_db.SetMaxOpenConns(10) // Set maximum number of open connections
    tx_db.SetMaxIdleConns(5) // Set maximum number of idle connections

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "token"}
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
