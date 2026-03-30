// @title Transport Payments API
// @version 1.0
// @description REST API for transport card payment authorization
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"fmt"
	"net/http"

	_ "transport-payments/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"transport-payments/db"
	"transport-payments/handlers"
	"transport-payments/middleware"
)

func main() {
	database, err := db.Open()
	if err != nil {
		fmt.Println("database open error:", err)
		return
	}
	defer database.Close()

	err = db.InitSchema(database)
	if err != nil {
		fmt.Println("schema init error:", err)
		return
	}
	err = db.SeedData(database)
	if err != nil {
		fmt.Println("seed data error:", err)
		return
	}

	err = db.SeedUsers(database)
	if err != nil {
		fmt.Println("seed error:", err)
		return
	}

	fmt.Println("Database connected and users table is ready")

	http.HandleFunc("/api/v1/ping", handlers.PingHandler)
	http.HandleFunc("/api/v1/echo", handlers.EchoHandler)
	http.HandleFunc("/api/v1/health", handlers.HealthHandler)

	http.HandleFunc("/api/v1/users", middleware.RequireAuth(
		middleware.RequireAdmin(handlers.UsersHandler(database)),
	))

	http.HandleFunc("/api/v1/users/", middleware.RequireAuth(
		handlers.UserByIDHandler(database),
	))

	http.HandleFunc("/api/v1/login", handlers.LoginHandler(database))
	http.HandleFunc("/api/v1/profile", middleware.RequireAuth(handlers.ProfileHandler))

	http.Handle("/api/v1/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/api/v1/swagger/doc.json"),
	))

	http.HandleFunc("/api/v1/terminal/keys", handlers.TerminalKeysHandler(database))

	http.HandleFunc("/api/v1/terminal/authorize", handlers.TerminalAuthorizeHandler(database))

	http.HandleFunc("/api/v1/cards", middleware.RequireAuth(
		middleware.RequireAdmin(handlers.CardsHandler(database)),
	))

	http.HandleFunc("/api/v1/cards/", middleware.RequireAuth(
		middleware.RequireAdmin(handlers.CardByIDHandler(database)),
	))

	http.HandleFunc("/api/v1/keys", middleware.RequireAuth(
		middleware.RequireAdmin(handlers.KeysHandler(database)),
	))

	http.HandleFunc("/api/v1/terminals", middleware.RequireAuth(
		middleware.RequireAdmin(handlers.TerminalsHandler(database)),
	))

	http.HandleFunc("/api/v1/transactions", middleware.RequireAuth(
		middleware.RequireAdmin(handlers.TransactionsHandler(database)),
	))

	http.HandleFunc("/api/v1/keys/", middleware.RequireAuth(
		middleware.RequireAdmin(handlers.KeyByIDHandler(database)),
	))

	http.HandleFunc("/api/v1/terminals/", middleware.RequireAuth(
		middleware.RequireAdmin(handlers.TerminalByIDHandler(database)),
	))

	fmt.Println("Server started on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
