package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestTodoListRedis(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := &gin.Context{}

	r := NewTodoListRedis(ctx, db)

	type args struct {
		userId int
		listId int
	}

	type mockBehavior func(args args, want string)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		want         string
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(args args, want string) {
				mock.ExpectHGet(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("list:%d", args.listId)).
					SetVal(want)
			},
			args: args{
				userId: 1,
				listId: 2,
			},
			want: "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\", \"done\":true}`",
		},
		{
			name: "Not needed listId",
			mockBehavior: func(args args, want string) {
				mock.ExpectHGet(fmt.Sprintf("user:%d", args.userId), "lists").
					SetVal(want)
			},
			args: args{
				userId: 1,
				listId: -1,
			},
			want: "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\"}`",
		},
		{
			name: "Error HGet",
			mockBehavior: func(args args, want string) {
				mock.ExpectHGet(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("list:%d", args.listId)).
					SetErr(errors.New("Error HGet"))
			},
			args: args{
				userId: 1,
				listId: 2,
			},
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.want)

			value, err := r.HGet(testCase.args.userId, testCase.args.listId)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, value)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
