package types

const (
	KermesseStatusStarted  string = "STARTED"
	KermesseStatusFinished string = "FINISHED"
)

type Kermesse struct {
	Id          int    `json:"id" db:"id"`
	UserId      int    `json:"user_id" db:"user_id"`
	Name        string `json:"name" db:"name"`
	Status      string `json:"status" db:"status"`
	Description string `json:"description" db:"description"`
}
