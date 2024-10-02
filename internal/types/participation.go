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
