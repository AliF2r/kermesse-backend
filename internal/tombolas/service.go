package tombolas

import (
	"context"
	"database/sql"
	goErrors "errors"
	"github.com/kermesse-backend/internal/kermesses"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/utils"
)

type TombolaService interface {
	GetAllTombolas(params map[string]interface{}) ([]types.Tombola, error)
	GetTombolaById(id int) (types.Tombola, error)
	AddTombola(ctx context.Context, input map[string]interface{}) error
	ModifyTombola(ctx context.Context, id int, input map[string]interface{}) error
	FinishTombola(ctx context.Context, id int) error
}

type Service struct {
	tombolasRepository  TombolaRepository
	kermessesRepository kermesses.KermessesRepository
}

func NewTombolasService(tombolasRepository TombolaRepository, kermessesRepository kermesses.KermessesRepository) *Service {
	return &Service{
		tombolasRepository:  tombolasRepository,
		kermessesRepository: kermessesRepository,
	}
}

func (service *Service) GetAllTombolas(params map[string]interface{}) ([]types.Tombola, error) {
	filters := make(map[string]interface{})
	if kermesseId, exists := params["kermesse_id"]; exists {
		filters["kermesse_id"] = kermesseId
	}
	tombolas, err := service.tombolasRepository.GetAllTombolas(filters)
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if tombolas == nil {
		return []types.Tombola{}, nil
	}

	return tombolas, nil
}

func (service *Service) GetTombolaById(id int) (types.Tombola, error) {
	tombola, err := service.tombolasRepository.GetTombolaById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return tombola, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return tombola, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return tombola, nil
}

func (service *Service) AddTombola(ctx context.Context, input map[string]interface{}) error {
	kermesseId, err := utils.ConvertToInt(input, "kermesse_id")
	if err != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: err,
		}
	}
	kermesse, err := service.kermessesRepository.GetKermesseById(kermesseId)
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

	if kermesse.Status == types.KermesseStatusFinished {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse is finished"),
		}
	}

	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok || kermesse.UserId != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("unauthorized"),
		}
	}

	err = service.tombolasRepository.AddTombola(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (service *Service) ModifyTombola(ctx context.Context, id int, input map[string]interface{}) error {
	tombola, err := service.tombolasRepository.GetTombolaById(id)
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

	kermesse, err := service.kermessesRepository.GetKermesseById(tombola.KermesseId)
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

	if kermesse.Status == types.KermesseStatusFinished {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("cannot modify tombola in a finished kermesse"),
		}
	}

	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok || kermesse.UserId != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("unauthorized"),
		}
	}

	err = service.tombolasRepository.ModifyTombola(id, input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (service *Service) FinishTombola(ctx context.Context, id int) error {
	tombola, err := service.tombolasRepository.GetTombolaById(id)
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

	kermesse, err := service.kermessesRepository.GetKermesseById(tombola.KermesseId)
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

	if kermesse.Status == types.KermesseStatusFinished {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("cannot end tombola in a finished kermesse"),
		}
	}

	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok || kermesse.UserId != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("unauthorized"),
		}
	}

	if tombola.Status != types.TombolaStatusStarted {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("tombola is not started"),
		}
	}
	err = service.tombolasRepository.SelectWinner(id)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}
