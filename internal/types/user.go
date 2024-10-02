package types

type SessionKey string

const (
	UserIDSessionKey   SessionKey = "session_user_id"
	UserRoleSessionKey SessionKey = "session_user_role"
)

const (
	UserRoleParent      string = "PARENT"
	UserRoleStudent     string = "STUDENT"
	UserRoleOrganizer   string = "ORGANIZER"
	UserRoleStandHolder string = "STAND_HOLDER"
)

type User struct {
	Id       int    `json:"id" db:"id"`
	ParentId *int   `json:"parentId" db:"parent_id"`
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email"`
	Balance  int    `json:"balance" db:"balance"`
	Password string `json:"password" db:"password"`
	Role     string `json:"role" db:"role"`
}

type UserBasic struct {
	Id      int    `json:"id" db:"id"`
	Name    string `json:"name" db:"name"`
	Email   string `json:"email" db:"email"`
	Balance int    `json:"balance" db:"balance"`
	Role    string `json:"role" db:"role"`
}

type UserWithAuthToken struct {
	Id      int    `json:"id" db:"id"`
	Name    string `json:"name" db:"name"`
	Email   string `json:"email" db:"email"`
	Role    string `json:"role" db:"role"`
	Balance int    `json:"balance" db:"balance"`
	Token   string `json:"token"`
}
