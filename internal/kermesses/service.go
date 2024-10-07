package kermesses

import (
	"context"
	"database/sql"
	goErrors "errors"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/internal/users"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/utils"
)

type KermessesService interface {
	GetAllKermesses(ctx context.Context) ([]types.Kermesse, error)
	GetKermesseById(ctx context.Context, id int) (types.KermesseWithStatistics, error)
	AddKermesse(ctx context.Context, input map[string]interface{}) error
	UpdateKermesse(ctx context.Context, id int, input map[string]interface{}) error
	MarkKermesseAsComplete(ctx context.Context, id int) error
	AssignUserToKermesse(ctx context.Context, input map[string]interface{}) error
	AssignStandToKermesse(ctx context.Context, input map[string]interface{}) error
	GetUsersForInvitation(kermesseId int) ([]types.UserBasic, error)
}

type Service struct {
	kermessesRepository KermessesRepository
	usersRepository     users.UsersRepository
}

func NewKermessesService(kermessesRepository KermessesRepository, usersRepository users.UsersRepository) *Service {
	return &Service{
		kermessesRepository: kermessesRepository,
		usersRepository:     usersRepository,
	}
}

func (service *Service) GetAllKermesses(ctx context.Context) ([]types.Kermesse, error) {

	userRole, ok := ctx.Value(types.UserRoleSessionKey).(string)
	if !ok {
		return nil, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user role not found"),
		}
	}

	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return nil, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user Id not found"),
		}
	}

	filters := make(map[string]interface{})
	switch userRole {
	case types.UserRoleStudent:
		filters["student_id"] = userId
	case types.UserRoleOrganizer:
		filters["organizer_id"] = userId
	case types.UserRoleStandHolder:
		filters["stand_holder_id"] = userId
	case types.UserRoleParent:
		filters["parent_id"] = userId
	}

	kermesses, err := service.kermessesRepository.GetAllKermesses(filters)
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return kermesses, nil
}

func (service *Service) GetKermesseById(ctx context.Context, id int) (types.KermesseWithStatistics, error) {

	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return types.KermesseWithStatistics{}, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found in context"),
		}
	}
	userRole, ok := ctx.Value(types.UserRoleSessionKey).(string)
	if !ok {
		return types.KermesseWithStatistics{}, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user role not found in context"),
		}
	}

	kermesse, err := service.kermessesRepository.GetKermesseById(id)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return types.KermesseWithStatistics{}, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return types.KermesseWithStatistics{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	filters := make(map[string]interface{})
	switch userRole {
	case types.UserRoleStudent:
		filters["student_id"] = userId
	case types.UserRoleOrganizer:
		filters["organizer_id"] = userId
	case types.UserRoleStandHolder:
		filters["stand_holder_id"] = userId
	case types.UserRoleParent:
		filters["parent_id"] = userId
	}

	statistics, err := service.kermessesRepository.getStatistics(id, filters)
	if err != nil {
		return types.KermesseWithStatistics{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	KermesseWithStatistics := types.KermesseWithStatistics{
		Id:                   kermesse.Id,
		Name:                 kermesse.Name,
		UserId:               kermesse.UserId,
		Status:               kermesse.Status,
		Description:          kermesse.Description,
		UserNumber:           statistics.UserNumber,
		StandNumber:          statistics.StandNumber,
		TombolaNumber:        statistics.TombolaNumber,
		TombolaBenefit:       statistics.TombolaBenefit,
		ParticipationNumber:  statistics.ParticipationNumber,
		ParticipationBenefit: statistics.ParticipationBenefit,
		Points:               statistics.Points,
	}

	return KermesseWithStatistics, nil
}

func (service *Service) AddKermesse(ctx context.Context, input map[string]interface{}) error {
	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("unable to fetch user id from context"),
		}
	}
	input["user_id"] = userId

	err := service.kermessesRepository.AddKermesse(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (service *Service) UpdateKermesse(ctx context.Context, id int, input map[string]interface{}) error {
	kermesse, err := service.kermessesRepository.GetKermesseById(id)
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
			Err: goErrors.New("cannot update a completed kermesse"),
		}
	}

	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("unable to fetch user id"),
		}
	}
	if kermesse.UserId != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("user is not authorized to modify this kermesse"),
		}
	}

	err = service.kermessesRepository.ModifyKermesse(id, input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (service *Service) MarkKermesseAsComplete(ctx context.Context, id int) error {
	kermesse, err := service.kermessesRepository.GetKermesseById(id)
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
			Err: goErrors.New("kermesse is already marked as complete"),
		}
	}

	canComplete, err := service.kermessesRepository.IsStandLinkable(id)
	if err != nil || !canComplete {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("kermesse cannot be marked as complete because there is at least"),
		}
	}

	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("unable to fetch user id from context"),
		}
	}
	if kermesse.UserId != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("user is not authorized to mark this kermesse as complete"),
		}
	}

	err = service.kermessesRepository.CompleteKermesse(id)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (service *Service) AssignUserToKermesse(ctx context.Context, input map[string]interface{}) error {
	kermesse, err := service.kermessesRepository.GetKermesseById(input["kermesse_id"].(int))
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
			Err: goErrors.New("cannot assign users to a completed kermesse"),
		}
	}

	organizerId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found"),
		}
	}
	if kermesse.UserId != organizerId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	studentId, err := utils.ConvertToInt(input, "user_id")
	if err != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: err,
		}
	}

	student, err := service.usersRepository.GetUserById(studentId)
	if err != nil || student.Role != types.UserRoleStudent {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("invalid user role for this operation"),
		}
	}

	if student.Role != types.UserRoleStudent {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("user is not a child"),
		}
	}

	err = service.kermessesRepository.LinkUserToKermesse(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if student.ParentId != nil {
		input["user_id"] = student.ParentId
		err = service.kermessesRepository.LinkUserToKermesse(input)
		if err != nil {
			return errors.CustomError{
				Key: errors.InternalServerError,
				Err: err,
			}
		}
	}

	return nil
}

func (s *Service) AssignStandToKermesse(ctx context.Context, input map[string]interface{}) error {
	kermesse, err := s.kermessesRepository.GetKermesseById(input["kermesse_id"].(int))
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
			Err: goErrors.New("cannot assign stands to a completed kermesse"),
		}
	}

	standId, err := utils.ConvertToInt(input, "stand_id")
	if err != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: err,
		}
	}
	isStandLinkable, err := s.kermessesRepository.IsStandLinkable(standId)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if !isStandLinkable {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("stand is already linked to a kermesse"),
		}
	}

	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user id not found"),
		}
	}
	if kermesse.UserId != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	err = s.kermessesRepository.LinkStandToKermesse(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (service *Service) GetUsersForInvitation(id int) ([]types.UserBasic, error) {
	users, err := service.kermessesRepository.GetUsersForInvitation(id)
	if err != nil {
		return nil, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if users == nil {
		return []types.UserBasic{}, nil
	}

	return users, nil
}
