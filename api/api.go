package api

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/kermesse-backend/api/handler"
	"github.com/kermesse-backend/internal/kermesses"
	"github.com/kermesse-backend/internal/participations"
	"github.com/kermesse-backend/internal/stands"
	"github.com/kermesse-backend/internal/tickets"
	"github.com/kermesse-backend/internal/tombolas"
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

	standRepository := stands.NewStandsRepository(s.db)
	standService := stands.NewStandsService(standRepository)
	standHandler := handler.NewStandsHandler(standService, userRepository)
	standHandler.RegisterRoutes(router)

	kermesseRepository := kermesses.NewkermessesRepository(s.db)
	kermesseService := kermesses.NewKermessesService(kermesseRepository, userRepository)
	kermesseHandler := handler.NewKermessesHandler(kermesseService, userRepository)
	kermesseHandler.RegisterRoutes(router)

	participationRepository := participations.NewParticipationsRepository(s.db)
	participationService := participations.NewParticipationsService(userRepository, kermesseRepository, participationRepository, standRepository)
	participationHandler := handler.NewParticipationsHandler(participationService, userRepository)
	participationHandler.RegisterRoutes(router)

	tombolaRepository := tombolas.NewTombolasRepository(s.db)
	tombolaService := tombolas.NewTombolasService(tombolaRepository, kermesseRepository)
	tombolaHandler := handler.NewTombolasHandler(tombolaService, userRepository)
	tombolaHandler.RegisterRoutes(router)

	ticketRepository := tickets.NewTicketsRepository(s.db)
	ticketService := tickets.NewTicketsService(ticketRepository, tombolaRepository, userRepository, kermesseRepository)
	ticketHandler := handler.NewTicketsHandler(ticketService, userRepository)
	ticketHandler.RegisterRoutes(router)

	router.HandleFunc("/webhook", handler.HandleWebhook(userService)).Methods(http.MethodPost)

	websocketHandler := handler.NewWebSocketHandler()
	router.HandleFunc("/ws", websocketHandler.HandleWebSocket).Methods(http.MethodGet)

	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	log.Printf("🚀 Starting server on %s", s.address)
	return http.ListenAndServe(s.address, cors(router))
}
