package handler

import (
	"bytes"
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

	var getContext = func(userId interface{}) *gin.Context { // функция передачи userId в контекст, для использования в getUserId
		ctx := &gin.Context{}
		ctx.Set(userCtx, userId)
		return ctx
	}

	type mockBehaviorCreate func(s *mock_service.MockTodoList, userId int, list todo.TodoList)

	type mockBehaviorHDel func(s *mock_service.MockTodoListCach, userId int)

	testTable := []struct {
		name                 string
		ctx                  *gin.Context
		shouidFail           bool
		userId               int
		Id                   int
		inputBody            string
		inputList            todo.TodoList
		mockBehaviorCreate   mockBehaviorCreate
		mockBehaviorHDel     mockBehaviorHDel
		expectedStatusCode   int // статус код ответа
		expectedResponseBody string
	}{
		{
			name:      "OK",
			ctx:       getContext(666),
			userId:    666,
			inputBody: `{"title":"test Bind", "description":"by testing BindJSON"}`,
			inputList: todo.TodoList{
				Title:       "test Create",
				Description: "by testing func Create",
			},
			mockBehaviorCreate: func(s *mock_service.MockTodoList, userId int, list todo.TodoList) {
				s.EXPECT().Create(userId, list).Return(1, nil)
			},
			mockBehaviorHDel: func(s *mock_service.MockTodoListCach, userId int) {
				s.EXPECT().HDelete(userId).Return(nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":1}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			// Init Deps
			c := gomock.NewController(t)
			defer c.Finish()

			userId, err := getUserId(testCase.ctx)

			errYes := false

			if err != nil {
				errYes = true
			}

			assert.NoError(t, err)

			todolist := mock_service.NewMockTodoList(c)
			testCase.mockBehaviorCreate(todolist, testCase.userId, testCase.inputList)

			todolistcach := mock_service.NewMockTodoListCach(c)
			testCase.mockBehaviorHDel(todolistcach, testCase.userId)

			services := &service.Service{TodoList: todolist, TodoListCach: todolistcach}
			handler := NewHandler(services)

			// Test Server
			r := gin.New()
			r.POST("/lists", handler.createList)

			// Test Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/lists", bytes.NewBufferString(testCase.inputBody))

			// Perform Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, "", w.Body.String())
			assert.Equal(t, testCase.userId, userId)
			assert.Equal(t, testCase.shouidFail, errYes)

		})
	}
}
