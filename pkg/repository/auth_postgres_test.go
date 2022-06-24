package repository

import (
	"errors"
	"testing"
	"todo-app"

	"github.com/stretchr/testify/assert"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
)

func TestAuthPostgres_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := NewAuthPostgres(db)

	type mockBehavior func(user todo.User, id int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		user         todo.User
		id           int
		wantErr      bool
	}{
		{
			name: "OK",
			user: todo.User{
				Name:     "Alex",
				Username: "alexkomzzz",
				Password: "qwerty",
			},
			id: 55,
			mockBehavior: func(user todo.User, id int) {
				row := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(user.Name, user.Username, user.Password).WillReturnRows(row)
			},
		},
		{
			name: "Error Create",
			user: todo.User{
				Name:     "Alex",
				Username: "alexkomzzz",
				Password: "qwerty",
			},
			mockBehavior: func(user todo.User, id int) {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(user.Name, user.Username, user.Password).WillReturnError(errors.New("Error Create"))
			},
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.user, testCase.id)

			gotId, err := r.CreateUser(testCase.user)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.id, gotId)
			}
		})
	}
}

func TestAuthPostgres_GeteUser(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := NewAuthPostgres(db)

	type input struct {
		username string
		password string
	}

	type mockBehavior func(input input, user todo.User)

	testTable := []struct {
		name         string
		user         todo.User
		input        input
		mockBehavior mockBehavior
		wantErr      bool
	}{
		{
			name: "OK",
			user: todo.User{
				Id:       22,
				Name:     "Alex",
				Username: "alexkomzzz",
				Password: "qwerty",
			},
			input: input{
				username: "alexkomzzz",
				password: "qwerty",
			},
			mockBehavior: func(input input, user todo.User) {
				rows := sqlmock.NewRows([]string{"id", "name", "username", "password"}).
					AddRow(user.Id, user.Name, user.Username, user.Password)
				mock.ExpectQuery("SELECT id FROM users").
					WithArgs(input.username, input.password).WillReturnRows(rows)
			},
		},
		{
			name: "Error GetUser",
			input: input{
				username: "alexkomzzz",
				password: "qwerty",
			},
			mockBehavior: func(input input, user todo.User) {
				mock.ExpectQuery("SELECT id FROM users").
					WithArgs(input.username, input.password).WillReturnError(errors.New("Error GetUser"))
			},
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.input, testCase.user)

			gotUser, err := r.GetUser(testCase.input.username, testCase.input.password)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.user, gotUser)
			}
		})
	}
}
