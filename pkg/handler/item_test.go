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
	"github.com/magiconair/properties/assert"
)

func TestHandler_createItem(t *testing.T) {

	type field struct {
		mockBehaviorCreate *mock_service.MockTodoItem
		mockBehaviorHDel   *mock_service.MockTodoItemCach
	}

	type args struct {
		userId    int
		listId    int
		inputItem todo.TodoItem
	}

	testTable := []struct {
		name                 string
		CtxNil               bool
		ErrId                bool
		args                 args
		Id                   int
		inputBody            string
		prepare              func(f *field, args args, Id int)
		expectedStatusCode   int // статус код ответа
		expectedResponseBody string
	}{
		{
			name: "OK",
			args: args{
				userId: 2,
				listId: 3,
				inputItem: todo.TodoItem{
					Title:       "test",
					Description: "by testing",
					Done:        true,
				},
			},
			Id:        5,
			inputBody: `{"title":"test","description":"by testing","done":true}`,
			prepare: func(f *field, args args, Id int) {
				gomock.InOrder(
					f.mockBehaviorCreate.EXPECT().Create(args.userId, args.listId, args.inputItem).Return(Id, nil),
					f.mockBehaviorHDel.EXPECT().HDelete(args.userId, args.listId).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":5}`,
		},
		{
			name:                 "Error getUserId",
			CtxNil:               true,
			prepare:              func(f *field, args args, Id int) {},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"user id not found\"}",
		},
		{
			name:                 "Error Atoi Id",
			ErrId:                true,
			prepare:              func(f *field, args args, Id int) {},
			expectedStatusCode:   400,
			expectedResponseBody: "{\"message\":\"invalid list id param\"}",
		},
		{
			name: "OK Body only Title",
			args: args{
				userId: 2,
				listId: 3,
				inputItem: todo.TodoItem{
					Title: "test",
				},
			},
			Id:        88,
			inputBody: `{"title":"test"}`,
			prepare: func(f *field, args args, Id int) {
				gomock.InOrder(
					f.mockBehaviorCreate.EXPECT().Create(args.userId, args.listId, args.inputItem).Return(Id, nil),
					f.mockBehaviorHDel.EXPECT().HDelete(args.userId, args.listId).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":88}`,
		},
		{
			name: "Empty Fields",
			args: args{
				userId: 2,
				listId: 3,
			},
			inputBody:            `{}`,
			prepare:              func(f *field, args args, Id int) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Key: 'TodoItem.Title' Error:Field validation for 'Title' failed on the 'required' tag"}`,
		},
		{
			name: "Empty Field Title",
			args: args{
				userId: 2,
				listId: 3,
			},
			inputBody:            `{"description":"by testing","done":true}`,
			prepare:              func(f *field, args args, Id int) {},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"Key: 'TodoItem.Title' Error:Field validation for 'Title' failed on the 'required' tag"}`,
		},
		{
			name: "Error Create",
			args: args{
				userId: 2,
				listId: 3,
				inputItem: todo.TodoItem{
					Title:       "test",
					Description: "by testing",
					Done:        true,
				},
			},
			inputBody: `{"title":"test","description":"by testing","done":true}`,
			prepare: func(f *field, args args, Id int) {
				f.mockBehaviorCreate.EXPECT().Create(args.userId, args.listId, args.inputItem).Return(0, errors.New("Error Create"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error Create"}`,
		},
		{
			name: "Error HDel",
			args: args{
				userId: 2,
				listId: 3,
				inputItem: todo.TodoItem{
					Title:       "test",
					Description: "by testing",
					Done:        true,
				},
			},
			Id:        5,
			inputBody: `{"title":"test","description":"by testing","done":true}`,
			prepare: func(f *field, args args, Id int) {
				gomock.InOrder(
					f.mockBehaviorCreate.EXPECT().Create(args.userId, args.listId, args.inputItem).Return(Id, nil),
					f.mockBehaviorHDel.EXPECT().HDelete(args.userId, args.listId).Return(errors.New("Error HDel")),
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
				mockBehaviorCreate: mock_service.NewMockTodoItem(c),
				mockBehaviorHDel:   mock_service.NewMockTodoItemCach(c),
			}

			if testCase.prepare != nil {
				testCase.prepare(&f, testCase.args, testCase.Id)
			}

			services := &service.Service{TodoItem: f.mockBehaviorCreate, TodoItemCach: f.mockBehaviorHDel}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			if testCase.CtxNil {
				r.POST("/lists/:id/items", handler.createItem)
			} else {
				r.POST("/lists/:id/items", func(c *gin.Context) { c.Set(userCtx, testCase.args.userId) }, handler.createItem)
			}

			// Test Request
			var req *http.Request
			w := httptest.NewRecorder()
			if testCase.ErrId {
				req = httptest.NewRequest("POST", "/lists/err/items", nil)
			} else {
				req = httptest.NewRequest("POST", fmt.Sprintf("/lists/%d/items", testCase.args.listId), bytes.NewBufferString(testCase.inputBody))
			}

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_getAllItems(t *testing.T) {

	type field struct {
		mockBehaviorGetAll *mock_service.MockTodoItem
		mockBehaviorH      *mock_service.MockTodoItemCach
	}

	type args struct {
		userId               int
		listId               int
		ReturnHGet_InputHSet string
		ReturnGetAll         []todo.TodoItem
	}

	testTable := []struct {
		name                 string
		CtxNil               bool
		ErrId                bool
		args                 args
		prepare              func(f *field, args args)
		expectedStatusCode   int // статус код ответа
		expectedResponseBody string
	}{
		{
			name: "OK HGet",
			args: args{
				userId:               55,
				listId:               44,
				ReturnHGet_InputHSet: "[{\"id\":1,\"title\":\"test1\",\"description\":\"by testing 1\",\"done\":true},{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\",\"done\":false}]",
			},
			prepare: func(f *field, args args) {
				f.mockBehaviorH.EXPECT().HGet(args.userId, args.listId, -1).Return(args.ReturnHGet_InputHSet, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "[{\"id\":1,\"title\":\"test1\",\"description\":\"by testing 1\",\"done\":true},{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\",\"done\":false}]",
		},
		{
			name:                 "Error getUserId",
			CtxNil:               true,
			prepare:              func(f *field, args args) {},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"user id not found\"}",
		},
		{
			name:                 "Error Atoi Id",
			ErrId:                true,
			prepare:              func(f *field, args args) {},
			expectedStatusCode:   400,
			expectedResponseBody: "{\"message\":\"invalid list id param\"}",
		},
		{
			name: "Error HGet",
			args: args{
				userId: 55,
				listId: 44,
			},
			prepare: func(f *field, args args) {
				f.mockBehaviorH.EXPECT().HGet(args.userId, args.listId, -1).Return("", errors.New("Error HGet"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error HGet"}`,
		},
		{
			name: "OK redis.Nil",
			args: args{
				userId:               55,
				listId:               44,
				ReturnHGet_InputHSet: "[{\"id\":1,\"title\":\"test1\",\"description\":\"by testing 1\",\"done\":true},{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\",\"done\":false}]",
				ReturnGetAll: []todo.TodoItem{
					{
						Id:          1,
						Title:       "test1",
						Description: "by testing 1",
						Done:        true,
					},
					{
						Id:          4,
						Title:       "test4",
						Description: "by testing 4",
						Done:        false,
					},
				},
			},
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(args.userId, args.listId, -1).Return("", redis.Nil),
					f.mockBehaviorGetAll.EXPECT().GetAll(args.userId, args.listId).Return(args.ReturnGetAll, nil),
					f.mockBehaviorH.EXPECT().HSet(args.userId, args.listId, -1, args.ReturnHGet_InputHSet).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "[{\"id\":1,\"title\":\"test1\",\"description\":\"by testing 1\",\"done\":true},{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\",\"done\":false}]",
		},
		{
			name: "Error GetAll",
			args: args{
				userId: 55,
				listId: 44,
			},
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(args.userId, args.listId, -1).Return("", redis.Nil),
					f.mockBehaviorGetAll.EXPECT().GetAll(args.userId, args.listId).Return([]todo.TodoItem{}, errors.New("Error GetAll")),
				)
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error GetAll"}`,
		},
		{
			name: "Error HSet",
			args: args{
				userId:               55,
				listId:               44,
				ReturnHGet_InputHSet: "[{\"id\":1,\"title\":\"test1\",\"description\":\"by testing 1\",\"done\":true},{\"id\":4,\"title\":\"test4\",\"description\":\"by testing 4\",\"done\":false}]",
				ReturnGetAll: []todo.TodoItem{
					{
						Id:          1,
						Title:       "test1",
						Description: "by testing 1",
						Done:        true,
					},
					{
						Id:          4,
						Title:       "test4",
						Description: "by testing 4",
						Done:        false,
					},
				},
			},
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(args.userId, args.listId, -1).Return("", redis.Nil),
					f.mockBehaviorGetAll.EXPECT().GetAll(args.userId, args.listId).Return(args.ReturnGetAll, nil),
					f.mockBehaviorH.EXPECT().HSet(args.userId, args.listId, -1, args.ReturnHGet_InputHSet).Return(errors.New("Error HSet")),
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
				mockBehaviorGetAll: mock_service.NewMockTodoItem(c),
				mockBehaviorH:      mock_service.NewMockTodoItemCach(c),
			}

			if testCase.prepare != nil {
				testCase.prepare(&f, testCase.args)
			}

			services := &service.Service{TodoItem: f.mockBehaviorGetAll, TodoItemCach: f.mockBehaviorH}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			if testCase.CtxNil {
				r.GET("/lists/:id/items", handler.getAllItems)
			} else {
				r.GET("/lists/:id/items", func(c *gin.Context) { c.Set(userCtx, testCase.args.userId) }, handler.getAllItems)
			}

			// Test Request
			var req *http.Request
			w := httptest.NewRecorder()
			if testCase.ErrId {
				req = httptest.NewRequest("GET", "/lists/err/items", nil)
			} else {
				req = httptest.NewRequest("GET", fmt.Sprintf("/lists/%d/items", testCase.args.listId), nil)
			}

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_getItemById(t *testing.T) {

	type field struct {
		mockBehaviorGetById *mock_service.MockTodoItem
		mockBehaviorH       *mock_service.MockTodoItemCach
	}

	type args struct {
		userId               int
		itemId               int
		ReturnHGet_InputHSet string
		ReturnGetById        todo.TodoItem
	}

	testTable := []struct {
		name                 string
		CtxNil               bool
		ErrId                bool
		args                 args
		prepare              func(f *field, args args)
		expectedStatusCode   int // статус код ответа
		expectedResponseBody string
	}{
		{
			name: "OK HGet",
			args: args{
				userId:               55,
				itemId:               44,
				ReturnHGet_InputHSet: "{\"id\":44,\"title\":\"test44\",\"description\":\"by testing 44\",\"done\":true}",
			},
			prepare: func(f *field, args args) {
				f.mockBehaviorH.EXPECT().HGet(args.userId, -1, args.itemId).Return(args.ReturnHGet_InputHSet, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"id\":44,\"title\":\"test44\",\"description\":\"by testing 44\",\"done\":true}",
		},
		{
			name:                 "Error getUserId",
			CtxNil:               true,
			prepare:              func(f *field, args args) {},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"user id not found\"}",
		},
		{
			name:                 "Error Atoi Id",
			ErrId:                true,
			prepare:              func(f *field, args args) {},
			expectedStatusCode:   400,
			expectedResponseBody: "{\"message\":\"invalid list id param\"}",
		},
		{
			name: "Error HGet",
			args: args{
				userId: 55,
				itemId: 44,
			},
			prepare: func(f *field, args args) {
				f.mockBehaviorH.EXPECT().HGet(args.userId, -1, args.itemId).Return("", errors.New("Error HGet"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error HGet"}`,
		},
		{
			name: "OK redis.Nil",
			args: args{
				userId: 55,
				itemId: 44,
				ReturnGetById: todo.TodoItem{
					Id:          44,
					Title:       "test44",
					Description: "by testing 44",
					Done:        true,
				},
				ReturnHGet_InputHSet: "{\"id\":44,\"title\":\"test44\",\"description\":\"by testing 44\",\"done\":true}",
			},
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(args.userId, -1, args.itemId).Return("", redis.Nil),
					f.mockBehaviorGetById.EXPECT().GetById(args.userId, args.itemId).Return(args.ReturnGetById, nil),
					f.mockBehaviorH.EXPECT().HSet(args.userId, -1, args.itemId, args.ReturnHGet_InputHSet).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"id\":44,\"title\":\"test44\",\"description\":\"by testing 44\",\"done\":true}",
		},
		{
			name: "Error GetById",
			args: args{
				userId: 55,
				itemId: 44,
			},
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(args.userId, -1, args.itemId).Return("", redis.Nil),
					f.mockBehaviorGetById.EXPECT().GetById(args.userId, args.itemId).Return(args.ReturnGetById, errors.New("Error GetById")),
				)
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error GetById"}`,
		},
		{
			name: "Error HSet",
			args: args{
				userId: 55,
				itemId: 44,
				ReturnGetById: todo.TodoItem{
					Id:          44,
					Title:       "test44",
					Description: "by testing 44",
					Done:        true,
				},
				ReturnHGet_InputHSet: "{\"id\":44,\"title\":\"test44\",\"description\":\"by testing 44\",\"done\":true}",
			},
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorH.EXPECT().HGet(args.userId, -1, args.itemId).Return("", redis.Nil),
					f.mockBehaviorGetById.EXPECT().GetById(args.userId, args.itemId).Return(args.ReturnGetById, nil),
					f.mockBehaviorH.EXPECT().HSet(args.userId, -1, args.itemId, args.ReturnHGet_InputHSet).Return(errors.New("Error HSet")),
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
				mockBehaviorGetById: mock_service.NewMockTodoItem(c),
				mockBehaviorH:       mock_service.NewMockTodoItemCach(c),
			}

			if testCase.prepare != nil {
				testCase.prepare(&f, testCase.args)
			}

			services := &service.Service{TodoItem: f.mockBehaviorGetById, TodoItemCach: f.mockBehaviorH}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			if testCase.CtxNil {
				r.GET("/items/:id", handler.getItemById)
			} else {
				r.GET("/items/:id", func(c *gin.Context) { c.Set(userCtx, testCase.args.userId) }, handler.getItemById)
			}

			// Test Request
			var req *http.Request
			w := httptest.NewRecorder()
			if testCase.ErrId {
				req = httptest.NewRequest("GET", "/items/err", nil)
			} else {
				req = httptest.NewRequest("GET", fmt.Sprintf("/items/%d", testCase.args.itemId), nil)
			}

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_updateItem(t *testing.T) {

	type field struct {
		mockBehaviorUpdate *mock_service.MockTodoItem
		mockBehaviorH      *mock_service.MockTodoItemCach
	}

	type args struct {
		userId    int
		itemId    int
		inputItem todo.UpdateItemInput
	}

	testTable := []struct {
		name                 string
		CtxNil               bool
		ErrId                bool
		inputBody            string
		args                 args
		prepare              func(f *field, args args)
		expectedStatusCode   int // статус код ответа
		expectedResponseBody string
	}{
		{
			name:                 "Error getUserId",
			CtxNil:               true,
			prepare:              func(f *field, args args) {},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"user id not found\"}",
		},
		{
			name:                 "Error Atoi Id",
			ErrId:                true,
			prepare:              func(f *field, args args) {},
			expectedStatusCode:   400,
			expectedResponseBody: "{\"message\":\"invalid id param\"}",
		},
		{
			name: "OK",
			args: args{
				userId: 55,
				itemId: 44,
				inputItem: todo.UpdateItemInput{
					Title:       stringPointers("test44"),
					Description: stringPointers("by testing 44"),
					Done:        boolPointers(true),
				},
			},
			inputBody: `{"title":"test44","description":"by testing 44","done":true}`,
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorUpdate.EXPECT().Update(args.userId, args.itemId, args.inputItem).Return(nil),
					f.mockBehaviorH.EXPECT().Delete(args.userId).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"status\":\"ok\"}",
		},
		{
			name: "OK Empty Title",
			args: args{
				userId: 55,
				itemId: 44,
				inputItem: todo.UpdateItemInput{
					Description: stringPointers("by testing 44"),
					Done:        boolPointers(true),
				},
			},
			inputBody: `{"description":"by testing 44","done":true}`,
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorUpdate.EXPECT().Update(args.userId, args.itemId, args.inputItem).Return(nil),
					f.mockBehaviorH.EXPECT().Delete(args.userId).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"status\":\"ok\"}",
		},
		{
			name: "OK Empty Title, Description",
			args: args{
				userId: 55,
				itemId: 44,
				inputItem: todo.UpdateItemInput{
					Done: boolPointers(true),
				},
			},
			inputBody: `{"done":true}`,
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorUpdate.EXPECT().Update(args.userId, args.itemId, args.inputItem).Return(nil),
					f.mockBehaviorH.EXPECT().Delete(args.userId).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"status\":\"ok\"}",
		},
		{
			name: "Empty Fields",
			args: args{
				userId: 55,
				itemId: 44,
			},
			inputBody: `{}`,
			prepare: func(f *field, args args) {
				f.mockBehaviorUpdate.EXPECT().Update(args.userId, args.itemId, args.inputItem).Return(errors.New("update structure has no values"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"update structure has no values"}`,
		},
		{
			name: "Error Update",
			args: args{
				userId: 55,
				itemId: 44,
				inputItem: todo.UpdateItemInput{
					Title:       stringPointers("test44"),
					Description: stringPointers("by testing 44"),
					Done:        boolPointers(true),
				},
			},
			inputBody: `{"title":"test44","description":"by testing 44","done":true}`,
			prepare: func(f *field, args args) {
				f.mockBehaviorUpdate.EXPECT().Update(args.userId, args.itemId, args.inputItem).Return(errors.New("Error Update"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"Error Update"}`,
		},
		{
			name: "Error Delete",
			args: args{
				userId: 55,
				itemId: 44,
				inputItem: todo.UpdateItemInput{
					Title:       stringPointers("test44"),
					Description: stringPointers("by testing 44"),
					Done:        boolPointers(true),
				},
			},
			inputBody: `{"title":"test44","description":"by testing 44","done":true}`,
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorUpdate.EXPECT().Update(args.userId, args.itemId, args.inputItem).Return(nil),
					f.mockBehaviorH.EXPECT().Delete(args.userId).Return(errors.New("Error Delete")),
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
				mockBehaviorUpdate: mock_service.NewMockTodoItem(c),
				mockBehaviorH:      mock_service.NewMockTodoItemCach(c),
			}

			if testCase.prepare != nil {
				testCase.prepare(&f, testCase.args)
			}

			services := &service.Service{TodoItem: f.mockBehaviorUpdate, TodoItemCach: f.mockBehaviorH}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			if testCase.CtxNil {
				r.PUT("/items/:id", handler.updateItem)
			} else {
				r.PUT("/items/:id", func(c *gin.Context) { c.Set(userCtx, testCase.args.userId) }, handler.updateItem)
			}

			// Test Request
			var req *http.Request
			w := httptest.NewRecorder()
			if testCase.ErrId {
				req = httptest.NewRequest("PUT", "/items/err", nil)
			} else {
				req = httptest.NewRequest("PUT", fmt.Sprintf("/items/%d", testCase.args.itemId), bytes.NewBufferString(testCase.inputBody))
			}

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_deleteItem(t *testing.T) {

	type field struct {
		mockBehaviorDelete *mock_service.MockTodoItem
		mockBehaviorH      *mock_service.MockTodoItemCach
	}

	type args struct {
		userId int
		itemId int
	}

	testTable := []struct {
		name                 string
		CtxNil               bool
		ErrId                bool
		args                 args
		prepare              func(f *field, args args)
		expectedStatusCode   int // статус код ответа
		expectedResponseBody string
	}{
		{
			name:                 "Error getUserId",
			CtxNil:               true,
			prepare:              func(f *field, args args) {},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"user id not found\"}",
		},
		{
			name:                 "Error Atoi Id",
			ErrId:                true,
			prepare:              func(f *field, args args) {},
			expectedStatusCode:   400,
			expectedResponseBody: "{\"message\":\"invalid id param\"}",
		},
		{
			name: "OK",
			args: args{
				userId: 55,
				itemId: 44,
			},
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorDelete.EXPECT().Delete(args.userId, args.itemId).Return(nil),
					f.mockBehaviorH.EXPECT().Delete(args.userId).Return(nil),
				)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "{\"status\":\"ok\"}",
		},
		{
			name: "Error Item Delete",
			args: args{
				userId: 55,
				itemId: 44,
			},
			prepare: func(f *field, args args) {
				f.mockBehaviorDelete.EXPECT().Delete(args.userId, args.itemId).Return(errors.New("Error Item Delete"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"Error Item Delete\"}",
		},
		{
			name: "Error Cach Delete",
			args: args{
				userId: 55,
				itemId: 44,
			},
			prepare: func(f *field, args args) {
				gomock.InOrder(
					f.mockBehaviorDelete.EXPECT().Delete(args.userId, args.itemId).Return(nil),
					f.mockBehaviorH.EXPECT().Delete(args.userId).Return(errors.New("Error Cach Delete")),
				)
			},
			expectedStatusCode:   500,
			expectedResponseBody: "{\"message\":\"Error Cach Delete\"}",
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Init Deps
			c := gomock.NewController(t)
			defer c.Finish()

			f := field{
				mockBehaviorDelete: mock_service.NewMockTodoItem(c),
				mockBehaviorH:      mock_service.NewMockTodoItemCach(c),
			}

			if testCase.prepare != nil {
				testCase.prepare(&f, testCase.args)
			}

			services := &service.Service{TodoItem: f.mockBehaviorDelete, TodoItemCach: f.mockBehaviorH}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			if testCase.CtxNil {
				r.DELETE("/items/:id", handler.deleteItem)
			} else {
				r.DELETE("/items/:id", func(c *gin.Context) { c.Set(userCtx, testCase.args.userId) }, handler.deleteItem)
			}

			// Test Request
			var req *http.Request
			w := httptest.NewRecorder()
			if testCase.ErrId {
				req = httptest.NewRequest("DELETE", "/items/err", nil)
			} else {
				req = httptest.NewRequest("DELETE", fmt.Sprintf("/items/%d", testCase.args.itemId), nil)
			}

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedResponseBody, w.Body.String())
		})
	}
}

func boolPointers(b bool) *bool {
	return &b
}
