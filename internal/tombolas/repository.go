package tombolas

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/internal/types"
	"strings"
)

type TombolaRepository interface {
	GetAllTombolas(filters map[string]interface{}) ([]types.Tombola, error)
	GetTombolaById(id int) (types.Tombola, error)
	AddTombola(input map[string]interface{}) error
	ModifyTombola(id int, input map[string]interface{}) error
	SelectWinner(id int) error
}

type Repository struct {
	db *sqlx.DB
}

func NewTombolasRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (repository *Repository) GetAllTombolas(filters map[string]interface{}) ([]types.Tombola, error) {
	var tombolas []types.Tombola
	baseQuery := `
		SELECT DISTINCT
			t.id AS id,
			t.kermesse_id AS kermesse_id,
			t.name AS name,
			t.prize AS prize,
			t.price AS price,
			t.status AS status
		FROM tombolas t WHERE 1=1
	`

	var conditions []string
	if kermesseId, ok := filters["kermesse_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("t.kermesse_id = %v", kermesseId))
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}
	err := repository.db.Select(&tombolas, baseQuery)
	return tombolas, err
}

func (repository *Repository) GetTombolaById(id int) (types.Tombola, error) {
	var tombola types.Tombola
	query := "SELECT * FROM tombolas WHERE id=$1"
	err := repository.db.Get(&tombola, query, id)

	return tombola, err
}

func (repository *Repository) AddTombola(input map[string]interface{}) error {
	query := "INSERT INTO tombolas (kermesse_id, name, price, prize) VALUES ($1, $2, $3, $4)"
	_, err := repository.db.Exec(query, input["kermesse_id"], input["name"], input["price"], input["prize"])

	return err
}

func (repository *Repository) ModifyTombola(id int, input map[string]interface{}) error {
	query := "UPDATE tombolas SET name=$1, price=$2, prize=$3 WHERE id=$4"
	_, err := repository.db.Exec(query, input["name"], input["price"], input["prize"], id)
	return err
}

func (repository *Repository) SelectWinner(id int) error {
	tx, err := repository.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	query := "UPDATE tombolas SET status='FINISHED' WHERE id=$1"
	_, err = tx.Exec(query, id)
	if err != nil {
		return err
	}

	query = `
		UPDATE tickets
		SET is_winner = true
		WHERE id = ( SELECT id FROM tickets WHERE tombola_id = $1 ORDER BY RANDOM() LIMIT 1 )
		AND tombola_id = $1
	`
	_, err = tx.Exec(query, id)
	if err != nil {
		return err
	}
	return err
}
