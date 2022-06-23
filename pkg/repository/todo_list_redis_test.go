package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestTodoListRedis_HGet(t *testing.T) {
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
			want: "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\"}`",
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

func TestTodoListRedis_HSet(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := &gin.Context{}

	r := NewTodoListRedis(ctx, db)

	type args struct {
		userId int
		listId int
	}

	type mockBehavior func(args args, data string)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		data         string
		wantErr      bool
	}{
		{
			name: "OK listId",
			mockBehavior: func(args args, data string) {
				mock.ExpectHSetNX(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("list:%d", args.listId), data).SetVal(true)
				mock.ExpectExpire(fmt.Sprintf("user:%d", args.userId), duration).SetVal(true)
			},
			args: args{
				userId: 1,
				listId: 2,
			},
			data: "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\", \"done\":true}`",
		},
		{
			name: "OK NO listId",
			mockBehavior: func(args args, data string) {
				mock.ExpectHSetNX(fmt.Sprintf("user:%d", args.userId), "lists", data).SetVal(true)
				mock.ExpectExpire(fmt.Sprintf("user:%d", args.userId), duration).SetVal(true)
			},
			args: args{
				userId: 1,
				listId: -9,
			},
			data: "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\", \"done\":true}`",
		},
		{
			name: "Error HSet",
			mockBehavior: func(args args, data string) {
				mock.ExpectHSetNX(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("list:%d", args.listId), data).SetErr(errors.New("Error"))
			},
			args: args{
				userId: 1,
				listId: 2,
			},
			data:    "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\", \"done\":true}`",
			wantErr: true,
		},
		{
			name: "Error Expire",
			mockBehavior: func(args args, data string) {
				mock.ExpectHSetNX(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("list:%d", args.listId), data).SetVal(true)
				mock.ExpectExpire(fmt.Sprintf("user:%d", args.userId), duration).SetErr(errors.New("Error"))
			},
			args: args{
				userId: 1,
				listId: 2,
			},
			data:    "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\", \"done\":true}`",
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.data)

			err := r.HSet(testCase.args.userId, testCase.args.listId, testCase.data)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTodoListRedis_HDelete(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := &gin.Context{}

	r := NewTodoListRedis(ctx, db)

	type mockBehavior func(userId int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		userId       int
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(userId int) {
				mock.ExpectHDel(fmt.Sprintf("user:%d", userId), "lists").SetVal(1)
			},
			userId: 55,
		},
		{
			name: "Error",
			mockBehavior: func(userId int) {
				mock.ExpectHDel(fmt.Sprintf("user:%d", userId), "lists").SetErr(errors.New("Error"))
			},
			userId:  55,
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.userId)

			err := r.HDelete(testCase.userId)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTodoListRedis_Delete(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := &gin.Context{}

	r := NewTodoListRedis(ctx, db)

	type mockBehavior func(userId int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		userId       int
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(userId int) {
				mock.ExpectDel(fmt.Sprintf("user:%d", userId)).SetVal(1)
			},
			userId: 66,
		},
		{
			name: "Error",
			mockBehavior: func(userId int) {
				mock.ExpectDel(fmt.Sprintf("user:%d", userId)).SetErr(errors.New("Error"))
			},
			userId:  66,
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.userId)

			err := r.Delete(testCase.userId)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
