package users

import (
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/internal/types"
)

type UsersRepository interface {
	GetUserById(userId int) (types.User, error)
	GetUserByEmail(email string) (types.User, error)
	Create(newUser map[string]interface{}) error
	AlterBalance(userId int, newBalance int) error
}

type Repository struct {
	db *sqlx.DB
}

func NewUsersRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (repository *Repository) Create(newUser map[string]interface{}) error {
	query := "INSERT INTO users (parent_id, name, email, password, role) VALUES ($1, $2, $3, $4, $5)"
	_, err := repository.db.Exec(query, newUser["parent_id"], newUser["name"], newUser["email"], newUser["password"], newUser["role"])
	return err
}

func (repository *Repository) GetUserById(userId int) (types.User, error) {
	var user types.User
	query := "SELECT * FROM users WHERE id=$1"
	err := repository.db.Get(&user, query, userId)
	return user, err
}

func (repository *Repository) GetUserByEmail(email string) (types.User, error) {
	var user types.User
	query := "SELECT * FROM users WHERE email=$1"
	err := repository.db.Get(&user, query, email)
	return user, err
}

func (repository *Repository) AlterBalance(userId int, newBalance int) error {
	query := "UPDATE users SET balance = balance + $1 WHERE id = $2"
	_, err := repository.db.Exec(query, newBalance, userId)
	return err
}
