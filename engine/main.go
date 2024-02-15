package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Poomon001/day-trading-package/identification"
	_ "github.com/lib/pq"
)

const (
	host     = "database"
	port     =   5432
	user     = "nt_user"
	password = "db123"
	dbname   = "nt_db"
)

// Define the structure of the request body for placing a stock order
type PlaceStockOrderRequest struct {
	StockID    int     `json:"stock_id" binding:"required"`
	IsBuy      *bool    `json:"is_buy" binding:"required"`
	OrderType  string  `json:"order_type" binding:"required"`
	Quantity   int     `json:"quantity" binding:"required"`
	Price      *float64 `json:"price"`
}

// Define the structure of the response body for placing a stock order
type PlaceStockOrderResponse struct {
	Success bool `json:"success"`
	Data    interface{} `json:"data"`
}

// HandlePlaceStockOrder is the handler for the /placeStockOrder endpoint
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

	// Check if IsBuy is not nil before dereferencing it
	if request.IsBuy != nil && !*request.IsBuy {
		err := updateUserStockQuantity(userName.(string), request.StockID, request.Quantity)
		if err != nil {
			handleError(c, http.StatusInternalServerError, "Failed to update user stock quantity", err)
			return
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

func main() {
	router := gin.Default()

	router.POST("/placeStockOrder", identification.Identification, HandlePlaceStockOrder)

	router.Run(":8585")
}
