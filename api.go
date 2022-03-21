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
	router.PUT("/enduser/:email", HandlePutUser)
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
	resultOfFind := db.Table("enduser").Where("email = ?", email).Take(&user)
	if errorExists(resultOfFind.Error) {
		context.IndentedJSON(http.StatusNotFound, resultOfFind.Error)
	} else {
		context.IndentedJSON(http.StatusOK, &user)
	}
}

func HandleAddUser(context *gin.Context) {
	user := map[string]interface{}{}
	context.BindJSON(&user)
	result := db.Table("enduser").Create(&user)
	if errorExists(result.Error) {
		context.IndentedJSON(http.StatusConflict, result.Error)
	} else {
		context.IndentedJSON(http.StatusOK, `{"success" : "true"}`)
	}
}

func HandlePutUser(context *gin.Context) {
	email := context.Param("email")
	if !userExists(email) {
		context.IndentedJSON(http.StatusNotFound, `{"success" : "false", "error" : "user not found"}`)
	} else {
		user := map[string]interface{}{}
		context.BindJSON(&user)
		resultOfUpdate := db.Table("enduser").Where("email = ?", email).Update("password", user["password"])
		if errorExists(resultOfUpdate.Error) {
			context.IndentedJSON(http.StatusBadRequest, `{"success" : "false", "error" : "Invalid format"}`)
		} else {
			context.IndentedJSON(http.StatusOK, `{"success" : "true"}`)
		}
	}
}

func userExists(email string) bool {
	user := map[string]interface{}{}
	resultOfReadUser := db.Table("enduser").Where("email = ?", email).Take(&user)
	fmt.Printf("Read user error: %v\n", resultOfReadUser.Error)
	return !errorExists(resultOfReadUser.Error)
}

func errorExists(err error) bool {
	return err != nil
}
