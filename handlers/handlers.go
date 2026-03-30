package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"transport-payments/auth"
	"transport-payments/db"
	"transport-payments/middleware"
	"transport-payments/models"
)

func writeJSON(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(data)
}

func PingHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
			Error: "method not allowed",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	response := models.PingResponse{ //здесь идёт заполнение json
		Message: "pong",
		Status:  "ok",
	}

	err := writeJSON(w, http.StatusOK, response)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
			Error: "method not allowed",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	var req models.EchoRequest
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid json",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.Message == "" {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "message is required",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	response := models.EchoResponse{
		Message: req.Message,
		Status:  "received",
	}

	err = writeJSON(w, http.StatusOK, response)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
			Error: "method not allowed",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	response := models.HealthResponse{ //здесь идёт заполнение json
		Status: "up",
	}

	err := writeJSON(w, http.StatusOK, response)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func UsersHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetUsers(w, r, database)
		case http.MethodPost:
			handleCreateUser(w, r, database)
		default:
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
		}
	}
}

// handleGetUsers godoc
// @Summary List users
// @Description Returns all users
// @Tags users
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} models.ErrorResponse
// @Router /users [get]
// @Security BearerAuth
func handleGetUsers(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	users, err := db.GetAllUsers(database)
	if err != nil {
		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to get users",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, users)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

// handleCreateUser godoc
// @Summary Create user
// @Description Creates a new user
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "Create user request"
// @Success 201 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [post]
// @Security BearerAuth
func handleCreateUser(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var req models.CreateUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid json",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.Login == "" || req.Password == "" || req.FullName == "" {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "login, password and full_name are required",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.IsAdmin != 0 && req.IsAdmin != 1 {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "is_admin must be 0 or 1",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	user, err := db.CreateUser(database, req)
	if err != nil {
		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to create user",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusCreated, user)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func UserByIDHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idPart := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
		if idPart == "" {
			err := writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "user id is required",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		id, err := strconv.ParseInt(idPart, 10, 64)
		if err != nil {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "invalid user id",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		if id <= 0 {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "user id must be positive",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleGetUserByID(w, r, database, id)

		case http.MethodPut:
			if !middleware.IsAdmin(r) {
				err := writeJSON(w, http.StatusForbidden, models.ErrorResponse{
					Error: "admin access required",
				})
				if err != nil {
					fmt.Println("write error:", err)
				}
				return
			}
			handleUpdateUserByID(w, r, database, id)

		case http.MethodDelete:
			if !middleware.IsAdmin(r) {
				err := writeJSON(w, http.StatusForbidden, models.ErrorResponse{
					Error: "admin access required",
				})
				if err != nil {
					fmt.Println("write error:", err)
				}
				return
			}
			handleDeleteUserByID(w, r, database, id)

		default:
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
		}
	}
}

// handleDeleteUserByID godoc
// @Summary Delete user
// @Description Deletes one user by id
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [delete]
// @Security BearerAuth
func handleDeleteUserByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	err := db.DeleteUserByID(database, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "user not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		if id == 1 {
			err = writeJSON(w, http.StatusForbidden, models.ErrorResponse{
				Error: "cannot delete main admin",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to delete user",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, models.MessageResponse{
		Message: "user deleted",
	})
	if err != nil {
		fmt.Println("write error:", err)
	}
}

// handleGetUserByID godoc
// @Summary Get user by ID
// @Description Returns one user by id
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [get]
// @Security BearerAuth
func handleGetUserByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	if !middleware.IsAdmin(r) {
		currentUserID, ok := middleware.GetUserID(r)
		if !ok {
			err := writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid token payload",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		if currentUserID != id {
			err := writeJSON(w, http.StatusForbidden, models.ErrorResponse{
				Error: "you can access only your own user",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}
	}

	user, err := db.GetUserByID(database, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "user not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to get user",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, user)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

// handleUpdateUserByID godoc
// @Summary Update user
// @Description Updates one user by id
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body models.UpdateUserRequest true "Update user request"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [put]
// @Security BearerAuth
func handleUpdateUserByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	var req models.UpdateUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid json",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.Login == "" || req.FullName == "" {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "login and full_name are required",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.IsAdmin != 0 && req.IsAdmin != 1 {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "is_admin must be 0 or 1",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if id == 1 && req.IsAdmin == 0 {
		err = writeJSON(w, http.StatusForbidden, models.ErrorResponse{
			Error: "main admin must remain admin",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	user, err := db.UpdateUserByID(database, id, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "user not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to update user",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, user)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

// LoginHandler godoc
// @Summary Login
// @Description Authenticates user and returns JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login request"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 405 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /login [post]
func LoginHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		var req models.LoginRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "invalid json",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		if req.Login == "" || req.Password == "" {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "login and password are required",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		user, err := db.GetUserByLoginAndPassword(database, req.Login, req.Password)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{
					Error: "invalid login or password",
				})
				if err != nil {
					fmt.Println("write error:", err)
				}
				return
			}

			err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
				Error: "failed to login",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		token, err := auth.GenerateToken(user)
		if err != nil {
			err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
				Error: "failed to generate token",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusOK, models.LoginResponse{
			Token:   token,
			Login:   req.Login,
			IsAdmin: user.IsAdmin,
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
	}
}

// ProfileHandler godoc
// @Summary Get profile
// @Description Returns current user from JWT
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ProfileResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 405 {object} models.ErrorResponse
// @Router /profile [get]
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
			Error: "method not allowed",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	claims, ok := middleware.GetClaims(r)
	if !ok {
		err := writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{
			Error: "missing claims in context",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		err := writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{
			Error: "invalid token payload",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	login, ok := claims["login"].(string)
	if !ok {
		err := writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{
			Error: "invalid token payload",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	isAdminFloat, ok := claims["is_admin"].(float64)
	if !ok {
		err := writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{
			Error: "invalid token payload",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	response := models.ProfileResponse{
		ID:      int64(userIDFloat),
		Login:   login,
		IsAdmin: int(isAdminFloat),
	}

	err := writeJSON(w, http.StatusOK, response)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func TerminalAuthorizeHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		var req models.TerminalAuthorizeRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "invalid json",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		if req.TerminalSerial == "" || req.CardNumber == "" || req.Amount <= 0 {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "terminal_serial, card_number and positive amount are required",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		response, err := db.AuthorizeTerminalPayment(database, req)
		if err != nil {
			err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
				Error: "failed to authorize payment",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusOK, response)
		if err != nil {
			fmt.Println("write error:", err)
		}
	}
}

func TerminalKeysHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		var req models.TerminalKeysRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "invalid json",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		if req.TerminalSerial == "" {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "terminal_serial is required",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		response, err := db.LoadKeysForTerminal(database, req.TerminalSerial)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
					Error: "terminal not found",
				})
				if err != nil {
					fmt.Println("write error:", err)
				}
				return
			}

			err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
				Error: "failed to load keys",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusOK, response)
		if err != nil {
			fmt.Println("write error:", err)
		}
	}
}

func CardsHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetCards(w, r, database)
		case http.MethodPost:
			handleCreateCard(w, r, database)
		default:
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
		}
	}
}

func handleGetCards(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	cards, err := db.GetAllCards(database)
	if err != nil {
		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to get cards",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, models.CardsResponse{
		Cards: cards,
	})
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func handleCreateCard(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var req models.CreateCardRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid json",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.CardNumber == "" || req.OwnerName == "" || req.KeyID <= 0 {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "card_number, owner_name and positive key_id are required",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.Balance < 0 {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "balance cannot be negative",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.IsBlocked != 0 && req.IsBlocked != 1 {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "is_blocked must be 0 or 1",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	_, err = db.GetKeyByID(database, req.KeyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "key not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to validate key",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	card, err := db.CreateCard(database, req)
	if err != nil {
		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to create card",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusCreated, card)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func CardByIDHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idPart := strings.TrimPrefix(r.URL.Path, "/api/v1/cards/")
		if idPart == "" {
			err := writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "card id is required",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		id, err := strconv.ParseInt(idPart, 10, 64)
		if err != nil {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "invalid card id",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		if id <= 0 {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "card id must be positive",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleGetCardByID(w, r, database, id)
		case http.MethodPut:
			handleUpdateCardByID(w, r, database, id)
		case http.MethodDelete:
			handleDeleteCardByID(w, r, database, id)
		default:
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
		}
	}
}

func handleGetCardByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	card, err := db.GetCardByID(database, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "card not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to get card",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, card)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func handleUpdateCardByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	var req models.UpdateCardRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid json",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.CardNumber == "" || req.OwnerName == "" || req.KeyID <= 0 {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "card_number, owner_name and positive key_id are required",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.Balance < 0 {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "balance cannot be negative",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.IsBlocked != 0 && req.IsBlocked != 1 {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "is_blocked must be 0 or 1",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	_, err = db.GetKeyByID(database, req.KeyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "key not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to validate key",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	card, err := db.UpdateCardByID(database, id, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "card not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to update card",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, card)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func handleDeleteCardByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	err := db.DeleteCardByID(database, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "card not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to delete card",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, models.MessageResponse{
		Message: "card deleted",
	})
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func KeysHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetKeys(w, r, database)
		case http.MethodPost:
			handleCreateKey(w, r, database)
		default:
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
		}
	}
}

func handleGetKeys(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	keys, err := db.GetAllKeys(database)
	if err != nil {
		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to get keys",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, models.KeysResponse{
		Keys: keys,
	})
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func handleCreateKey(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var req models.CreateKeyRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid json",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.KeyValue == "" || req.KeyName == "" {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "key_value and key_name are required",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	key, err := db.CreateKey(database, req)
	if err != nil {
		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to create key",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusCreated, key)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func TerminalsHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetTerminals(w, r, database)
		case http.MethodPost:
			handleCreateTerminal(w, r, database)
		default:
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
		}
	}
}

func handleGetTerminals(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	terminals, err := db.GetAllTerminals(database)
	if err != nil {
		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to get terminals",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, models.TerminalsResponse{
		Terminals: terminals,
	})
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func handleCreateTerminal(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var req models.CreateTerminalRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid json",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.SerialNumber == "" || req.Name == "" || req.Address == "" {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "serial_number, name and address are required",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	terminal, err := db.CreateTerminal(database, req)
	if err != nil {
		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to create terminal",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusCreated, terminal)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func TransactionsHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		transactions, err := db.GetAllTransactions(database)
		if err != nil {
			err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
				Error: "failed to get transactions",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusOK, models.TransactionsResponse{
			Transactions: transactions,
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
	}
}

func KeyByIDHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idPart := strings.TrimPrefix(r.URL.Path, "/api/v1/keys/")
		if idPart == "" {
			err := writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "key id is required",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		id, err := strconv.ParseInt(idPart, 10, 64)
		if err != nil {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "invalid key id",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		if id <= 0 {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "key id must be positive",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleGetKeyByID(w, r, database, id)
		case http.MethodPut:
			handleUpdateKeyByID(w, r, database, id)
		case http.MethodDelete:
			handleDeleteKeyByID(w, r, database, id)
		default:
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
		}
	}
}

func handleGetKeyByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	key, err := db.GetKeyByID(database, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "key not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to get key",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, key)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func handleUpdateKeyByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	var req models.UpdateKeyRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid json",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.KeyValue == "" || req.KeyName == "" {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "key_value and key_name are required",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	key, err := db.UpdateKeyByID(database, id, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "key not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to update key",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, key)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func handleDeleteKeyByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	usageCount, err := db.GetKeyUsageCount(database, id)
	if err != nil {
		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to check key usage",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if usageCount > 0 {
		err = writeJSON(w, http.StatusForbidden, models.ErrorResponse{
			Error: "cannot delete key that is used by cards",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = db.DeleteKeyByID(database, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "key not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to delete key",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, models.MessageResponse{
		Message: "key deleted",
	})
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func TerminalByIDHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idPart := strings.TrimPrefix(r.URL.Path, "/api/v1/terminals/")
		if idPart == "" {
			err := writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "terminal id is required",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		id, err := strconv.ParseInt(idPart, 10, 64)
		if err != nil {
			err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "invalid terminal id",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		if id <= 0 {
			err := writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
				Error: "terminal id must be positive",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleGetTerminalByID(w, r, database, id)
		case http.MethodPut:
			handleUpdateTerminalByID(w, r, database, id)
		case http.MethodDelete:
			handleDeleteTerminalByID(w, r, database, id)
		default:
			err := writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{
				Error: "method not allowed",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
		}
	}
}

func handleGetTerminalByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	terminal, err := db.GetTerminalByID(database, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "terminal not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to get terminal",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, terminal)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func handleUpdateTerminalByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	var req models.UpdateTerminalRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid json",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if req.SerialNumber == "" || req.Name == "" || req.Address == "" {
		err = writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: "serial_number, name and address are required",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	terminal, err := db.UpdateTerminalByID(database, id, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "terminal not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to update terminal",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, terminal)
	if err != nil {
		fmt.Println("write error:", err)
	}
}

func handleDeleteTerminalByID(w http.ResponseWriter, r *http.Request, database *sql.DB, id int64) {
	usageCount, err := db.GetTerminalUsageCount(database, id)
	if err != nil {
		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to check terminal usage",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	if usageCount > 0 {
		err = writeJSON(w, http.StatusForbidden, models.ErrorResponse{
			Error: "cannot delete terminal that is used by transactions",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = db.DeleteTerminalByID(database, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = writeJSON(w, http.StatusNotFound, models.ErrorResponse{
				Error: "terminal not found",
			})
			if err != nil {
				fmt.Println("write error:", err)
			}
			return
		}

		err = writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to delete terminal",
		})
		if err != nil {
			fmt.Println("write error:", err)
		}
		return
	}

	err = writeJSON(w, http.StatusOK, models.MessageResponse{
		Message: "terminal deleted",
	})
	if err != nil {
		fmt.Println("write error:", err)
	}
}
