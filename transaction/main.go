package main
import ("fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/dgrijalva/jwt-go"
	"example/Transactions/middleware")

// TODO: need env to store a secret key
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

type AddMoney struct {
	Amount int `json:"amount"`
}

type PostResponse struct {
	Success bool `json:"success"`
	Data    *string `json:"data"`
}

func handleError(c *gin.Context, statusCode int, message string, err error) {
	errorResponse := Error{
		Success: false,
		Data:    nil,
		Message: fmt.Sprintf("%s: %v", message, err),
	}
	c.IndentedJSON(statusCode, errorResponse)
}

func addMoneyToWallet(c *gin.Context) {
	user_name, err := c.Get("user_name")
	
	if !err {
		handleError(c, http.StatusBadRequest, "Failed to obtain the user name", nil)
		return
	}

	var addMoney AddMoney
	if err := c.ShouldBindJSON(&addMoney); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// TODO: add the money to the user's wallet in database
	fmt.Println("User: ", user_name)
	c.IndentedJSON(http.StatusOK, addMoney)
}

func getCookies(c *gin.Context) {
	cookie, err := c.Cookie("session_token")
	if err != nil {
	  handleError(c, http.StatusBadRequest, "Unauthorized", err)
	  return
	}
	c.String(http.StatusOK, "Cookie: " + cookie)
  }

func main() {
	router := gin.Default()
  	router.Use(cors.Default())
	router.Use(middleware.Identification) // apply this middle where to any routes below
	router.POST("/addMoneyToWallet", addMoneyToWallet)
	router.GET("/eatCookies", getCookies)
	router.Run(":5000")
}