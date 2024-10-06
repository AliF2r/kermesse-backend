package types

type Ticket struct {
	Id        int  `json:"id" db:"id"`
	UserId    int  `json:"user_id" db:"user_id"`
	TombolaId int  `json:"tombola_id" db:"tombola_id"`
	IsWinner  bool `json:"is_winner" db:"is_winner"`
}

type TicketUser struct {
	Id    int    `json:"id" db:"id"`
	Name  string `json:"name" db:"name"`
	Email string `json:"email" db:"email"`
	Role  string `json:"role" db:"role"`
}

type TicketTombola struct {
	Id     int    `json:"id" db:"id"`
	Name   string `json:"name" db:"name"`
	Status string `json:"status" db:"status"`
	Prize  string `json:"prize" db:"prize"`
	Price  int    `json:"price" db:"price"`
}

type TicketKermesse struct {
	Id          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Status      string `json:"status" db:"status"`
}

type TicketCompleteModel struct {
	Id       int            `json:"id" db:"id"`
	IsWinner bool           `json:"is_winner" db:"is_winner"`
	User     TicketUser     `json:"user" db:"user"`
	Tombola  TicketTombola  `json:"tombola" db:"tombola"`
	Kermesse TicketKermesse `json:"kermesse" db:"kermesse"`
}
