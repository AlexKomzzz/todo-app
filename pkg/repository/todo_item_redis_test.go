package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestTodoItemRedis_HGet(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := &gin.Context{}

	r := NewTodoItemRedis(ctx, db)

	type args struct {
		userId int
		listId int
		itemId int
	}

	type mockBehavior func(args args, want string)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		want         string
		wantErr      bool
		Error        error
	}{
		{
			name: "OK by listId",
			mockBehavior: func(args args, want string) {
				mock.ExpectHGet(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("items:list%d", args.listId)).
					SetVal(want)
			},
			args: args{
				userId: 1,
				listId: 2,
				itemId: -3,
			},
			want: "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\", \"done\":true}, {\"id\":2, \"title\":\"test2 title\", \"description\":\"test2 description\", \"done\":false}`",
		},
		{
			name: "OK by itemId",
			mockBehavior: func(args args, want string) {
				mock.ExpectHGet(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("item:%d", args.itemId)).
					SetVal(want)
			},
			args: args{
				userId: 1,
				listId: -2,
				itemId: 3,
			},
			want: "`{\"id\":3, \"title\":\"test3 title\", \"description\":\"test3 description\", \"done\":true}`",
		},
		{
			name: "invalide input1",
			args: args{
				userId: 1,
				listId: 2,
				itemId: 3,
			},
			want:    "",
			wantErr: true,
			Error:   errors.New("invalide input params func HSet"),
		},
		{
			name: "invalide input2",
			args: args{
				userId: 1,
				listId: -2,
				itemId: -3,
			},
			want:    "",
			wantErr: true,
			Error:   errors.New("invalide input params func HSet"),
		},
		{
			name: "Error HGet by listId",
			mockBehavior: func(args args, want string) {
				mock.ExpectHGet(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("items:list%d", args.listId)).
					SetErr(errors.New("Error HGet by listId"))
			},
			args: args{
				userId: 1,
				listId: 2,
				itemId: -3,
			},
			wantErr: true,
		},
		{
			name: "Error HGet by itemId",
			mockBehavior: func(args args, want string) {
				mock.ExpectHGet(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("item:%d", args.itemId)).
					SetErr(errors.New("Error HGet by itemId"))
			},
			args: args{
				userId: 1,
				listId: -2,
				itemId: 3,
			},
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {

			if testCase.args.itemId <= 0 && testCase.args.listId > 0 || testCase.args.itemId > 0 && testCase.args.listId <= 0 {
				testCase.mockBehavior(testCase.args, testCase.want)

				value, err := r.HGet(testCase.args.userId, testCase.args.listId, testCase.args.itemId)
				if testCase.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, testCase.want, value)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			} else {
				value, err := r.HGet(testCase.args.userId, testCase.args.listId, testCase.args.itemId)
				assert.Equal(t, testCase.Error, err)
				assert.Equal(t, testCase.want, value)
			}
		})
	}
}

func TestTodoItemRedis_HSet(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := &gin.Context{}

	r := NewTodoItemRedis(ctx, db)

	type args struct {
		userId int
		listId int
		itemId int
	}

	type mockBehavior func(args args, data string)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		data         string
		wantErr      bool
		Error        error
	}{
		{
			name: "OK by listId",
			mockBehavior: func(args args, data string) {
				mock.ExpectHSetNX(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("items:list%d", args.listId), data).SetVal(true)
				mock.ExpectExpire(fmt.Sprintf("user:%d", args.userId), duration).SetVal(true)
			},
			args: args{
				userId: 1,
				listId: 2,
				itemId: -3,
			},
			data: "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\", \"done\":true}, {\"id\":2, \"title\":\"test2 title\", \"description\":\"test2 description\", \"done\":false}`",
		},
		{
			name: "OK by itemId",
			mockBehavior: func(args args, data string) {
				mock.ExpectHSetNX(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("item:%d", args.itemId), data).SetVal(true)
				mock.ExpectExpire(fmt.Sprintf("user:%d", args.userId), duration).SetVal(true)
			},
			args: args{
				userId: 1,
				listId: -2,
				itemId: 3,
			},
			data: "`{\"id\":3, \"title\":\"test3 title\", \"description\":\"test3 description\", \"done\":true}`",
		},
		{
			name: "Error HSet by listId",
			mockBehavior: func(args args, data string) {
				mock.ExpectHSetNX(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("items:list%d", args.listId), data).SetErr(errors.New("Error HSet by listId"))
			},
			args: args{
				userId: 1,
				listId: 2,
				itemId: -3,
			},
			data:    "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\", \"done\":true}, {\"id\":2, \"title\":\"test2 title\", \"description\":\"test2 description\", \"done\":false}`",
			wantErr: true,
		},
		{
			name: "Error HSet by itemId",
			mockBehavior: func(args args, data string) {
				mock.ExpectHSetNX(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("item:%d", args.itemId), data).SetErr(errors.New("Error HSet by itemId"))
			},
			args: args{
				userId: 1,
				listId: -2,
				itemId: 3,
			},
			data:    "`{\"id\":3, \"title\":\"test3 title\", \"description\":\"test3 description\", \"done\":true}`",
			wantErr: true,
		},
		{
			name: "Error Expire by listId",
			mockBehavior: func(args args, data string) {
				mock.ExpectHSetNX(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("items:list%d", args.listId), data).SetVal(true)
				mock.ExpectExpire(fmt.Sprintf("user:%d", args.userId), duration).SetErr(errors.New("Error Expire by listId"))
			},
			args: args{
				userId: 1,
				listId: 2,
				itemId: -3,
			},
			data:    "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\", \"done\":true}, {\"id\":2, \"title\":\"test2 title\", \"description\":\"test2 description\", \"done\":false}`",
			wantErr: true,
		},
		{
			name: "Error Expire by itemId",
			mockBehavior: func(args args, data string) {
				mock.ExpectHSetNX(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("item:%d", args.itemId), data).SetVal(true)
				mock.ExpectExpire(fmt.Sprintf("user:%d", args.userId), duration).SetErr(errors.New("Error Expire by itemId"))
			},
			args: args{
				userId: 1,
				listId: -2,
				itemId: 3,
			},
			data:    "`{\"id\":3, \"title\":\"test3 title\", \"description\":\"test3 description\", \"done\":true}`",
			wantErr: true,
		},
		{
			name: "invalide input1",
			args: args{
				userId: 1,
				listId: 2,
				itemId: 3,
			},
			data:    "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\", \"done\":true}, {\"id\":2, \"title\":\"test2 title\", \"description\":\"test2 description\", \"done\":false}`",
			wantErr: true,
			Error:   errors.New("invalide input params func HSet"),
		},
		{
			name: "invalide input2",
			args: args{
				userId: 1,
				listId: -2,
				itemId: -3,
			},
			data:    "`{\"id\":1, \"title\":\"test title\", \"description\":\"test description\", \"done\":true}, {\"id\":2, \"title\":\"test2 title\", \"description\":\"test2 description\", \"done\":false}`",
			wantErr: true,
			Error:   errors.New("invalide input params func HSet"),
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {

			if testCase.args.itemId <= 0 && testCase.args.listId > 0 || testCase.args.itemId > 0 && testCase.args.listId <= 0 {
				testCase.mockBehavior(testCase.args, testCase.data)

				err := r.HSet(testCase.args.userId, testCase.args.listId, testCase.args.itemId, testCase.data)
				if testCase.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			} else {
				err := r.HSet(testCase.args.userId, testCase.args.listId, testCase.args.itemId, testCase.data)
				assert.Equal(t, testCase.Error, err)
			}
		})
	}
}

func TestTodoItemRedis_HDelete(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := &gin.Context{}

	r := NewTodoItemRedis(ctx, db)

	type args struct {
		userId int
		listId int
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		wantErr      bool
	}{
		{
			name: "OK",
			mockBehavior: func(args args) {
				mock.ExpectHDel(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("items:list%d", args.listId)).SetVal(1)
			},
			args: args{
				userId: 55,
				listId: 44,
			},
		},
		{
			name: "Error",
			mockBehavior: func(args args) {
				mock.ExpectHDel(fmt.Sprintf("user:%d", args.userId), fmt.Sprintf("items:list%d", args.listId)).SetErr(errors.New("Error"))
			},
			args: args{
				userId: 55,
				listId: 44,
			},
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			err := r.HDelete(testCase.args.userId, testCase.args.listId)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTodoItemRedis_Delete(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := &gin.Context{}

	r := NewTodoItemRedis(ctx, db)

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
