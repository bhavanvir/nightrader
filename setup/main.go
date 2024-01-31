package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Stock struct {
	ID        string `json:"stock_id"`
	StockName string `json:"stock_name"`
}

var stocks = []Stock{}

const jsonFilePath = "stocks.json"

func generateUUID() (string, error) {
	uuidBytes := make([]byte, 16)
	_, err := rand.Read(uuidBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(uuidBytes), nil
}

func createStock(c *gin.Context) {
	var json struct {
		StockName string `json:"stock_name"`
	}

	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stockID, err := generateUUID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate UUID"})
		return
	}

	newStock := Stock{
		ID:        stockID,
		StockName: json.StockName,
	}

	stocks = append(stocks, newStock)

	// Save stocks to JSON file
	if err := saveStocksToFile(stocks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save stocks to file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"stock_id": stockID,
		},
	})
}

// TODO: Instead of saving to a JSON file we need to save to a database
func saveStocksToFile(stocks []Stock) error {
	// Convert stocks slice to JSON
	jsonData, err := json.Marshal(stocks)
	if err != nil {
		return err
	}

	// Write JSON data to file
	err = ioutil.WriteFile(jsonFilePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Load existing stocks from JSON file
	if jsonData, err := ioutil.ReadFile(jsonFilePath); err == nil {
		err := json.Unmarshal(jsonData, &stocks)
		if err != nil {
			panic(err)
		}
	}

	router := gin.Default()
	router.Use(cors.Default())

	router.POST("/createStock", createStock)

	router.Run(":8080")
}
