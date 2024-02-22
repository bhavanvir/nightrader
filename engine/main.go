package main

// TODO: seperate into module: queue, buy, sell, matching

import (
	"container/heap"
	"database/sql"
	"github.com/gin-contrib/cors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Poomon001/day-trading-package/identification"
	_ "github.com/lib/pq"
	"github.com/google/uuid"
)

const (
	// Use localhost for local testing
	host     = "database"
	port     = 5432
	user     = "nt_user"
	password = "db123"
	dbname   = "nt_db"

	namespaceUUID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
)

// Define the structure of the request body for placing a stock order
type PlaceStockOrderRequest struct {
	StockID    int     `json:"stock_id" binding:"required"`
	IsBuy      *bool   `json:"is_buy" binding:"required"`
	OrderType  string  `json:"order_type" binding:"required"`
	Quantity   int     `json:"quantity" binding:"required"`
	Price      *float64 `json:"price"`
}

// Define the structure of the response body for placing a stock order
type PlaceStockOrderResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// Define the structure of the request body for cancelling a stock transaction
type CancelStockTransactionRequest struct {
	StockTxID string `json:"stock_tx_id" binding:"required"`
}

// Define the structure of the response body for cancelling a stock transaction
type CancelStockTransactionResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type Order struct {
	StockTxID  string  `json:"stock_tx_id"`
	StockID    int     `json:"stock_id"`
	WalletTxID string  `json:"wallet_tx_id"`
	IsBuy      bool    `json:"is_buy"`
	OrderType  string  `json:"order_type"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
	TimeStamp  string  `json:"time_stamp"`
	Status     string  `json:"status"`
}

// Define the order book
type OrderBook struct {
	BuyOrders  PriorityQueue
	SellOrders PriorityQueue
	mu         sync.Mutex
}

// PriorityQueue
type PriorityQueue struct {
	Order []*Order
	LessFunc func(i, j float64) bool
}

// handleError is a helper function to send error responses
func handleError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := map[string]interface{}{
		"success": false,
		"data":    nil,
		"message": message,
	}
	if err != nil {
		errorResponse["message"] = fmt.Sprintf("%s: %v", message, err)
	}
	c.JSON(statusCode, errorResponse)
}

func openConnection() (*sql.DB, error) {
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	return sql.Open("postgres", postgresqlDbInfo)
}

/** standard heap interface **/
func (pq PriorityQueue) Len() int { return len(pq.Order) }
func (pq PriorityQueue) Swap(i, j int) { pq.Order[i], pq.Order[j] = pq.Order[j], pq.Order[i] }
func (pq PriorityQueue) Less(i, j int) bool { return pq.LessFunc(pq.Order[i].Price, pq.Order[j].Price) }
func highPriorityLess(i, j float64) bool { return i > j }
func lowPriorityLess(i, j float64) bool { return i < j }

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Order)
	pq.Order = append(pq.Order, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := pq.Order
	n := len(old)
	if n == 0 {
		return nil
	}
	item := old[n-1]
	pq.Order = old[0 : n-1]
	return item
}
/** standard heap interface END **/

// print the queue
func (pq *PriorityQueue) Printn() {
	temp := PriorityQueue{Order: make([]*Order, len(pq.Order)), LessFunc: pq.LessFunc}
	copy(temp.Order, pq.Order)
	for temp.Len() > 0 {
		item := heap.Pop(&temp).(*Order)
		fmt.Printf("Stock Tx ID: %s, StockID: %d, WalletTxID: %s, Price: %.2f, Quantity: %d, Status: %s, TimeStamp: %s\n", item.StockTxID, item.StockID, item.WalletTxID, item.Price, item.Quantity, item.Status, item.TimeStamp)
	}
}

// update modifies the priority and value in the queue
// func (pq *PriorityQueue) update(order *Order, quantity int, timestamp string, status string) {
// 	order.Quantity = quantity
// 	order.TimeStamp = timestamp
// 	order.Status = status
// 	heap.Fix(pq, order.StockTxID)
// }

// generateOrderID generates a unique ID for each order
func generateOrderID() string {
	return uuid.New().String()
}
	
// Generate a unique wallet ID for the user
func generateWalletID(userName string) string {
	return uuid.NewSHA1(uuid.Must(uuid.NewRandom()), []byte(userName)).String()
}

func HandlePlaceStockOrder(c *gin.Context) {
	userName, exists := c.Get("user_name")
	if !exists || userName == nil {
		handleError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var request PlaceStockOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if request.OrderType == "MARKET" && request.Price != nil {
		handleError(c, http.StatusBadRequest, "Price must be null for market orders", nil)
		return
	} else if request.OrderType == "LIMIT" && request.Price == nil {
		handleError(c, http.StatusBadRequest, "Price must not be null for limit orders", nil)
		return
	}

	// Create a new order
	order := Order{
		StockTxID:  generateOrderID(),
		StockID:    request.StockID,
		WalletTxID: generateWalletID(userName.(string)),
		IsBuy:      request.IsBuy != nil && *request.IsBuy,
		OrderType:  request.OrderType,
		Quantity:   request.Quantity,
		Price:      *request.Price,
		TimeStamp:  time.Now().Format(time.RFC3339Nano),
		Status:     "IN_PROGRESS",
	}

	fmt.Printf("Order: %+v\n", order)

	// Add the order to the order book corresponding to the stock ID
	orderBookMap.mu.Lock()
	book, ok := orderBookMap.OrderBooks[order.StockID]
	if !ok {
		// If the order book for this stock does not exist, create a new one
		book = &OrderBook{
			BuyOrders:  PriorityQueue{Order: make([]*Order, 0), LessFunc: highPriorityLess},
			SellOrders: PriorityQueue{Order: make([]*Order, 0), LessFunc: lowPriorityLess},
		}
		orderBookMap.OrderBooks[order.StockID] = book
	}
	orderBookMap.mu.Unlock()

	// Add the order to the appropriate queue in the order book
	book.mu.Lock()
	if order.IsBuy {
		heap.Push(&book.BuyOrders, &order)
	} else {
		heap.Push(&book.SellOrders, &order)
	}
	book.mu.Unlock()

	// Update the user's stock quantity in the database
	err := updateUserStockQuantity(userName.(string), order.StockID, order.Quantity, order.IsBuy)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to update user stock quantity", err)
		return
	}

	// Print the order book after adding the order
	orderBookMap.mu.Lock()
	defer orderBookMap.mu.Unlock()
	fmt.Println("\n === Current Sell Queue === \n")
	book.SellOrders.Printn()
	fmt.Println("\n ====== \n")
	fmt.Println("\n === Current Buy Queue === \n")
	book.BuyOrders.Printn()
	fmt.Println("\n ====== \n")

	response := PlaceStockOrderResponse{
		Success: true,
		Data:    nil,
	}

	c.IndentedJSON(http.StatusOK, response)
}

// updateUserStockQuantity updates the user's stock quantity in the database
func updateUserStockQuantity(userName string, stockID int, quantity int, isBuy bool) error {
	var query string

	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if isBuy {
		query = "UPDATE user_stocks SET quantity = quantity + $1 WHERE user_name = $2 AND stock_id = $3"
	} else {
		query = "UPDATE user_stocks SET quantity = quantity - $1 WHERE user_name = $2 AND stock_id = $3"
	}

	_, err = tx.Exec(query, quantity, userName, stockID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func TraverseOrderBook(StockTxID string, book *OrderBook, bookType string) (response CancelStockTransactionResponse) {
	response = CancelStockTransactionResponse{
		Success: false,
		Data:    nil,
	}

    var bookOrders *PriorityQueue
    if bookType == "buy" {
        bookOrders = &book.BuyOrders
    } else {
        bookOrders = &book.SellOrders
    }

    // Find the index of the order to remove
    indexToRemove := -1
    for i, order := range bookOrders.Order {
        if order.StockTxID == StockTxID && order.Status == "IN_PROGRESS" && order.OrderType == "LIMIT"{
            indexToRemove = i
            break
        }
    }

    // If the order was found, remove it from the heap
    if indexToRemove != -1 {
        heap.Remove(bookOrders, indexToRemove)
        response.Success = true
    }

	return response
}

func HandleCancelStockTransaction(c *gin.Context) {
    userName, exists := c.Get("user_name")
    if !exists || userName == nil {
        handleError(c, http.StatusUnauthorized, "User not authenticated", nil)
        return
    }

    var request CancelStockTransactionRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        handleError(c, http.StatusBadRequest, "Invalid request body", err)
        return
    }

    StockTxID := request.StockTxID

    orderBookMap.mu.Lock()
    defer orderBookMap.mu.Unlock()
    // Find which order book the order is in
    for _, book := range orderBookMap.OrderBooks {
        book.mu.Lock()
		defer book.mu.Unlock()

        foundBuy := TraverseOrderBook(StockTxID, book, "buy")
        foundSell := TraverseOrderBook(StockTxID, book, "sell")

		// Inside TraverseOrderBook, after removing the item
		fmt.Println("\n --- Current Sell Queue --- \n")
		book.SellOrders.Printn()
		fmt.Println("\n ------ \n")
		fmt.Println("\n --- Current Buy Queue --- \n")
		book.BuyOrders.Printn()
		fmt.Println("\n ------ \n")

		if foundBuy.Success || foundSell.Success {
			response := CancelStockTransactionResponse{
				Success: true,
				Data:    nil,
			}
			c.IndentedJSON(http.StatusOK, response)
			return
		}
    }

    handleError(c, http.StatusBadRequest, "Order not found", nil)
}

// Store completed stock transactions in the database
func setStockTransaction(c *gin.Context, tx Order) {
	userName, _ := c.Get("user_name")

	if userName == nil {
		handleError(c, http.StatusBadRequest, "Failed to obtain the user name", nil)
		return
	}

	// Connect to database
	db, err := openConnection()
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to connect to the database", err)
		return
	}
	defer db.Close()

	// TODO: How to determine is_debit? Default is true for now.
	// Insert transaction to wallet transactions
	_, err = db.Exec(`
		INSERT INTO wallet_transactions (wallet_tx_id, user_name, is_debit, amount)
		VALUES ($1, $2, $3, $4)`, tx.WalletTxID, userName, true, tx.Quantity)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to insert stock transaction", err)
		return
	}

	// Insert transaction to stock transactions
	_, err = db.Exec(`
		INSERT INTO stock_transactions (stock_tx_id, user_name, stock_id, wallet_tx_id, order_status, is_buy, order_type, stock_price, quantity)
	    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, tx.StockTxID, userName, tx.StockID, tx.WalletTxID, tx.Status, tx.IsBuy, tx.OrderType, tx.Price, tx.Quantity)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to insert stock transaction", err)
		return
	}
}

// Define the structure of the order book map
type OrderBookMap struct {
	OrderBooks map[int]*OrderBook // Map of stock ID to order book
	mu         sync.Mutex         // Mutex to synchronize access to the map
}

// Initialize the order book map
var orderBookMap = OrderBookMap{
	OrderBooks: make(map[int]*OrderBook),
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
	router.POST("/placeStockOrder", identification.Identification, HandlePlaceStockOrder)
	router.POST("/cancelStockTransaction", identification.Identification, HandleCancelStockTransaction)
	router.Run(":8585")
}
