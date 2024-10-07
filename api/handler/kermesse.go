package handler

import (
	"github.com/gorilla/mux"
	"github.com/kermesse-backend/api/middleware"
	"github.com/kermesse-backend/internal/kermesses"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/internal/users"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/json"
	"net/http"
	"strconv"
)

type KermessesHandler struct {
	kermessesService kermesses.KermessesService
	usersRepository  users.UsersRepository
}

func NewKermessesHandler(kermessesService kermesses.KermessesService, usersRepository users.UsersRepository) *KermessesHandler {
	return &KermessesHandler{
		kermessesService: kermessesService,
		usersRepository:  usersRepository,
	}
}

func (handler *KermessesHandler) RegisterRoutes(router *mux.Router) {
	router.Handle("/kermesses", errors.ErrorHandler(middleware.IsAuth(handler.GetAllKermesses, handler.usersRepository))).Methods(http.MethodGet)
	router.Handle("/kermesses", errors.ErrorHandler(middleware.IsAuth(handler.CreateKermesse, handler.usersRepository, types.UserRoleOrganizer))).Methods(http.MethodPost)
	router.Handle("/kermesses/{id}", errors.ErrorHandler(middleware.IsAuth(handler.GetKermesseById, handler.usersRepository))).Methods(http.MethodGet)
	router.Handle("/kermesses/{id}", errors.ErrorHandler(middleware.IsAuth(handler.ModifyKermesse, handler.usersRepository, types.UserRoleOrganizer))).Methods(http.MethodPatch)
	router.Handle("/kermesses/{id}/complete", errors.ErrorHandler(middleware.IsAuth(handler.CompleteKermesse, handler.usersRepository, types.UserRoleOrganizer))).Methods(http.MethodPatch)
	router.Handle("/kermesses/{id}/add-user", errors.ErrorHandler(middleware.IsAuth(handler.AssignUserToKermesse, handler.usersRepository, types.UserRoleOrganizer))).Methods(http.MethodPatch)
	router.Handle("/kermesses/{id}/users", errors.ErrorHandler(middleware.IsAuth(handler.GetUsersForInvitation, handler.usersRepository))).Methods(http.MethodGet)
	router.Handle("/kermesses/{id}/add-stand", errors.ErrorHandler(middleware.IsAuth(handler.AssignStandToKermesse, handler.usersRepository, types.UserRoleOrganizer))).Methods(http.MethodPatch)
}

func (handler *KermessesHandler) GetAllKermesses(w http.ResponseWriter, r *http.Request) error {
	kermesses, err := handler.kermessesService.GetAllKermesses(r.Context())
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, kermesses); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (handler *KermessesHandler) CreateKermesse(w http.ResponseWriter, r *http.Request) error {
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if err := handler.kermessesService.AddKermesse(r.Context(), input); err != nil {
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

func (handler *KermessesHandler) GetKermesseById(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	kermesse, err := handler.kermessesService.GetKermesseById(r.Context(), id)
	if err != nil {
		return err
	}

	if err := json.Write(w, http.StatusOK, kermesse); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (handler *KermessesHandler) ModifyKermesse(w http.ResponseWriter, r *http.Request) error {
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
	if err := handler.kermessesService.UpdateKermesse(r.Context(), id, input); err != nil {
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

func (handler *KermessesHandler) CompleteKermesse(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if err := handler.kermessesService.MarkKermesseAsComplete(r.Context(), id); err != nil {
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

func (handler *KermessesHandler) AssignUserToKermesse(w http.ResponseWriter, r *http.Request) error {
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
	input["kermesse_id"] = id
	if err := handler.kermessesService.AssignUserToKermesse(r.Context(), input); err != nil {
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

func (handler *KermessesHandler) AssignStandToKermesse(w http.ResponseWriter, r *http.Request) error {
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
	input["kermesse_id"] = id
	if err := handler.kermessesService.AssignStandToKermesse(r.Context(), input); err != nil {
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

func (handler *KermessesHandler) GetUsersForInvitation(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	users, err := handler.kermessesService.GetUsersForInvitation(id)
	if err != nil {
		return err
	}

	if err := json.Write(w, http.StatusOK, users); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}
