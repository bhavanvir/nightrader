package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"example/Transactions/middleware"
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
	user_name, err := c.Get("user_name")
	
	if !err {
		handleError(c, http.StatusBadRequest, "Failed to obtain the user name", nil)
		return
	}

	var addMoney AddMoney
	if err := c.ShouldBindJSON(&addMoney); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// TODO: add the money to the user's wallet in database
	fmt.Println("User: ", userName)
	c.IndentedJSON(http.StatusOK, addMoney)
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
	fmt.Println(middleware.Identification)
	router.POST("/addMoneyToWallet", middleware.Identification, addMoneyToWallet)
	router.GET("/eatCookies", middleware.Identification, getCookies)
	router.Run(":5000")
}
