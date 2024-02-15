package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default() // Create a new router
	router.Run(":8585")     // Run the router on port  8181
}
