package handler

import (
	"github.com/gorilla/mux"
	"github.com/kermesse-backend/api/middleware"
	"github.com/kermesse-backend/internal/participations"
	"github.com/kermesse-backend/internal/types"
	"github.com/kermesse-backend/internal/users"
	"github.com/kermesse-backend/pkg/errors"
	"github.com/kermesse-backend/pkg/json"
	"net/http"
	"strconv"
)

type ParticipationsHandler struct {
	participationService participations.ParticipationsService
	userRepository       users.UsersRepository
}

func NewParticipationsHandler(participationService participations.ParticipationsService, userRepository users.UsersRepository) *ParticipationsHandler {
	return &ParticipationsHandler{
		participationService: participationService,
		userRepository:       userRepository,
	}
}

func (handler *ParticipationsHandler) RegisterRoutes(router *mux.Router) {
	router.Handle("/participations", errors.ErrorHandler(middleware.IsAuth(handler.GetAllParticipations, handler.userRepository))).Methods(http.MethodGet)
	router.Handle("/participations", errors.ErrorHandler(middleware.IsAuth(handler.AddParticipation, handler.userRepository, types.UserRoleParent, types.UserRoleStudent))).Methods(http.MethodPost)
	router.Handle("/participations/{id}", errors.ErrorHandler(middleware.IsAuth(handler.GetParticipationById, handler.userRepository))).Methods(http.MethodGet)
	router.Handle("/participations/{id}", errors.ErrorHandler(middleware.IsAuth(handler.ModifyParticipation, handler.userRepository, types.UserRoleStandHolder))).Methods(http.MethodPatch)
}

func (handler *ParticipationsHandler) GetAllParticipations(w http.ResponseWriter, r *http.Request) error {
	participations, err := handler.participationService.GetAllParticipations()
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, participations); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	return nil
}

func (handler *ParticipationsHandler) GetParticipationById(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	participation, err := handler.participationService.GetParticipationById(id)
	if err != nil {
		return err
	}
	if err := json.Write(w, http.StatusOK, participation); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}

	return nil
}

func (handler *ParticipationsHandler) AddParticipation(w http.ResponseWriter, r *http.Request) error {
	var input map[string]interface{}
	if err := json.Parse(r, &input); err != nil {
		return errors.CustomError{
			Key: errors.InternalServerError,
			Err: err,
		}
	}
	if err := handler.participationService.AddParticipation(r.Context(), input); err != nil {
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

func (handler *ParticipationsHandler) ModifyParticipation(w http.ResponseWriter, r *http.Request) error {
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
	if err := handler.participationService.ModifyParticipation(r.Context(), id, input); err != nil {
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
