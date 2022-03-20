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

func TestAddUser(test *testing.T) {
	test.Run("Add New User", func(subtest *testing.T) {
		InitDB()
		addUserRecorder := httptest.NewRecorder()
		s1 := rand.NewSource(time.Now().UnixNano())
		userId := rand.New(s1).Intn(100000)
		email := fmt.Sprintf("ronald%d@gmail.com", userId)
		addUserJson := fmt.Sprintf(`{"id" : %d, "password" : "randompass", "email" : "%s"}`, userId, email)
		addUserContext := createContextWithData(addUserRecorder, addUserJson)
		HandleAddUser(addUserContext)
		checkCorrectErrorCode(test, 200, addUserRecorder.Code)
		readUserRecorder := httptest.NewRecorder()
		readUserContext := createContextWithEmailEncoded(readUserRecorder, email)
		readUserJson := getReadUserJsonResult(readUserRecorder, readUserContext)
		checkCorrectJsonOutput(test, addUserJson, readUserJson)
	})
}

func createContextWithData(recorder *httptest.ResponseRecorder, givenData string) *gin.Context {
	context, _ := gin.CreateTestContext(recorder)
	context.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(givenData)))
	return context
}

func getReadUserJsonResult(readUserRecorder *httptest.ResponseRecorder, readUserContext *gin.Context) string {
	HandleGetUserByEmail(readUserContext)
	return readUserRecorder.Body.String()
}

func TestReadUser(test *testing.T) {
	InitDB()
	recorder := httptest.NewRecorder()
	context := createContextWithEmailEncoded(recorder, "somebody@gmail.com")
	HandleGetUserByEmail(context)
	checkCorrectErrorCode(test, 200, recorder.Code)
	observedJson := recorder.Body.String()
	expectedJson := `{"id" : 10, "password" : "something", "email" : "somebody@gmail.com"}`
	checkCorrectJsonOutput(test, expectedJson, observedJson)
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

func checkCorrectErrorCode(subtest *testing.T, expectedErrCode int, observedErrCode int) {
	subtest.Helper()
	if expectedErrCode != observedErrCode {
		subtest.Fatalf("Expected error code of %d, instead got %d", expectedErrCode, observedErrCode)
	}
}

func checkCorrectJsonOutput(subtest *testing.T, expectedJson string, observedJson string) {
	subtest.Helper()
	require.JSONEq(subtest, expectedJson, observedJson)
}
