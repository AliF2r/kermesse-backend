package participations

import (
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/internal/types"
)

type ParticipationsRepository interface {
	GetAllParticipations() ([]types.Participation, error)
	GetParticipationById(id int) (types.Participation, error)
	AddParticipation(input map[string]interface{}) error
	UpdateParticipation(id int, input map[string]interface{}) error
	IsEligibleForCreation(input map[string]interface{}) (bool, error)
}

type Repository struct {
	db *sqlx.DB
}

func NewParticipationsRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (repository *Repository) GetAllParticipations() ([]types.Participation, error) {
	var participations []types.Participation
	query := "SELECT * FROM participations"
	err := repository.db.Select(&participations, query)
	return participations, err
}

func (repository *Repository) GetParticipationById(id int) (types.Participation, error) {
	var participation types.Participation
	query := "SELECT * FROM participations WHERE id=$1"
	err := repository.db.Get(&participation, query, id)
	return participation, err
}

func (repository *Repository) AddParticipation(input map[string]interface{}) error {
	query := "INSERT INTO participations (user_id, kermesse_id, stand_id, category, balance) VALUES ($1, $2, $3, $4, $5)"
	_, err := repository.db.Exec(query, input["user_id"], input["kermesse_id"], input["stand_id"], input["category"], input["balance"])

	return err
}

func (repository *Repository) UpdateParticipation(id int, input map[string]interface{}) error {
	query := "UPDATE participations SET status=$1, point=$2 WHERE id=$3"
	_, err := repository.db.Exec(query, input["status"], input["point"], id)

	return err
}

func (repository *Repository) IsEligibleForCreation(input map[string]interface{}) (bool, error) {
	var isEligible bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM kermesses_users ku
  			JOIN kermesses_stands ks ON ku.kermesse_id = ks.kermesse_id
			JOIN kermesses k ON ku.kermesse_id = k.id
  			WHERE ku.user_id = $1 AND ks.stand_id = $2 AND k.status = 'STARTED'
		) AS is_associated
 	`
	err := repository.db.QueryRow(query, input["user_id"], input["stand_id"]).Scan(&isEligible)
	return isEligible, err
}
