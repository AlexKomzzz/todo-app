package repository

import (
	"errors"
	"testing"
	"todo-app"

	"github.com/stretchr/testify/assert"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
)

func TestTodoListPostgres_Create(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := NewTodoListPostgres(db)

	type args struct {
		userId int
		list   todo.TodoList
	}

	type mockBehavior func(args args, id int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		id           int
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userId: 1,
				list: todo.TodoList{
					Title:       "test title",
					Description: "test description",
				},
			},
			id: 99,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin() // Откроем транзакцию

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("INSERT INTO todo_lists").
					WithArgs(args.list.Title, args.list.Description).WillReturnRows(rows)

				mock.ExpectExec("INSERT INTO user_lists").WithArgs(args.userId, id).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
		},
		{
			name: "Empty Field Title",
			args: args{
				userId: 1,
				list: todo.TodoList{
					Title:       "",
					Description: "test description",
				},
			},
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin() // Откроем транзакцию

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id).RowError(1, errors.New("some error"))
				mock.ExpectQuery("INSERT INTO todo_lists").
					WithArgs(args.list.Title, args.list.Description).WillReturnRows(rows)

				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Empty Field Description",
			args: args{
				userId: 1,
				list: todo.TodoList{
					Title:       "test title",
					Description: "",
				},
			},
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id).RowError(1, errors.New("some error"))
				mock.ExpectQuery("INSERT INTO todo_lists").
					WithArgs(args.list.Title, args.list.Description).WillReturnRows(rows)

				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "2nd Insert Error",
			args: args{
				userId: 1,
				list: todo.TodoList{
					Title:       "test title",
					Description: "test description",
				},
			},
			id: 99,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("INSERT INTO todo_lists").
					WithArgs(args.list.Title, args.list.Description).WillReturnRows(rows)

				mock.ExpectExec("INSERT INTO user_lists").WithArgs(args.userId, id).
					WillReturnError(errors.New("some error"))

				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Error Begin",
			args: args{
				userId: 1,
				list: todo.TodoList{
					Title:       "test title",
					Description: "test description",
				},
			},
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin().WillReturnError(errors.New("Error Begin"))
			},
			wantErr: true,
		},
		{
			name: "Error Commit",
			args: args{
				userId: 1,
				list: todo.TodoList{
					Title:       "test title",
					Description: "test description",
				},
			},
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("INSERT INTO todo_lists").
					WithArgs(args.list.Title, args.list.Description).WillReturnRows(rows)

				mock.ExpectExec("INSERT INTO user_lists").WithArgs(args.userId, id).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit().WillReturnError(errors.New("Error Commit"))
			},
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.id)

			got, err := r.Create(testCase.args.userId, testCase.args.list)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.id, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTodoListPostgres_GetAll(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := NewTodoListPostgres(db)

	type mockBehavior func(userId int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		userId       int
		lists        []todo.TodoList
		wantErr      bool
	}{
		{
			name: "Ok",
			mockBehavior: func(userId int) {
				rows := sqlmock.NewRows([]string{"id", "title", "description"}).
					AddRow(1, "title1", "description1").
					AddRow(2, "title2", "description2").
					AddRow(3, "title3", "description3")

				mock.ExpectQuery("SELECT tl.id, tl.title, tl.description FROM ").
					WithArgs(userId).WillReturnRows(rows)
			},
			userId: 88,
			lists: []todo.TodoList{
				{Id: 1, Title: "title1", Description: "description1"},
				{Id: 2, Title: "title2", Description: "description2"},
				{Id: 3, Title: "title3", Description: "description3"},
			},
		},
		{
			name: "No Records",
			mockBehavior: func(userId int) {
				rows := sqlmock.NewRows([]string{"id", "title", "description"})

				mock.ExpectQuery("SELECT tl.id, tl.title, tl.description FROM ").
					WithArgs(userId).WillReturnRows(rows)
			},
			userId: 88,
		},
		{
			name: "Error Select",
			mockBehavior: func(userId int) {
				mock.ExpectQuery("SELECT tl.id, tl.title, tl.description FROM ").
					WithArgs(userId).WillReturnError(errors.New("some error"))
			},
			userId:  88,
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.userId)

			got, err := r.GetAll(testCase.userId)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.lists, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTodoLisrPostgres_GetById(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := NewTodoListPostgres(db)

	type args struct {
		userId int
		listId int
	}

	type mockBehavior func(userId, listId int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		input        args
		want         todo.TodoList
		wantErr      bool
	}{
		{
			name: "Ok",
			mockBehavior: func(userId, listId int) {
				rows := sqlmock.NewRows([]string{"id", "title", "description"}).
					AddRow(1, "title1", "description1")

				mock.ExpectQuery("SELECT tl.id, tl.title, tl.description FROM").
					WithArgs(userId, listId).WillReturnRows(rows)
			},
			input: args{
				listId: 1,
				userId: 1,
			},
			want: todo.TodoList{Id: 1, Title: "title1", Description: "description1"},
		},
		{
			name: "Not Found",
			mockBehavior: func(userId, listId int) {
				rows := sqlmock.NewRows([]string{"id", "title", "description"})

				mock.ExpectQuery("SELECT tl.id, tl.title, tl.description FROM").
					WithArgs(userId, listId).WillReturnRows(rows)
			},
			input: args{
				listId: 404,
				userId: 1,
			},
			wantErr: true,
		},
		{
			name: "Error Select",
			mockBehavior: func(userId, listId int) {
				mock.ExpectQuery("SELECT tl.id, tl.title, tl.description FROM").
					WithArgs(userId, listId).WillReturnError(errors.New("Error SELECT"))
			},
			input: args{
				listId: 500,
				userId: 1,
			},
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.input.userId, testCase.input.listId)

			got, err := r.GetById(testCase.input.userId, testCase.input.listId)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
