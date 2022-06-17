package handler

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"
	"todo-app"
	"todo-app/pkg/service"
	mock_service "todo-app/pkg/service/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
)

func TestHandler_signUp(t *testing.T) { //Тест функция метода handler`а sign-up

	type mockBehavior func(s *mock_service.MockAuthorization, user todo.User) // функция mock`а

	testTable := []struct { // Создание тестовой таблицы, содержит структуру из...
		name                 string    // имя теста
		inputBody            string    // тело запроса
		inputUser            todo.User // структура пользователя
		mockBehavior         mockBehavior
		expectedStatusCode   int    // статус код ответа
		expectedResponseBody string // ожидаемое тело ответа.
	}{
		{ // позитивный сценарий
			name:      "OK",
			inputBody: `{"name":"Test", "username":"test","password":"qwerty"}`,
			inputUser: todo.User{
				Name:     "Test",
				Username: "test",
				Password: "qwerty",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user todo.User) {
				s.EXPECT().CreateUser(user).Return(1, nil) // ожидаем получить функцию CreateUser, которая возвращает (1, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":1}`,
		},
		{
			name:                 "Empty Fields",
			inputBody:            `{"username":"test","password":"qwerty"}`,
			mockBehavior:         func(s *mock_service.MockAuthorization, user todo.User) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "Service Failure",
			inputBody: `{"name":"Test", "username":"test","password":"qwerty"}`,
			inputUser: todo.User{
				Name:     "Test",
				Username: "test",
				Password: "qwerty",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user todo.User) {
				s.EXPECT().CreateUser(user).Return(1, errors.New("service failure"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service failure"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Init Deps
			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_service.NewMockAuthorization(c)
			testCase.mockBehavior(auth, testCase.inputUser)

			services := &service.Service{Authorization: auth}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			r.POST("/sign-up", handler.signUp)

			// Test Request
			w := httptest.NewRecorder() // Захватывает возвращенный HTTP-ответ
			req := httptest.NewRequest("POST", "/sign-up",
				bytes.NewBufferString(testCase.inputBody))

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, w.Code, testCase.expectedStatusCode)
			assert.Equal(t, w.Body.String(), testCase.expectedResponseBody)
		})
	}
}

func TestHandler_signIn(t *testing.T) { //Тест функция метода handler`а sign-in

	type mockBehavior func(s *mock_service.MockAuthorization, username string, password string) // функция mock`а

	testTable := []struct { // Создание тестовой таблицы, содержит структуру из...
		name         string // имя теста
		inputBody    string // тело запроса
		username     string
		password     string
		token        string
		mockBehavior mockBehavior

		expectedStatusCode   int    // статус код ответа
		expectedResponseBody string // ожидаемое тело ответа.
	}{
		{ // позитивный сценарий
			name:      "OK",
			inputBody: `{"username":"test","password":"qwerty"}`,
			username:  "test",
			password:  "qwerty",
			token:     "token",
			mockBehavior: func(s *mock_service.MockAuthorization, username string, password string) {
				s.EXPECT().GenerateToken(username, password).Return("token", nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"token":"token"}`,
		},
		{
			name:      "Empty Field password",
			inputBody: `{"username":"test"}`,
			mockBehavior: func(s *mock_service.MockAuthorization, username string, password string) {
			},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body: Key: 'signInInput.Password' Error:Field validation for 'Password' failed on the 'required' tag"}`,
		},
		{
			name:      "Empty Field username",
			inputBody: `{"password":"qwerty"}`,
			mockBehavior: func(s *mock_service.MockAuthorization, username string, password string) {
			},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body: Key: 'signInInput.Username' Error:Field validation for 'Username' failed on the 'required' tag"}`,
		},
		{
			name: "Empty Fields",
			mockBehavior: func(s *mock_service.MockAuthorization, username string, password string) {
			},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body: EOF"}`,
		},
		{
			name:      "Service Failure",
			inputBody: `{"name":"Test", "username":"test","password":"qwerty"}`,
			username:  "test",
			password:  "qwerty",
			mockBehavior: func(s *mock_service.MockAuthorization, username string, password string) {
				s.EXPECT().GenerateToken(username, password).Return("", errors.New("token invalid"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"token invalid"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Init Deps
			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_service.NewMockAuthorization(c)
			testCase.mockBehavior(auth, testCase.username, testCase.password)

			services := &service.Service{Authorization: auth}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			r.POST("/sign-in", handler.signIn)

			// Test Request
			w := httptest.NewRecorder() // Захватывает возвращенный HTTP-ответ
			req := httptest.NewRequest("POST", "/sign-in",
				bytes.NewBufferString(testCase.inputBody))

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, w.Code, testCase.expectedStatusCode)
			assert.Equal(t, w.Body.String(), testCase.expectedResponseBody)
		})
	}
}
