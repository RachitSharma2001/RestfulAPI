package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var db *gorm.DB

func HandleGetUserByEmail(context *gin.Context) {
	email := context.Param("email")
	user := map[string]interface{}{}
	db.Table("enduser").Where("email = ?", email).Take(&user)
	context.IndentedJSON(http.StatusOK, &user)
}

func InitDB() {
	var err error
	db, err = GetDatabase()
	if err != nil {
		fmt.Printf("Unexpected connection error: %v", err)
		os.Exit(3)
	}
}

func main() {
	InitDB()
	router := gin.Default()
	router.GET("/enduser/:email", HandleGetUserByEmail)
	router.Run(":8080")
}
