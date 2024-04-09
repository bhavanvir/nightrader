package main

import (
    "container/heap"
    "database/sql"
    "fmt"
    "net/http"
    "sync"
    "time"

    "github.com/gin-contrib/cors"

    "github.com/Poomon001/day-trading-package/identification"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    _ "github.com/lib/pq"
)

var db *sql.DB

var (
    stmtUpdateWalletTransaction *sql.Stmt
    stmtUpdateMoneyWallet       *sql.Stmt
    stmtCheckUserStocks         *sql.Stmt
    stmtDeleteUserStocks        *sql.Stmt
    stmtInsertUserStocks        *sql.Stmt
    stmtSetWalletTransaction    *sql.Stmt
    stmtDeleteWalletTransaction *sql.Stmt
    stmtGetWalletTransactionsAmount *sql.Stmt
    stmtSetStockTransaction     *sql.Stmt
    stmtDeleteStockTransaction  *sql.Stmt
    stmtSetStatus               *sql.Stmt
    stmtUpdateWalletTxId        *sql.Stmt
    stmtVerifyWalletBeforeTransaction *sql.Stmt
    stmtVerifyStockBeforeTransaction  *sql.Stmt
    stmtUpdateMarketStockPrice        *sql.Stmt
    stmtUpdateUserStocks              *sql.Stmt
    stmtCheckWalletTransaction      *sql.Stmt
)

const (
    host = "database"
    // host     = "localhost" // for local testing
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
    StockID   string   `json:"stock_id" binding:"required"`
    IsBuy     *bool    `json:"is_buy" binding:"required"`
    OrderType string   `json:"order_type" binding:"required"`
    Quantity  float64      `json:"quantity" binding:"required"`
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

func createInitOrder(request *PlaceStockOrderRequest, userName string) (Order, error) {
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
} // createInitOrder

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
        handleError(c, http.StatusOK, err.Error(), err)
        return
    }

    order, e := createInitOrder(&request, userName)
    if e != nil {
        handleError(c, http.StatusInternalServerError, "Failed to create order", e)
        return
    }

    book, bookerr := initializePriorityQueue(order)
    if bookerr != nil {
        handleError(c, http.StatusInternalServerError, "Failed to push order to priority queue", bookerr)
        return
    }

    if err := verifyQueueBeforeMarketTransaction(book, order); err != nil {
        handleError(c, http.StatusBadRequest, "Fail to place Market order: ", err)
        return
    }

    orderPrice := getStockOrderPrice(book, order);
    amount := (*orderPrice) * float64(order.Quantity)

    // to be safe, lock here
    book.mu.Lock()
    defer book.mu.Unlock()

    if order.IsBuy {
        if err := verifyWalletBeforeTransaction(userName, book, order); err != nil {
            handleError(c, http.StatusBadRequest, "Failed to verify Wallet", err)
            return
        }

        if err := updateMoneyWallet(userName, amount, false); err != nil {
            handleError(c, http.StatusInternalServerError, "Failed to deduct money from user's wallet", err)
            return
        }

        if err := setWalletTransaction(userName, order.WalletTxID, order.TimeStamp, orderPrice, order.Quantity, false); err != nil {
            handleError(c, http.StatusInternalServerError, "Buy Order setWalletTx Error: "+err.Error(), err)
            return
        }

        if err := setStockTransaction(userName, order, orderPrice, order.Quantity); err != nil {
            handleError(c, http.StatusInternalServerError, "Buy Order setStockTx Error: "+err.Error(), err)
            return
        }

        processOrder(book, order)
        LogBuyOrder(order)
    } else {
        if err := verifyStockBeforeTransaction(userName, order); err != nil {
            handleError(c, http.StatusBadRequest, "Failed to verify stocks", err)
            return
        }

        if err := updateStockPortfolio(userName, order, order.Quantity, false); err != nil {
            handleError(c, http.StatusInternalServerError, "Failed to deduct stock from user's portfolio", err)
            return
        }

        if err := setStockTransaction(userName, order, orderPrice, order.Quantity); err != nil {
            handleError(c, http.StatusInternalServerError, "Sell Order setStockTx Error: "+err.Error(), err)
            return
        }

        processOrder(book, order)
        LogSellOrder(order)
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
    heap.Remove(bookOrders, indexToRemove)

    if order.IsBuy {
        postprocessingRemoveBuyOrder(order)
    } else {
        postprocessingRemoveSellOrder(order)
    }
}

// Only for Limit orders
func postprocessingRemoveBuyOrder(order Order) {
    amount := (*order.Price) * float64(order.Quantity)

    if order.Status == "IN_PROGRESS" {
        // refund all dedeucted money back to wallet
        if err := updateMoneyWallet(order.UserName, amount, true); err != nil {
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
        fmt.Println("Remove PARTIAL_FULFILLED buy order")
        if err := updateMoneyWallet(order.UserName, amount, true); err != nil {
            fmt.Println("Error updating wallet: ", err)
        }

        // remove transaction from stock_transactions
        if err := deleteWalletTransaction(order.UserName, order); err != nil {
            fmt.Println("Error deleting wallet transaction: ", err)
        }
    }
}

func postprocessingRemoveSellOrder(order Order) {
    if order.Status == "IN_PROGRESS" {
        // refund all dedeucted stock back to portfolio
        if err := updateStockPortfolio(order.UserName, order, order.Quantity, true); err != nil {
            fmt.Println("Error updating stock portfolio: ", err)
        }

        // remove transaction from stock_transactions
        if err := deleteStockTransaction(order.UserName, order); err != nil {
            fmt.Println("Error deleting stock transaction: ", err)
        }
    } else {
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

        if foundBuy.Success || foundSell.Success {
            response := CancelStockTransactionResponse{
                Success: true,
                Data:    nil,
            }
            c.IndentedJSON(http.StatusOK, response)
            return
        }
    }

    errorMessage := fmt.Sprintf("Order [StockTxID: %s] not found", StockTxID)
    handleError(c, http.StatusOK, errorMessage, nil)
}

// Define the structure of the order book map
type OrderBookMap struct {
    OrderBooks map[string]*OrderBook // Map of stock ID to order book
    mu         sync.Mutex            // Mutex to synchronize access to the map
}

// Initialize the order book map
var orderBookMap = OrderBookMap{
    OrderBooks: make(map[string]*OrderBook),
}

/** === BUY Order === **/
func matchLimitBuyOrder(book *OrderBook, order Order) {
    // Add the buy order to the buy queue
    heap.Push(&book.BuyOrders, &order)
    highestBuyOrder := book.BuyOrders.Order[0]

    // If the buy order is a limit order, match it with the lowest sell order that is less than or equal to the buy order price
    for highestBuyOrder.Quantity > 0 && book.SellOrders.Len() > 0 {
        lowestSellOrder := book.SellOrders.Order[0]

        // If the lowest sell order price is less than or equal to the buy order price, execute the trade
        if *lowestSellOrder.Price <= *highestBuyOrder.Price {
            buyPrice := getStockOrderPrice(book, *highestBuyOrder);
            sellPrice := getStockOrderPrice(book, *lowestSellOrder);

            // execute the trade
            executeBuyTrade(highestBuyOrder, lowestSellOrder, buyPrice, sellPrice)

            // If the sell order quantity is empty, pop it from the queue
            if lowestSellOrder.Quantity == 0 {
                lowestSellOrder = heap.Pop(&book.SellOrders).(*Order)
            }
        } else {
            // If the lowest sell order price is greater than the buy order price, put it back in the sell queue
            break
        }
    }
    highestBuyOrder = heap.Pop(&book.BuyOrders).(*Order)

    // If the buy order was not fully executed, put it back in the buy queue
    if highestBuyOrder.Quantity > 0 {
        heap.Push(&book.BuyOrders, highestBuyOrder)
    }
}

/*
    Assumption: The Sell order price will be equal and unchanged thoughout the trading process
                means there is enough Sell orders quantity at the exact MARKET price to complete one Market order.
                Thus, Cannot support Marekt order with different prices. 
                e.g Market order Quan 100 will not work for Limit $5 with Quantity 50 and Limit $10 with Quantity 50
    Error Handling: Refund money, remove transaction from wallet_transactions, stock_transactions, exit with error
*/
func matchMarketBuyOrder(book *OrderBook, order Order) {
    if book.SellOrders.Len() <= 0 {
        // refund money, remove transaction from wallet_transactions, stock_transactions, exit with error
    }
    // Match the buy order with the lowest Sell order that is less than or equal to the buy order price
    for order.Quantity > 0 && book.SellOrders.Len() > 0 {
        lowestSellOrder := book.SellOrders.Order[0]

        buyPrice := getStockOrderPrice(book, order);
        sellPrice := getStockOrderPrice(book, *lowestSellOrder);

        // execute the trade
        executeBuyTrade(&order, lowestSellOrder, buyPrice, sellPrice)
        
        // If the buy order quantity is empty, pop it from the queue
        if lowestSellOrder.Quantity == 0 {
            lowestSellOrder = heap.Pop(&book.SellOrders).(*Order)
        }
    }
}

func executeBuyTrade(buyOrder *Order, sellOrder *Order, buyPrice *float64, sellPrice *float64) {
    tradeQuantity := min(buyOrder.Quantity, sellOrder.Quantity)

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

func completeBuyOrder(buyOrder *Order, tradeQuantity float64, buyPrice *float64, sellPrice *float64) {
    // Calculate refund amount
    refundAmount := (*buyPrice - *sellPrice) * float64(tradeQuantity)

    // Refund deducted money to the Buy user's wallet, adjusting for any price differences
    if refundAmount > 0 {
        if err := updateMoneyWallet(buyOrder.UserName, refundAmount, true); err != nil {
            fmt.Println("Error updating different price refund to wallet: ", err)
        }

        // Update wallet transactions from BUY order
        if err := updateWalletTransaction(buyOrder.UserName, *buyOrder, (*sellPrice)*tradeQuantity); err != nil {
            fmt.Println("Error updating wallet transaction: ", err)
        }
    }

    // Update stock portfolio
    if err := updateStockPortfolio(buyOrder.UserName, *buyOrder, tradeQuantity, true); err != nil {
        fmt.Println("Error updating stock portfolio: ", err)
    }

    // Set order status to COMPLETED
    if err := setStatus(buyOrder, "COMPLETED", false); err != nil {
        fmt.Println("Error setting status: ", err)
    }
}

/** === END BUY Order === **/

/** === SELL Order === **/
func matchLimitSellOrder(book *OrderBook, order Order) {
    // initialize the market price if there isn't one yet
    if book.SellOrders.Len() == 0 {
        lastSoldStockPrice := getStockOrderPrice(book, order)
        if err := updateMarketStockPrice(order.StockID, lastSoldStockPrice); err != nil {
            fmt.Println("Failed to update Market Stock Price after Limit Sell: ", err)
        }
    }

    // Add the Sell order to the sell queue
    heap.Push(&book.SellOrders, &order)
    lowestSellOrder := book.SellOrders.Order[0]

    for lowestSellOrder.Quantity > 0 && book.BuyOrders.Len() > 0 {
        highestBuyOrder := book.BuyOrders.Order[0]

        // If the lowest sell order price is less than or equal to the buy order price, execute the trade
        if *lowestSellOrder.Price <= *highestBuyOrder.Price {
            buyPrice := getStockOrderPrice(book, *highestBuyOrder);
            sellPrice := getStockOrderPrice(book, *lowestSellOrder);

            // execute the trade
            executeSellTrade(highestBuyOrder, lowestSellOrder, buyPrice, sellPrice)
            
            if highestBuyOrder.Quantity == 0 {
                highestBuyOrder = heap.Pop(&book.BuyOrders).(*Order)
            }
        } else {
            break
        }
    }

    lowestSellOrder = heap.Pop(&book.SellOrders).(*Order)

    if lowestSellOrder.Quantity > 0 {
        heap.Push(&book.SellOrders, lowestSellOrder)
    }
}

/*
*

    Assumption: The Buy order price will be equal and unchanged thoughout the Market trading process
                means there is enough Buy orders quantity at the exact MARKET price to complete one Market order.
                Thus, Cannot support Marekt order with different prices. 
                e.g Market order Quan 100 will not work for Limit $5 with Quantity 50 and Limit $10 with Quantity 50
    Error Handling: Refund stock, remove stock_transactions, exit with error

*
*/
func matchMarketSellOrder(book *OrderBook, order Order) {
    //TODO: refund stock, remove transaction from wallet_transactions, stock_transactions, exit with error????

    // Match the Sell order with the highest Buy order that is greater than or equal to the sell order price
    for order.Quantity > 0 && book.BuyOrders.Len() > 0 {
        highestBuyOrder := book.BuyOrders.Order[0]

        buyPrice := getStockOrderPrice(book, *highestBuyOrder);
        sellPrice := getStockOrderPrice(book, order);

        // execute the trade
        executeSellTrade(highestBuyOrder, &order, buyPrice, sellPrice)

        // if highestBuyOrder.Quantity <= order.Quantity {
        if highestBuyOrder.Quantity == 0 {
            highestBuyOrder = heap.Pop(&book.BuyOrders).(*Order)
        }
    }
}

func executeSellTrade(buyOrder *Order, sellOrder *Order, buyPrice *float64, sellPrice *float64) {
    tradeQuantity := min(buyOrder.Quantity, sellOrder.Quantity)

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
        StockTxID:  generateOrderID(),
        StockID:    sellOrder.StockID,
        WalletTxID: generateWalletID(),
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
    if err := setWalletTransaction(order.UserName, completedOrder.WalletTxID, completedOrder.TimeStamp, sellPrice, tradeQuantity, false); err != nil {
        fmt.Println("Error setting wallet transaction: ", err)
    }

    if err := setStockTransaction(order.UserName, completedOrder, sellPrice, tradeQuantity); err != nil {
        fmt.Println("Error setting stock transaction: ", err)
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

/** === END SELL Order === **/

/** === BUY/SELL Order === **/
func updateWalletTransaction(userName string, order Order, amount float64) error {
    // Update the wallet transaction
    _, err := stmtUpdateWalletTransaction.Exec(amount, userName, order.WalletTxID)
    if err != nil {
        return fmt.Errorf("Failed to update wallet transaction: %w", err)
    }

    return nil
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

/** === END SELL Order === **/

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

func deleteWalletTransaction(userName string, order Order) error {
    // Insert transaction to wallet transactions
    _, err := stmtDeleteWalletTransaction.Exec(userName, order.WalletTxID)
    if err != nil {
        return fmt.Errorf("Failed to delete wallet transaction: %w", err)
    }
    return nil
}

func getWalletTransactionsAmount(userName string, walletTxID string) (float64, error) {
    // Query the database to get the total amount of wallet transactions for the specified user and wallet ID
    var totalAmount float64
    err := stmtGetWalletTransactionsAmount.QueryRow(userName, walletTxID).Scan(&totalAmount)
    if err != nil {
        return 0, fmt.Errorf("Failed to get wallet transactions amount: %w", err)
    }

    return totalAmount, nil
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

func deleteStockTransaction(userName string, order Order) error {
    if order.Status != "IN_PROGRESS" {
        return fmt.Errorf("Cannot delete completed or partially completed transactions")
    }

    // Insert transaction to wallet transactions
    _, err := stmtDeleteStockTransaction.Exec(userName, order.StockTxID)
    if err != nil {
        return fmt.Errorf("Failed to delete stock transaction: %w", err)
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

// get stock Order price for Limit or Market order
// if order is MARKET, get the top of the queue price
// if order is LIMIT, get the price of the order
func getStockOrderPrice(book *OrderBook, order Order) *float64 {
    if order.OrderType == "MARKET" {
        if book.SellOrders.Len() > 0 && order.IsBuy {
            return book.SellOrders.Order[0].Price
        } 
        
        if book.BuyOrders.Len() > 0 && !order.IsBuy{
            // impossible case by an assumption that there should not be Buy order with no Sell orders in queue
            return book.BuyOrders.Order[0].Price
        }
    }
    return order.Price
}

func verifyWalletBeforeTransaction(userName string, book *OrderBook, order Order) error {
    // Execute the SQL query
    row := stmtVerifyWalletBeforeTransaction.QueryRow(userName, order.StockID)

    // Declare variables to store the results
    var stockID string
    var wallet float64

    // Scan the results into variables
    err := row.Scan(&stockID, &wallet)
    if err != nil {
        return fmt.Errorf("Failed to get stock ID and user wallet: %w", err)
    }

    // Calculate the order price
    price := getStockOrderPrice(book, order)

    // Check if user has enough funds to buy the stock
    if wallet < (*price)*float64(order.Quantity) {
        return fmt.Errorf("Insufficient funds")
    }

    return nil
}

func verifyQueueBeforeMarketTransaction(book *OrderBook, order Order) error {
    if order.OrderType == "MARKET" && order.IsBuy && book.SellOrders.Len() <= 0 {
        return fmt.Errorf("No Sell orders available")
    }

    if order.OrderType == "MARKET" && !order.IsBuy && book.BuyOrders.Len() <= 0 {
        return fmt.Errorf("No Buy orders available")
    }

    // check if stocks in sell orders in queue with the same price is enough to fulfill the buy order
    // Assumption: The Market Buy order price will be equal and unchanged thoughout the trading process
    if order.OrderType == "MARKET" && order.IsBuy {
        var totalSellQuantity float64
        for i := 0; i < book.SellOrders.Len(); i++ {
            if *book.SellOrders.Order[i].Price == *(getStockOrderPrice(book, order)) {
                totalSellQuantity += book.SellOrders.Order[i].Quantity
            }
        }
        if totalSellQuantity < order.Quantity {
            return fmt.Errorf("Insufficient Sell stocks")
        }
    }

    // check if stocks in buy orders in queue with the same price is enough to fulfill the sell order
    // Assumption: The Market Sell order price will be equal and unchanged thoughout the trading process
    if order.OrderType == "MARKET" && !order.IsBuy {
        var totalBuyQuantity float64
        for i := 0; i < book.BuyOrders.Len(); i++ {
            if *book.BuyOrders.Order[i].Price == *(getStockOrderPrice(book, order)) {
                totalBuyQuantity += book.BuyOrders.Order[i].Quantity
            }
        }
        if totalBuyQuantity < order.Quantity {
            return fmt.Errorf("Insufficient Buy stocks")
        }
    }
        
    return nil
}

func verifyStockBeforeTransaction(userName string, order Order) error {
    // Get stock id and check if it exists
    var quantity float64
    err := stmtVerifyStockBeforeTransaction.QueryRow(userName, order.StockID).Scan(&quantity)
    if err != nil {
        return fmt.Errorf("failed to get user stock portfolio: %w", err)
    }

    // Check if user has enough stock to sell
    if quantity < order.Quantity {
        return fmt.Errorf("insufficient stock")
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

func isOrderExpired(order *Order) bool {
    // Parse the timestamp of the order
    orderTime, err := time.Parse(time.RFC3339Nano, order.TimeStamp)
    if err != nil {
        // Handle error
        return false
    }

    // Check if the order is older than 15 minutes
    return time.Since(orderTime) > 14*time.Minute
}

func prepareStatements() error {
    var err error

    stmtUpdateWalletTransaction, err = db.Prepare(`
        UPDATE wallet_transactions SET amount = $1 WHERE user_name = $2 AND wallet_tx_id = $3`)
    if err != nil {
        return fmt.Errorf("failed to prepare update wallet transaction statement: %v", err)
    }

    stmtUpdateMoneyWallet, err = db.Prepare(`
        UPDATE users SET wallet = wallet + $1 WHERE user_name = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare update money wallet statement: %v", err)
    }

    stmtCheckUserStocks, err = db.Prepare(`
        SELECT quantity FROM user_stocks WHERE user_name = $1 AND stock_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare check user stocks statement: %v", err)
    }

    stmtDeleteUserStocks, err = db.Prepare(`
        DELETE FROM user_stocks WHERE user_name = $1 AND stock_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare delete user stocks statement: %v", err)
    }

    stmtInsertUserStocks, err = db.Prepare(`
        INSERT INTO user_stocks VALUES ($1, $2, $3)`)
    if err != nil {
        return fmt.Errorf("failed to prepare insert user stocks statement: %v", err)
    }

    stmtSetWalletTransaction, err = db.Prepare(`
        INSERT INTO wallet_transactions (wallet_tx_id, user_name, is_debit, amount, time_stamp)
        VALUES ($1, $2, $3, $4, $5)`)
    if err != nil {
        return fmt.Errorf("failed to prepare set wallet transaction statement: %v", err)
    }

    stmtDeleteWalletTransaction, err = db.Prepare(`
        DELETE FROM wallet_transactions WHERE user_name = $1 AND wallet_tx_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare delete wallet transaction statement: %v", err)
    }

    stmtGetWalletTransactionsAmount, err = db.Prepare(`
        SELECT SUM(amount) FROM wallet_transactions WHERE user_name = $1 AND wallet_tx_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare get wallet transactions amount statement: %v", err)
    }

    stmtSetStockTransaction, err = db.Prepare(`
        INSERT INTO stock_transactions (stock_tx_id, user_name, stock_id, wallet_tx_id, order_status, parent_stock_tx_id, is_buy, order_type, stock_price, quantity,  time_stamp)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`)
    if err != nil {
        return fmt.Errorf("failed to prepare set stock transaction statement: %v", err)
    }

    stmtDeleteStockTransaction, err = db.Prepare(`
        DELETE FROM stock_transactions WHERE user_name = $1 AND stock_tx_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare delete stock transaction statement: %v", err)
    }

    stmtSetStatus, err = db.Prepare(`
        UPDATE stock_transactions SET order_status = $1 WHERE user_name = $2 AND stock_tx_id = $3`)
    if err != nil {
        return fmt.Errorf("failed to prepare set status statement: %v", err)
    }

    stmtUpdateWalletTxId, err = db.Prepare(`
        UPDATE stock_transactions SET wallet_tx_id = $1 WHERE user_name = $2 AND stock_tx_id = $3`)
    if err != nil {
        return fmt.Errorf("failed to prepare update wallet transaction ID statement: %v", err)
    }

    stmtVerifyWalletBeforeTransaction, err = db.Prepare(`
        SELECT s.stock_id, u.wallet
        FROM stocks s
        JOIN users u ON u.user_name = $1
        WHERE s.stock_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare verify wallet before transaction statement: %v", err)
    }

    stmtVerifyStockBeforeTransaction, err = db.Prepare(`
        SELECT quantity FROM user_stocks WHERE user_name = $1 AND stock_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare verify stock before transaction statement: %v", err)
    }

    stmtUpdateMarketStockPrice, err = db.Prepare(`
        UPDATE stocks SET current_price = $1 WHERE stock_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare update market stock price statement: %v", err)
    }

    stmtUpdateUserStocks, err = db.Prepare(`
        UPDATE user_stocks SET quantity = quantity + $1 WHERE user_name = $2 AND stock_id = $3`)
    if err != nil {
        return fmt.Errorf("failed to prepare update user stocks statement: %v", err)
    }

    stmtCheckWalletTransaction, err = db.Prepare(`
		SELECT wallet_tx_id FROM wallet_transactions WHERE user_name = $1 AND wallet_tx_id = $2`)
    if err != nil {
        return fmt.Errorf("failed to prepare check wallet transaction statement: %v", err)
    }

    return nil
}


func initializeDB() error {
    postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
    var err error
    db, err = sql.Open("postgres", postgresqlDbInfo)
    if err != nil {
        return fmt.Errorf("failed to connect to the database: %v", err)
    }

    // Ensure the database connection is fully established
    for {
        err = db.Ping()
        if err == nil {
            break
        }
        fmt.Println("Waiting for the database connection to be established...")
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
    defer db.Close()

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
    defer stmtDeleteWalletTransaction.Close()
    defer stmtGetWalletTransactionsAmount.Close()
    defer stmtSetStockTransaction.Close()
    defer stmtDeleteStockTransaction.Close()
    defer stmtSetStatus.Close()
    defer stmtUpdateWalletTxId.Close()
    defer stmtVerifyWalletBeforeTransaction.Close()
    defer stmtVerifyStockBeforeTransaction.Close()
    defer stmtUpdateUserStocks.Close()
    defer stmtCheckWalletTransaction.Close()


    db.SetMaxOpenConns(10) // Set maximum number of open connections
    db.SetMaxIdleConns(5) // Set maximum number of idle connections

    router := gin.Default()

    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:3000", "http://localhost"}
    config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "token"}
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
