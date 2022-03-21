package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type enduser struct {
	email    string
	id       int
	password string
}

func (e *enduser) getJsonRepresentingUser() string {
	return fmt.Sprintf(`{"id" : %d, "password" : "%s", "email" : "%s"}`, e.id, e.password, e.email)
}

func TestDelete(test *testing.T) {
	newUser := createNewUser()
	addUserToDb(test, newUser)
	deleteUser(test, newUser)
	verifyUserDeleted(test, newUser)
}

func addUserToDb(test *testing.T, userToAdd enduser) {
	test.Helper()
	InitDB()
	addUserRecorder := httptest.NewRecorder()
	callAddUserEndpoint(test, addUserRecorder, userToAdd)
	CloseDB()
}

func deleteUser(test *testing.T, userToDelete enduser) {
	test.Helper()
	InitDB()
	deleteUserRecorder := httptest.NewRecorder()
	deleteUserContext := createContextWithEmailEncoded(deleteUserRecorder, userToDelete.email)
	HandleDeleteUser(deleteUserContext)
	CloseDB()
}

func verifyUserDeleted(test *testing.T, user enduser) {
	test.Helper()
	InitDB()
	readUserRecorder := httptest.NewRecorder()
	callReadUserEndpoint(test, readUserRecorder, user.email)
	verifyNotFoundErrThrown(test, readUserRecorder)
	CloseDB()
}

func TestPost(test *testing.T) {
	test.Run("Add New User", func(subtest *testing.T) {
		addUserRecorder := httptest.NewRecorder()
		userToAdd := createNewUser()
		InitDB()
		callAddUserEndpoint(subtest, addUserRecorder, userToAdd)
		verifyUserAdded(subtest, addUserRecorder, userToAdd)
	})
	test.Run("Add Existent User", func(subtest *testing.T) {
		addUserRecorder := httptest.NewRecorder()
		userToAdd := createExistentUser()
		InitDB()
		callAddUserEndpoint(subtest, addUserRecorder, userToAdd)
		verifyConflictErrThrown(subtest, addUserRecorder)
	})
}

func createNewUser() enduser {
	userId := createRandomUserInteger()
	password := "randompass"
	email := fmt.Sprintf("ronald%d@gmail.com", userId)
	return enduser{id: userId, password: password, email: email}
}

func createRandomUserInteger() int {
	randomSeed := time.Now().UnixNano()
	maxId := 100000
	source := rand.NewSource(randomSeed)
	return rand.New(source).Intn(maxId)
}

func createExistentUser() enduser {
	return enduser{id: 10, password: "something", email: "somebody@gmail.com"}
}

func callAddUserEndpoint(test *testing.T, addUserRecorder *httptest.ResponseRecorder, userToAdd enduser) {
	test.Helper()
	jsonOfUserToAdd := userToAdd.getJsonRepresentingUser()
	addUserContext := createContextWithData(addUserRecorder, jsonOfUserToAdd)
	HandleAddUser(addUserContext)
}

func createContextWithData(recorder *httptest.ResponseRecorder, data string) *gin.Context {
	context, _ := gin.CreateTestContext(recorder)
	context.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(data)))
	return context
}

func verifyUserAdded(test *testing.T, addUserRecorder *httptest.ResponseRecorder, userToAdd enduser) {
	test.Helper()
	verifyNoErrThrown(test, addUserRecorder)
	verifyUserExistsInDb(test, userToAdd)
}

func verifyUserExistsInDb(test *testing.T, userToAdd enduser) {
	test.Helper()
	readUserJson := readUserFromDb(test, userToAdd.email)
	checkCorrectJsonOutput(test, userToAdd.getJsonRepresentingUser(), readUserJson)
}

func readUserFromDb(test *testing.T, userEmail string) string {
	test.Helper()
	readUserRecorder := httptest.NewRecorder()
	callReadUserEndpoint(test, readUserRecorder, userEmail)
	return readUserRecorder.Body.String()
}

func TestGet(test *testing.T) {
	test.Run("Reading an existent user", func(subtest *testing.T) {
		recorder := httptest.NewRecorder()
		userEmail := "james@gmail.com"
		InitDB()
		callReadUserEndpoint(subtest, recorder, userEmail)
		verifyCorrectRead(subtest, recorder)
	})
	test.Run("Reading a nonexistent user", func(subtest *testing.T) {
		recorder := httptest.NewRecorder()
		userEmail := "notanyonehere@gmail.com"
		InitDB()
		callReadUserEndpoint(subtest, recorder, userEmail)
		verifyNotFoundErrThrown(subtest, recorder)
	})
}

func callReadUserEndpoint(test *testing.T, recorder *httptest.ResponseRecorder, userEmail string) {
	test.Helper()
	contextWithEmail := createContextWithEmailEncoded(recorder, userEmail)
	HandleGetUserByEmail(contextWithEmail)
}

func createContextWithEmailEncoded(recorder *httptest.ResponseRecorder, givenEmail string) *gin.Context {
	context, _ := gin.CreateTestContext(recorder)
	context.Params = []gin.Param{
		{
			Key:   "email",
			Value: givenEmail,
		},
	}
	return context
}

func verifyCorrectRead(test *testing.T, recorder *httptest.ResponseRecorder) {
	test.Helper()
	verifyNoErrThrown(test, recorder)
	verifyCorrectUserRead(test, recorder)
}

func verifyNoErrThrown(test *testing.T, recorder *httptest.ResponseRecorder) {
	test.Helper()
	expectedErrCode := http.StatusOK
	observedErrCode := recorder.Code
	checkCorrectErrorCode(test, expectedErrCode, observedErrCode)
}

func verifyNotFoundErrThrown(test *testing.T, recorder *httptest.ResponseRecorder) {
	test.Helper()
	expectedErrCode := http.StatusNotFound
	observedErrCode := recorder.Code
	checkCorrectErrorCode(test, expectedErrCode, observedErrCode)
}

func verifyConflictErrThrown(test *testing.T, recorder *httptest.ResponseRecorder) {
	test.Helper()
	expectedErrCode := http.StatusConflict
	observedErrCode := recorder.Code
	checkCorrectErrorCode(test, expectedErrCode, observedErrCode)
}

func checkCorrectErrorCode(test *testing.T, expectedErrCode int, observedErrCode int) {
	test.Helper()
	if !errorCodesMatch(expectedErrCode, observedErrCode) {
		test.Fatalf("Expected error code of %d, instead got %d", expectedErrCode, observedErrCode)
	}
}

func errorCodesMatch(expectedErrCode, observedErrCode int) bool {
	return expectedErrCode == observedErrCode
}

func verifyCorrectUserRead(test *testing.T, recorder *httptest.ResponseRecorder) {
	test.Helper()
	observedJson := recorder.Body.String()
	expectedJson := `{"id" : 5, "password" : "sfsdfs", "email" : "james@gmail.com"}`
	checkCorrectJsonOutput(test, expectedJson, observedJson)
}

func checkCorrectJsonOutput(test *testing.T, expectedJson string, observedJson string) {
	test.Helper()
	require.JSONEq(test, expectedJson, observedJson)
}

func TestPut(test *testing.T) {
	test.Run("Update existing user password", func(subtest *testing.T) {
		addUserRecorder := httptest.NewRecorder()
		userEmail := "somebody@gmail.com"
		newPassword := createRandomPassword()
		InitDB()
		changeUserPassword(subtest, addUserRecorder, userEmail, newPassword)
		verifyUpdateOccurred(subtest, addUserRecorder, userEmail, newPassword)
	})
	test.Run("Update non-existing user password", func(subtest *testing.T) {
		addUserRecorder := httptest.NewRecorder()
		userEmail := "adfsdfsad@gmail.com"
		newPassword := createRandomPassword()
		InitDB()
		changeUserPassword(subtest, addUserRecorder, userEmail, newPassword)
		verifyNotFoundErrThrown(subtest, addUserRecorder)
	})
}

func createRandomPassword() string {
	randomId := createRandomUserInteger()
	return fmt.Sprintf("pass%d", randomId)
}

func changeUserPassword(test *testing.T, addUserRecorder *httptest.ResponseRecorder, userEmail string, newPassword string) {
	putUserContext := createContextWithEmailEncoded(addUserRecorder, userEmail)
	putUserContext.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(fmt.Sprintf(`{"password" : "%s"}`, newPassword))))
	HandlePutUser(putUserContext)
}

func verifyUpdateOccurred(test *testing.T, addUserRecorder *httptest.ResponseRecorder, userEmail string, newPassword string) {
	test.Helper()
	verifyNoErrThrown(test, addUserRecorder)
	verifyUserPasswordChanged(test, userEmail, newPassword)
}

func verifyUserPasswordChanged(test *testing.T, userEmail string, newPassword string) {
	test.Helper()
	readUserRecorder := httptest.NewRecorder()
	callReadUserEndpoint(test, readUserRecorder, userEmail)
	verifyCorrectPassword(test, readUserRecorder, newPassword)
}

func verifyCorrectPassword(test *testing.T, readUserRecorder *httptest.ResponseRecorder, newPassword string) {
	test.Helper()
	observedPassword := getUserPassword(readUserRecorder.Body.String())
	if observedPassword != newPassword {
		test.Errorf("Expected password of %q, instead got %q", newPassword, observedPassword)
	}
}

func getUserPassword(userInfo string) string {
	var userMap map[string]string
	json.Unmarshal([]byte(userInfo), &userMap)
	return userMap["password"]
}
