package kermesses

import (
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/internal/types"
)

type KermessesRepository interface {
	AddKermesse(input map[string]interface{}) error
	GetAllKermesses() ([]types.Kermesse, error)
	GetKermesseById(id int) (types.Kermesse, error)
	ModifyKermesse(id int, input map[string]interface{}) error
	CompleteKermesse(id int) error
	IsStandLinkable(standId int) (bool, error)
	LinkStandToKermesse(input map[string]interface{}) error
	IsCompletionAllowed(id int) (bool, error)
	LinkUserToKermesse(input map[string]interface{}) error
}

type Repository struct {
	db *sqlx.DB
}

func NewkermessesRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (repository *Repository) AddKermesse(input map[string]interface{}) error {
	query := "INSERT INTO kermesses (user_id, name, description) VALUES ($1, $2, $3)"
	_, err := repository.db.Exec(query, input["user_id"], input["name"], input["description"])
	return err
}

func (repository *Repository) GetAllKermesses() ([]types.Kermesse, error) {
	var kermesses []types.Kermesse
	query := "SELECT * FROM kermesses"
	err := repository.db.Select(&kermesses, query)
	return kermesses, err
}

func (repository *Repository) GetKermesseById(id int) (types.Kermesse, error) {
	var kermesse types.Kermesse
	query := "SELECT * FROM kermesses WHERE id=$1"
	err := repository.db.Get(&kermesse, query, id)
	return kermesse, err
}

func (repository *Repository) ModifyKermesse(id int, input map[string]interface{}) error {
	query := "UPDATE kermesses SET name=$1, description=$2 WHERE id=$3"
	_, err := repository.db.Exec(query, input["name"], input["description"], id)

	return err
}

func (repository *Repository) CompleteKermesse(id int) error {
	query := "UPDATE kermesses SET status='FINISHED' WHERE id=$1"
	_, err := repository.db.Exec(query, id)
	return err
}

func (repository *Repository) IsStandLinkable(standId int) (bool, error) {
	var canLink bool
	query := `SELECT EXISTS ( SELECT 1 FROM kermesses_stands ks JOIN kermesses k ON ks.kermesse_id = k.id WHERE ks.stand_id = $1 AND k.status = 'STARTED' ) AS is_linkable`
	err := repository.db.QueryRow(query, standId).Scan(&canLink)
	return !canLink, err
}

func (repository *Repository) LinkStandToKermesse(input map[string]interface{}) error {
	query := "INSERT INTO kermesses_stands (kermesse_id, stand_id) VALUES ($1, $2)"
	_, err := repository.db.Exec(query, input["kermesse_id"], input["stand_id"])
	return err
}

func (r *Repository) IsCompletionAllowed(id int) (bool, error) {
	var completionAllowed bool
	query := "SELECT EXISTS ( SELECT 1 FROM tombolas WHERE kermesse_id = $1 AND status = 'STARTED' ) AS can_end"
	err := r.db.QueryRow(query, id).Scan(&completionAllowed)
	return !completionAllowed, err
}

func (repository *Repository) LinkUserToKermesse(input map[string]interface{}) error {
	query := "INSERT INTO kermesses_users (kermesse_id, user_id) VALUES ($1, $2)"
	_, err := repository.db.Exec(query, input["kermesse_id"], input["user_id"])
	return err
}
