package handler

import (
	"github.com/kermesse-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kermesse-backend/api/middleware"
	"github.com/kermesse-backend/internal/stands"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/internal/users"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/json"
)

type StandsHandler struct {
	standService    stands.StandsService
	usersRepository users.UsersRepository
}

func NewStandsHandler(standService stands.StandsService, usersRepository users.UsersRepository) *StandsHandler {
	return &StandsHandler{
		standService:    standService,
		usersRepository: usersRepository,
	}
}

func (handler *StandsHandler) RegisterRoutes(router *mux.Router) {
	router.Handle("/stands", errors.ErrorHandler(middleware.IsAuth(handler.AddStand, handler.usersRepository, types.UserRoleStandHolder))).Methods(http.MethodPost)
	router.Handle("/stands", errors.ErrorHandler(middleware.IsAuth(handler.GetAllStands, handler.usersRepository))).Methods(http.MethodGet)
	router.Handle("/stands/owner", errors.ErrorHandler(middleware.IsAuth(handler.GetOwnStand, handler.usersRepository, types.UserRoleStandHolder))).Methods(http.MethodGet)
	router.Handle("/stands/{id}", errors.ErrorHandler(middleware.IsAuth(handler.GetStandById, handler.usersRepository))).Methods(http.MethodGet)
	router.Handle("/stands/modify", errors.ErrorHandler(middleware.IsAuth(handler.ModifyStand, handler.usersRepository, types.UserRoleStandHolder))).Methods(http.MethodPatch)
}

func (handler *StandsHandler) AddStand(w http.ResponseWriter, r *http.Request) error {
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if err := handler.standService.AddStand(r.Context(), input); err != nil {
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

func (handler *StandsHandler) GetAllStands(w http.ResponseWriter, r *http.Request) error {
	stands, err := handler.standService.GetAllStands(utils.GetParams(r))
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, stands); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (handler *StandsHandler) GetStandById(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	stand, err := handler.standService.GetStandById(id)
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, stand); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (handler *StandsHandler) ModifyStand(w http.ResponseWriter, r *http.Request) error {
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	if err := handler.standService.ModifyStand(r.Context(), input); err != nil {
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

func (handler *StandsHandler) GetOwnStand(w http.ResponseWriter, r *http.Request) error {
	stand, err := handler.standService.GetOwnStand(r.Context())
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, stand); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}
