package kermesses

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/internal/types"
	"strings"
)

type KermessesRepository interface {
	AddKermesse(input map[string]interface{}) error
	GetAllKermesses(filters map[string]interface{}) ([]types.Kermesse, error)
	GetKermesseById(id int) (types.Kermesse, error)
	ModifyKermesse(id int, input map[string]interface{}) error
	CompleteKermesse(id int) error
	IsStandLinkable(standId int) (bool, error)
	LinkStandToKermesse(input map[string]interface{}) error
	IsCompletionAllowed(id int) (bool, error)
	LinkUserToKermesse(input map[string]interface{}) error
	GetUsersForInvitation(kermesseId int) ([]types.UserBasic, error)
	getStatistics(id int, filters map[string]interface{}) (types.KermesseStatistics, error)
	IsAllTombolaFinished(kermesseId int) (bool, error)
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

func (repository *Repository) GetAllKermesses(filters map[string]interface{}) ([]types.Kermesse, error) {
	kermesses := []types.Kermesse{}
	baseQuery := `
		SELECT DISTINCT
			k.id AS id,
			k.user_id AS user_id,
			k.name AS name,
			k.description AS description,
			k.status AS status
		FROM kermesses k
		    FULL OUTER JOIN kermesses_stands ks ON ks.kermesse_id = k.id
			FULL OUTER JOIN kermesses_users ku ON ku.kermesse_id = k.id
			FULL OUTER JOIN stands s ON ks.stand_id = s.id
			WHERE 1=1
		`

	var conditions []string

	if studentId, ok := filters["student_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("ku.user_id = %v", studentId))
	}
	if organizerId, ok := filters["organizer_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("k.user_id = %v", organizerId))
	}
	if parentId, ok := filters["parent_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("ku.user_id = %v", parentId))
	}
	if standHolderId, ok := filters["stand_holder_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("ks.stand_id IS NOT NULL AND s.user_id = %v", standHolderId))
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	err := repository.db.Select(&kermesses, baseQuery)

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

func (repository *Repository) IsAllTombolaFinished(kermesseId int) (bool, error) {
	var allFinished bool
	query := `
        SELECT COUNT(*) = 0
        FROM tombolas
        WHERE kermesse_id = $1
        AND status != 'FINISHED';
    `
	err := repository.db.QueryRow(query, kermesseId).Scan(&allFinished)
	if err != nil {
		return false, err
	}
	return allFinished, nil
}

func (repository *Repository) LinkStandToKermesse(input map[string]interface{}) error {
	query := "INSERT INTO kermesses_stands (kermesse_id, stand_id) VALUES ($1, $2)"
	_, err := repository.db.Exec(query, input["kermesse_id"], input["stand_id"])
	return err
}

func (repository *Repository) IsCompletionAllowed(id int) (bool, error) {
	var completionAllowed bool
	query := "SELECT EXISTS ( SELECT 1 FROM tombolas WHERE kermesse_id = $1 AND status = 'STARTED' ) AS can_end"
	err := repository.db.QueryRow(query, id).Scan(&completionAllowed)
	return !completionAllowed, err
}

func (repository *Repository) LinkUserToKermesse(input map[string]interface{}) error {
	query := "INSERT INTO kermesses_users (kermesse_id, user_id) VALUES ($1, $2)"
	_, err := repository.db.Exec(query, input["kermesse_id"], input["user_id"])
	return err
}

func (repository *Repository) GetUsersForInvitation(kermesseId int) ([]types.UserBasic, error) {
	var users []types.UserBasic
	query := `
		SELECT DISTINCT
			u.id AS id,
			u.name AS name,
			u.email AS email,
			u.balance AS balance,
			u.role AS role
		FROM users u
		LEFT JOIN kermesses_users ku ON u.id = ku.user_id
		WHERE u.id IS NOT NULL
		AND u.role = 'STUDENT'
		AND (ku.kermesse_id IS NULL OR ku.kermesse_id != $1)
	`
	err := repository.db.Select(&users, query, kermesseId)

	return users, err
}

func (repository *Repository) getStatistics(id int, filters map[string]interface{}) (types.KermesseStatistics, error) {
	statistics := types.KermesseStatistics{}

	if err := repository.getStandNumber(id, &statistics.StandNumber); err != nil {
		return types.KermesseStatistics{}, err
	}

	if err := repository.getTombolaNumber(id, &statistics.TombolaNumber); err != nil {
		return types.KermesseStatistics{}, err
	}

	if err := repository.getUserNumber(id, filters, &statistics.UserNumber); err != nil {
		return types.KermesseStatistics{}, err
	}

	if err := repository.getParticipationStatistics(id, filters, &statistics.ParticipationNumber, &statistics.ParticipationBenefit); err != nil {
		return types.KermesseStatistics{}, err
	}

	if filters["organizer_id"] != nil {
		if err := repository.getTombolaBenefits(id, &statistics.TombolaBenefit); err != nil {
			return types.KermesseStatistics{}, err
		}
	}

	if filters["student_id"] != nil {
		if err := repository.getPoints(id, filters["student_id"].(int), &statistics.Points); err != nil {
			return types.KermesseStatistics{}, err
		}
	}

	return types.KermesseStatistics{
		UserNumber:           statistics.UserNumber,
		StandNumber:          statistics.StandNumber,
		TombolaNumber:        statistics.TombolaNumber,
		TombolaBenefit:       statistics.TombolaBenefit,
		ParticipationNumber:  statistics.ParticipationNumber,
		ParticipationBenefit: statistics.ParticipationBenefit,
		Points:               statistics.Points,
	}, nil
}

func (repository *Repository) getStandNumber(kermesseId int, standNumber *int) error {
	query := "SELECT COUNT(*) FROM kermesses_stands WHERE kermesse_id=$1"
	return repository.db.Get(standNumber, query, kermesseId)
}

func (repository *Repository) getTombolaNumber(kermesseId int, tombolaNumber *int) error {
	query := "SELECT COUNT(*) FROM tombolas WHERE kermesse_id=$1"
	return repository.db.Get(tombolaNumber, query, kermesseId)
}

func (repository *Repository) getUserNumber(kermesseId int, filters map[string]interface{}, userNumber *int) error {
	query := `SELECT COUNT(*) FROM kermesses_users ku JOIN users u ON ku.user_id = u.id WHERE ku.kermesse_id=$1`
	if filters["parent_id"] != nil {
		query += fmt.Sprintf(" AND u.role='%v' AND u.parent_id=%v", types.UserRoleStudent, filters["parent_id"])
	}
	return repository.db.Get(userNumber, query, kermesseId)
}

func (repository *Repository) getParticipationStatistics(kermesseId int, filters map[string]interface{}, participationNumber *int, participationBenefits *int) error {
	query := `SELECT COUNT(*) FROM participations p JOIN stands s ON p.stand_id = s.id WHERE p.kermesse_id=$1`
	if filters["stand_holder_id"] != nil {
		query += fmt.Sprintf(" AND s.user_id=%v", filters["stand_holder_id"])
	}
	err := repository.db.Get(participationNumber, query, kermesseId)
	if err != nil {
		return err
	}

	query = `SELECT COALESCE(SUM(p.balance), 0) FROM participations p JOIN stands s ON p.stand_id = s.id WHERE p.kermesse_id=$1`
	if filters["stand_holder_id"] != nil {
		query += fmt.Sprintf(" AND s.user_id=%v", filters["stand_holder_id"])
	}
	return repository.db.Get(participationBenefits, query, kermesseId)
}

func (repository *Repository) getTombolaBenefits(kermesseId int, tombolaBenefits *int) error {
	query := `SELECT COALESCE(SUM(tb.price), 0) FROM tickets t JOIN tombolas tb ON t.tombola_id = tb.id WHERE tb.kermesse_id=$1`
	return repository.db.Get(tombolaBenefits, query, kermesseId)
}

func (repository *Repository) getPoints(kermesseId int, userId int, points *int) error {
	query := "SELECT COALESCE(SUM(point), 0) FROM participations WHERE kermesse_id=$1 AND user_id=$2"
	return repository.db.Get(points, query, kermesseId, userId)
}
