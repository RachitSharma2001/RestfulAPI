package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var db *gorm.DB
var userNotFoundMsg map[string]interface{} = gin.H{"success": "false", "error": "user not found"}
var userAlreadyExistsMsg map[string]interface{} = gin.H{"success": "false", "error": "user already exists"}
var successMsg map[string]interface{} = gin.H{"success": "true"}
var invalidPasswordDataMsg map[string]interface{} = gin.H{"success": "false", "error": "Invalid password data given"}

func main() {
	InitDB()
	router := gin.Default()
	router.GET("/enduser/:email", HandleGetUserByEmail)
	router.POST("/enduser", HandleAddUser)
	router.PUT("/enduser/:email", HandlePutUser)
	router.DELETE("/enduser/:email", HandleDeleteUser)
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
	if !userExists(email) {
		context.IndentedJSON(http.StatusNotFound, userNotFoundMsg)
	} else {
		user := map[string]interface{}{}
		db.Table("enduser").Where("email = ?", email).Take(&user)
		context.IndentedJSON(http.StatusOK, &user)
	}
}

func HandlePutUser(context *gin.Context) {
	email := context.Param("email")
	user := map[string]interface{}{}
	bindErr := context.ShouldBindJSON(&user)
	if !userExists(email) {
		context.IndentedJSON(http.StatusNotFound, userNotFoundMsg)
	} else if errorExists(bindErr) {
		context.IndentedJSON(http.StatusBadRequest, invalidPasswordDataMsg)
	} else {
		db.Table("enduser").Where("email = ?", email).Update("password", user["password"])
		context.IndentedJSON(http.StatusOK, successMsg)
	}
}

func HandleDeleteUser(context *gin.Context) {
	email := context.Param("email")
	if !userExists(email) {
		context.IndentedJSON(http.StatusNotFound, userNotFoundMsg)
	} else {
		user := map[string]interface{}{}
		db.Table("enduser").Where("email = ?", email).Delete(&user)
		context.IndentedJSON(http.StatusOK, successMsg)
	}
}

func userExists(email string) bool {
	user := map[string]interface{}{}
	resultOfReadUser := db.Table("enduser").Where("email = ?", email).Take(&user)
	fmt.Printf("Read user error: %v\n", resultOfReadUser.Error)
	return !errorExists(resultOfReadUser.Error)
}

func HandleAddUser(context *gin.Context) {
	user := map[string]interface{}{}
	context.BindJSON(&user)
	result := db.Table("enduser").Create(&user)
	if errorExists(result.Error) {
		context.IndentedJSON(http.StatusConflict, userAlreadyExistsMsg)
	} else {
		context.IndentedJSON(http.StatusOK, successMsg)
	}
}

func errorExists(err error) bool {
	return err != nil
}
