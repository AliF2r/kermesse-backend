package users

import (
	"context"
	"database/sql"
	goErrors "errors"
	goJwt "github.com/golang-jwt/jwt/v5"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/hasher"
	"github.com/kermesse-backend/pkg/jwt"
	"github.com/kermesse-backend/pkg/utils"
	"os"
	"strconv"
)

type UsersService interface {
	GetUserById(userID int) (types.UserBasic, error)
	Register(input map[string]interface{}) error
	Login(input map[string]interface{}) (types.UserWithAuthToken, error)
	GetLoggedInUser(ctx context.Context) (types.UserWithAuthToken, error)
	InviteStudent(ctx context.Context, input map[string]interface{}) error
	UpdatePassword(ctx context.Context, id int, input map[string]interface{}) error
	MakePayment(ctx context.Context, input map[string]interface{}) error
	GetAllStudentByParentId(ctx context.Context, params map[string]interface{}) ([]types.UserBasic, error)
	GetAllUsers(params map[string]interface{}) ([]types.UserBasic, error)
	ModifyBalanceFromStripe(userId int, balance int) error
}

type Service struct {
	usersRepository UsersRepository
}

func NewUsersService(usersRepository UsersRepository) *Service {
	return &Service{
		usersRepository: usersRepository,
	}
}

func (service *Service) ModifyBalanceFromStripe(userId int, balance int) error {
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
	if user.Role == types.UserRoleStudent {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	err = service.usersRepository.ModifyBalanceFromStripe(userId, balance)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (service *Service) GetUserById(userID int) (types.UserBasic, error) {
	user, err := service.usersRepository.GetUserById(userID)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return types.UserBasic{}, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return types.UserBasic{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	totalPoint, err := service.usersRepository.GetTotalPoints(userID)
	if err != nil {
		return types.UserBasic{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return types.UserBasic{
		Id:         user.Id,
		Name:       user.Name,
		Email:      user.Email,
		Role:       user.Role,
		Balance:    user.Balance,
		TotalPoint: totalPoint,
	}, nil
}

func (service *Service) GetAllStudentByParentId(ctx context.Context, params map[string]interface{}) ([]types.UserBasic, error) {
	userId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return nil, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("user ID not found in context"),
		}
	}
	filters := make(map[string]interface{})
	if kermesseId, exists := params["kermesse_id"]; exists {
		filters["kermesse_id"] = kermesseId
	}
	users, err := service.usersRepository.GetAllStudentByParentId(userId, filters)
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

func (service *Service) GetAllUsers(params map[string]interface{}) ([]types.UserBasic, error) {

	filters := make(map[string]interface{})
	if kermesseId, exists := params["kermesse_id"]; exists {
		filters["kermesse_id"] = kermesseId
	}

	users, err := service.usersRepository.GetAllUsers(filters)
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

func (service *Service) Register(input map[string]interface{}) error {
	_, err := service.usersRepository.GetUserByEmail(input["email"].(string))
	if err == nil {
		return errors.CustomError{
			Key: errors.EmailAlreadyExists,
			Err: goErrors.New("email already exists"),
		}
	}

	hashedPassword, err := hasher.Hash(input["password"].(string))
	if err != nil {
		return err
	}
	input["password"] = hashedPassword
	input["parent_id"] = nil

	if input["role"] == types.UserRoleStudent {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("role cannot be student"),
		}
	}

	err = service.usersRepository.Create(input)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (service *Service) Login(input map[string]interface{}) (types.UserWithAuthToken, error) {
	user, err := service.usersRepository.GetUserByEmail(input["email"].(string))
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return types.UserWithAuthToken{}, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return types.UserWithAuthToken{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if !hasher.Compare(user.Password, input["password"].(string)) {
		return types.UserWithAuthToken{}, errors.CustomError{
			Key: errors.InvalidCredentials,
			Err: goErrors.New("invalid credentials"),
		}
	}

	expiresIn, err := strconv.Atoi(os.Getenv("JWT_EXPIRES_IN"))
	if err != nil {
		return types.UserWithAuthToken{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	token, err := jwt.Create(os.Getenv("JWT_SECRET"), expiresIn, user.Id)
	if err != nil {
		if goErrors.Is(err, goJwt.ErrTokenExpired) || goErrors.Is(err, goJwt.ErrSignatureInvalid) {
			return types.UserWithAuthToken{}, errors.CustomError{
				Key: errors.Unauthorized,
				Err: err,
			}
		}
		return types.UserWithAuthToken{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	withStand, err := service.usersRepository.AnyStandWithUserId(user.Id)
	if err != nil {
		return types.UserWithAuthToken{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return types.UserWithAuthToken{
		Id:        user.Id,
		Name:      user.Name,
		Email:     user.Email,
		Balance:   user.Balance,
		Role:      user.Role,
		Token:     token,
		WithStand: withStand,
	}, nil
}

func (service *Service) GetLoggedInUser(ctx context.Context) (types.UserWithAuthToken, error) {
	userID, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return types.UserWithAuthToken{}, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("users id not found in context"),
		}
	}

	user, err := service.usersRepository.GetUserById(userID)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return types.UserWithAuthToken{}, errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return types.UserWithAuthToken{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	withStand, err := service.usersRepository.AnyStandWithUserId(user.Id)
	if err != nil {
		return types.UserWithAuthToken{}, errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return types.UserWithAuthToken{
		Id:        user.Id,
		Name:      user.Name,
		Email:     user.Email,
		Balance:   user.Balance,
		Role:      user.Role,
		WithStand: withStand,
	}, nil
}

func (service *Service) InviteStudent(ctx context.Context, input map[string]interface{}) error {
	_, err := service.usersRepository.GetUserByEmail(input["email"].(string))
	if err == nil {
		return errors.CustomError{
			Key: errors.EmailAlreadyExists,
			Err: goErrors.New("email already exists"),
		}
	}

	hashedPassword, err := hasher.Hash("esgi-kermesse")
	if err != nil {
		return err
	}

	parentId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("parent id not found"),
		}
	}

	err = service.usersRepository.Create(map[string]interface{}{
		"parent_id": parentId,
		"name":      input["name"],
		"email":     input["email"],
		"password":  hashedPassword,
		"role":      types.UserRoleStudent,
	})
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (service *Service) MakePayment(ctx context.Context, input map[string]interface{}) error {
	studentId, err := utils.ConvertToInt(input, "student_id")
	if err != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: err,
		}
	}
	student, err := service.usersRepository.GetUserById(studentId)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("student not found"),
		}
	}

	parentId, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("parent id not found"),
		}
	}
	parent, err := service.usersRepository.GetUserById(parentId)
	if err != nil {
		if goErrors.Is(err, sql.ErrNoRows) {
			return errors.CustomError{
				Key: errors.NotFound,
				Err: err,
			}
		}
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("parent not found"),
		}
	}

	if student.ParentId == nil || *student.ParentId != parent.Id {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("not allowed"),
		}
	}

	newBalance, err := utils.ConvertToInt(input, "balance")
	if err != nil {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: err,
		}
	}
	if parent.Balance < newBalance {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("insufficient balance"),
		}
	}

	err = service.usersRepository.AlterBalance(studentId, newBalance)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	err = service.usersRepository.AlterBalance(parentId, -newBalance)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (service *Service) UpdatePassword(ctx context.Context, id int, input map[string]interface{}) error {

	user, err := service.usersRepository.GetUserById(id)
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
			Err: goErrors.New("user ID not found"),
		}
	}

	if user.Id != userId {
		return errors.CustomError{
			Key: errors.Forbidden,
			Err: goErrors.New("forbidden"),
		}
	}

	if !hasher.Compare(user.Password, input["password"].(string)) {
		return errors.CustomError{
			Key: errors.BadRequest,
			Err: goErrors.New("invalid password"),
		}
	}

	hashedPassword, err := hasher.Hash(input["new_password"].(string))
	if err != nil {
		return err
	}
	input["new_password"] = hashedPassword

	if err := service.usersRepository.UpdatePassword(id, input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}
