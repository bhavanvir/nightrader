package main
import ("net/http"
		"github.com/gin-gonic/gin"
		"github.com/gin-contrib/cors")

type User struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Cash int `json:"cash"`
}

var users = []User {
	{ID: "1", Name: "John", Cash: 1000},
	{ID: "2", Name: "Jane", Cash: 2000},
	{ID: "3", Name: "Jack", Cash: 3000},
}

func getUsers(c *gin.Context) {
	c.JSON(http.StatusOK, users)
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/users", getUsers)
	router.Run(":8080")
}

