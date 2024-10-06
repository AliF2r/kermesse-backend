package users

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/internal/types"
	"strings"
)

type UsersRepository interface {
	GetUserById(userId int) (types.User, error)
	GetUserByEmail(email string) (types.User, error)
	Create(newUser map[string]interface{}) error
	UpdatePassword(id int, input map[string]interface{}) error
	AlterBalance(userId int, newBalance int) error
	AnyStandWithUserId(id int) (bool, error)
	GetAllUsers(filters map[string]interface{}) ([]types.UserBasic, error)
	GetAllStudentByParentId(id int, filters map[string]interface{}) ([]types.UserBasic, error)
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

func (repository *Repository) GetAllUsers(filters map[string]interface{}) ([]types.UserBasic, error) {
	var users []types.UserBasic
	baseQuery := `
		SELECT DISTINCT
			u.id AS id,
			u.name AS name,
			u.email AS email,
			u.balance AS balance,
			u.role AS role
		FROM users u
		FULL OUTER JOIN kermesses_users ku ON ku.user_id = u.id
		WHERE 1=1
	`

	var conditions []string
	if kermesseID, ok := filters["kermesse_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("ku.kermesse_id = %v", kermesseID))
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	err := repository.db.Select(&users, baseQuery)
	return users, err
}

func (repository *Repository) GetAllStudentByParentId(id int, filters map[string]interface{}) ([]types.UserBasic, error) {
	var users []types.UserBasic
	baseQuery := `
		SELECT DISTINCT
			u.id AS id,
			u.name AS name,
			u.email AS email,
			u.balance AS balance,
			u.role AS role
		FROM users u
		FULL OUTER JOIN kermesses_users ku ON ku.user_id = u.id
		WHERE u.role = 'STUDENT' AND u.parent_id = $1
	`

	if kermesseId, ok := filters["kermesse_id"]; ok {
		baseQuery += fmt.Sprintf(" AND ku.kermesse_id = %v", kermesseId)
	}

	err := repository.db.Select(&users, baseQuery, id)

	return users, err
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

func (repository *Repository) UpdatePassword(id int, input map[string]interface{}) error {
	query := "UPDATE users SET password=$1 WHERE id=$2"
	_, err := repository.db.Exec(query, input["new_password"], id)
	return err
}

func (repository *Repository) AlterBalance(userId int, newBalance int) error {
	query := "UPDATE users SET balance = balance + $1 WHERE id = $2"
	_, err := repository.db.Exec(query, newBalance, userId)
	return err
}

func (repository *Repository) AnyStandWithUserId(id int) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM stands 
		WHERE user_id = $1 OR user_id IS NULL
	`
	err := repository.db.Get(&count, query, id)
	return count >= 1, err
}
