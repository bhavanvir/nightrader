package main

import (
	"fmt"
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"github.com/Poomon001/day-trading-package/identification"
	"github.com/Poomon001/day-trading-package/tester"
	_ "github.com/lib/pq"
)

// TODO: need env to store secret key
var secretKey = []byte("secret")

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

// user_name is a primary key in the DB used to identify user
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
	Success bool    `json:"success"`
	Data    *string `json:"data"`
}

type Claims struct {
	Name     string `json:"name"`
	UserName string `json:"user_name"`
	jwt.StandardClaims
}

func handleError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := Error{
		Success: false,
		Data:    nil,
		Message: fmt.Sprintf("%s: %v", message, err),
	}
	c.IndentedJSON(statusCode, errorResponse)
}

func createToken(name string, username string, expirationTime time.Time) (string, error) {
	claims := &Claims{
		Name: name,
		UserName: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func createSession(c *gin.Context, token string, expirationTime time.Duration) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("session_token", token, int(expirationTime.Seconds()), "/", "http://localhost:3000", false, false)
}

func postLogin(c *gin.Context) {
	var login Login

	// Verify request body
	if err := c.BindJSON(&login); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", postgresqlDbInfo)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to connect to the database", err)
		return
	}
	defer db.Close()

    fmt.Println("Successfully connected to the database")

	// Login the user 
	var is_valid bool
	err = db.QueryRow("SELECT login($1, $2)", login.UserName, login.Password).Scan(&is_valid)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to query the database", err)
		return
	}
	if !is_valid {
		handleError(c, http.StatusBadRequest, "Password is incorrect", nil)
		return
	}

	// NOTE February 14 2024
	//
	// Made a function in database that handles login, blocks below commented out for now
	//
	// If this creates a problem, simply uncomment the blocks below and comment out the login
	// block above and it should work again


	// // Check if the username exists in DB
	// var count int
	// err = db.QueryRow("SELECT COUNT(*) FROM users WHERE user_name = $1", login.UserName).Scan(&count)
	// if err != nil {
	// 	handleError(c, http.StatusInternalServerError, "Failed to query the database", err)
	// 	return
	// }
	// if count == 0 {
	// 	handleError(c, http.StatusBadRequest, "Username does not exist", nil)
	// 	return
	// }

	// // Check password for the username
	// var correctPassword bool
	// err = db.QueryRow("SELECT (user_pass = crypt($1, user_pass)) AS is_valid FROM users WHERE user_name = $2", login.Password, login.UserName).Scan(&correctPassword)
	// if err != nil {
	// 	handleError(c, http.StatusInternalServerError, "Failed to query the database", err)
	// 	return
	// }
	// if !correctPassword {
	// 	handleError(c, http.StatusBadRequest, "Incorrect password", nil)
	// 	return
	// }

	// // Get the user's name
	// var name string
	// err = db.QueryRow("SELECT name FROM users WHERE user_name = $1", login.UserName).Scan(&name)
	// if err != nil {
	// 	handleError(c, http.StatusInternalServerError, "Failed to query the database", err)
	// 	return
	// }

	// Create token
	minutes := 10 * time.Minute
	expirationTime := time.Now().Add(minutes)
	token, err := createToken(name, login.UserName, expirationTime)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to create token", err)
		return
	}

	// Create a cookie session
	createSession(c, token, minutes)

	// Respond
	loginResponse := Response{
		Success: true,
		Data:    &token,
	}

	c.IndentedJSON(http.StatusOK, loginResponse)
}

func postRegister(c *gin.Context) {
	postgresqlDbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", postgresqlDbInfo)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to connect to the database", err)
		return
	}
	defer db.Close()

    fmt.Println("Successfully connected to the database")

	var newRegister Register

	if err := c.BindJSON(&newRegister); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Check if the username already exists in DB
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE user_name = $1", newRegister.UserName).Scan(&count)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to query the database", err)
		return
	}
	if count > 0 {
		handleError(c, http.StatusBadRequest, "Username already exists", nil)
		return
	}
	
	// Insert new user to DB
	_, err = db.Exec("INSERT INTO users (user_name, name, user_pass) VALUES ($1, $2, $3)", newRegister.UserName, newRegister.Name, newRegister.Password)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to insert new user to the database", err)
		return
	}

	// Format JSON response
	successResponse := Response{
		Success: true,
		Data:	nil,
	}

	c.IndentedJSON(http.StatusCreated, successResponse)
}

func getCookies(c *gin.Context) {
	cookie, err := c.Cookie("session_token")
	if err != nil {
		handleError(c, http.StatusBadRequest, "Unauthorized", err)
		return
	}
	c.String(http.StatusOK, "Cookie: "+cookie)
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
	tester.TestUser() // example how to use function from a package 
	router.POST("/login", identification.TestMiddleware, postLogin) // example how to use middlware from a package
	router.POST("/register", postRegister)
	router.GET("/eatCookies", getCookies)
	router.Run(":8888")
}
