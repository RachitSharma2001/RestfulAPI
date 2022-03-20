package main

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestReadUser(test *testing.T) {
	InitDB()
	recorder := httptest.NewRecorder()
	context := createContextWithEmail(recorder, "somebody@gmail.com")
	HandleGetUserByEmail(context)
	checkCorrectErrorCode(test, 200, recorder.Code)
	observedJson := recorder.Body.String()
	expectedJson := `{"id" : 10, "password" : "something", "email" : "somebody@gmail.com"}`
	checkCorrectJsonOutput(test, expectedJson, observedJson)
}

func createContextWithEmail(recorder *httptest.ResponseRecorder, givenEmail string) *gin.Context {
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
