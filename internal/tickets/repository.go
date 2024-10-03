package tickets

import (
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/internal/types"
)

type TicketRepository interface {
	GetAllTickets() ([]types.Ticket, error)
	GetTicketById(id int) (types.Ticket, error)
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

func (repository *Repository) GetAllTickets() ([]types.Ticket, error) {
	var tickets []types.Ticket
	query := "SELECT * FROM tickets"
	err := repository.db.Select(&tickets, query)
	return tickets, err
}

func (repository *Repository) GetTicketById(id int) (types.Ticket, error) {
	var ticket types.Ticket
	query := "SELECT * FROM tickets WHERE id=$1"
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
