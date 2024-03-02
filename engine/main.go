package main

import (
	"container/heap"
	"database/sql"
	"fmt"
	"github.com/gin-contrib/cors"
	"net/http"
	"sync"
	"time"

	"github.com/Poomon001/day-trading-package/identification"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	// host = "database"
	host     = "localhost" // for local testing
	port     = 5432
	user     = "nt_user"
	password = "db123"
	dbname   = "nt_db"

	namespaceUUID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
)

type ErrorResponse struct {
	Success bool              `json:"success"`
	Data    map[string]string `json:"data"`
}

// TODO: Why do we need *bool?
// Define the structure of the request body for placing a stock order
type PlaceStockOrderRequest struct {
	StockID   int      `json:"stock_id" binding:"required"`
	IsBuy     *bool    `json:"is_buy" binding:"required"`
	OrderType string   `json:"order_type" binding:"required"`
	Quantity  int      `json:"quantity" binding:"required"`
	Price     *float64 `json:"price"`
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
	ParentTxID *string `json:"parent_tx_id"`
	IsBuy      bool    `json:"is_buy"`
	OrderType  string  `json:"order_type"`
	Quantity   int     `json:"quantity"`
	Price      *float64 `json:"price"`
	TimeStamp  string  `json:"time_stamp"`
	Status     string  `json:"status"`
	UserName   string  `json:"user_name"`
}

// Define the order book
type OrderBook struct {
	BuyOrders  PriorityQueue
	SellOrders PriorityQueue
	mu         sync.Mutex
}

// PriorityQueue
type PriorityQueue struct {
	Order    []*Order
	LessFunc func(i, j float64) bool
}

// handleError is a helper function to send error responses
func handleError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := ErrorResponse{
		Success: false,
		Data:    map[string]string{"error": message + err.Error()},
	}
	c.IndentedJSON(statusCode, errorResponse)
}

func openConnection() (*sql.DB, error) {
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	return sql.Open("postgres", postgresqlDbInfo)
}

/** standard heap interface **/
func (pq PriorityQueue) Len() int      { return len(pq.Order) }
func (pq PriorityQueue) Swap(i, j int) { pq.Order[i], pq.Order[j] = pq.Order[j], pq.Order[i] }
func (pq PriorityQueue) Less(i, j int) bool { 
	if *pq.Order[i].Price == *pq.Order[j].Price {
		return pq.Order[i].TimeStamp < pq.Order[j].TimeStamp
	}
	return pq.LessFunc(*pq.Order[i].Price, *pq.Order[j].Price) 
}
func highPriorityLess(i, j float64) bool { return i > j }
func lowPriorityLess(i, j float64) bool  { return i < j }

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
		fmt.Printf("Stock Tx ID: %s, StockID: %d, WalletTxID: %s, Price: %.2f, Quantity: %d, Status: %s, TimeStamp: %s\n", item.StockTxID, item.StockID, item.WalletTxID, *item.Price, item.Quantity, item.Status, item.TimeStamp)
	}
}

// generateOrderID generates a unique ID for each order
func generateOrderID() string {
	return uuid.New().String()
}

// Generate a unique wallet ID for the user
func generateWalletID() string {
	return uuid.New().String()
}

func validateOrderType(request *PlaceStockOrderRequest) error {
	if request.OrderType == "MARKET" && request.Price != nil {
		return fmt.Errorf("Price must be null for market orders")
	} else if request.OrderType == "LIMIT" && request.Price == nil {
		return fmt.Errorf("Price must not be null for limit orders")
	}
	return nil
} // validateOrderType

func createOrder(request *PlaceStockOrderRequest, userName string) (Order, error) {
	order := Order{
		StockTxID:  generateOrderID(),
		StockID:    request.StockID,
		WalletTxID: generateWalletID(),
		ParentTxID: nil,
		IsBuy:      request.IsBuy != nil && *request.IsBuy,
		OrderType:  request.OrderType,
		Quantity:   request.Quantity,
		Price:      request.Price,
		TimeStamp:  time.Now().Format(time.RFC3339Nano),
		Status:     "IN_PROGRESS",
		UserName:   userName,
	}
	return order, nil
} // createOrder

var existingOrderIDs = make(map[int]struct{})

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
		handleError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	order, e := createOrder(&request, userName)
	if e != nil {
		handleError(c, http.StatusInternalServerError, "Failed to create order", e)
		return
	}

	book, bookerr := initializePriorityQueue(order)
	if bookerr != nil {
		handleError(c, http.StatusInternalServerError, "Failed to push order to priority queue", bookerr)
		return
	}

	if _, exists := existingOrderIDs[order.StockID]; !exists {
    updateMarketStockPrice(order.StockID, *order.Price)
    // Add the order ID to the hashset to mark it as processed
    existingOrderIDs[order.StockID] = struct{}{}
   	}

	// to be safe, lock here
	book.mu.Lock()
	defer book.mu.Unlock()

	if order.IsBuy {
		if err := verifyWalletBeforeTransaction(userName, order); err != nil {
			handleError(c, http.StatusBadRequest, "Failed to verify Wallet", err)
			return
		}
		
		if err := updateMoneyWallet(userName, order, order.Price, order.Quantity, false); err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to deduct money from user's wallet", err)
			return
		}

		if err := setWalletTransaction(userName, order, order.Price, order.Quantity, false); err != nil {
			handleError(c, http.StatusInternalServerError,  "Buy Order setWalletTx Error: " + err.Error(), err)
			return
		}

		if err := setStockTransaction(userName, order, order.Price, order.Quantity); err != nil {
			handleError(c, http.StatusInternalServerError, "Buy Order setStockTx Error: " + err.Error(), err)
			return
		}

		processOrder(book, order)

		printq(book)
	} else {
		if err := verifyStockBeforeTransaction(userName, order); err != nil {
			handleError(c, http.StatusBadRequest, "Failed to verify stocks", err)
			return
		}

		if err := updateStockPortfolio(userName, order, order.Quantity, false); err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to deduct stock from user's portfolio", err)
			return
		}

		if err := setStockTransaction(userName, order, order.Price, order.Quantity); err != nil {
			handleError(c, http.StatusInternalServerError, "Sell Order setStockTx Error: " + err.Error(), err)
			return
		}

		processOrder(book, order)

		printq(book)
	}

	response := PlaceStockOrderResponse{
		Success: true,
		Data:    nil,
	}

	c.IndentedJSON(http.StatusOK, response)
} // HandlePlaceStockOrder

func TraverseOrderBook(StockTxID string, book *OrderBook, bookType string) (response CancelStockTransactionResponse) {
	response = CancelStockTransactionResponse{
		Success: false,
		Data:    nil,
	}

	fmt.Println("Traversing order book:")

	var bookOrders *PriorityQueue
	if bookType == "buy" {
		bookOrders = &book.BuyOrders
	} else {
		bookOrders = &book.SellOrders
	}

	// Find the index of the order to remove
	indexToRemove := -1
	removeOrder := Order{}
	for i, order := range bookOrders.Order {
		if order.StockTxID == StockTxID && order.Status != "COMPLETED" && order.OrderType == "LIMIT" {
			indexToRemove = i
			removeOrder = *order
			break
		}
	}

	// If the order was found, remove it from the heap
	if indexToRemove != -1 {
		executeRemoveOrder(removeOrder, bookOrders, indexToRemove)
		response.Success = true
	}

	return response
}

func executeRemoveOrder(order Order, bookOrders *PriorityQueue, indexToRemove int) {
	fmt.Println("Removing order: ", order.StockTxID)
	heap.Remove(bookOrders, indexToRemove)

	if order.IsBuy {
		postprocessingRemoveBuyOrder(order)
	}else {
		postprocessingRemoveSellOrder(order)
	}
}

func postprocessingRemoveBuyOrder(order Order) {
	if order.Status == "IN_PROGRESS" {
		fmt.Println("Remove IN_PROGRESS buy order")
		// refund all dedeucted money back to wallet
		if err := updateMoneyWallet(order.UserName, order, order.Price, order.Quantity, true); err != nil {
			fmt.Println("Error updating wallet: ", err)
		}

		// remove transaction from wallet_transactions
		if err := deleteWalletTransaction(order.UserName, order); err != nil {
			fmt.Println("Error deleting wallet transaction: ", err)
		}

		// remove transaction from stock_transactions
		if err := deleteStockTransaction(order.UserName, order); err != nil {
			fmt.Println("Error deleting stock transaction: ", err)
		}
	} else {
		fmt.Println("Remove PARTIALLY_FULFILLED buy order")
		if err := updateMoneyWallet(order.UserName, order, order.Price, order.Quantity, true); err != nil {
			fmt.Println("Error updating wallet: ", err)
		}

		// remove transaction from stock_transactions
		if err := deleteWalletTransaction(order.UserName, order); err != nil {
			fmt.Println("Error deleting wallet transaction: ", err)
		}
	}
}

func postprocessingRemoveSellOrder(order Order) {
	fmt.Println("status: ", order.Status)	
	if order.Status == "IN_PROGRESS" {
		fmt.Println("Remove IN_PROGRESS sell order")
		// refund all dedeucted stock back to portfolio
		if err := updateStockPortfolio(order.UserName, order, order.Quantity, true); err != nil {
			fmt.Println("Error updating stock portfolio: ", err)
		}

		// remove transaction from stock_transactions
		if err := deleteStockTransaction(order.UserName, order); err != nil {
			fmt.Println("Error deleting stock transaction: ", err)
		}
	} else {
		fmt.Println("Remove PARTIALLY_FULFILLED sell order")
		if err := updateStockPortfolio(order.UserName, order, order.Quantity, true); err != nil {
			fmt.Println("Error updating stock portfolio: ", err)
		}
	}
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
// Assumption: match when price is equal
func matchLimitBuyOrder(book *OrderBook, order Order) {
	// Add the buy order to the buy queue
	heap.Push(&book.BuyOrders, &order)
	highestBuyOrder := heap.Pop(&book.BuyOrders).(*Order)

	// If the buy order is a limit order, match it with the lowest sell order that is less than or equal to the buy order price
	for highestBuyOrder.Quantity > 0 && book.SellOrders.Len() > 0 {
		lowestSellOrder := heap.Pop(&book.SellOrders).(*Order)

		// If the lowest sell order price is less than or equal to the buy order price, execute the trade
		if *lowestSellOrder.Price == *highestBuyOrder.Price {
			executeBuyTrade(book, highestBuyOrder , lowestSellOrder)
			db, err := openConnection()
            if err != nil {
                fmt.Printf("Failed to connect to database: %v\n", err)
                return
            }
    
            _, err = db.Exec("UPDATE stocks SET current_price = $1 WHERE stock_id = $2", *order.Price, order.StockID)
            if err != nil {
                fmt.Printf("Failed to update stock price: %v\n", err)
                return
            }
			fmt.Printf("\nTrade Executed - Buy Order: ID=%s, Quantity=%d, Price=$%.2f | Sell Order: ID=%s, Quantity=%d, Price=$%.2f\n", 
			highestBuyOrder.StockTxID, highestBuyOrder.Quantity, *highestBuyOrder.Price, lowestSellOrder.StockTxID, lowestSellOrder.Quantity, *lowestSellOrder.Price)
		} else {
			// If the lowest sell order price is greater than the buy order price, put it back in the sell queue
			fmt.Println("No match found, putting back in the buy queue")
			heap.Push(&book.SellOrders, lowestSellOrder)
			break
		}
	}

	// If the buy order was not fully executed, put it back in the buy queue
	if highestBuyOrder.Quantity > 0 {
		heap.Push(&book.BuyOrders, highestBuyOrder)
	}
}

/** 
	Assumption: There is always sufficient Sell LIMIT orders in the queue to match Buy MARKET order demands 
			  : The Sell order price will be the unchanged thoughout the trading process
			  	means there is enough Sell orders quantity at the exact MARKET price. 
			    Thus, no partial fulfillment at different prices.
	Error Handling: Refund money, remove transaction from wallet_transactions, stock_transactions, exit with error
**/
func matchMarketBuyOrder(book *OrderBook, order Order) {
	if book.SellOrders.Len() <= 0 {
		// refund money, remove transaction from wallet_transactions, stock_transactions, exit with error
		fmt.Println("Cancel order: No Sell orders available")
	}
	// Match the buy order with the lowest Sell order that is less than or equal to the buy order price
	for order.Quantity > 0 && book.SellOrders.Len() > 0 {
		lowestSellOrder := heap.Pop(&book.SellOrders).(*Order)
		executeBuyTrade(book, &order, lowestSellOrder)
		// db, err := openConnection()
		// if err != nil {
		// 	fmt.Printf("Failed to connect to database: %v\n", err)
		// 	return
		// }

		// _, err = db.Exec("UPDATE stocks SET current_price = $1 WHERE stock_id = $2", *order.Price, order.StockID)
		// if err != nil {
		// 	fmt.Printf("Failed to update stock price: %v\n", err)
		// 	return
		// }
		fmt.Printf("\nTrade Executed - Buy Order: ID=%s, Quantity=%d, Price=$%.2f | Sell Order: ID=%s, Quantity=%d, Price=$%.2f\n", 
		order.StockTxID, order.Quantity, *lowestSellOrder.Price, lowestSellOrder.StockTxID, lowestSellOrder.Quantity, *lowestSellOrder.Price)
	}
}

func executeBuyTrade(book *OrderBook, order *Order, sellOrder *Order){
	tradeQuantity := min(order.Quantity, sellOrder.Quantity)
	tradingPrice := sellOrder.Price

	updateMarketStockPrice(order.StockID, *tradingPrice)

	if order.Quantity > sellOrder.Quantity {
		// execute partial trade for buy order and complete trade for sell order
		order.Quantity -= tradeQuantity
		sellOrder.Quantity = 0
		partialFulfillBuyOrder(book, order, tradeQuantity)
		completeSellOrder(book, sellOrder, tradeQuantity, tradingPrice)
	} else if order.Quantity < sellOrder.Quantity {
		// execute partial trade for sell order and complete trade for buy order
		sellOrder.Quantity -= tradeQuantity
		order.Quantity = 0
		heap.Push(&book.SellOrders, sellOrder)
		completeBuyOrder(book, order, tradeQuantity)
		partialFulfillSellOrder(book, sellOrder, tradeQuantity)
	} else {
		// execute complete trade for both buy and sell orders
		order.Quantity = 0
		sellOrder.Quantity = 0
		completeBuyOrder(book, order, tradeQuantity)
		completeSellOrder(book, sellOrder, tradeQuantity, tradingPrice)
	}
}

func completeBuyOrder(book *OrderBook, buyOrder *Order, tradeQuantity int) {
	fmt.Println("Buy User: ", buyOrder.UserName)
	if err:= updateStockPortfolio(buyOrder.UserName, *buyOrder, tradeQuantity, true); err != nil {
		fmt.Println("Error updating stock portfolio: ", err)
	}

	if err := setStatus(buyOrder, "COMPLETED", false); err != nil {
		fmt.Println("Error setting status: ", err)
	}
}
/** === END BUY Order === **/

/** === SELL Order === **/
// Assumption: match when price is equal
func matchLimitSellOrder(book *OrderBook, order Order) {
	// Add the Sell order to the sell queue
	heap.Push(&book.SellOrders, &order)
	lowestSellOrder := heap.Pop(&book.SellOrders).(*Order)

	for lowestSellOrder.Quantity > 0 && book.BuyOrders.Len() > 0 {
		highestBuyOrder := heap.Pop(&book.BuyOrders).(*Order)

		if *highestBuyOrder.Price == *lowestSellOrder.Price {
			executeSellTrade(book, highestBuyOrder, lowestSellOrder)
			db, err := openConnection()
            if err != nil {
                fmt.Printf("Failed to connect to database: %v\n", err)
                return
            }
    
            _, err = db.Exec("UPDATE stocks SET current_price = $1 WHERE stock_id = $2", *order.Price, order.StockID)
            if err != nil {
                fmt.Printf("Failed to update stock price: %v\n", err)
                return
            }
			fmt.Printf("\nTrade Executed - Buy Order: ID=%s, Quantity=%d, Price=$%.2f | Sell Order: ID=%s, Quantity=%d, Price=$%.2f\n", 
			highestBuyOrder.StockTxID, highestBuyOrder.Quantity, *highestBuyOrder.Price, lowestSellOrder.StockTxID, lowestSellOrder.Quantity, *lowestSellOrder.Price)
		} else {
			fmt.Println("No match found, putting back in the buy queue")
			heap.Push(&book.BuyOrders, highestBuyOrder)
			break
		}
	}

	if lowestSellOrder.Quantity > 0 {
		heap.Push(&book.SellOrders, lowestSellOrder)
	}
}

/** 
	Assumption: There is always sufficient Buy LIMIT orders in the queue to match Sell MARKET order demands 
			  : The Buy order price will be the unchanged thoughout the trading process
			  	means there is enough Buy orders quantity at the exact MARKET price. 
			    Thus, no partial fulfillment at different prices.
	Error Handling: Refund stock, remove stock_transactions, exit with error
**/
func matchMarketSellOrder(book *OrderBook, order Order) {
	if book.BuyOrders.Len() <= 0 {
		// refund stock, remove transaction from wallet_transactions, stock_transactions, exit with error
		fmt.Println("Cancel order: No Sell orders available")
	}

	// Match the Sell order with the highest Buy order that is greater than or equal to the sell order price
	for order.Quantity > 0 && book.BuyOrders.Len() > 0 {
		highestBuyOrder := heap.Pop(&book.BuyOrders).(*Order)

		executeSellTrade(book, highestBuyOrder, &order)
		db, err := openConnection()
		if err != nil {
			fmt.Printf("Failed to connect to database: %v\n", err)
			return
		}

		_, err = db.Exec("UPDATE stocks SET current_price = $1 WHERE stock_id = $2", *order.Price, order.StockID)
		if err != nil {
			fmt.Printf("Failed to update stock price: %v\n", err)
			return
		}
		fmt.Println("Trade executed")
		fmt.Printf("\nTrade Executed - Buy Order: ID=%s, Quantity=%d, Price=$%.2f | Sell Order: ID=%s, Quantity=%d, Price=$%.2f\n", 
		highestBuyOrder.StockTxID, highestBuyOrder.Quantity, *highestBuyOrder.Price, order.StockTxID, order.Quantity, *highestBuyOrder.Price)
	}
}

func executeSellTrade(book *OrderBook, buyOrder *Order, order *Order){
	tradeQuantity := min(buyOrder.Quantity, order.Quantity)
	tradingPrice := buyOrder.Price

	updateMarketStockPrice(order.StockID, *tradingPrice)

	if buyOrder.Quantity > order.Quantity {
		// execute partial trade for buy order and complete trade for sell order
		buyOrder.Quantity -= tradeQuantity
		order.Quantity = 0
		heap.Push(&book.BuyOrders, buyOrder)
		completeSellOrder(book, order, tradeQuantity, tradingPrice)
		partialFulfillBuyOrder(book, buyOrder, tradeQuantity)
	} else if buyOrder.Quantity < order.Quantity {
		// execute partial trade for sell order and complete trade for buy order
		order.Quantity -= tradeQuantity
		buyOrder.Quantity = 0
		completeBuyOrder(book, buyOrder, tradeQuantity)
		partialFulfillSellOrder(book, order, tradeQuantity)
	} else {
		// execute complete trade for both buy and sell orders
		buyOrder.Quantity = 0
		order.Quantity = 0
		completeBuyOrder(book, buyOrder, tradeQuantity)
		completeSellOrder(book, order, tradeQuantity, tradingPrice)
	}
}

func partialFulfillSellOrder(book *OrderBook, order *Order, tradeQuantity int) {
	fmt.Println("Sell User: ", order.UserName)
	if err:= updateMoneyWallet(order.UserName, *order, order.Price, tradeQuantity, true); err != nil {
		fmt.Println("Error updating wallet: ", err)
	}

	if err := setStatus(order, "PARTIALLY_FULFILLED", false); err != nil {
		fmt.Println("Error setting status: ", err)
	}

	completedOrder := Order{
		StockTxID:  generateOrderID(),
		StockID:    order.StockID,
		WalletTxID: generateWalletID(),
		ParentTxID: &order.StockTxID,
		IsBuy:      order.IsBuy,
		OrderType:  order.OrderType,
		Quantity:   tradeQuantity,
		Price:      order.Price,
		TimeStamp:  time.Now().Format(time.RFC3339Nano),
		Status:     "COMPLETED",
		UserName:   order.UserName,
	}

	fmt.Println("Completed wallet tx: ", completedOrder.WalletTxID)

	// setWalletTransaction should always be before the setStockTransaction
	if err := setWalletTransaction(order.UserName, completedOrder, order.Price, tradeQuantity, true); err != nil {
		fmt.Println("Error setting wallet transaction: ", err)
	}

	if err := setStockTransaction(order.UserName, completedOrder, order.Price, tradeQuantity); err != nil {
		fmt.Println("Error setting stock transaction: ", err)
	}
}

func partialFulfillBuyOrder(book *OrderBook, order *Order, tradeQuantity int) {
	fmt.Println("Buy User: ", order.UserName)
	if err:= updateStockPortfolio(order.UserName, *order, tradeQuantity, true); err != nil {
		fmt.Println("Error updating stock portfolio: ", err)
	}

	if err := setStatus(order, "PARTIALLY_FULFILLED", false); err != nil {
		fmt.Println("Error setting status: ", err)
	}

	completedOrder := Order{
		StockTxID:  generateOrderID(),
		StockID:    order.StockID,
		WalletTxID: generateWalletID(),
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
	if err := setWalletTransaction(order.UserName, completedOrder, order.Price, tradeQuantity, false); err != nil {
		fmt.Println("Error setting wallet transaction: ", err)
	}

	if err := setStockTransaction(order.UserName, completedOrder, order.Price, tradeQuantity); err != nil {
		fmt.Println("Error setting stock transaction: ", err)
	}
	
}

// Assumption: assume that completed sell has wallet_tx_id
func completeSellOrder(book *OrderBook, order *Order, tradeQuantity int, tradingPrice *float64) {
	fmt.Println("Sell User: ", order.UserName)
	if err:= updateMoneyWallet(order.UserName, *order, tradingPrice, tradeQuantity, true); err != nil {
		fmt.Println("Error updating wallet: ", err)
	}

	if err := setStatus(order, "COMPLETED", true); err != nil {
		fmt.Println("Error setting status: ", err)
	}

	if err := setWalletTransaction(order.UserName, *order, tradingPrice, tradeQuantity, true); err != nil {
		fmt.Println("Error setting wallet transaction: ", err)
	}
}
/** === END SELL Order === **/

/** === BUY/SELL Order === **/
func updateMoneyWallet(userName string, order Order, price *float64, quantity int, isAdded bool) error {
	fmt.Println("Deducting money from wallet")

	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	var newPrice float64
	if order.OrderType == "MARKET" {
		newPrice, err = getMarketStockPrice(order.StockID)
		if err != nil {
			return fmt.Errorf("Failed to get market stock price: %w", err)
		}
	} else {
		newPrice = *price
	}

	// Calculate total to be added or deducted
	total := newPrice * float64(quantity)
	if !isAdded {
		total = total * (-1) // Reduce funds if buying
	}

	// Dont remove: Total = 0 when there is no Sell orders in Queue
	if total == 0 {
		fmt.Println("Total: ", total)
		return fmt.Errorf("No Sell stock orders in the queue - Market Price is 0")
	}

	// Update the user's wallet
	_, err = db.Exec(`
		UPDATE users SET wallet = wallet + $1 WHERE user_name = $2`, total, userName)
	if err != nil {
		return fmt.Errorf("Failed to update wallet: %w", err)
	}
	return nil
}
/** === END SELL Order === **/

func updateStockPortfolio(userName string, order Order, quantity int, isAdded bool) error {
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
		var amount int
		if err := rows.Scan(&amount); err != nil {
			return fmt.Errorf("Error while scanning row: %w", err)
		}
		if total < 0 && (amount + total) <= 0 {
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

// Store completed wallet transactions in the database
func setWalletTransaction(userName string, tx Order, price *float64, quantity int, isAdded bool) error {
	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to insert stock transaction: %w", err)
	}
	defer db.Close()

	var newPrice float64
	if tx.OrderType == "MARKET" {
		newPrice, err = getMarketStockPrice(tx.StockID)
		if err != nil {
			return fmt.Errorf("Failed to get market stock price: %w", err)
		}
	} else {
		newPrice = *price
	}

	// isDebit = True if money is deducted from wallet. 
	// isDebit = False if money is added to wallet
	var isDebit bool
	isDebit = !isAdded

	amount := newPrice * float64(quantity)

	// Insert transaction to wallet transactions
	_, err = db.Exec(`
		INSERT INTO wallet_transactions (wallet_tx_id, user_name, is_debit, amount, time_stamp)
		VALUES ($1, $2, $3, $4, $5)`, tx.WalletTxID, userName, isDebit, amount, tx.TimeStamp)
	if err != nil {
		return fmt.Errorf("Failed to commit transaction: %w", err)
	}
	return nil
}

func deleteWalletTransaction(userName string, order Order) error {
	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	// Insert transaction to wallet transactions
	_, err = db.Exec(`
		DELETE FROM wallet_transactions WHERE user_name = $1 AND wallet_tx_id = $2`, userName, order.WalletTxID)
	if err != nil {
		return fmt.Errorf("Failed to delete wallet transaction: %w", err)
	}
	return nil
}

// Store completed stock transactions in the database
func setStockTransaction(userName string, tx Order, price *float64, quantity int) error {
	fmt.Println("Setting stock transaction")
	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to insert stock transaction: %w", err)
	}
	defer db.Close()

	var newPrice float64
	if tx.OrderType == "MARKET" {
		newPrice, err = getMarketStockPrice(tx.StockID)
		if err != nil {
			return fmt.Errorf("Failed to get market stock price: %w", err)
		}
	} else {
		newPrice = *price
	
	}

	// Check if a wallet transaction has been made for this order yet
	rows, err := db.Query(`
		SELECT wallet_tx_id FROM wallet_transactions WHERE user_name = $1 AND wallet_tx_id = $2`, userName, tx.WalletTxID)
	if err != nil {
		return fmt.Errorf("Error querying wallet transactions: %w", err)
	}
	defer rows.Close()

	wallet_tx_id := ""

	// if there is a wallet transaction, then need to pair with it OR,
	// if status is COMPLETED, the stock transaction need to pair with a wallet transaction
	if rows.Next() || tx.Status == "COMPLETED"{
		fmt.Println("Wallet transaction: ", wallet_tx_id)
		wallet_tx_id = tx.WalletTxID
	}

	// Insert transaction to stock transactions
	_, err = db.Exec(`
		INSERT INTO stock_transactions (stock_tx_id, user_name, stock_id, wallet_tx_id, order_status, parent_tx_id, is_buy, order_type, stock_price, quantity,  time_stamp)
	    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`, tx.StockTxID, userName, tx.StockID, wallet_tx_id, tx.Status, tx.ParentTxID, tx.IsBuy, tx.OrderType, newPrice, quantity, tx.TimeStamp)
	if err != nil {
		return fmt.Errorf("Failed to commit transaction: %w", err)
	}
	return nil
}

func deleteStockTransaction(userName string, order Order) error {
	if order.Status != "IN_PROGRESS" {
		return fmt.Errorf("Cannot delete completed or partially completed transactions")
	}

	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	// Insert transaction to wallet transactions
	_, err = db.Exec(`
		DELETE FROM stock_transactions WHERE user_name = $1 AND stock_tx_id = $2`, userName, order.StockTxID)
	if err != nil {
		return fmt.Errorf("Failed to delete stock transaction: %w", err)
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

	if status == "PARTIALLY_FULFILLED" {
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

// ProcessOrder processes a buy or sell order based on the order type
func processOrder(book *OrderBook, order Order) {
	if order.IsBuy {
		if order.OrderType == "MARKET" {
			matchMarketBuyOrder(book, order)
		} else {
			matchLimitBuyOrder(book, order)
		}
	} else {
		if order.OrderType == "MARKET" {
			matchMarketSellOrder(book, order)
		} else {
			matchLimitSellOrder(book, order)
		}
	}
}

func updateMarketStockPrice(stockID int, price float64) error {
	fmt.Println("Updating stock price")
	// Connect to database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	// Update the stock price
	_, err = db.Exec("UPDATE stocks SET current_price = $1 WHERE stock_id = $2", price, stockID)
	if err != nil {
		return fmt.Errorf("Failed to update stock price: %w", err)
	}
	return nil
}

// getMarketStockPrice retrieves the current market stock price from the database.
func getMarketStockPrice(stockID int) (float64, error) {
	// Connect to the database
	db, err := openConnection()
	if err != nil {
		return 0, fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	// Query the database to get the current price for the specified stock ID
	var currentPrice float64
	err = db.QueryRow("SELECT current_price FROM stocks WHERE stock_id = $1", stockID).Scan(&currentPrice)
	if err != nil {
		return 0, fmt.Errorf("Failed to get market stock price: %w", err)
	}

	return currentPrice, nil
}

func verifyWalletBeforeTransaction(userName string, order Order) error {
	// Connect to the database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	// get stock id 
	var stockID int
	err = db.QueryRow("SELECT stock_id FROM stocks WHERE stock_id = $1", order.StockID).Scan(&stockID)
	if err != nil {
		return fmt.Errorf("Stock not exist - Failed to get stock id: %w", err)
	}

	var price float64
	if order.OrderType == "MARKET" {
		price, err = getMarketStockPrice(order.StockID)
		if err != nil {
			return fmt.Errorf("Failed to get market stock price: %w", err)
		}
	} else {
		price = *order.Price
	}

	// get user wallet
	var wallet float64
	err = db.QueryRow("SELECT wallet FROM users WHERE user_name = $1", userName).Scan(&wallet)
	if err != nil {
		return fmt.Errorf("Failed to get user wallet: %w", err)
	}

	// check if user has enough money to buy the stock
	if wallet < price * float64(order.Quantity) {
		return fmt.Errorf("Insufficient funds")
	}

	return nil
}

func verifyStockBeforeTransaction(userName string, order Order) error {
	// Connect to the database
	db, err := openConnection()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	defer db.Close()

	// get stock id 
	var stockID int
	err = db.QueryRow("SELECT stock_id FROM stocks WHERE stock_id = $1", order.StockID).Scan(&stockID)
	if err != nil {
		return fmt.Errorf("Stock not exist - Failed to get stock id: %w", err)
	}

	// get user stock portfolio
	var quantity int
	err = db.QueryRow("SELECT quantity FROM user_stocks WHERE user_name = $1 AND stock_id = $2", userName, order.StockID).Scan(&quantity)
	if err != nil {
		return fmt.Errorf("Failed to get user stock portfolio: %w", err)
	}

	// check if user has enough stock to sell
	if quantity < order.Quantity {
		return fmt.Errorf("Insufficient stock")
	}

	return nil
}

func checkAndRemoveExpiredOrders() {
    // Iterate over each order book and check for expired orders
    for _, book := range orderBookMap.OrderBooks {
        book.mu.Lock()
        defer book.mu.Unlock()

        // Iterate over buy orders
        for i := 0; i < book.BuyOrders.Len(); {
            order := book.BuyOrders.Order[i]
            if isOrderExpired(order) {
                // Execute the function to remove the expired order and perform post-processing
                executeRemoveOrder(*order, &book.BuyOrders, i)
            } else {
                i++
            }
        }

        // Iterate over sell orders
        for i := 0; i < book.SellOrders.Len(); {
            order := book.SellOrders.Order[i]
            if isOrderExpired(order) {
                // Execute the function to remove the expired order and perform post-processing
                executeRemoveOrder(*order, &book.SellOrders, i)
            } else {
                i++
            }
        }
    }
}



func refundMoney(userName string, order Order) error {
    // Connect to the database
    db, err := openConnection()
    if err != nil {
        return fmt.Errorf("Failed to connect to database: %w", err)
    }
    defer db.Close()

    // Calculate the total amount to be refunded
    total := *order.Price * float64(order.Quantity)

    // Update the user's wallet by adding the refunded amount
    _, err = db.Exec(`
        UPDATE users SET wallet = wallet + $1 WHERE user_name = $2`, total, userName)
    if err != nil {
        return fmt.Errorf("Failed to update wallet: %w", err)
    }

    return nil
}

func refundStocks(userName string, order Order) error {
    // Connect to the database
    db, err := openConnection()
    if err != nil {
        return fmt.Errorf("Failed to connect to database: %w", err)
    }
    defer db.Close()

    // Update the user's stock portfolio by adding the refunded quantity
    _, err = db.Exec(`
        UPDATE user_stocks SET quantity = quantity + $1 
        WHERE user_name = $2 AND stock_id = $3`, order.Quantity, userName, order.StockID)
    if err != nil {
        return fmt.Errorf("Failed to update stock portfolio: %w", err)
    }

    return nil
}


func isOrderExpired(order *Order) bool {
    // Parse the timestamp of the order
    orderTime, err := time.Parse(time.RFC3339Nano, order.TimeStamp)
    if err != nil {
        // Handle error
        return false
    }

    // Check if the order is older than 15 minutes
    return time.Since(orderTime) > 15*time.Minute
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

    // Start a background goroutine to periodically check and remove expired orders
    go func() {
        for {
            time.Sleep(time.Minute)
            checkAndRemoveExpiredOrders()
        }
    }()

	router.Run(":8585")
}
