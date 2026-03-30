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

// PingHandler godoc
// @Summary Ping
// @Description Health check ping endpoint
// @Tags system
// @Produce json
// @Success 200 {object} models.PingResponse
// @Failure 405 {object} models.ErrorResponse
// @Router /ping [get]
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

// EchoHandler godoc
// @Summary Echo
// @Description Echoes message back to client
// @Tags system
// @Accept json
// @Produce json
// @Param request body models.EchoRequest true "Echo request"
// @Success 200 {object} models.EchoResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 405 {object} models.ErrorResponse
// @Router /echo [post]
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

// HealthHandler godoc
// @Summary Health
// @Description Returns service health status
// @Tags system
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Failure 405 {object} models.ErrorResponse
// @Router /health [get]
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

// TerminalAuthorizeHandler godoc
// @Summary Authorize terminal payment
// @Description Checks card, balance and authorizes payment transaction
// @Tags terminal
// @Accept json
// @Produce json
// @Param request body models.TerminalAuthorizeRequest true "Terminal authorize request"
// @Success 200 {object} models.TerminalAuthorizeResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 405 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /terminal/authorize [post]
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

// TerminalKeysHandler godoc
// @Summary Load keys for terminal
// @Description Returns all available keys for terminal card decryption
// @Tags terminal
// @Accept json
// @Produce json
// @Param request body models.TerminalKeysRequest true "Terminal keys request"
// @Success 200 {object} models.TerminalKeysResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 405 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /terminal/keys [post]
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

// CardsHandler godoc
// @Summary List cards
// @Description Returns all transport cards
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.CardsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 405 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cards [get]
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

// handleGetCards godoc
// @Summary List cards
// @Description Returns all transport cards
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.CardsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cards [get]
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

// handleCreateCard godoc
// @Summary Create card
// @Description Creates a new transport card
// @Tags cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateCardRequest true "Create card request"
// @Success 201 {object} models.Card
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cards [post]
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

// handleGetCardByID godoc
// @Summary Get card by ID
// @Description Returns one transport card by id
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 200 {object} models.Card
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cards/{id} [get]
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

// handleUpdateCardByID godoc
// @Summary Update card
// @Description Updates transport card by id
// @Tags cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Param request body models.UpdateCardRequest true "Update card request"
// @Success 200 {object} models.Card
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cards/{id} [put]
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

// handleDeleteCardByID godoc
// @Summary Delete card
// @Description Deletes transport card by id
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param id path int true "Card ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cards/{id} [delete]
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

// handleGetKeys godoc
// @Summary List keys
// @Description Returns all card keys
// @Tags keys
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.KeysResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /keys [get]
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

// handleCreateKey godoc
// @Summary Create key
// @Description Creates a new card key
// @Tags keys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateKeyRequest true "Create key request"
// @Success 201 {object} models.Key
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /keys [post]
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

// handleGetTerminals godoc
// @Summary List terminals
// @Description Returns all terminals
// @Tags terminals
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.TerminalsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /terminals [get]
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

// handleCreateTerminal godoc
// @Summary Create terminal
// @Description Creates a new terminal
// @Tags terminals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateTerminalRequest true "Create terminal request"
// @Success 201 {object} models.Terminal
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /terminals [post]
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

// TransactionsHandler godoc
// @Summary List transactions
// @Description Returns all transactions ordered from newest to oldest
// @Tags transactions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.TransactionsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 405 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /transactions [get]
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

// handleGetKeyByID godoc
// @Summary Get key by ID
// @Description Returns one key by id
// @Tags keys
// @Produce json
// @Security BearerAuth
// @Param id path int true "Key ID"
// @Success 200 {object} models.Key
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /keys/{id} [get]
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

// handleUpdateKeyByID godoc
// @Summary Update key
// @Description Updates key by id
// @Tags keys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Key ID"
// @Param request body models.UpdateKeyRequest true "Update key request"
// @Success 200 {object} models.Key
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /keys/{id} [put]
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

// handleDeleteKeyByID godoc
// @Summary Delete key
// @Description Deletes key by id if it is not used by cards
// @Tags keys
// @Produce json
// @Security BearerAuth
// @Param id path int true "Key ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /keys/{id} [delete]
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

// handleGetTerminalByID godoc
// @Summary Get terminal by ID
// @Description Returns one terminal by id
// @Tags terminals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Terminal ID"
// @Success 200 {object} models.Terminal
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /terminals/{id} [get]
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

// handleUpdateTerminalByID godoc
// @Summary Update terminal
// @Description Updates terminal by id
// @Tags terminals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Terminal ID"
// @Param request body models.UpdateTerminalRequest true "Update terminal request"
// @Success 200 {object} models.Terminal
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /terminals/{id} [put]
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

// handleDeleteTerminalByID godoc
// @Summary Delete terminal
// @Description Deletes terminal by id if it is not used by transactions
// @Tags terminals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Terminal ID"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /terminals/{id} [delete]
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
