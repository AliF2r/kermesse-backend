package stands

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/internal/types"
	"strings"
)

type StandsRepository interface {
	GetAllStands(filters map[string]interface{}) ([]types.Stand, error)
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

func (repository *Repository) GetAllStands(filters map[string]interface{}) ([]types.Stand, error) {
	var stands []types.Stand
	baseQuery := `
		SELECT DISTINCT
			s.id AS id,
			s.user_id AS user_id,
			s.name AS name,
			s.price AS price,
			s.stock AS stock,
			s.description AS description,
			s.category AS category
		FROM stands s
		LEFT JOIN kermesses_stands ks ON ks.stand_id = s.id
		WHERE 1=1 AND s.id IS NOT NULL
	`

	var conditions []string
	if isReady, ok := filters["is_ready"]; ok && isReady != nil {
		conditions = append(conditions, `
			AND (
				ks.kermesse_id IS NULL
				OR s.id NOT IN (
					SELECT ks_inner.stand_id 
					FROM kermesses_stands ks_inner
					JOIN kermesses k ON ks_inner.kermesse_id = k.id
					WHERE k.status = 'STARTED'
				)
			)
		`)
	}
	if kermesseId, ok := filters["kermesse_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("ks.kermesse_id IS NOT NULL AND ks.kermesse_id = %v", kermesseId))
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	err := repository.db.Select(&stands, baseQuery)

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
