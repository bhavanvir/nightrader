package main
import (
  "time"
  "github.com/dgrijalva/jwt-go"
  "net/http"
  "github.com/gin-contrib/cors"
  "github.com/gin-gonic/gin")

// TODO: need env to store secret key
var secretKey = []byte("secret")

type Error struct {
	Success bool    `json:"success"`
	Data    *string `json:"data"`
	Message string  `json:"message"`
}


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

func postLogin(c *gin.Context) {
  var login Login

  // verify request body
  if err := c.BindJSON(&login); err != nil {
    errorResponse := Error{
      Success: false,
      Data:    nil,
      Message: err.Error(),
    }
    c.IndentedJSON(http.StatusBadRequest, errorResponse) 
    return
  }

  expectedPassword, ok := userToPassword[login.UserName]

  // verify username and password
  if !ok || expectedPassword != login.Password {
    errorResponse := Error{
      Success: false,
      Data:    nil,
      Message: "Insuccessful Authentication",
    }
    c.IndentedJSON(http.StatusBadRequest, errorResponse) 
    return
  }

  // create token
  expirationTime := time.Now().Add(10 * time.Minute)
  claims := &Claims{
    UserName: login.UserName,
    StandardClaims: jwt.StandardClaims{
      ExpiresAt: expirationTime.Unix(),
    },
  }
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  tokenString, err := token.SignedString(secretKey)

  if err != nil {
    errorResponse := Error{
      Success: false,
      Data:    nil,
      Message: "Fail to create token",
    }
    c.IndentedJSON(http.StatusInternalServerError, errorResponse)
    return
  }
    
  loginResponse := Response{
    Success: true,
    Data:    &tokenString,
  }

  // Create a cookie session
	c.SetCookie("session_token", tokenString, int(expirationTime.Unix()), "/", "", false, true)

  c.IndentedJSON(http.StatusOK, loginResponse) 
}

func postRegister(c *gin.Context) {

  var newRegister Register
  // TODO: Check if the username already exists in DB
  // yes? return error, otherwise continue

  if err := c.BindJSON(&newRegister); err != nil {
    errorResponse := Error{
      Success: false,
      Data:    nil,
      Message: err.Error(),
    }
    c.IndentedJSON(http.StatusBadRequest, errorResponse) 
    return
  }

  // TODO: Insert new user to DB
  // TODO: format return response
  c.IndentedJSON(http.StatusCreated, newRegister)
}

func getCookies(c *gin.Context) {
  cookie, err := c.Cookie("session_token")
  if err != nil {
    c.String(http.StatusUnauthorized, "Unauthorized")
    return
  }
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