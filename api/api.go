package api

import (
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/api/handler"
	"github.com/kermesse-backend/internal/users"
	"log"
	"net/http"
)

type APIServer struct {
	address string
	db      *sqlx.DB
}

func NewAPIServer(address string, db *sqlx.DB) *APIServer {
	return &APIServer{
		address: address,
		db:      db,
	}
}

func (s *APIServer) Start() error {
	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	userRepository := users.NewUsersRepository(s.db)
	userService := users.NewUsersService(userRepository)
	userHandler := handler.NewUserHandler(userService, userRepository)
	userHandler.RegisterRoutes(router)

	log.Printf("🚀 Starting server on %s", s.address)
	return http.ListenAndServe(s.address, router)
}
