package handler

import (
	"github.com/gorilla/mux"
	"github.com/kermesse-backend/api/middleware"
	"github.com/kermesse-backend/internal/tombolas"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/internal/users"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/json"
	"github.com/kermesse-backend/pkg/utils"
	"net/http"
	"strconv"
)

type TombolasHandler struct {
	tombolasService tombolas.TombolaService
	usersRepository users.UsersRepository
}

func NewTombolasHandler(tombolasService tombolas.TombolaService, usersRepository users.UsersRepository) *TombolasHandler {
	return &TombolasHandler{
		tombolasService: tombolasService,
		usersRepository: usersRepository,
	}
}

func (handler *TombolasHandler) RegisterRoutes(router *mux.Router) {
	router.Handle("/tombolas", errors.ErrorHandler(middleware.IsAuth(handler.GetAllTombolas, handler.usersRepository))).Methods(http.MethodGet)
	router.Handle("/tombolas", errors.ErrorHandler(middleware.IsAuth(handler.AddTombola, handler.usersRepository, types.UserRoleOrganizer))).Methods(http.MethodPost)
	router.Handle("/tombolas/{id}", errors.ErrorHandler(middleware.IsAuth(handler.GetTombolaById, handler.usersRepository))).Methods(http.MethodGet)
	router.Handle("/tombolas/{id}", errors.ErrorHandler(middleware.IsAuth(handler.ModifyTombola, handler.usersRepository, types.UserRoleOrganizer))).Methods(http.MethodPatch)
	router.Handle("/tombolas/{id}/finish-winner", errors.ErrorHandler(middleware.IsAuth(handler.FinishTombola, handler.usersRepository, types.UserRoleOrganizer))).Methods(http.MethodPatch)
}

func (handler *TombolasHandler) GetAllTombolas(w http.ResponseWriter, r *http.Request) error {
	tombolas, err := handler.tombolasService.GetAllTombolas(utils.GetParams(r))
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, tombolas); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (handler *TombolasHandler) GetTombolaById(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	tombola, err := handler.tombolasService.GetTombolaById(id)
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, tombola); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (handler *TombolasHandler) AddTombola(w http.ResponseWriter, r *http.Request) error {
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if err := handler.tombolasService.AddTombola(r.Context(), input); err != nil {
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

func (handler *TombolasHandler) ModifyTombola(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if err := handler.tombolasService.ModifyTombola(r.Context(), id, input); err != nil {
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

func (handler *TombolasHandler) FinishTombola(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if err := handler.tombolasService.FinishTombola(r.Context(), id); err != nil {
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
