package models

type PingResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type EchoRequest struct {
	Message string `json:"message"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type EchoResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

//бд
type User struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	FullName string `json:"full_name"`
	IsAdmin  int    `json:"is_admin"`
}

type CreateUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	IsAdmin  int    `json:"is_admin"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
type UpdateUserRequest struct {
	Login    string `json:"login"`
	FullName string `json:"full_name"`
	IsAdmin  int    `json:"is_admin"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Login   string `json:"login"`
	IsAdmin int    `json:"is_admin"`
}

type ProfileResponse struct {
	ID      int64  `json:"id"`
	Login   string `json:"login"`
	IsAdmin int    `json:"is_admin"`
}

//бд
type Terminal struct {
	ID           int64  `json:"id"`
	SerialNumber string `json:"serial_number"`
	Name         string `json:"name"`
	Address      string `json:"address"`
}

type Key struct {
	ID       int64  `json:"id"`
	KeyValue string `json:"key_value"`
	KeyName  string `json:"key_name"`
}

type Card struct {
	ID         int64   `json:"id"`
	CardNumber string  `json:"card_number"`
	Balance    float64 `json:"balance"`
	IsBlocked  int     `json:"is_blocked"`
	OwnerName  string  `json:"owner_name"`
	KeyID      int64   `json:"key_id"`
}

type Transaction struct {
	ID         int64   `json:"id"`
	Amount     float64 `json:"amount"`
	CardID     int64   `json:"card_id"`
	TerminalID int64   `json:"terminal_id"`
	CreatedAt  string  `json:"created_at"`
}

//не бд

type TerminalAuthorizeRequest struct {
	TerminalSerial string  `json:"terminal_serial"`
	CardNumber     string  `json:"card_number"`
	Amount         float64 `json:"amount"`
}

type TerminalAuthorizeResponse struct {
	Approved bool    `json:"approved"`
	Message  string  `json:"message"`
	Balance  float64 `json:"balance"`
}

type TerminalKeysRequest struct {
	TerminalSerial string `json:"terminal_serial"`
}

type TerminalKeysResponse struct {
	TerminalSerial string `json:"terminal_serial"`
	Keys           []Key  `json:"keys"`
}

type CreateCardRequest struct {
	CardNumber string  `json:"card_number"`
	Balance    float64 `json:"balance"`
	IsBlocked  int     `json:"is_blocked"`
	OwnerName  string  `json:"owner_name"`
	KeyID      int64   `json:"key_id"`
}

type CardsResponse struct {
	Cards []Card `json:"cards"`
}

type UpdateCardRequest struct {
	CardNumber string  `json:"card_number"`
	Balance    float64 `json:"balance"`
	IsBlocked  int     `json:"is_blocked"`
	OwnerName  string  `json:"owner_name"`
	KeyID      int64   `json:"key_id"`
}

type CreateKeyRequest struct {
	KeyValue string `json:"key_value"`
	KeyName  string `json:"key_name"`
}

type KeysResponse struct {
	Keys []Key `json:"keys"`
}

type CreateTerminalRequest struct {
	SerialNumber string `json:"serial_number"`
	Name         string `json:"name"`
	Address      string `json:"address"`
}

type TerminalsResponse struct {
	Terminals []Terminal `json:"terminals"`
}
