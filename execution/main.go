package main

import (
    "database/sql"
	"net/http"
    "fmt"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"
)

// Global variable for the database connection
var db *sql.DB

const (
	// host = "database"
	host     = "localhost" // for local testing
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
	BuyOrder  *Order   `json:"buy_order"`
	SellOrder *Order   `json:"sell_order"`
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

func executeBuyTrade(c *gin.Context) {
	var payload TradePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	buyOrder := payload.BuyOrder
	sellOrder := payload.SellOrder

	// Handle the buy trade execution here
	fmt.Printf("\nBuy Trade Executed - Sell Order: ID=%s, Quantity=%.2f, Price=$%.2f | Buy Order: ID=%s, Quantity=%.2f, Price=$%.2f\n",
	sellOrder.StockTxID, sellOrder.Quantity, *sellOrder.Price, buyOrder.StockTxID, buyOrder.Quantity, *buyOrder.Price)
	
	// For example:
	c.JSON(http.StatusOK, gin.H{"message": "Buy trade executed successfully"})
}

func executeSellTrade(c *gin.Context) {
	var payload TradePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	buyOrder := payload.BuyOrder
	sellOrder := payload.SellOrder

	// Handle the sell trade execution here
	fmt.Printf("\nSell Trade Executed - Sell Order: ID=%s, Quantity=%.2f, Price=$%.2f | Buy Order: ID=%s, Quantity=%.2f, Price=$%.2f\n",
	sellOrder.StockTxID, sellOrder.Quantity, *sellOrder.Price, buyOrder.StockTxID, buyOrder.Quantity, *buyOrder.Price)
	
	// For example:
	c.JSON(http.StatusOK, gin.H{"message": "Sell trade executed successfully"})
}


func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "token"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	router.SetTrustedProxies([]string{"127.0.0.1"})

	router.POST("/executeSellTrade", executeSellTrade)
	router.POST("/executeBuyTrade", executeBuyTrade)

	router.Run(":5555")
}