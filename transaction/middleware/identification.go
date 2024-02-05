package middleware

import (
"net/http"
"fmt"
"github.com/dgrijalva/jwt-go"
"github.com/gin-gonic/gin"
)

// TODO: need env to store secret key
var secretKey = []byte("secret")

type Error struct {
	Success bool    `json:"success"`
	Data    *string `json:"data"`
	Message string  `json:"message"`
}

type Claims struct {
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

func Identification(c *gin.Context){
	fmt.Println("Identification Middleware: ")
	cookie, err := c.Cookie("session_token")
	if err != nil {
		handleError(c, http.StatusBadRequest, "Failed to obtain the authentication token", err)
		return
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(cookie, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			handleError(c, http.StatusBadRequest, "Unauthorized Access", err)
			return
		}
		handleError(c, http.StatusBadRequest, "Failed to parse claims", err)
		return
	}

	if !token.Valid {
		handleError(c, http.StatusBadRequest, "Invalid token", err)
		return
	}

	fmt.Println(claims.ExpiresAt)
	fmt.Println(claims.UserName)
	c.Set("user_name", claims.UserName)
	c.Next()
}