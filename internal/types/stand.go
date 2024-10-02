package types

type Stand struct {
	Id          int    `json:"id" db:"id"`
	UserId      int    `json:"user_id" db:"user_id"`
	Name        string `json:"name" db:"name"`
	Category    string `json:"category" db:"category"`
	Stock       int    `json:"stock" db:"stock"`
	Price       int    `json:"price" db:"price"`
	Description string `json:"description" db:"description"`
}
