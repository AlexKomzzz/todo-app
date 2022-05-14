package repository

type Authorization struct {
}

type TodoList struct {
}

type TodoItem struct {
}

type Repository struct {
	Authorization
	TodoList
	TodoItem
}

func NewRepository() *Repository {
	return &Repository{}
}
