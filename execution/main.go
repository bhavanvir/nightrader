package main

import (
    "database/sql"
    "fmt"
	"encoding/json"
    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"
	"github.com/google/uuid"
	"time"
	"os"
	"os/signal"
	"syscall"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Global variable for the database connection
var user_db *sql.DB
var stock_db *sql.DB
var tx_db *sql.DB

var (
    stmtUpdateWalletTransaction *sql.Stmt
    stmtUpdateMoneyWallet       *sql.Stmt
    stmtCheckUserStocks         *sql.Stmt
    stmtDeleteUserStocks        *sql.Stmt
    stmtInsertUserStocks        *sql.Stmt
    stmtSetWalletTransaction    *sql.Stmt
    stmtGetWalletTransactionsAmount *sql.Stmt
    stmtSetStockTransaction     *sql.Stmt
    stmtSetStatus               *sql.Stmt
    stmtUpdateWalletTxId        *sql.Stmt
    stmtUpdateMarketStockPrice        *sql.Stmt
    stmtUpdateUserStocks              *sql.Stmt
    stmtCheckWalletTransaction      *sql.Stmt
)

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

    namespaceUUID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
)

const (
	rabbitHost = "rabbitmq3"
	// rabbitHost     = "localhost" // for local testing
	rabbitPort     = "5672"
	rabbitUser     = "guest"
	rabbitPassword = "guest"
    rabbitRoutingKey = "order_execution_queue"
)

var (
    rabbitMQConnect *amqp.Connection
    rabbitMQChannel *amqp.Channel
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
	IsBuyExecuted bool `json:"is_buy_executed"`
}

type ResponsePayload struct {
	BuyQuantity float64 `json:"buy_quantity"`
	SellQuantity float64 `json:"sell_quantity"`
	BuyStatus string `json:"buy_status"`
	SellStatus string `json:"sell_status"`
	IsBuyExecuted bool `json:"is_buy_executed"`
}

func createRabbitMQChannel() error {
	var err error
    if rabbitMQChannel == nil {
        rabbitMQChannel, err = rabbitMQConnect.Channel()
        if err != nil {
			return fmt.Errorf("Failed to open a channel: %v", err)
        }
    }
	return nil
}

func createRabbitMQQueue(queueName string) error {
    _, err := rabbitMQChannel.QueueDeclare(queueName, false, false, false, false, nil)
    if err != nil {
        fmt.Println("Failed to declare a queue: ", err)
		return err
    }
    return nil
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
	refundAmount := (*buyPrice - *sellPrice) * float64(tradeQuantity)

	if refundAmount > 0 {
		newWalletTxAmount, err := getWalletTransactionsAmount(order.UserName, order.WalletTxID)
		if err != nil {
			fmt.Println("Error getting wallet transaction amount: ", err)
		}
		
		// Refund deducted money to the Buy user's wallet, adjusting for any price differences
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

	// setWalletTransaction should always be before the setStockTransaction
	if err := setWalletTransaction(sellOrder.UserName, completedOrder.WalletTxID, completedOrder.TimeStamp, sellPrice, tradeQuantity, true); err != nil {
		fmt.Println("Error setting wallet transaction: ", err)
	}

	if err := setStockTransaction(sellOrder.UserName, completedOrder, sellPrice, tradeQuantity); err != nil {
		fmt.Println("Error setting stock transaction: ", err)
	}
}

func completeBuyOrder(buyOrder *Order, tradeQuantity float64, buyPrice *float64, sellPrice *float64) {
	refundAmount := (*buyPrice - *sellPrice) * float64(tradeQuantity)

	if refundAmount > 0 {
		totalSoldAmount := (*sellPrice) * float64(tradeQuantity)

		// Refund deducted money to the Buy user's wallet, adjusting for any price differences
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

func executeTrade() {
	msgs, err := rabbitMQChannel.Consume(
		rabbitRoutingKey, // queue
		"",        // consumer
		true,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		fmt.Printf("Failed to consume messages: %v", err)
	}

	// Consume messages in a separate goroutine
	go func() {
		for msg := range msgs {
			var payload TradePayload
			err := json.Unmarshal(msg.Body, &payload)
			if err != nil {
				// Handle JSON parsing error
				fmt.Println("Failed to parse JSON:", err)
				continue // Move to the next message
			}

			buyOrder := &payload.BuyOrder
			sellOrder := &payload.SellOrder

			// Handle the buy trade execution here
			if payload.IsBuyExecuted {
				handleBuyTrade(buyOrder, sellOrder)
			} else {
				handleSellTrade(buyOrder, sellOrder)
			}

			// Send response message
			responsePayload := ResponsePayload{
				BuyQuantity: buyOrder.Quantity,
				SellQuantity: sellOrder.Quantity,
				BuyStatus: buyOrder.Status,
				SellStatus: sellOrder.Status,
				IsBuyExecuted: payload.IsBuyExecuted,
			}

			responseDataBytes, err := json.Marshal(responsePayload)
			if err != nil {
				fmt.Println("Failed to marshal response data:", err)
			}

			err = rabbitMQChannel.Publish(
				"",                // exchange
				"response_queue",  // routing key
				false,             // mandatory
				false,             // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        responseDataBytes,
				},
			)
			if err != nil {
				fmt.Println("Failed to publish response message:", err)
			}
		}
	}()

	fmt.Println("Press Ctrl + C to exit")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("Shutting down...")
}

/** === BUY/SELL support === **/

func getWalletTransactionsAmount(userName string, walletTxID string) (float64, error) {
    // Query the database to get the total amount of wallet transactions for the specified user and wallet ID
    var totalAmount float64
    err := stmtGetWalletTransactionsAmount.QueryRow(userName, walletTxID).Scan(&totalAmount)
    if err != nil {
        return 0, fmt.Errorf("Failed to get wallet transactions amount: %w", err)
    }

    return totalAmount, nil
}

func updateMoneyWallet(userName string, amount float64, isAdded bool) error {
    // Adjust the amount based on the transaction type
    if !isAdded {
        amount *= -1 // Deduct funds if buying
    }
    _, err := stmtUpdateMoneyWallet.Exec(amount, userName)
    if err != nil {
        return fmt.Errorf("Failed to update wallet: %w", err)
    }
    return nil
}

func updateWalletTransaction(userName string, order Order, amount float64) error {
    // Update the wallet transaction
    _, err := stmtUpdateWalletTransaction.Exec(amount, userName, order.WalletTxID)
    if err != nil {
        return fmt.Errorf("Failed to update wallet transaction: %w", err)
    }

    return nil
}

func updateStockPortfolio(userName string, order Order, quantity float64, isAdded bool) error {
    // Calculate the total quantity to be added or deducted
    total := quantity
    if !isAdded {
        total *= -1 // Reduce stocks if selling
    }

    // Check if user already owns this stock
    var currentQuantity float64
    err := stmtCheckUserStocks.QueryRow(userName, order.StockID).Scan(&currentQuantity)
    if err != nil && err != sql.ErrNoRows {
        return fmt.Errorf("Failed to query user stocks: %w", err)
    }

    if currentQuantity+total <= 0 {
        // Delete user's stock if the total quantity becomes zero or negative
        _, err = stmtDeleteUserStocks.Exec(userName, order.StockID)
    } else if currentQuantity > 0 {
        // Update user's stock quantity
        _, err = stmtUpdateUserStocks.Exec(total, userName, order.StockID)
    } else {
        // Insert new user's stock
        _, err = stmtInsertUserStocks.Exec(userName, order.StockID, quantity)
    }
    if err != nil {
        return fmt.Errorf("Failed to update user stocks: %w", err)
    }
    return nil
}

func setStatus(order *Order, status string, isUpdateWalletTxId bool) error {
    if status == "PARTIAL_FULFILLED" {
        order.Status = status
    }

    // Insert transaction to wallet transactions
    _, err := stmtSetStatus.Exec(status, order.UserName, order.StockTxID)
    if err != nil {
        return fmt.Errorf("Failed to update status: %w", err)
    }

    // assign wallet_tx_id to stock_tx_id if the Sell order is completed
    if isUpdateWalletTxId {
        _, err = stmtUpdateWalletTxId.Exec(order.WalletTxID, order.UserName, order.StockTxID)
    }

    return nil
}

// Store completed wallet transactions based on order matched
func setWalletTransaction(userName string, walletTxID string, timestamp string, price *float64, quantity float64, isAdded bool) error {
    amount := *price * quantity // Calculate transaction amount
    isDebit := !isAdded         // Determine if it's a debit transaction

    _, err := stmtSetWalletTransaction.Exec(walletTxID, userName, isDebit, amount, timestamp)
    if err != nil {
        return fmt.Errorf("Failed to commit wallet transaction: %w", err)
    }
    return nil
}

// Store transaction based on the order user created
func setStockTransaction(userName string, tx Order, price *float64, quantity float64) error {
    // Check if a wallet transaction has been made for this order yet
    rows, err := stmtCheckWalletTransaction.Query(userName, tx.WalletTxID)
    if err != nil {
        return fmt.Errorf("Error querying wallet transactions: %w", err)
    }
    defer rows.Close()

    var wallet_tx_id *string

    // if a wallet transaction is found in wallet_transaction table db, then add it to stock_transaction table OR,
    // if status is COMPLETED, the stock transaction need to a wallet transaction
    if rows.Next() || tx.Status == "COMPLETED" {
        wallet_tx_id = &tx.WalletTxID
    }

    // Insert transaction to stock transactions
    _, err = stmtSetStockTransaction.Exec(tx.StockTxID, userName, tx.StockID, wallet_tx_id, tx.Status, tx.ParentTxID, tx.IsBuy, tx.OrderType, *price, quantity, tx.TimeStamp)
    if err != nil {
        return fmt.Errorf("Failed to commit transaction: %w", err)
    }
    return nil
}

// Update db of Market price of a stock X to the last sold price of a stock X
// For UI display only, backend will NOT use the last sold price to find Market price
// Backend will use the top of the queue for the Market price 
func updateMarketStockPrice(stockID string, price *float64) error {
    // Update the stock price
    _, err := stmtUpdateMarketStockPrice.Exec(*price, stockID)
    if err != nil {
        return fmt.Errorf("Failed to update stock price: %w", err)
    }
    return nil
}

/** === END BUY/SELL support === **/

func prepareStatements() error {
    var err error

    stmtUpdateWalletTransaction, err = tx_db.Prepare(`
        UPDATE wallet_transactions SET amount = $1 WHERE user_name = $2 AND wallet_tx_id = $3`)
    if err != nil {
        return fmt.Errorf("failed to prepare update wallet transaction statement: %v", err)
    }

    stmtUpdateMoneyWallet, err = user_db.Prepare(`
        UPDATE users SET wallet = wallet + $1 WHERE user_name = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare update money wallet statement: %v", err)
    }

    stmtCheckUserStocks, err = stock_db.Prepare(`
        SELECT quantity FROM user_stocks WHERE user_name = $1 AND stock_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare check user stocks statement: %v", err)
    }

    stmtDeleteUserStocks, err = stock_db.Prepare(`
        DELETE FROM user_stocks WHERE user_name = $1 AND stock_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare delete user stocks statement: %v", err)
    }

    stmtInsertUserStocks, err = stock_db.Prepare(`
        INSERT INTO user_stocks VALUES ($1, $2, $3)`)
    if err != nil {
        return fmt.Errorf("failed to prepare insert user stocks statement: %v", err)
    }

    stmtSetWalletTransaction, err = tx_db.Prepare(`
        INSERT INTO wallet_transactions (wallet_tx_id, user_name, is_debit, amount, time_stamp)
        VALUES ($1, $2, $3, $4, $5)`)
    if err != nil {
        return fmt.Errorf("failed to prepare set wallet transaction statement: %v", err)
    }

    stmtGetWalletTransactionsAmount, err = tx_db.Prepare(`
        SELECT SUM(amount) FROM wallet_transactions WHERE user_name = $1 AND wallet_tx_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare get wallet transactions amount statement: %v", err)
    }

    stmtSetStockTransaction, err = tx_db.Prepare(`
        INSERT INTO stock_transactions (stock_tx_id, user_name, stock_id, wallet_tx_id, order_status, parent_stock_tx_id, is_buy, order_type, stock_price, quantity,  time_stamp)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`)
    if err != nil {
        return fmt.Errorf("failed to prepare set stock transaction statement: %v", err)
    }

    stmtSetStatus, err = tx_db.Prepare(`
        UPDATE stock_transactions SET order_status = $1 WHERE user_name = $2 AND stock_tx_id = $3`)
    if err != nil {
        return fmt.Errorf("failed to prepare set status statement: %v", err)
    }

    stmtUpdateWalletTxId, err = tx_db.Prepare(`
        UPDATE stock_transactions SET wallet_tx_id = $1 WHERE user_name = $2 AND stock_tx_id = $3`)
    if err != nil {
        return fmt.Errorf("failed to prepare update wallet transaction ID statement: %v", err)
    }

    stmtUpdateMarketStockPrice, err = stock_db.Prepare(`
        UPDATE stocks SET current_price = $1 WHERE stock_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare update market stock price statement: %v", err)
    }

    stmtUpdateUserStocks, err = stock_db.Prepare(`
        UPDATE user_stocks SET quantity = quantity + $1 WHERE user_name = $2 AND stock_id = $3`)
    if err != nil {
        return fmt.Errorf("failed to prepare update user stocks statement: %v", err)
    }

    stmtCheckWalletTransaction, err = tx_db.Prepare(`
		SELECT wallet_tx_id FROM wallet_transactions WHERE user_name = $1 AND wallet_tx_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare check wallet transaction statement: %v", err)
    }

    return nil
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
    defer stmtUpdateWalletTransaction.Close()
    defer stmtUpdateMoneyWallet.Close()
    defer stmtCheckUserStocks.Close()
    defer stmtDeleteUserStocks.Close()
    defer stmtInsertUserStocks.Close()
    defer stmtSetWalletTransaction.Close()
    defer stmtGetWalletTransactionsAmount.Close()
    defer stmtSetStockTransaction.Close()
    defer stmtSetStatus.Close()
    defer stmtUpdateWalletTxId.Close()
    defer stmtUpdateUserStocks.Close()
    defer stmtCheckWalletTransaction.Close()

    user_db.SetMaxOpenConns(10) // Set maximum number of open connections
    user_db.SetMaxIdleConns(5) // Set maximum number of idle connections

    stock_db.SetMaxOpenConns(10) // Set maximum number of open connections
    stock_db.SetMaxIdleConns(5) // Set maximum number of idle connections

    tx_db.SetMaxOpenConns(10) // Set maximum number of open connections
    tx_db.SetMaxIdleConns(5) // Set maximum number of idle connections

	if rabbitMQConnect == nil {
        amqpURI := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitUser, rabbitPassword, rabbitHost, rabbitPort)
        rabbitMQConnect, err = amqp.Dial(amqpURI)
        if err != nil {
			fmt.Println("Failed to connect to RabbitMQ", err)
			return
        }
    }
    err = createRabbitMQChannel()
    if err != nil {
        fmt.Println("Failed to create RabbitMQ channels", err)
        return
    }

	err = createRabbitMQQueue("response_queue")
    if err != nil {
        fmt.Println("Failed to create RabbitMQ queues", err)
        return
    }
	
    defer rabbitMQConnect.Close()
    defer rabbitMQChannel.Close()

	executeTrade()
}