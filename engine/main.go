package main

import (
	"container/heap"
	"database/sql"
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
	host     = "database"
	port     = 5432
	user     = "nt_user"
	password = "db123"
	dbname   = "nt_db"
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

type Order struct {
	ID         string  `json:"id"`
	StockID    int     `json:"stock_id"`
	IsBuy      bool    `json:"is_buy"`
	OrderType  string  `json:"order_type"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
	Timestamp  int64   `json:"timestamp"`
	Status     string  `json:"status"`
}

// Define the order book
type OrderBook struct {
	BuyOrders  PriorityQueue
	SellOrders PriorityQueue
	mu         sync.Mutex
}

// PriorityQueue is a min-heap of orders.
type PriorityQueue []*Order

// Len is the number of elements in the collection.
func (pq PriorityQueue) Len() int { return len(pq) }

// index i should sort before the element with index j.
func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than for price.
	return pq[i].Price < pq[j].Price
}

// Swap the elements with indexes i and j.
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

// Push pushes the element x onto the heap.
func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Order)
	*pq = append(*pq, item)
}

// Pop removes and returns the minimum element from the heap.
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// generateOrderID generates a unique ID for each order
func generateOrderID() string {
    id := uuid.New()
    return id.String()
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
		ID:         generateOrderID(),
		StockID:    request.StockID,
		IsBuy:      request.IsBuy != nil && *request.IsBuy,
		OrderType:  request.OrderType,
		Quantity:   request.Quantity,
		Price:      *request.Price,
		Timestamp:  time.Now().UnixNano(),
		Status:     "IN_PROGRESS", // Set the initial status to IN_PROGRESS
	}

	// Add the order to the order book corresponding to the stock ID
	orderBookMap.mu.Lock()
	book, ok := orderBookMap.OrderBooks[order.StockID]
	if !ok {
		// If the order book for this stock does not exist, create a new one
		book = &OrderBook{
			BuyOrders:  make(PriorityQueue,  0),
			SellOrders: make(PriorityQueue,  0),
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
	err := updateUserStockQuantity(userName.(string), order.StockID, order.Quantity)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to update user stock quantity", err)
		return
	}

	// Print the order book after adding the order
	orderBookMap.mu.Lock()
	defer orderBookMap.mu.Unlock()
	fmt.Println("Order Books:")
	for stockID, book := range orderBookMap.OrderBooks {
		fmt.Printf("Stock ID: %d\n", stockID)
		fmt.Println("Buy Orders:")
		for _, order := range book.BuyOrders {
			fmt.Printf("ID: %s, StockID: %d, Price: %.2f, Quantity: %d, Status: %s\n", order.ID, order.StockID, order.Price, order.Quantity, order.Status)
		}
		fmt.Println("Sell Orders:")
		for _, order := range book.SellOrders {
			fmt.Printf("ID: %s, StockID: %d, Price: %.2f, Quantity: %d, Status: %s\n", order.ID, order.StockID, order.Price, order.Quantity, order.Status)
		}
	}

	response := PlaceStockOrderResponse{
		Success: true,
		Data:    nil,
	}

	c.IndentedJSON(http.StatusOK, response)
}


// updateUserStockQuantity updates the user's stock quantity in the database
func updateUserStockQuantity(userName string, stockID int, quantity int) error {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname))
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE user_stocks SET quantity = quantity - $1 WHERE user_name = $2 AND stock_id = $3", quantity, userName, stockID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
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

	router.POST("/placeStockOrder", identification.Identification, HandlePlaceStockOrder)

	router.Run(":8585")
}
