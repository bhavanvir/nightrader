package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

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

func Identification(c *gin.Context) {
	fmt.Println("Identification Middleware: ")

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		handleError(c, http.StatusBadRequest, "Authorization header missing", nil)
		c.Abort()
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		handleError(c, http.StatusBadRequest, "Invalid authorization header format", nil)
		c.Abort()
		return
	}
	tokenString := parts[1]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			handleError(c, http.StatusBadRequest, "Unauthorized Access", err)
			c.Abort()
			return
		}
		handleError(c, http.StatusBadRequest, "Failed to parse claims", err)
		c.Abort()
		return
	}

	if !token.Valid {
		handleError(c, http.StatusBadRequest, "Invalid token", err)
		c.Abort()
		return
	}

	fmt.Println(claims.ExpiresAt)
	fmt.Println(claims.UserName)
	c.Set("user_name", claims.UserName)
	c.Next()
}
