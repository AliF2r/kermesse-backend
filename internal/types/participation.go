package types

const (
	ParticipationStatusStarted  string = "STARTED"
	ParticipationStatusFinished string = "FINISHED"
	ParticipationTypeGame       string = "GAME"
	ParticipationTypeFood       string = "FOOD"
)

type Participation struct {
	Id         int    `json:"id" db:"id"`
	KermesseId int    `json:"kermesse_id" db:"kermesse_id"`
	StandId    int    `json:"stand_id" db:"stand_id"`
	UserId     int    `json:"user_id" db:"user_id"`
	Category   string `json:"category" db:"category"`
	Balance    int    `json:"balance" db:"balance"`
	Point      int    `json:"point" db:"point"`
	Status     string `json:"status" db:"status"`
}

type ParticipatedUser struct {
	Id    int    `json:"id" db:"id"`
	Name  string `json:"name" db:"name"`
	Email string `json:"email" db:"email"`
	Role  string `json:"role" db:"role"`
}

type ParticipatedKermesse struct {
	Id          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Status      string `json:"status" db:"status"`
}

type ParticipatedStand struct {
	Id          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Type        string `json:"type" db:"type"`
	Price       int    `json:"price" db:"price"`
	Description string `json:"description" db:"description"`
}

type ParticipationCompleteModel struct {
	Id       int                  `json:"id" db:"id"`
	Type     string               `json:"type" db:"type"`
	Balance  int                  `json:"balance" db:"balance"`
	Point    int                  `json:"point" db:"point"`
	Status   string               `json:"status" db:"status"`
	User     ParticipatedUser     `json:"user" db:"user"`
	Kermesse ParticipatedKermesse `json:"kermesse" db:"kermesse"`
	Stand    ParticipatedStand    `json:"stand" db:"stand"`
}

type ParticipationUserStand struct {
	Id      int               `json:"id" db:"id"`
	Type    string            `json:"type" db:"type"`
	Balance int               `json:"balance" db:"balance"`
	Point   int               `json:"point" db:"point"`
	Status  string            `json:"status" db:"status"`
	User    ParticipatedUser  `json:"user" db:"user"`
	Stand   ParticipatedStand `json:"stand" db:"stand"`
}
