package main
import (
  "fmt"
  "time"
  "github.com/dgrijalva/jwt-go"
  "net/http"
  "github.com/gin-contrib/cors"
  "github.com/gin-gonic/gin"
  "github.com/Poomon001/day-trader/tree/poom/test")

// TODO: need env to store secret key
var secretKey = []byte("secret")

type Error struct {
	Success bool    `json:"success"`
	Data    *string `json:"data"`
	Message string  `json:"message"`
}


// user_name is a primary key in the DB used to identify user
type Register struct {
  UserName string `json:"user_name"`
  Password string `json:"password"`
  Name string `json:"name"`
}

type Login struct {
  UserName string `json:"user_name"`
  Password string `json:"password"`
}

type Response struct {
  Success bool `json:"success"`
  Data    *string `json:"data"`
}

type Claims struct {
  UserName string `json:"user_name"`
  jwt.StandardClaims
}

var userToPassword = map[string]string{
  "test1": "test001",
  "test2": "test002",
  "test3": "test003",
}

func handleError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := Error{
		Success: false,
		Data:    nil,
		Message: fmt.Sprintf("%s: %v", message, err),
	}
	c.IndentedJSON(statusCode, errorResponse)
}

func createToken(username string, expirationTime time.Time) (string, error) {
	claims := &Claims{
		UserName: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func createSession(c *gin.Context, token string, expirationTime time.Time) {
	c.SetCookie("session_token", token, int(expirationTime.Unix()), "/", "", false, true)
}

func postLogin(c *gin.Context) {
	fmt.Println("Authentication: %s", secretKey)
	var login Login

	// Verify request body
	if err := c.BindJSON(&login); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

  // TODO: Check if the username already exists in real DB instead of userToPassword map

	// Verify username and password
	expectedPassword, ok := userToPassword[login.UserName]
	if !ok || expectedPassword != login.Password {
		handleError(c, http.StatusBadRequest, "Unsuccessful Authentication", nil)
		return
	}

	// Create token
	expirationTime := time.Now().Add(10 * time.Minute)
	token, err := createToken(login.UserName, expirationTime)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to create token", err)
		return
	}

	// Create a cookie session
	createSession(c, token, expirationTime)

	// Respond
	loginResponse := Response{
		Success: true,
		Data:    &token,
	}

	c.IndentedJSON(http.StatusOK, loginResponse)
}


func postRegister(c *gin.Context) {

  var newRegister Register
  // TODO: Check if the username already exists in DB
  // yes? return error, otherwise continue

  if err := c.BindJSON(&newRegister); err != nil {
    handleError(c, http.StatusBadRequest, "Invalid request body", err)
    return
  }

  // TODO: Insert new user to DB
  // TODO: format return response
  c.IndentedJSON(http.StatusCreated, newRegister)
}

func getCookies(c *gin.Context) {
  cookie, err := c.Cookie("session_token")
  if err != nil {
	handleError(c, http.StatusBadRequest, "Unauthorized", err)
    return
  }
  fmt.Println(test.Test())
  c.String(http.StatusOK, "Cookie: " + cookie)
}

func main() {
  router := gin.Default()
  router.Use(cors.Default())
  router.POST("/login", postLogin)
  router.POST("/register", postRegister)
  router.GET("/eatCookies", getCookies)
  router.Run(":8888")
}