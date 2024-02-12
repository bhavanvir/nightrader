package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/Poomon001/day-trading-package/identification"
	_ "github.com/lib/pq"
)

const (
	host     = "database"
	port     = 5432
	user     = "nt_user"
	password = "db123"
	dbname   = "nt_db"
)

type Error struct {
	Success bool    `json:"success"`
	Data    *string `json:"data"`
	Message string  `json:"message"`
}

type AddMoney struct {
	Amount int `json:"amount"`
}

type PostResponse struct {
	Success bool    `json:"success"`
	Data    *string `json:"data"`
}

func handleError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := Error{
		Success: false,
		Data:    nil,
		Message: fmt.Sprintf("%s: %v", message, err),
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

	// PostgreSQL connection info
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// Connect to the PostgreSQL database
	db, err := sql.Open("postgres", postgresqlDbInfo)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to connect to the database", err)
		return
	}
	defer db.Close()

	// Execute SQL query to update user's wallet
	_, err = db.Exec("UPDATE users SET wallet = wallet + $1 WHERE user_name = $2", addMoney.Amount, userName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to update wallet", err)
		return
	}

	// If everything succeeded, return success response
	response := PostResponse{
		Success: true,
		Data:    nil,
	}
	c.IndentedJSON(http.StatusOK, response)
}

func getCookies(c *gin.Context) {
	cookie := c.GetHeader("Authorization")

	if cookie == "" {
		handleError(c, http.StatusBadRequest, "Authorization token missing", nil)
		return
	}

	c.String(http.StatusOK, "Authorization token: "+cookie)
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	identification.Test()
	router.POST("/addMoneyToWallet", identification.Identification, addMoneyToWallet)
	router.GET("/eatCookies", getCookies)
	router.Run(":5000")
}
