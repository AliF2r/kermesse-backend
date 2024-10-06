package stands

import (
	"context"
	"database/sql"
	goErrors "errors"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/pkg/errors"
)

type StandsService interface {
	GetAllStands(params map[string]interface{}) ([]types.Stand, error)
	GetStandById(id int) (types.Stand, error)
	AddStand(ctx context.Context, input map[string]interface{}) error
	ModifyStand(ctx context.Context, id int, input map[string]interface{}) error // TODO: to get by userID
}

type Service struct {
	standsRepository StandsRepository
}

func NewStandsService(standsRepository StandsRepository) *Service {
	return &Service{
		standsRepository: standsRepository,
	}
}

func (service *Service) GetAllStands(params map[string]interface{}) ([]types.Stand, error) {
	filters := make(map[string]interface{})
	if kermesseId, exists := params["kermesse_id"]; exists {
		filters["kermesse_id"] = kermesseId
	}
	if isReady, exists := params["is_ready"]; exists {
		filters["is_ready"] = isReady
	}
	stands, err := service.standsRepository.GetAllStands(filters)
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return stands, nil
}

func (service *Service) GetStandById(id int) (types.Stand, error) {
	stand, err := service.standsRepository.GetStandById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return stand, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return stand, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return stand, nil
}

func (service *Service) AddStand(ctx context.Context, input map[string]interface{}) error {
	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}

	input["user_id"] = userId
	err := service.standsRepository.AddStand(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (service *Service) ModifyStand(ctx context.Context, id int, input map[string]interface{}) error {
	stand, err := service.standsRepository.GetStandById(id)
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
			Err: goErrors.New("user id not found"),
		}
	}
	if stand.UserId != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("user is not the holder"),
		}
	}
	err = service.standsRepository.ModifyStand(id, input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}
