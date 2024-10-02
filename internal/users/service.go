package users

import (
	"context"
	"database/sql"
	goErrors "errors"
	goJwt "github.com/golang-jwt/jwt/v5"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/generator"
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
	GetLoggedInUser(ctx context.Context) (types.UserBasic, error)
	InviteStudent(ctx context.Context, input map[string]interface{}) error
	MakePayment(ctx context.Context, input map[string]interface{}) error
}

type Service struct {
	usersRepository UsersRepository
}

func NewUsersService(usersRepository UsersRepository) *Service {
	return &Service{
		usersRepository: usersRepository,
	}
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

	return types.UserBasic{
		Id:      user.Id,
		Name:    user.Name,
		Email:   user.Email,
		Role:    user.Role,
		Balance: user.Balance,
	}, nil
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

	return types.UserWithAuthToken{
		Id:      user.Id,
		Name:    user.Name,
		Email:   user.Email,
		Balance: user.Balance,
		Role:    user.Role,
		Token:   token,
	}, nil
}

func (service *Service) GetLoggedInUser(ctx context.Context) (types.UserBasic, error) {
	userID, ok := ctx.Value(types.UserIDSessionKey).(int)
	if !ok {
		return types.UserBasic{}, errors.CustomError{
			Key: errors.Unauthorized,
			Err: goErrors.New("users id not found in context"),
		}
	}

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

	return types.UserBasic{
		Id:      user.Id,
		Name:    user.Name,
		Email:   user.Email,
		Balance: user.Balance,
		Role:    user.Role,
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

	randomPassword, err := generator.RandomPassword(8)
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	hashedPassword, err := hasher.Hash(randomPassword)
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

	//TODO: Send Email
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
