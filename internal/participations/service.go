package participations

import (
	"context"
	"database/sql"
	goErrors "errors"

	"github.com/kermesse-backend/internal/kermesses"
	"github.com/kermesse-backend/internal/stands"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/internal/users"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/utils"
)

type ParticipationsService interface {
	GetAllParticipations(ctx context.Context, params map[string]interface{}) ([]types.ParticipationUserStand, error)
	GetParticipationById(id int) (types.ParticipationCompleteModel, error)
	AddParticipation(ctx context.Context, input map[string]interface{}) error
	ModifyParticipation(ctx context.Context, id int, input map[string]interface{}) error
}

type Service struct {
	participationsRepository ParticipationsRepository
	kermessesRepository      kermesses.KermessesRepository
	usersRepository          users.UsersRepository
	standsRepository         stands.StandsRepository
}

func NewParticipationsService(usersRepository users.UsersRepository, kermessesRepository kermesses.KermessesRepository, participationsRepository ParticipationsRepository, standsRepository stands.StandsRepository) *Service {
	return &Service{
		participationsRepository: participationsRepository,
		kermessesRepository:      kermessesRepository,
		usersRepository:          usersRepository,
		standsRepository:         standsRepository,
	}
}

func (service *Service) GetAllParticipations(ctx context.Context, params map[string]interface{}) ([]types.ParticipationUserStand, error) {
	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return nil, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user ID not found"),
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
	case types.UserRoleStandHolder:
		filters["stand_holder_id"] = userId
	}

	if kermesseId, exists := params["kermesse_id"]; exists {
		filters["kermesse_id"] = kermesseId
	}

	participations, err := service.participationsRepository.GetAllParticipations(filters)
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if participations == nil {
		return []types.ParticipationUserStand{}, nil
	}

	return participations, nil
}

func (service *Service) GetParticipationById(id int) (types.ParticipationCompleteModel, error) {
	participation, err := service.participationsRepository.GetParticipationById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return participation, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return participation, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return participation, nil
}

func (service *Service) AddParticipation(ctx context.Context, input map[string]interface{}) error {
	standId, err := utils.ConvertToInt(input, "stand_id")
	if err != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: err,
		}
	}
	stand, err := service.standsRepository.GetStandById(standId)
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
	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("unable to retrieve user id"),
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

	canBeCreated, err := service.participationsRepository.IsEligibleForCreation(map[string]interface{}{
		"user_id":  userId,
		"stand_id": standId,
	})
	if err != nil || !canBeCreated {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("participation creation is not allowed"),
		}
	}

	totalPrice := stand.Price
	quantity := 1
	if stand.Category == types.ParticipationTypeFood {
		quantity, err = utils.ConvertToInt(input, "quantity")
		if err != nil {
			return errors.CustomError{
				Key: errors.BadRequest,
				Err: err,
			}
		}
		totalPrice *= quantity
	}

	if stand.Category == types.ParticipationTypeFood && stand.Stock < totalPrice {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("insufficient stock"),
		}
	}

	if user.Balance < totalPrice {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("insufficient balance"),
		}
	}

	if stand.Category == types.ParticipationTypeFood {
		err = service.standsRepository.AdjustStock(standId, -quantity)
		if err != nil {
			return errors.CustomError{
				Key: errors.InternalServerError,
				Err: err,
			}
		}
	}

	err = service.usersRepository.AlterBalance(userId, -totalPrice)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	err = service.usersRepository.AlterBalance(stand.UserId, totalPrice)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	input["balance"] = totalPrice
	input["user_id"] = userId
	input["category"] = stand.Category

	err = service.participationsRepository.AddParticipation(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (service *Service) ModifyParticipation(ctx context.Context, id int, input map[string]interface{}) error {
	participation, err := service.participationsRepository.GetParticipationById(id)
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

	if participation.Stand.Type != types.ParticipationTypeGame {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("participation type is not activity"),
		}
	}

	kermesse, err := service.kermessesRepository.GetKermesseById(participation.Kermesse.Id)
	if err != nil || kermesse.Status == types.KermesseStatusFinished {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("cannot modify participation in a finished kermesse"),
		}
	}

	stand, err := service.standsRepository.GetStandById(participation.Stand.Id)
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

	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok || stand.UserId != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("user is not authorized to modify this participation"),
		}
	}

	err = service.participationsRepository.UpdateParticipation(id, map[string]interface{}{
		"point":  input["point"],
		"status": types.ParticipationStatusFinished,
	})
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}
