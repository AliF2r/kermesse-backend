package tickets

import (
	"context"
	"database/sql"
	goErrors "errors"
	"github.com/kermesse-backend/internal/tombolas"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/internal/users"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/utils"
)

type TicketService interface {
	GetAllTickets(ctx context.Context) ([]types.TicketCompleteModel, error)
	GetTicketById(id int) (types.TicketCompleteModel, error)
	CreateTicket(ctx context.Context, input map[string]interface{}) error
}

type Service struct {
	ticketsRepository  TicketRepository
	tombolasRepository tombolas.TombolaRepository
	usersRepository    users.UsersRepository
}

func NewTicketsService(ticketsRepository TicketRepository, tombolasRepository tombolas.TombolaRepository, usersRepository users.UsersRepository) *Service {
	return &Service{
		ticketsRepository:  ticketsRepository,
		tombolasRepository: tombolasRepository,
		usersRepository:    usersRepository,
	}
}

func (service *Service) GetAllTickets(ctx context.Context) ([]types.TicketCompleteModel, error) {
	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return nil, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user Id not found"),
		}
	}
	userRole, ok := ctx.Value(types.UserRoleSessionKey).(string)
	if !ok {
		return nil, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user role not found"),
		}
	}

	filters := make(map[string]interface{})
	switch userRole {
	case types.UserRoleStudent:
		filters["student_id"] = userId
	case types.UserRoleParent:
		filters["parent_id"] = userId
	case types.UserRoleOrganizer:
		filters["organizer_id"] = userId
	}

	tickets, err := service.ticketsRepository.GetAllTickets(filters)
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if tickets == nil {
		return []types.TicketCompleteModel{}, nil
	}

	return tickets, nil
}

func (service *Service) GetTicketById(id int) (types.TicketCompleteModel, error) {
	ticket, err := service.ticketsRepository.GetTicketById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return ticket, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return ticket, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return ticket, nil
}

func (service *Service) CreateTicket(ctx context.Context, input map[string]interface{}) error {
	tombolaId, err := utils.ConvertToInt(input, "tombola_id")
	if err != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: err,
		}
	}
	tombola, err := service.tombolasRepository.GetTombolaById(tombolaId)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if tombola.Status != types.TombolaStatusStarted {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("tombola is not active or has ended"),
		}
	}
	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user ID not found in context"),
		}
	}
	user, err := service.usersRepository.GetUserById(userId)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if user.Balance < tombola.Price {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("insufficient balance"),
		}
	}

	canBeCreated, err := service.ticketsRepository.IsEligibleForTicketCreation(map[string]interface{}{
		"kermesse_id": tombola.KermesseId,
		"user_id":     userId,
	})
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if !canBeCreated {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("not eligible to create ticket"),
		}
	}
	err = service.usersRepository.AlterBalance(userId, -tombola.Price)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	input["user_id"] = userId
	err = service.ticketsRepository.AddTicket(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}
