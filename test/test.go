package test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	)

func Test(c *gin.Context) {
	fmt.Println("Test Middleware: ")
}