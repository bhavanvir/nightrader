package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Global variables for the prepared statements
var (
	stmtLogin  *sql.Stmt
	stmtExist  *sql.Stmt
	stmtInsert *sql.Stmt
)

// Global variable for the database connection
var user_db *sql.DB
var stock_db *sql.DB
var tx_db *sql.DB

// TODO: need env to store secret key
var secretKey = []byte("secret")

const (
    user_host = "user_database"
    stock_host = "stock_database"
    tx_host = "tx_database"
    // host     = "localhost" // for local testing
    user_port     = 5432
    stock_port    = 5431
    tx_port      = 5430
    user     = "nt_user"
    password = "db123"
    dbname   = "nt_db"
)

type ErrorResponse struct {
	Success bool              `json:"success"`
	Data    map[string]string `json:"data"`
}

type Register struct {
	UserName string `json:"user_name"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Login struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type Response struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
}

type Claims struct {
	Name     string `json:"name"`
	UserName string `json:"user_name"`
	jwt.StandardClaims
}

func handleError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := ErrorResponse{
		Success: false,
		Data:    map[string]string{"error": message},
	}
	c.IndentedJSON(statusCode, errorResponse)
}

func createToken(name string, username string, expirationTime time.Time) (string, error) {
	claims := &Claims{
		Name:     name,
		UserName: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func postLogin(c *gin.Context) {
	var login Login

	if err := c.BindJSON(&login); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	var name string
	var correctPassword bool
	err := stmtLogin.QueryRow(login.Password, login.UserName).Scan(&name, &correctPassword)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to query the database", err)
		return
	}
	if !correctPassword {
		handleError(c, http.StatusOK, "Incorrect password", nil)
		return
	}

	minutes := 30 * time.Minute
	expirationTime := time.Now().Add(minutes)
	token, err := createToken(name, login.UserName, expirationTime)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to create token", err)
		return
	}

	loginResponse := Response{
		Success: true,
		Data:    map[string]interface{}{"token": token},
	}

	c.IndentedJSON(http.StatusOK, loginResponse)
}

func postRegister(c *gin.Context) {
	var newRegister Register

	if err := c.BindJSON(&newRegister); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	var count int
	err := stmtExist.QueryRow(newRegister.UserName).Scan(&count)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to query the database", err)
		return
	}
	if count > 0 {
		handleError(c, http.StatusOK, "Username already exists", nil)
		return
	}

	_, err = stmtInsert.Exec(newRegister.UserName, newRegister.Name, newRegister.Password)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to insert new user to the database", err)
		return
	}

	successResponse := Response{
		Success: true,
		Data:    nil,
	}

	c.IndentedJSON(http.StatusCreated, successResponse)
}

func initializeDB() error {
	var err error
    postgresqlUserDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", user_host, user_port, user, password, dbname)
    user_db, err = sql.Open("postgres", postgresqlUserDbInfo)
    if err != nil {
        return fmt.Errorf("failed to connect to the user database: %v", err)
    }

    postgresqlStockDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", stock_host, stock_port, user, password, dbname)
    stock_db, err = sql.Open("postgres", postgresqlStockDbInfo)
    if err != nil {
        return fmt.Errorf("failed to connect to the stock database: %v", err)
    }

    postgresqlTxDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", tx_host, tx_port, user, password, dbname)
    tx_db, err = sql.Open("postgres", postgresqlTxDbInfo)
    if err != nil {
        return fmt.Errorf("failed to connect to the transaction database: %v", err)
    }

	// Ensure the database connection is fully established
    for {
        err = user_db.Ping()
        if err == nil {
            break
        }
        fmt.Println("Waiting for the user database connection to be established...")
        time.Sleep(1 * time.Second)
    }

    for {
        err = stock_db.Ping()
        if err == nil {
            break
        }
        fmt.Println("Waiting for the stock database connection to be established...")
        time.Sleep(1 * time.Second)
    }

    for {
        err = tx_db.Ping()
        if err == nil {
            break
        }
        fmt.Println("Waiting for the transaction database connection to be established...")
        time.Sleep(1 * time.Second)
    }

	return nil
}

func prepareStatements() error {
	var err error

	stmtLogin, err = user_db.Prepare("SELECT name, (user_pass = crypt($1, user_pass)) AS is_valid FROM users WHERE user_name = $2")
	if err != nil {
		return fmt.Errorf("failed to prepare login statement: %v", err)
	}

	stmtExist, err = user_db.Prepare("SELECT COUNT(*) FROM users WHERE user_name = $1")
	if err != nil {
		return fmt.Errorf("failed to prepare exist statement: %v", err)
	}

	stmtInsert, err = user_db.Prepare("INSERT INTO users (user_name, name, user_pass) VALUES ($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %v", err)
	}

	return nil
}

func main() {
	err := initializeDB()
	if err != nil {
		fmt.Printf("Failed to initialize the database: %v\n", err)
		return
	}
    defer user_db.Close()
    defer stock_db.Close()
    defer tx_db.Close()

	err = prepareStatements()
	if err != nil {
		fmt.Printf("Failed to prepare SQL statements: %v\n", err)
		return
	}
	defer stmtLogin.Close()
	defer stmtExist.Close()
	defer stmtInsert.Close()

    user_db.SetMaxOpenConns(10) // Set maximum number of open connections
    user_db.SetMaxIdleConns(5) // Set maximum number of idle connections

    stock_db.SetMaxOpenConns(10) // Set maximum number of open connections
    stock_db.SetMaxIdleConns(5) // Set maximum number of idle connections

    tx_db.SetMaxOpenConns(10) // Set maximum number of open connections
    tx_db.SetMaxIdleConns(5) // Set maximum number of idle connections

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "token"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	router.POST("/login", postLogin)
	router.POST("/register", postRegister)
	router.Run(":8888")
}