package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	InitDB()
	router := gin.Default()
	router.GET("/enduser/:email", HandleGetUserByEmail)
	router.POST("/enduser", HandleAddUser)
	router.Run(":8080")
}

func InitDB() {
	var err error
	db, err = GetDatabase()
	if errorExists(err) {
		throwConnectionError(err)
	}
}

func throwConnectionError(err error) {
	fmt.Printf("Unexpected connection error: %v", err)
	os.Exit(3)
}

func HandleGetUserByEmail(context *gin.Context) {
	email := context.Param("email")
	user := map[string]interface{}{}
	db.Table("enduser").Where("email = ?", email).Take(&user)
	context.IndentedJSON(http.StatusOK, &user)
}

func HandleAddUser(context *gin.Context) {
	user := map[string]interface{}{}
	context.BindJSON(&user)
	result := db.Table("enduser").Create(&user)
	if errorExists(result.Error) {
		context.IndentedJSON(http.StatusBadRequest, result.Error)
	} else {
		context.IndentedJSON(http.StatusOK, `{"success" : "true"}`)
	}
}

func errorExists(err error) bool {
	return err != nil
}
