package main

import (
    "database/sql"
	"net/http"
    "fmt"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"
	"github.com/google/uuid"
	"time"
)

// Global variable for the database connection
var db *sql.DB

const (
	host = "database"
	// host     = "localhost" // for local testing
	port     = 5432
	user     = "nt_user"
	password = "db123"
	dbname   = "nt_db"
)

type Order struct {
	StockTxID  string   `json:"stock_tx_id"`
	StockID    string   `json:"stock_id"`
	WalletTxID string   `json:"wallet_tx_id"`
	ParentTxID *string  `json:"parent_stock_tx_id"`
	IsBuy      bool     `json:"is_buy"`
	OrderType  string   `json:"order_type"`
	Quantity   float64  `json:"quantity"`
	Price      *float64 `json:"price"`
	TimeStamp  string   `json:"time_stamp"`
	Status     string   `json:"status"`
	UserName   string   `json:"user_name"`
}

type ErrorResponse struct {
    Success bool              `json:"success"`
    Data    map[string]string `json:"data"`
}

type TradePayload struct {
	BuyOrder  Order   `json:"buy_order"`
	SellOrder Order   `json:"sell_order"`
}

func openConnection() (*sql.DB, error) {
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	return sql.Open("postgres", postgresqlDbInfo)
}

// handleError is a helper function to send error responses
func handleError(c *gin.Context, statusCode int, message string, err error) {
	errorMessage := message
	if err != nil {
		errorMessage += err.Error()
	}
	errorResponse := ErrorResponse{
		Success: false,
		Data:    map[string]string{"error": errorMessage},
	}
	c.IndentedJSON(statusCode, errorResponse)
}

func handleSellTrade(buyOrder *Order, sellOrder *Order) {
	tradeQuantity := min(buyOrder.Quantity, sellOrder.Quantity)
	buyPrice := buyOrder.Price
	sellPrice := sellOrder.Price

	if buyOrder.Quantity > sellOrder.Quantity {
		// execute partial trade for buy order and complete trade for sell order
		buyOrder.Quantity -= tradeQuantity
		sellOrder.Quantity = 0
		completeSellOrder(sellOrder, tradeQuantity, sellPrice)
		partialFulfillBuyOrder(buyOrder, tradeQuantity, buyPrice, sellPrice)
	} else if buyOrder.Quantity < sellOrder.Quantity {
		// execute partial trade for sell order and complete trade for buy order
		sellOrder.Quantity -= tradeQuantity
		buyOrder.Quantity = 0
		completeBuyOrder(buyOrder, tradeQuantity, buyPrice, sellPrice)
		partialFulfillSellOrder(sellOrder, tradeQuantity, sellPrice)
	} else {
		// execute complete trade for both buy and sell orders
		buyOrder.Quantity = 0
		sellOrder.Quantity = 0
		completeBuyOrder(buyOrder, tradeQuantity, buyPrice, sellPrice)
		completeSellOrder(sellOrder, tradeQuantity, sellPrice)
	}
}

func handleBuyTrade(buyOrder *Order, sellOrder *Order) {
	tradeQuantity := min(buyOrder.Quantity, sellOrder.Quantity)
	buyPrice := buyOrder.Price
	sellPrice := sellOrder.Price

	if buyOrder.Quantity > sellOrder.Quantity {
		// execute partial trade for buy order and complete trade for sell order
		buyOrder.Quantity -= tradeQuantity
		sellOrder.Quantity = 0
		partialFulfillBuyOrder(buyOrder, tradeQuantity, buyPrice, sellPrice)
		completeSellOrder(sellOrder, tradeQuantity, sellPrice)
	} else if buyOrder.Quantity < sellOrder.Quantity {
		// execute partial trade for sell order and complete trade for buy order
		sellOrder.Quantity -= tradeQuantity
		buyOrder.Quantity = 0
		completeBuyOrder(buyOrder, tradeQuantity, buyPrice, sellPrice)
		partialFulfillSellOrder(sellOrder, tradeQuantity, sellPrice)
	} else {
		// execute complete trade for both buy and sell orders
		buyOrder.Quantity = 0
		sellOrder.Quantity = 0
		completeBuyOrder(buyOrder, tradeQuantity, buyPrice, sellPrice)
		completeSellOrder(sellOrder, tradeQuantity, sellPrice)
	}
}

// generateOrderID generates a unique ID
func generateUID() string {
	return uuid.New().String()
}

/** === BUY/SELL Order === **/

func partialFulfillBuyOrder(order *Order, tradeQuantity float64, buyPrice *float64, sellPrice *float64) {
	fmt.Println("Buy User: ", order.UserName)
	fmt.Println("Buy Wallet_tx: === ", order.WalletTxID)

	refundAmount := (*buyPrice - *sellPrice) * float64(tradeQuantity)

	if refundAmount > 0 {
		newWalletTxAmount, err := getWalletTransactionsAmount(order.UserName, order.WalletTxID)
		if err != nil {
			fmt.Println("Error getting wallet transaction amount: ", err)
		}
		
		// Refund deducted money to the Buy user's wallet, adjusting for any price differences
		fmt.Printf("Refund Amount: [%f] to User: [%s]\n", refundAmount, order.UserName)
		if err := updateMoneyWallet(order.UserName, refundAmount, true); err != nil {
			fmt.Println("Error updating different price refund to wallet: ", err)
		}

		// update wallet_transactions from BUY order
		if err := updateWalletTransaction(order.UserName, *order, newWalletTxAmount - refundAmount); err != nil {
			fmt.Println("Error updating wallet transaction: ", err)
		}
	} 

	if err := updateStockPortfolio(order.UserName, *order, tradeQuantity, true); err != nil {
		fmt.Println("Error updating stock portfolio: ", err)
	}

	if err := setStatus(order, "PARTIAL_FULFILLED", false); err != nil {
		fmt.Println("Error setting status: ", err)
	}

	completedOrder := Order{
		StockTxID:  generateUID(),
		StockID:    order.StockID,
		WalletTxID: generateUID(),
		ParentTxID: &order.StockTxID,
		IsBuy:      order.IsBuy,
		OrderType:  order.OrderType,
		Quantity:   tradeQuantity,
		Price:      order.Price,
		TimeStamp:  time.Now().Format(time.RFC3339Nano),
		Status:     "COMPLETED",
		UserName:   order.UserName,
	}

	// setWalletTransaction should always be before the setStockTransaction
	if err := setWalletTransaction(order.UserName, completedOrder.WalletTxID, completedOrder.TimeStamp, sellPrice, tradeQuantity, false); err != nil {
		fmt.Println("Error setting wallet transaction: ", err)
	}

	if err := setStockTransaction(order.UserName, completedOrder, sellPrice, tradeQuantity); err != nil {
		fmt.Println("Error setting stock transaction: ", err)
	}
}

func partialFulfillSellOrder(sellOrder *Order, tradeQuantity float64, sellPrice *float64) {
	fmt.Println("Sell User: ", sellOrder.UserName)

	if err := updateMarketStockPrice(sellOrder.StockID, sellPrice); err != nil {
		fmt.Println("Failed to update Market Stock Price after Limit Sell: ", err)
	}

	amount := (*sellPrice) * float64(tradeQuantity)
	if err := updateMoneyWallet(sellOrder.UserName, amount, true); err != nil {
		fmt.Println("Error updating wallet: ", err)
	}

	if err := setStatus(sellOrder, "PARTIAL_FULFILLED", false); err != nil {
		fmt.Println("Error setting status: ", err)
	}

	completedOrder := Order{
		StockTxID:  generateUID(),
		StockID:    sellOrder.StockID,
		WalletTxID: generateUID(),
		ParentTxID: &sellOrder.StockTxID,
		IsBuy:      sellOrder.IsBuy,
		OrderType:  sellOrder.OrderType,
		Quantity:   tradeQuantity,
		Price:      sellOrder.Price,
		TimeStamp:  time.Now().Format(time.RFC3339Nano),
		Status:     "COMPLETED",
		UserName:   sellOrder.UserName,
	}

	fmt.Println("Completed wallet tx: ", completedOrder.WalletTxID)

	// setWalletTransaction should always be before the setStockTransaction
	if err := setWalletTransaction(sellOrder.UserName, completedOrder.WalletTxID, completedOrder.TimeStamp, sellPrice, tradeQuantity, true); err != nil {
		fmt.Println("Error setting wallet transaction: ", err)
	}

	if err := setStockTransaction(sellOrder.UserName, completedOrder, sellPrice, tradeQuantity); err != nil {
		fmt.Println("Error setting stock transaction: ", err)
	}
}

func completeBuyOrder(buyOrder *Order, tradeQuantity float64, buyPrice *float64, sellPrice *float64) {
	fmt.Println("Buy User: ", buyOrder.UserName)
	fmt.Println("Buy Wallet_tx: === ", buyOrder.WalletTxID)

	refundAmount := (*buyPrice - *sellPrice) * float64(tradeQuantity)

	if refundAmount > 0 {
		totalSoldAmount := (*sellPrice) * float64(tradeQuantity)

		// Refund deducted money to the Buy user's wallet, adjusting for any price differences
		fmt.Printf("Refund Amount: [%f] to User: [%s]\n", refundAmount, buyOrder.UserName)
		if err := updateMoneyWallet(buyOrder.UserName, refundAmount, true); err != nil {
			fmt.Println("Error updating different price refund to wallet: ", err)
		}

		// update wallet_transactions from BUY order
		if err := updateWalletTransaction(buyOrder.UserName, *buyOrder, totalSoldAmount); err != nil {
			fmt.Println("Error updating wallet transaction: ", err)
		}
	} 

	if err := updateStockPortfolio(buyOrder.UserName, *buyOrder, tradeQuantity, true); err != nil {
		fmt.Println("Error updating stock portfolio: ", err)
	}

	if err := setStatus(buyOrder, "COMPLETED", false); err != nil {
		fmt.Println("Error setting status: ", err)
	}
}

func completeSellOrder(sellOrder *Order, tradeQuantity float64, sellPrice *float64) {
	fmt.Println("Sell User: ", sellOrder.UserName)
	fmt.Println("Sell Wallet_tx: === ", sellOrder.WalletTxID)

	if err := updateMarketStockPrice(sellOrder.StockID, sellPrice); err != nil {
		fmt.Println("Failed to update Market Stock Price after Limit Sell: ", err)
	}

	amount := (*sellPrice) * float64(tradeQuantity)
	if err := updateMoneyWallet(sellOrder.UserName, amount, true); err != nil {
		fmt.Println("Error updating wallet: ", err)
	}

	if err := setStatus(sellOrder, "COMPLETED", true); err != nil {
		fmt.Println("Error setting status: ", err)
	}

	if err := setWalletTransaction(sellOrder.UserName, sellOrder.WalletTxID, sellOrder.TimeStamp, sellPrice, tradeQuantity, true); err != nil {
		fmt.Println("Error setting wallet transaction: ", err)
	}
}

/** === END BUY/SELL Order === **/

func executeBuyTrade(c *gin.Context) {
	var payload TradePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	buyOrder := &payload.BuyOrder
	sellOrder := &payload.SellOrder

	// Handle the buy trade execution here
	handleBuyTrade(buyOrder, sellOrder)

	fmt.Printf("\nBuy Trade Executed - Sell Order: ID=%s, Quantity=%.2f, Price=$%.2f | Buy Order: ID=%s, Quantity=%.2f, Price=$%.2f\n",
	sellOrder.StockTxID, sellOrder.Quantity, *sellOrder.Price, buyOrder.StockTxID, buyOrder.Quantity, *buyOrder.Price)
	
	responsepayload := TradePayload{
		BuyOrder:  *buyOrder,
		SellOrder: *sellOrder,
	}

	c.JSON(http.StatusOK, responsepayload)
}

func executeSellTrade(c *gin.Context) {
	var payload TradePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	buyOrder := &payload.BuyOrder
	sellOrder := &payload.SellOrder

	// Handle the sell trade execution here
	handleSellTrade(buyOrder, sellOrder)

	fmt.Printf("\nSell Trade Executed - Sell Order: ID=%s, Quantity=%.2f, Price=$%.2f | Buy Order: ID=%s, Quantity=%.2f, Price=$%.2f\n",
	sellOrder.StockTxID, sellOrder.Quantity, *sellOrder.Price, buyOrder.StockTxID, buyOrder.Quantity, *buyOrder.Price)
	
	responsepayload := TradePayload{
		BuyOrder:  *buyOrder,
		SellOrder: *sellOrder,
	}

	c.JSON(http.StatusOK, responsepayload)
}

/** === BUY/SELL support === **/

func getWalletTransactionsAmount(userName string, walletTxID string) (float64, error) {
	// Connect to database
	db, err := openConnection()
	if err != nil {
		return 0, fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	// Query the database to get the total amount of wallet transactions for the specified user and wallet ID
	var totalAmount float64
	err = db.QueryRow("SELECT SUM(amount) FROM wallet_transactions WHERE user_name = $1 AND wallet_tx_id = $2", userName, walletTxID).Scan(&totalAmount)
	if err != nil {
		return 0, fmt.Errorf("Failed to get wallet transactions amount: %w", err)
	}

	return totalAmount, nil
}

func updateMoneyWallet(userName string, amount float64, isAdded bool) error {
	fmt.Println("Deducting money from wallet")

	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	// Calculate total to be added or deducted
	if !isAdded {
		amount = amount * (-1) // Reduce funds if buying
	}

	// Update the user's wallet
	_, err = db.Exec(`
		UPDATE users SET wallet = wallet + $1 WHERE user_name = $2`, amount, userName)
	if err != nil {
		return fmt.Errorf("Failed to update wallet: %w", err)
	}
	return nil
}

func updateWalletTransaction(userName string, order Order, amount float64) error {
	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	// Update the wallet transaction
	_, err = db.Exec(`UPDATE wallet_transactions SET amount = $1 WHERE user_name = $2 AND wallet_tx_id = $3`, amount, userName, order.WalletTxID)
	if err != nil {
		return fmt.Errorf("Failed to update wallet transaction: %w", err)
	}

	return nil
}

func updateStockPortfolio(userName string, order Order, quantity float64, isAdded bool) error {
	fmt.Println("Deducting stock from portfolio")

	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	// Calculate total to be added or deducted
	total := quantity
	if !isAdded {
		total = total * (-1) // Reduce stocks if selling
	}

	rows, err := db.Query(`
		SELECT quantity FROM user_stocks WHERE user_name = $1 AND stock_id = $2`, userName, order.StockID)
	if err != nil {
		return fmt.Errorf("Failed to query user stocks: %w", err)
	}
	defer rows.Close()

	// User already owns this stock
	if rows.Next() {
		// Update the user's stocks
		var amount float64
		if err := rows.Scan(&amount); err != nil {
			return fmt.Errorf("Error while scanning row: %w", err)
		}
		if total < 0 && (amount+total) <= 0 {
			_, err = db.Exec(`
				DELETE FROM user_stocks WHERE user_name = $1 AND stock_id = $2`, userName, order.StockID)
			if err != nil {
				return fmt.Errorf("Failed to update user stocks: %w", err)
			}
		} else {
			_, err = db.Exec(`
				UPDATE user_stocks SET quantity = quantity + $1 WHERE user_name = $2 AND stock_id = $3`, total, userName, order.StockID)
			if err != nil {
				return fmt.Errorf("Failed to update user stocks: %w", err)
			}
		}
	} else {
		// For wallet transactions, update the wallet regardless of the order type
		if total <= 0 {
			return fmt.Errorf("No stocks to deduct")
		} else {
			_, err = db.Exec(`
				INSERT INTO user_stocks VALUES($1, $2, $3)`, userName, order.StockID, quantity)
			if err != nil {
				return fmt.Errorf("Failed to create user_stock: %w", err)
			}
		}
	}

	return nil
}

func setStatus(order *Order, status string, isUpdateWalletTxId bool) error {
	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	if status == "PARTIAL_FULFILLED" {
		order.Status = status
	}

	// Insert transaction to wallet transactions
	_, err = db.Exec(`
		UPDATE stock_transactions SET order_status = $1 WHERE user_name = $2 AND stock_tx_id = $3`, status, order.UserName, order.StockTxID)
	if err != nil {
		return fmt.Errorf("Failed to update status: %w", err)
	}

	// assign wallet_tx_id to stock_tx_id if the Sell order is completed
	if isUpdateWalletTxId {
		_, err = db.Exec(`
			UPDATE stock_transactions SET wallet_tx_id = $1 WHERE user_name = $2 AND stock_tx_id = $3`, order.WalletTxID, order.UserName, order.StockTxID)
	}

	return nil
}

// Store completed wallet transactions based on order matched
func setWalletTransaction(userName string, walletTxID string, timestamp string, price *float64, quantity float64, isAdded bool) error {
	fmt.Println("Setting wallet transaction")
	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to insert stock transaction: %w", err)
	}
	defer db.Close()

	// isDebit = True if money is deducted from wallet.
	// isDebit = False if money is added to wallet
	var isDebit bool
	isDebit = !isAdded

	amount := (*price) * float64(quantity)

	// Insert transaction to wallet transactions
	_, err = db.Exec(`
		INSERT INTO wallet_transactions (wallet_tx_id, user_name, is_debit, amount, time_stamp)
		VALUES ($1, $2, $3, $4, $5)`, walletTxID, userName, isDebit, amount, timestamp)
	if err != nil {
		return fmt.Errorf("Failed to commit transaction: %w", err)
	}
	return nil
}

// Store transaction based on the order user created
func setStockTransaction(userName string, tx Order, price *float64, quantity float64) error {
	fmt.Println("Setting stock transaction")
	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to insert stock transaction: %w", err)
	}
	defer db.Close()

	// Check if a wallet transaction has been made for this order yet
	rows, err := db.Query(`
		SELECT wallet_tx_id FROM wallet_transactions WHERE user_name = $1 AND wallet_tx_id = $2`, userName, tx.WalletTxID)
	if err != nil {
		return fmt.Errorf("Error querying wallet transactions: %w", err)
	}
	defer rows.Close()

	var wallet_tx_id *string

	// if a wallet transaction is found in wallet_transaction table db, then add it to stock_transaction table OR,
	// if status is COMPLETED, the stock transaction need to a wallet transaction
	if rows.Next() || tx.Status == "COMPLETED" {
		fmt.Println("Wallet transaction: ", wallet_tx_id)
		wallet_tx_id = &tx.WalletTxID
	}

	// Insert transaction to stock transactions
	_, err = db.Exec(`
		INSERT INTO stock_transactions (stock_tx_id, user_name, stock_id, wallet_tx_id, order_status, parent_stock_tx_id, is_buy, order_type, stock_price, quantity,  time_stamp)
	    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`, tx.StockTxID, userName, tx.StockID, wallet_tx_id, tx.Status, tx.ParentTxID, tx.IsBuy, tx.OrderType, *price, quantity, tx.TimeStamp)
	if err != nil {
		return fmt.Errorf("Failed to commit transaction: %w", err)
	}
	return nil
}

// Update db of Market price of a stock X to the last sold price of a stock X
// For UI display only, backend will NOT use the last sold price to find Market price
// Backend will use the top of the queue for the Market price 
func updateMarketStockPrice(stockID string, price *float64) error {
	
	fmt.Println("Starting Execution Service \n\n\n ===============")

	fmt.Println("Updating stock price")
	
	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	// Update the stock price
	_, err = db.Exec("UPDATE stocks SET current_price = $1 WHERE stock_id = $2", *price, stockID)
	if err != nil {
		return fmt.Errorf("Failed to update stock price: %w", err)
	}
	return nil
}

/** === END BUY/SELL support === **/

func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "token"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	router.POST("/executeSellTrade", executeSellTrade)
	router.POST("/executeBuyTrade", executeBuyTrade)

	router.Run(":5555")
}