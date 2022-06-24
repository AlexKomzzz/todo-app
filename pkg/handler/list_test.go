package handler

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"todo-app"
	"todo-app/pkg/service"
	mock_service "todo-app/pkg/service/mocks"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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
			name:                 "Error getUserId",
			CtxNil:               true,
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
		})
	}
}

func TestHandler_getAllLists(t *testing.T) {

	type field struct {
		mockBehaviorH      *mock_service.MockTodoListCach
		mockBehaviorGetAll *mock_service.MockTodoList
	}

	testTable := []struct {
		name                 string
		CtxNil               bool
		userId               int
		ReturnHGet_InputHSet string
		ReturnGetAll         []todo.TodoList
		prepare              func(f *field, userId int, ReturnHGet_InputHSet string, ReturnGetAll []todo.TodoList)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                 "OK HGet",
			userId:               55,
			ReturnHGet_InputHSet: "[{\"id\":1,\"title\":\"test1\",\"description\":\"by testing 1\"},{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\"}]",
			prepare: func(f *field, userId int, ReturnHGet_InputHSet string, ReturnGetAll []todo.TodoList) {
				f.mockBehaviorH.EXPECT().HGet(userId, -1).Return(ReturnHGet_InputHSet, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"data\":[{\"id\":1,\"title\":\"test1\",\"description\":\"by testing 1\"},{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\"}]}",
		},
		{
			name:                 "Error getUserId",
			CtxNil:               true,
			prepare:              func(f *field, userId int, ReturnHGet_InputHSet string, ReturnGetAll []todo.TodoList) {},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"user id not found\"}",
		},
		{
			name:   "Error HGet",
			userId: 55,
			prepare: func(f *field, userId int, ReturnHGet_InputHSet string, ReturnGetAll []todo.TodoList) {
				f.mockBehaviorH.EXPECT().HGet(userId, -1).Return("", errors.New("Error HGet"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error HGet"}`,
		},
		{
			name:                 "OK redis.Nil",
			userId:               55,
			ReturnHGet_InputHSet: "[{\"id\":1,\"title\":\"test1\",\"description\":\"by testing 1\"},{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\"}]",
			ReturnGetAll: []todo.TodoList{
				{
					Id:          1,
					Title:       "test1",
					Description: "by testing 1",
				},
				{
					Id:          4,
					Title:       "test4",
					Description: "by testing 4",
				},
			},
			prepare: func(f *field, userId int, ReturnHGet_InputHSet string, ReturnGetAll []todo.TodoList) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(userId, -1).Return("", redis.Nil),
					f.mockBehaviorGetAll.EXPECT().GetAll(userId).Return(ReturnGetAll, nil),
					f.mockBehaviorH.EXPECT().HSet(userId, -1, ReturnHGet_InputHSet).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"data\":[{\"id\":1,\"title\":\"test1\",\"description\":\"by testing 1\"},{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\"}]}",
		},
		{
			name:   "Error GetAll",
			userId: 55,
			prepare: func(f *field, userId int, ReturnHGet_InputHSet string, ReturnGetAll []todo.TodoList) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(userId, -1).Return("", redis.Nil),
					f.mockBehaviorGetAll.EXPECT().GetAll(userId).Return(nil, errors.New("Error GetAll")),
				)
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error GetAll"}`,
		},
		{
			name:                 "Error HSet",
			userId:               55,
			ReturnHGet_InputHSet: "[{\"id\":1,\"title\":\"test1\",\"description\":\"by testing 1\"},{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\"}]",
			ReturnGetAll: []todo.TodoList{
				{
					Id:          1,
					Title:       "test1",
					Description: "by testing 1",
				},
				{
					Id:          4,
					Title:       "test4",
					Description: "by testing 4",
				},
			},
			prepare: func(f *field, userId int, ReturnHGet_InputHSet string, ReturnGetAll []todo.TodoList) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(userId, -1).Return("", redis.Nil),
					f.mockBehaviorGetAll.EXPECT().GetAll(userId).Return(ReturnGetAll, nil),
					f.mockBehaviorH.EXPECT().HSet(userId, -1, ReturnHGet_InputHSet).Return(errors.New("Error HSet")),
				)
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error HSet"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Init Deps
			c := gomock.NewController(t)
			defer c.Finish()

			f := field{
				mockBehaviorH:      mock_service.NewMockTodoListCach(c),
				mockBehaviorGetAll: mock_service.NewMockTodoList(c),
			}

			if testCase.prepare != nil {
				testCase.prepare(&f, testCase.userId, testCase.ReturnHGet_InputHSet, testCase.ReturnGetAll)
			}

			services := &service.Service{TodoListCach: f.mockBehaviorH, TodoList: f.mockBehaviorGetAll}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			if testCase.CtxNil {
				r.GET("/lists", handler.getAllLists)
			} else {
				r.GET("/lists", func(c *gin.Context) { c.Set(userCtx, testCase.userId) }, handler.getAllLists)
			}

			// Test Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/lists", nil)

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_getListById(t *testing.T) {

	type field struct {
		mockBehaviorH       *mock_service.MockTodoListCach
		mockBehaviorGetById *mock_service.MockTodoList
	}

	testTable := []struct {
		name                 string
		CtxNil               bool
		ErrId                bool
		Id                   int
		userId               int
		ReturnHGet_InputHSet string
		ReturnGetById        todo.TodoList
		prepare              func(f *field, userId, Id int, ReturnHGet_InputHSet string, ReturnGetById todo.TodoList)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                 "OK HGet",
			Id:                   4,
			userId:               55,
			ReturnHGet_InputHSet: "{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\"}",
			prepare: func(f *field, userId, Id int, ReturnHGet_InputHSet string, ReturnGetById todo.TodoList) {
				f.mockBehaviorH.EXPECT().HGet(userId, Id).Return(ReturnHGet_InputHSet, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\"}",
		},
		{
			name:                 "Error getUserId",
			CtxNil:               true,
			prepare:              func(f *field, userId, Id int, ReturnHGet_InputHSet string, ReturnGetById todo.TodoList) {},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"user id not found\"}",
		},
		{
			name:                 "Error Atoi Id",
			ErrId:                true,
			prepare:              func(f *field, userId, Id int, ReturnHGet_InputHSet string, ReturnGetById todo.TodoList) {},
			expectedStatusCode:   400,
			expectedResponseBody: "{\"message\":\"invalid type list id\"}",
		},
		{
			name:   "Error HGet",
			Id:     44,
			userId: 55,
			prepare: func(f *field, userId, Id int, ReturnHGet_InputHSet string, ReturnGetById todo.TodoList) {
				f.mockBehaviorH.EXPECT().HGet(userId, Id).Return("", errors.New("Error HGet"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error HGet"}`,
		},
		{
			name:                 "OK redis.Nil",
			Id:                   33,
			userId:               55,
			ReturnHGet_InputHSet: "{\"id\":33,\"title\":\"test33\",\"description\":\"by testing 33\"}",
			ReturnGetById: todo.TodoList{
				Id:          33,
				Title:       "test33",
				Description: "by testing 33",
			},
			prepare: func(f *field, userId, Id int, ReturnHGet_InputHSet string, ReturnGetById todo.TodoList) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(userId, Id).Return("", redis.Nil),
					f.mockBehaviorGetById.EXPECT().GetById(userId, Id).Return(ReturnGetById, nil),
					f.mockBehaviorH.EXPECT().HSet(userId, Id, ReturnHGet_InputHSet).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"id\":33,\"title\":\"test33\",\"description\":\"by testing 33\"}",
		},
		{
			name:   "Error GetById",
			Id:     33,
			userId: 55,
			prepare: func(f *field, userId, Id int, ReturnHGet_InputHSet string, ReturnGetById todo.TodoList) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(userId, Id).Return("", redis.Nil),
					f.mockBehaviorGetById.EXPECT().GetById(userId, Id).Return(todo.TodoList{}, errors.New("Error GetById")),
				)
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error GetById"}`,
		},
		{
			name:                 "Error HSet",
			Id:                   33,
			userId:               55,
			ReturnHGet_InputHSet: "{\"id\":33,\"title\":\"test33\",\"description\":\"by testing 33\"}",
			ReturnGetById: todo.TodoList{
				Id:          33,
				Title:       "test33",
				Description: "by testing 33",
			},
			prepare: func(f *field, userId, Id int, ReturnHGet_InputHSet string, ReturnGetById todo.TodoList) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(userId, Id).Return("", redis.Nil),
					f.mockBehaviorGetById.EXPECT().GetById(userId, Id).Return(ReturnGetById, nil),
					f.mockBehaviorH.EXPECT().HSet(userId, Id, ReturnHGet_InputHSet).Return(errors.New("Error HSet")),
				)
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error HSet"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Init Deps
			c := gomock.NewController(t)
			defer c.Finish()

			f := field{
				mockBehaviorH:       mock_service.NewMockTodoListCach(c),
				mockBehaviorGetById: mock_service.NewMockTodoList(c),
			}

			if testCase.prepare != nil {
				testCase.prepare(&f, testCase.userId, testCase.Id, testCase.ReturnHGet_InputHSet, testCase.ReturnGetById)
			}

			services := &service.Service{TodoListCach: f.mockBehaviorH, TodoList: f.mockBehaviorGetById}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			if testCase.CtxNil {
				r.GET("/lists/:id", handler.getListById)
			} else {
				r.GET("/lists/:id", func(c *gin.Context) { c.Set(userCtx, testCase.userId) }, handler.getListById)
			}

			// Test Request
			var req *http.Request
			w := httptest.NewRecorder()
			if testCase.ErrId {
				req = httptest.NewRequest("GET", "/lists/err", nil)
			} else {
				req = httptest.NewRequest("GET", fmt.Sprintf("/lists/%d", testCase.Id), nil)
			}

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_updateList(t *testing.T) {

	type field struct {
		mockBehaviorH          *mock_service.MockTodoListCach
		mockBehaviorUpdateById *mock_service.MockTodoList
	}

	testTable := []struct {
		name                 string
		CtxNil               bool
		ErrId                bool
		Id                   int
		userId               int
		inputBody            string
		inputUpdate          todo.UpdateListInput
		ReturnUpdate         todo.TodoList
		prepare              func(f *field, userId, Id int, inputUpdate todo.UpdateListInput, ReturnUpdate todo.TodoList)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			Id:        4,
			userId:    5,
			inputBody: `{"title":"test4", "description":"by testing4"}`,
			inputUpdate: todo.UpdateListInput{
				Title:       stringPointers("test4"),
				Description: stringPointers("by testing4"),
			},
			ReturnUpdate: todo.TodoList{
				Id:          4,
				Title:       "test4",
				Description: "by testing4",
			},
			prepare: func(f *field, userId, Id int, inputUpdate todo.UpdateListInput, ReturnUpdate todo.TodoList) {
				gomock.InOrder(
					f.mockBehaviorUpdateById.EXPECT().UpdateById(userId, Id, inputUpdate).Return(ReturnUpdate, nil),
					f.mockBehaviorH.EXPECT().Delete(userId).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"id\":4,\"title\":\"test4\",\"description\":\"by testing4\"}",
		},
		{
			name:                 "Error getUserId",
			CtxNil:               true,
			prepare:              func(f *field, userId, Id int, inputUpdate todo.UpdateListInput, ReturnUpdate todo.TodoList) {},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"user id not found\"}",
		},
		{
			name:                 "Error Atoi Id",
			ErrId:                true,
			prepare:              func(f *field, userId, Id int, inputUpdate todo.UpdateListInput, ReturnUpdate todo.TodoList) {},
			expectedStatusCode:   400,
			expectedResponseBody: "{\"message\":\"invalid type list id\"}",
		},
		{
			name:      "Empty Fields",
			Id:        4,
			userId:    5,
			inputBody: `{}`,
			prepare: func(f *field, userId, Id int, inputUpdate todo.UpdateListInput, ReturnUpdate todo.TodoList) {
				f.mockBehaviorUpdateById.EXPECT().UpdateById(userId, Id, inputUpdate).Return(todo.TodoList{}, errors.New("update structure has no values"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"update structure has no values"}`,
		},
		{
			name:      "Error UpdateById",
			Id:        4,
			userId:    5,
			inputBody: `{"title":"test4", "description":"by testing4"}`,
			inputUpdate: todo.UpdateListInput{
				Title:       stringPointers("test4"),
				Description: stringPointers("by testing4"),
			},
			prepare: func(f *field, userId, Id int, inputUpdate todo.UpdateListInput, ReturnUpdate todo.TodoList) {
				f.mockBehaviorUpdateById.EXPECT().UpdateById(userId, Id, inputUpdate).Return(todo.TodoList{}, errors.New("Error UpdateById"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error UpdateById"}`,
		},
		{
			name:      "Error Delete",
			Id:        4,
			userId:    5,
			inputBody: `{"title":"test4", "description":"by testing4"}`,
			inputUpdate: todo.UpdateListInput{
				Title:       stringPointers("test4"),
				Description: stringPointers("by testing4"),
			},
			ReturnUpdate: todo.TodoList{
				Id:          4,
				Title:       "test4",
				Description: "by testing4",
			},
			prepare: func(f *field, userId, Id int, inputUpdate todo.UpdateListInput, ReturnUpdate todo.TodoList) {
				gomock.InOrder(
					f.mockBehaviorUpdateById.EXPECT().UpdateById(userId, Id, inputUpdate).Return(ReturnUpdate, nil),
					f.mockBehaviorH.EXPECT().Delete(userId).Return(errors.New("Error Delete")),
				)
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error Delete"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Init Deps
			c := gomock.NewController(t)
			defer c.Finish()

			f := field{
				mockBehaviorH:          mock_service.NewMockTodoListCach(c),
				mockBehaviorUpdateById: mock_service.NewMockTodoList(c),
			}

			if testCase.prepare != nil {
				testCase.prepare(&f, testCase.userId, testCase.Id, testCase.inputUpdate, testCase.ReturnUpdate)
			}

			services := &service.Service{TodoListCach: f.mockBehaviorH, TodoList: f.mockBehaviorUpdateById}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			if testCase.CtxNil {
				r.PUT("/lists/:id", handler.updateList)
			} else {
				r.PUT("/lists/:id", func(c *gin.Context) { c.Set(userCtx, testCase.userId) }, handler.updateList)
			}

			// Test Request
			var req *http.Request
			w := httptest.NewRecorder()
			if testCase.ErrId {
				req = httptest.NewRequest("PUT", "/lists/err", nil)
			} else {
				req = httptest.NewRequest("PUT", fmt.Sprintf("/lists/%d", testCase.Id), bytes.NewBufferString(testCase.inputBody))
			}

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_deleteList(t *testing.T) {

	type field struct {
		mockBehaviorH          *mock_service.MockTodoListCach
		mockBehaviorDeleteById *mock_service.MockTodoList
	}

	testTable := []struct {
		name                 string
		CtxNil               bool
		ErrId                bool
		Id                   int
		userId               int
		prepare              func(f *field, userId, Id int)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:   "OK",
			Id:     4,
			userId: 5,
			prepare: func(f *field, userId, Id int) {
				gomock.InOrder(
					f.mockBehaviorDeleteById.EXPECT().DeleteById(userId, Id).Return(nil),
					f.mockBehaviorH.EXPECT().Delete(userId).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"Ok\":\"deleted list by id: 4\"}",
		},
		{
			name:                 "Error getUserId",
			CtxNil:               true,
			prepare:              func(f *field, userId, Id int) {},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"user id not found\"}",
		},
		{
			name:                 "Error Atoi Id",
			ErrId:                true,
			prepare:              func(f *field, userId, Id int) {},
			expectedStatusCode:   400,
			expectedResponseBody: "{\"message\":\"invalid type list id\"}",
		},
		{
			name:   "Error DeleteById",
			Id:     4,
			userId: 5,
			prepare: func(f *field, userId, Id int) {
				f.mockBehaviorDeleteById.EXPECT().DeleteById(userId, Id).Return(errors.New("Error DeleteById"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"Error DeleteById\"}",
		},
		{
			name:   "Error Delete",
			Id:     4,
			userId: 5,
			prepare: func(f *field, userId, Id int) {
				gomock.InOrder(
					f.mockBehaviorDeleteById.EXPECT().DeleteById(userId, Id).Return(nil),
					f.mockBehaviorH.EXPECT().Delete(userId).Return(errors.New("Error Delete")),
				)
			},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"Error Delete\"}",
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Init Deps
			c := gomock.NewController(t)
			defer c.Finish()

			f := field{
				mockBehaviorH:          mock_service.NewMockTodoListCach(c),
				mockBehaviorDeleteById: mock_service.NewMockTodoList(c),
			}

			if testCase.prepare != nil {
				testCase.prepare(&f, testCase.userId, testCase.Id)
			}

			services := &service.Service{TodoListCach: f.mockBehaviorH, TodoList: f.mockBehaviorDeleteById}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			if testCase.CtxNil {
				r.DELETE("/lists/:id", handler.deleteList)
			} else {
				r.DELETE("/lists/:id", func(c *gin.Context) { c.Set(userCtx, testCase.userId) }, handler.deleteList)
			}

			// Test Request
			var req *http.Request
			w := httptest.NewRecorder()
			if testCase.ErrId {
				req = httptest.NewRequest("DELETE", "/lists/err", nil)
			} else {
				req = httptest.NewRequest("DELETE", fmt.Sprintf("/lists/%d", testCase.Id), nil)
			}

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func stringPointers(s string) *string {
	return &s
}
