package participations

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/internal/types"
	"strings"
)

type ParticipationsRepository interface {
	GetAllParticipations(filters map[string]interface{}) ([]types.ParticipationUserStand, error)
	GetParticipationById(id int) (types.ParticipationCompleteModel, error)
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

func (repository *Repository) GetAllParticipations(filters map[string]interface{}) ([]types.ParticipationUserStand, error) {
	var participations []types.ParticipationUserStand
	baseQuery := `
		SELECT DISTINCT
			p.id AS id,
			p.category AS category,
			p.status AS status,
			p.point AS point,
			p.balance AS balance,
			s.id AS "stand.id",
			s.name AS "stand.name",
			s.description AS "stand.description",
			s.price AS "stand.price",
			s.category AS "stand.category",
			u.id AS "user.id",
			u.name AS "user.name",
			u.email AS "user.email",
			u.role AS "user.role"
		FROM participations p
		JOIN users u ON p.user_id = u.id
		JOIN stands s ON p.stand_id = s.id
		WHERE 1=1
	`

	var conditions []string
	if parentId, ok := filters["parent_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("(u.id = %v OR u.parent_id = %v)", parentId, parentId))
	}
	if studentId, ok := filters["student_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("u.id = %v", studentId))
	}
	if kermesseId, ok := filters["kermesse_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("p.kermesse_id = %v", kermesseId))
	}
	if standHolderId, ok := filters["stand_holder_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("s.user_id = %v", standHolderId))
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	err := repository.db.Select(&participations, baseQuery)

	return participations, err
}

func (repository *Repository) GetParticipationById(id int) (types.ParticipationCompleteModel, error) {
	var participation types.ParticipationCompleteModel
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
