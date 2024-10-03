package handler

import (
	"github.com/gorilla/mux"
	"github.com/kermesse-backend/api/middleware"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/internal/users"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/json"
	"net/http"
	"strconv"
)

type UsersHandler struct {
	userService    users.UsersService
	userRepository users.UsersRepository
}

func NewUserHandler(userService users.UsersService, userRepository users.UsersRepository) *UsersHandler {
	return &UsersHandler{
		userService:    userService,
		userRepository: userRepository,
	}
}

func (handler *UsersHandler) RegisterRoutes(mux *mux.Router) {
	mux.Handle("/register", errors.ErrorHandler(handler.Register)).Methods(http.MethodPost)
	mux.Handle("/login", errors.ErrorHandler(handler.Login)).Methods(http.MethodPost)
	mux.Handle("/me", errors.ErrorHandler(middleware.IsAuth(handler.GetLoggedInUser, handler.userRepository))).Methods(http.MethodGet)
	mux.Handle("/users/{id}", errors.ErrorHandler(middleware.IsAuth(handler.GetUserById, handler.userRepository))).Methods(http.MethodGet)
	mux.Handle("/users/invite", errors.ErrorHandler(middleware.IsAuth(handler.InviteStudent, handler.userRepository))).Methods(http.MethodPost)
	mux.Handle("/users/send-jeton", errors.ErrorHandler(middleware.IsAuth(handler.MakePayment, handler.userRepository, types.UserRoleParent))).Methods(http.MethodPatch)
}

func (handler *UsersHandler) GetUserById(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	user, err := handler.userService.GetUserById(id)
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, user); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (handler *UsersHandler) InviteStudent(w http.ResponseWriter, r *http.Request) error {
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if err := handler.userService.InviteStudent(r.Context(), input); err != nil {
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

func (handler *UsersHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if err := handler.userService.Register(input); err != nil {
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

func (handler *UsersHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	response, err := handler.userService.Login(input)
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, response); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (handler *UsersHandler) GetLoggedInUser(w http.ResponseWriter, r *http.Request) error {
	response, err := handler.userService.GetLoggedInUser(r.Context())
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, response); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (handler *UsersHandler) MakePayment(w http.ResponseWriter, r *http.Request) error {
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if err := handler.userService.MakePayment(r.Context(), input); err != nil {
		return err
	}
	if err := json.Write(w, http.StatusAccepted, nil); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}
