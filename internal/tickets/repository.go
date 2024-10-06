package tickets

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/internal/types"
	"strings"
)

type TicketRepository interface {
	GetAllTickets(filters map[string]interface{}) ([]types.TicketCompleteModel, error)
	GetTicketById(id int) (types.TicketCompleteModel, error)
	AddTicket(input map[string]interface{}) error
	IsEligibleForTicketCreation(input map[string]interface{}) (bool, error)
}

type Repository struct {
	db *sqlx.DB
}

func NewTicketsRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (repository *Repository) GetAllTickets(filters map[string]interface{}) ([]types.TicketCompleteModel, error) {
	var tickets []types.TicketCompleteModel
	baseQuery := `
		SELECT DISTINCT
			u.id AS "user.id",
			u.name AS "user.name",
			u.email AS "user.email",
			u.role AS "user.role",
			k.id AS "kermesse.id",
			k.name AS "kermesse.name",
			k.description AS "kermesse.description",
			k.status AS "kermesse.status",
			t.id AS "tombola.id",
			t.name AS "tombola.name",
			t.status AS "tombola.status",
			t.price AS "tombola.price",
			t.prize AS "tombola.prize",
			ticket.id AS id,
			ticket.is_winner AS is_winner
		FROM tickets ticket
		JOIN users u ON ticket.user_id = u.id
		JOIN tombolas t ON ticket.tombola_id = t.id
		JOIN kermesses k ON t.kermesse_id = k.id
		WHERE 1=1
	`

	// Applying dynamic filters
	var conditions []string
	if organizerId, ok := filters["organizer_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("k.user_id IS NOT NULL AND k.user_id = %v", organizerId))
	}
	if studentId, ok := filters["student_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("ticket.user_id IS NOT NULL AND ticket.user_id = %v", studentId))
	}
	if parentId, ok := filters["parent_id"]; ok {
		conditions = append(conditions, fmt.Sprintf("u.parent_id IS NOT NULL AND u.parent_id = %v", parentId))
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}
	err := repository.db.Select(&tickets, baseQuery)
	return tickets, err
}

func (repository *Repository) GetTicketById(id int) (types.TicketCompleteModel, error) {
	var ticket types.TicketCompleteModel
	query := `
		SELECT
			ticket.id AS id,
			ticket.is_winner AS is_winner,
			t.id AS "tombola.id",
			t.name AS "tombola.name",
			t.prize AS "tombola.prize",
			t.price AS "tombola.price",
			t.status AS "tombola.status",
			u.id AS "user.id",
			u.email AS "user.email",
			u.name AS "user.name",
			u.role AS "user.role",
			k.id AS "kermesse.id",
			k.name AS "kermesse.name",
			k.description AS "kermesse.description",
			k.status AS "kermesse.status"
		FROM tickets ticket
		JOIN tombolas t ON ticket.tombola_id = t.id
		JOIN kermesses k ON t.kermesse_id = k.id
		JOIN users u ON ticket.user_id = u.id
		WHERE t.id=$1
	`
	err := repository.db.Get(&ticket, query, id)
	return ticket, err
}

func (repository *Repository) IsEligibleForTicketCreation(input map[string]interface{}) (bool, error) {
	var isEligible bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM kermesses_users ku
			JOIN kermesses k ON k.id = ku.kermesse_id
			WHERE ku.kermesse_id = $1 AND ku.user_id = $2 AND k.status = 'STARTED'
		) AS is_eligible
	`
	err := repository.db.QueryRow(query, input["kermesse_id"], input["user_id"]).Scan(&isEligible)
	return isEligible, err
}

func (repository *Repository) AddTicket(input map[string]interface{}) error {
	query := "INSERT INTO tickets (user_id, tombola_id) VALUES ($1, $2)"
	_, err := repository.db.Exec(query, input["user_id"], input["tombola_id"])
	return err
}
