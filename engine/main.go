package main
// TODO: seperate into module: queue, buy, sell, matching
// Clarification: getWalletTransactions and getStockTransactions - is_debit, wallet_tx_id, (duplicate) stock_tx_id

// TODO: seperate into module: queue, buy, sell, matching

import (
	"errors"
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
	// host     = "database"
	host     = "localhost" // for local testing
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
	ParentTxID *string  `json:"parent_tx_id"`
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
func (pq PriorityQueue) Less(i, j int) bool { 
	if pq.Order[i].Price == pq.Order[j].Price {
        return i < j
    }
	return pq.LessFunc(pq.Order[i].Price, pq.Order[j].Price) 
}
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
func printq(book *OrderBook) {
	// Print the order book after adding the order
	orderBookMap.mu.Lock()
	defer orderBookMap.mu.Unlock()
	fmt.Println("\n === Current Sell Queue === \n")
	book.SellOrders.Printn()
	fmt.Println("\n ====== \n")
	fmt.Println("\n === Current Buy Queue === \n")
	book.BuyOrders.Printn()
	fmt.Println("\n ====== \n")
}

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
	// return uuid.NewSHA1(uuid.Must(uuid.NewRandom()), []byte(userName)).String()
	return uuid.New().String()
}

func validateOrderType(request *PlaceStockOrderRequest) error {
    if request.OrderType == "MARKET" && request.Price != nil {
        return errors.New("Price must be null for market orders")
    } else if request.OrderType == "LIMIT" && request.Price == nil {
        return errors.New("Price must not be null for limit orders")
    }
    return nil
} // validateOrderType

func createOrder(request *PlaceStockOrderRequest, userName string) (Order, error) {
	order := Order{
		StockTxID:  generateOrderID(),
		StockID:    request.StockID,
		WalletTxID: generateWalletID(userName),
		ParentTxID: nil,
		IsBuy:      request.IsBuy != nil && *request.IsBuy,
		OrderType:  request.OrderType,
		Quantity:   request.Quantity,
		Price:      *request.Price,
		TimeStamp:  time.Now().Format(time.RFC3339Nano),
		Status:     "IN_PROGRESS",
	}
	return order, nil
} // createOrder

func HandlePlaceStockOrder(c *gin.Context) {
	user_name, exists := c.Get("user_name")
	if !exists || user_name == nil {
		handleError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userName, ok := user_name.(string)
	if !ok {
		handleError(c, http.StatusBadRequest, "Invalid user name type", nil)
		return
	}

	var request PlaceStockOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	if err := validateOrderType(&request); err != nil {
        handleError(c, http.StatusBadRequest, err.Error(), nil)
        return
    }

	order, e := createOrder(&request, userName)
	if e != nil {
		handleError(c, http.StatusInternalServerError, "Failed to create order", e)
		return
	}

	if order.IsBuy {
		if err := deductMoneyFromWallet(userName, order); err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to deduct money from user's wallet", err)
			return
		}

		if err := InsertWalletTransaction(userName, order); err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to insert wallet transaction", err)
			return
		}

		if err := InsertStockTransaction(userName, order); err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to insert stock transaction", err)
			return
		}

		book, bookerr := initializePriorityQueue(order)
		if bookerr != nil {
			handleError(c, http.StatusInternalServerError, "Failed to push order to priority queue", bookerr)
			return
		}

		trade := processBuyOrder(book, order)

		fmt.Println("====")
		fmt.Println(trade)
		fmt.Println("====")
		printq(book)
	} else {
		book, bookerr := initializePriorityQueue(order)
		if bookerr != nil {
			handleError(c, http.StatusInternalServerError, "Failed to push order to priority queue", bookerr)
			return
		}

		trade := processSellOrder(book, order)

		fmt.Println("====")
		fmt.Println(trade)
		fmt.Println("====")
		printq(book)
	}

	// Update the user's stock quantity in the database
	err := updateUserStockQuantity(userName, order.StockID, order.Quantity, order.IsBuy)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to update user stock quantity", err)
		return
	}

	response := PlaceStockOrderResponse{
		Success: true,
		Data:    nil,
	}

	c.IndentedJSON(http.StatusOK, response)
} // HandlePlaceStockOrder

// updateUserStockQuantity updates the user's stock quantity in the database
func updateUserStockQuantity(userName string, stockID int, quantity int, isBuy bool) error {
	var query string

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

// Define the structure of the order book map
type OrderBookMap struct {
	OrderBooks map[int]*OrderBook // Map of stock ID to order book
	mu         sync.Mutex         // Mutex to synchronize access to the map
}

// Initialize the order book map
var orderBookMap = OrderBookMap{
	OrderBooks: make(map[int]*OrderBook),
}

/** === BUY Order === **/
func deductMoneyFromWallet(userName string, order Order) error {
	fmt.Println("Deducting money from wallet")
	return nil
}


func InsertWalletTransaction(userName string, order Order) error {
	fmt.Println("Inserting wallet transaction")
	return nil
}

func InsertStockTransaction(userName string, order Order) error {
	fmt.Println("Inserting stock transaction")
	return nil
}

func processBuyOrder(book *OrderBook, order Order) (trade *Order) {	
	// If the buy order is a market order, match it with the lowest sell order
	if order.OrderType == "MARKET" {
		// trade = matchMarketBuyOrder(book, order)
	} else {
		trade = matchLimitBuyOrder(book, order)
	}
	return trade
}

func matchLimitBuyOrder(book *OrderBook, order Order) (trade *Order) {
	book.mu.Lock()
	defer book.mu.Unlock()

	// If the buy order is a limit order, match it with the lowest sell order that is less than or equal to the buy order price
	for order.Quantity > 0 && book.SellOrders.Len() > 0 {
		fmt.Println("Try matching limit buy order:")
		// Get the lowest sell order
		lowestSellOrder := heap.Pop(&book.SellOrders).(*Order)

		// If the lowest sell order price is less than or equal to the buy order price, execute the trade
		if lowestSellOrder.Price <= order.Price {
			trade = executeLimitTrade(book, &order, lowestSellOrder)
			if trade != nil {
				return trade
			}
			fmt.Println("Trade executed")
			fmt.Println("Buy Order: ", order)
			fmt.Println("Sell Order: ", lowestSellOrder)
			return lowestSellOrder
		} else {
			// If the lowest sell order price is greater than the buy order price, put it back in the sell queue
			fmt.Println("No match found, putting back in the buy queue")
			heap.Push(&book.SellOrders, lowestSellOrder)
			break
		}
	}
	

	// If no match is found or partial fulfill, add the buy order to the buy queue
	heap.Push(&book.BuyOrders, &order)
	return nil
}


/** === END BUY Order === **/


/** === SELL Order === **/
func processSellOrder(book *OrderBook, order Order) (trade *Order) {
	// If the sell order is a market order, match it with the highest buy order
	if order.OrderType == "MARKET" {
		// trade = matchMarketSellOrder(book, order)
	} else {
		trade = matchLimitSellOrder(book, order)
	}
	return trade
}

func matchLimitSellOrder(book *OrderBook, order Order) (trade *Order) {
	book.mu.Lock()
	defer book.mu.Unlock()
	for book.BuyOrders.Len() > 0 {}

	// If the sell order is a limit order, match it with the highest buy order that is greater than or equal to the sell order price
	for order.Quantity > 0 && book.BuyOrders.Len() > 0 {
		fmt.Println("Try matching limit sell order:")
		// Get the highest buy order
		highestBuyOrder := heap.Pop(&book.BuyOrders).(*Order)

		// If the highest buy order price is greater than or equal to the sell order price, execute the trade
		if highestBuyOrder.Price >= order.Price {
			trade = executeLimitTrade(book, highestBuyOrder, &order)
			if trade != nil {
				return trade
			}
			fmt.Println("Trade executed")
			fmt.Println("Buy Order: ", highestBuyOrder)
			fmt.Println("Sell Order: ", order)
			return highestBuyOrder
		} else {
			// If the highest buy order price is less than the sell order price, put it back in the buy queue
			fmt.Println("No match found, putting back in the buy queue")
			heap.Push(&book.BuyOrders, highestBuyOrder)
			break
		}
	}

	// If no match is found, add the sell order to the sell queue
	heap.Push(&book.SellOrders, &order)

	return nil
}
/** === END SELL Order === **/

/** === BUY/SELL Order === **/

func executeLimitTrade(book *OrderBook, buyOrder *Order, sellOrder *Order) (trade *Order) {
	tradeQuantity := min(buyOrder.Quantity, sellOrder.Quantity)
	if buyOrder.Quantity > sellOrder.Quantity {
		// execute partial trade for buy order and complete trade for sell order
		buyOrder.Quantity -= tradeQuantity
		heap.Push(&book.BuyOrders, buyOrder)

	} else if buyOrder.Quantity < sellOrder.Quantity  {
		// execute partial trade for sell order and complete trade for buy order
		sellOrder.Quantity -= tradeQuantity
		heap.Push(&book.SellOrders, sellOrder)
	} else {
		// execute complete trade for both buy and sell orders
		buyOrder.Quantity = 0
		sellOrder.Quantity = 0
	}

	return trade
}

func initializePriorityQueue(order Order) (*OrderBook, error) {
	// Add the order to the order book corresponding to the stock ID
	orderBookMap.mu.Lock()
	defer orderBookMap.mu.Unlock()
	book, ok := orderBookMap.OrderBooks[order.StockID]
	if !ok {
		// If the order book for this stock does not exist, create a new one
		book = &OrderBook{
			BuyOrders:  PriorityQueue{Order: make([]*Order, 0), LessFunc: highPriorityLess},
			SellOrders: PriorityQueue{Order: make([]*Order, 0), LessFunc: lowPriorityLess},
		}
		orderBookMap.OrderBooks[order.StockID] = book
	}
	return book, nil
}

/** === END BUY/SELL Order === **/

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
