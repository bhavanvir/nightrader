package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Stock struct {
	ID        string `json:"stock_id"`
	StockName string `json:"stock_name"`
}

const (
	host     = "database"
	port     = 5432
	user     = "nt_user"
	password = "db123"
	dbname   = "nt_db"
)

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

	// Save stocks to database
	if err := saveStockToDatabase(newStock); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save stock to database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"stock_id": stockID,
		},
	})
}

func saveStockToDatabase(stock Stock) error {
	// Define formatted string for database connection
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// Attempt to connect to database
	db, err := sql.Open("postgres", postgresqlDbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create the stocks table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS stocks (
			id UUID PRIMARY KEY,
			stock_name VARCHAR(255) NOT NULL
		)`)
	if err != nil {
		return err
	}

	// Insert stock into the stocks table
	_, err = db.Exec("INSERT INTO stocks (id, stock_name) VALUES ($1, $2)", stock.ID, stock.StockName)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Define formatted string for database connection
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// Attempt to connect to database
	db, err := sql.Open("postgres", postgresqlDbInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Established a successful connection!")

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
