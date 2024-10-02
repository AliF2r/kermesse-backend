package stands

import (
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/internal/types"
)

type StandsRepository interface {
	GetAllStands() ([]types.Stand, error)
	GetStandById(id int) (types.Stand, error)
	AddStand(input map[string]interface{}) error
	ModifyStand(id int, input map[string]interface{}) error
	AdjustStock(id int, quantity int) error
}

type Repository struct {
	db *sqlx.DB
}

func NewStandsRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (repository *Repository) GetAllStands() ([]types.Stand, error) {
	var stands []types.Stand
	query := "SELECT * FROM stands"
	err := repository.db.Select(&stands, query)
	return stands, err
}

func (repository *Repository) GetStandById(id int) (types.Stand, error) {
	var stand types.Stand
	query := "SELECT * FROM stands WHERE id=$1"
	err := repository.db.Get(&stand, query, id)
	return stand, err
}

func (repository *Repository) ModifyStand(id int, input map[string]interface{}) error {
	query := "UPDATE stands SET name=$1, description=$2, price=$3, stock=$4 WHERE id=$5"
	_, err := repository.db.Exec(query, input["name"], input["description"], input["price"], input["stock"], id)
	return err
}

func (repository *Repository) AdjustStock(id int, quantity int) error {
	query := "UPDATE stands SET stock=stock+$1 WHERE id=$2"
	_, err := repository.db.Exec(query, quantity, id)
	return err
}

func (repository *Repository) AddStand(input map[string]interface{}) error {
	query := "INSERT INTO stands (user_id, name, description, category, price, stock) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := repository.db.Exec(query, input["user_id"], input["name"], input["description"], input["category"], input["price"], input["stock"])
	return err
}
