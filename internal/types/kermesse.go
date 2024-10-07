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

type KermesseWithStatistics struct {
	Id                   int    `json:"id" db:"id"`
	Name                 string `json:"name" db:"name"`
	UserId               int    `json:"user_id" db:"user_id"`
	Status               string `json:"status" db:"status"`
	Description          string `json:"description" db:"description"`
	UserNumber           int    `json:"user_number"`
	StandNumber          int    `json:"stand_number"`
	TombolaNumber        int    `json:"tombola_number"`
	TombolaBenefit       int    `json:"tombola_benefit"`
	ParticipationNumber  int    `json:"participation_number"`
	ParticipationBenefit int    `json:"participation_benefit"`
	Points               int    `json:"points"`
}

type KermesseStatistics struct {
	UserNumber           int `json:"user_number"`
	StandNumber          int `json:"stand_number"`
	TombolaNumber        int `json:"tombola_number"`
	TombolaBenefit       int `json:"tombola_benefit"`
	ParticipationNumber  int `json:"participation_number"`
	ParticipationBenefit int `json:"participation_benefit"`
	Points               int `json:"points"`
}
