package main

import (
	"bytes"
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
	id       int
	password string
	email    string
}

func (e *enduser) getJsonOfUser() string {
	return fmt.Sprintf(`{"id" : %d, "password" : "%s", "email" : "%s"}`, e.id, e.password, e.email)
}

func TestAddUser(test *testing.T) {
	test.Run("Add New User", func(subtest *testing.T) {
		addUserRecorder := httptest.NewRecorder()
		userToAdd := createUserToAdd()
		InitDB()
		callAddUserEndpoint(subtest, addUserRecorder, userToAdd)
		verifyUserAdded(subtest, addUserRecorder, userToAdd)
	})
}

func createUserToAdd() enduser {
	userId := createRandomUserId()
	password := "randompass"
	email := fmt.Sprintf("ronald%d@gmail.com", userId)
	return enduser{id: userId, password: password, email: email}
}

func createRandomUserId() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	return rand.New(s1).Intn(100000)
}

func callAddUserEndpoint(test *testing.T, addUserRecorder *httptest.ResponseRecorder, userToAdd enduser) {
	jsonOfUserToAdd := userToAdd.getJsonOfUser()
	addUserContext := createContextWithData(addUserRecorder, jsonOfUserToAdd)
	HandleAddUser(addUserContext)
}

func createContextWithData(recorder *httptest.ResponseRecorder, data string) *gin.Context {
	context, _ := gin.CreateTestContext(recorder)
	context.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(data)))
	return context
}

func verifyUserAdded(test *testing.T, addUserRecorder *httptest.ResponseRecorder, userToAdd enduser) {
	checkCorrectErrorCode(test, http.StatusOK, addUserRecorder.Code)
	readUserRecorder := httptest.NewRecorder()
	readUserContext := createContextWithEmailEncoded(readUserRecorder, userToAdd.email)
	readUserJson := getReadUserJsonResult(readUserRecorder, readUserContext)
	checkCorrectJsonOutput(test, userToAdd.getJsonOfUser(), readUserJson)
}

func getReadUserJsonResult(readUserRecorder *httptest.ResponseRecorder, readUserContext *gin.Context) string {
	HandleGetUserByEmail(readUserContext)
	return readUserRecorder.Body.String()
}

func TestReadUser(test *testing.T) {
	recorder := httptest.NewRecorder()
	InitDB()
	callReadUserEndpoint(test, recorder)
	verifyCorrectRead(test, recorder)
}

func callReadUserEndpoint(test *testing.T, recorder *httptest.ResponseRecorder) {
	test.Helper()
	contextWithEmail := createContextWithEmailEncoded(recorder, "somebody@gmail.com")
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
	verifyCorrectErrCodeForGet(test, recorder)
	verifyCorrectOutputForGet(test, recorder)
}

func verifyCorrectErrCodeForGet(test *testing.T, recorder *httptest.ResponseRecorder) {
	test.Helper()
	expectedErrCode := http.StatusOK
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

func verifyCorrectOutputForGet(test *testing.T, recorder *httptest.ResponseRecorder) {
	test.Helper()
	observedJson := recorder.Body.String()
	expectedJson := `{"id" : 10, "password" : "something", "email" : "somebody@gmail.com"}`
	checkCorrectJsonOutput(test, expectedJson, observedJson)
}

func checkCorrectJsonOutput(test *testing.T, expectedJson string, observedJson string) {
	test.Helper()
	require.JSONEq(test, expectedJson, observedJson)
}
