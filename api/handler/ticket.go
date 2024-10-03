package handler

import (
	"github.com/gorilla/mux"
	"github.com/kermesse-backend/api/middleware"
	"github.com/kermesse-backend/internal/tickets"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/internal/users"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/json"
	"net/http"
	"strconv"
)

type TicketHandler struct {
	ticketsService  tickets.TicketService
	usersRepository users.UsersRepository
}

func NewTicketsHandler(ticketsService tickets.TicketService, usersRepository users.UsersRepository) *TicketHandler {
	return &TicketHandler{
		ticketsService:  ticketsService,
		usersRepository: usersRepository,
	}
}

func (h *TicketHandler) RegisterRoutes(mux *mux.Router) {
	mux.Handle("/tickets", errors.ErrorHandler(middleware.IsAuth(h.GetAllTickets, h.usersRepository))).Methods(http.MethodGet)
	mux.Handle("/tickets", errors.ErrorHandler(middleware.IsAuth(h.CreateTicket, h.usersRepository, types.UserRoleStudent))).Methods(http.MethodPost)
	mux.Handle("/tickets/{id}", errors.ErrorHandler(middleware.IsAuth(h.GetTicketById, h.usersRepository))).Methods(http.MethodGet)
}

func (h *TicketHandler) GetAllTickets(w http.ResponseWriter, r *http.Request) error {
	tickets, err := h.ticketsService.GetAllTickets()
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, tickets); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (h *TicketHandler) GetTicketById(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	ticket, err := h.ticketsService.GetTicketById(id)
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, ticket); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (h *TicketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) error {
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if err := h.ticketsService.CreateTicket(r.Context(), input); err != nil {
		return err
	}
	if err := json.Write(w, http.StatusCreated, nil); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}
