package main

import (
	"database/sql"
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
var badRequestMsg map[string]interface{} = gin.H{"success": "false", "error": "Invalid data given"}

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

func CloseDB() {
	sqlDb, err := db.DB()
	if errorExists(err) {
		throwCloseError(err)
	} else {
		closeSqlDB(sqlDb)
	}
}

func throwConnectionError(err error) {
	fmt.Printf("Unexpected connection error: %v", err)
	os.Exit(3)
}

func throwCloseError(err error) {
	fmt.Printf("Unexpected error while closing: %v", err)
	os.Exit(3)
}

func closeSqlDB(sqlDb *sql.DB) {
	sqlDb.Close()
}

func HandleAddUser(context *gin.Context) {
	user := map[string]interface{}{}
	bindErr := context.ShouldBindJSON(&user)
	createUserErr := db.Table("enduser").Create(&user).Error
	if errorExists(bindErr) {
		encodeBadRequestInContext(context)
	} else if errorExists(createUserErr) {
		encodeConflictErrInContext(context)
	} else {
		encodeSuccessInContext(context)
	}
}

func HandleGetUserByEmail(context *gin.Context) {
	email := context.Param("email")
	if !userExists(email) {
		encodeNotFoundErrInContext(context)
	} else {
		user := getUserFromDB(email)
		encodeUserInfoInContext(context, user)
	}
}

func getUserFromDB(email string) map[string]interface{} {
	user := map[string]interface{}{}
	db.Table("enduser").Where("email = ?", email).Take(&user)
	return user
}

func HandlePutUser(context *gin.Context) {
	email := context.Param("email")
	user := map[string]interface{}{}
	bindErr := context.ShouldBindJSON(&user)
	if !userExists(email) {
		encodeNotFoundErrInContext(context)
	} else if errorExists(bindErr) {
		encodeBadRequestInContext(context)
	} else {
		changeUserPasswordInDb(email, user["password"])
		encodeSuccessInContext(context)
	}
}

func changeUserPasswordInDb(email string, newPassword interface{}) {
	db.Table("enduser").Where("email = ?", email).Update("password", newPassword)
}

func HandleDeleteUser(context *gin.Context) {
	email := context.Param("email")
	if !userExists(email) {
		encodeNotFoundErrInContext(context)
	} else {
		deleteUserFromDb(email)
		encodeSuccessInContext(context)
	}
}

func deleteUserFromDb(email string) {
	user := map[string]interface{}{}
	db.Table("enduser").Where("email = ?", email).Delete(&user)
}

func userExists(email string) bool {
	user := map[string]interface{}{}
	resultOfReadUser := db.Table("enduser").Where("email = ?", email).Take(&user)
	return !errorExists(resultOfReadUser.Error)
}

func encodeNotFoundErrInContext(context *gin.Context) {
	context.IndentedJSON(http.StatusNotFound, userNotFoundMsg)
}

func encodeSuccessInContext(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, successMsg)
}

func encodeConflictErrInContext(context *gin.Context) {
	context.IndentedJSON(http.StatusConflict, userAlreadyExistsMsg)
}

func encodeUserInfoInContext(context *gin.Context, user map[string]interface{}) {
	context.IndentedJSON(http.StatusOK, &user)
}

func encodeBadRequestInContext(context *gin.Context) {
	context.IndentedJSON(http.StatusBadRequest, badRequestMsg)
}

func errorExists(err error) bool {
	return err != nil
}
