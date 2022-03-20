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

func (e *enduser) getJsonRepresentingUser() string {
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
	randomSeed := time.Now().UnixNano()
	maxId := 100000
	s1 := rand.NewSource(randomSeed)
	return rand.New(s1).Intn(maxId)
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

func TestReadUser(test *testing.T) {
	recorder := httptest.NewRecorder()
	userEmail := "somebody@gmail.com"
	InitDB()
	callReadUserEndpoint(test, recorder, userEmail)
	verifyCorrectRead(test, recorder)
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
	expectedJson := `{"id" : 10, "password" : "something", "email" : "somebody@gmail.com"}`
	checkCorrectJsonOutput(test, expectedJson, observedJson)
}

func checkCorrectJsonOutput(test *testing.T, expectedJson string, observedJson string) {
	test.Helper()
	require.JSONEq(test, expectedJson, observedJson)
}
