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
	"github.com/stretchr/testify/assert"
)

func TestHandler_createList(t *testing.T) {

	type field struct {
		mockBehaviorCreate *mock_service.MockTodoList
		mockBehaviorHDel   *mock_service.MockTodoListCach
	}

	testTable := []struct {
		name                 string
		CtxNil               bool
		shouidFail           bool
		userId               int
		Id                   int
		inputBody            string
		inputList            todo.TodoList
		prepare              func(f *field, userId int, list todo.TodoList, Id int)
		expectedStatusCode   int // статус код ответа
		expectedResponseBody string
	}{
		{
			name:      "OK",
			userId:    2,
			Id:        5,
			inputBody: `{"title":"test", "description":"by testing"}`,
			inputList: todo.TodoList{
				Title:       "test",
				Description: "by testing",
			},
			prepare: func(f *field, userId int, list todo.TodoList, Id int) {
				gomock.InOrder(
					f.mockBehaviorCreate.EXPECT().Create(userId, list).Return(Id, nil),
					f.mockBehaviorHDel.EXPECT().HDelete(userId).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":5}`,
		},
		{
			name:      "Error getUserId",
			CtxNil:    true,
			userId:    2,
			inputBody: `{"title":"test", "description":"by testing"}`,
			inputList: todo.TodoList{
				Title:       "test",
				Description: "by testing",
			},
			prepare:              func(f *field, userId int, list todo.TodoList, Id int) {},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"user id not found\"}",
		},
		{
			name:                 "Empty Fields",
			userId:               2,
			inputBody:            `{}`,
			inputList:            todo.TodoList{},
			prepare:              func(f *field, userId int, list todo.TodoList, Id int) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Key: 'TodoList.Title' Error:Field validation for 'Title' failed on the 'required' tag"}`,
		},
		{
			name:                 "Empty Field Title",
			userId:               2,
			inputBody:            `{"description":"by testing"}`,
			inputList:            todo.TodoList{},
			prepare:              func(f *field, userId int, list todo.TodoList, Id int) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Key: 'TodoList.Title' Error:Field validation for 'Title' failed on the 'required' tag"}`,
		},
		{
			name:      "OK Empty Description",
			userId:    2,
			Id:        8,
			inputBody: `{"title":"test"}`,
			inputList: todo.TodoList{
				Title: "test",
			},
			prepare: func(f *field, userId int, list todo.TodoList, Id int) {
				gomock.InOrder(
					f.mockBehaviorCreate.EXPECT().Create(userId, list).Return(Id, nil),
					f.mockBehaviorHDel.EXPECT().HDelete(userId).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":8}`,
		},
		{
			name:      "Error Create",
			userId:    2,
			inputBody: `{"title":"test", "description":"by testing"}`,
			inputList: todo.TodoList{
				Title:       "test",
				Description: "by testing",
			},
			prepare: func(f *field, userId int, list todo.TodoList, Id int) {
				f.mockBehaviorCreate.EXPECT().Create(userId, list).Return(1, errors.New("Error Create"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error Create"}`,
		},
		{
			name:      "Error HDel",
			userId:    2,
			Id:        5,
			inputBody: `{"title":"test", "description":"by testing"}`,
			inputList: todo.TodoList{
				Title:       "test",
				Description: "by testing",
			},
			prepare: func(f *field, userId int, list todo.TodoList, Id int) {
				gomock.InOrder(
					f.mockBehaviorCreate.EXPECT().Create(userId, list).Return(Id, nil),
					f.mockBehaviorHDel.EXPECT().HDelete(userId).Return(errors.New("Error HDel")),
				)
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error HDel"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Init Deps
			c := gomock.NewController(t)
			defer c.Finish()

			f := field{
				mockBehaviorCreate: mock_service.NewMockTodoList(c),
				mockBehaviorHDel:   mock_service.NewMockTodoListCach(c),
			}

			if testCase.prepare != nil {
				testCase.prepare(&f, testCase.userId, testCase.inputList, testCase.Id)
			}

			//testCase.prepare.field.mockBehaviorCreate(todolist, testCase.userId, testCase.inputList)
			//testCase.mockBehaviorHDel(todolistcach, testCase.userId)

			services := &service.Service{TodoList: f.mockBehaviorCreate, TodoListCach: f.mockBehaviorHDel}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			if testCase.CtxNil {
				r.POST("/lists", handler.createList)
			} else {
				r.POST("/lists", func(c *gin.Context) { c.Set(userCtx, testCase.userId) }, handler.createList)
			}

			// Test Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/lists", bytes.NewBufferString(testCase.inputBody))

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
			//assert.Equal(t, testCase.userId, userId)
			//assert.Equal(t, testCase.shouidFail, errYes)

		})
	}
}
