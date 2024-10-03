package types

const (
	TombolaStatusStarted  = "STARTED"
	TombolaStatusFinished = "FINISHED"
)

type Tombola struct {
	Id         int    `json:"id" db:"id"`
	KermesseId int    `json:"kermesse_id" db:"kermesse_id"`
	Prize      string `json:"prize" db:"prize"`
	Name       string `json:"name" db:"name"`
	Price      int    `json:"price" db:"price"`
	Status     string `json:"status" db:"status"`
}
