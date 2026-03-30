package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"transport-payments/models"
)

const dataSourceName = "db/transport.db"

func Open() (*sql.DB, error) {
	database, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	err = database.Ping()
	if err != nil {
		database.Close()
		return nil, err
	}

	return database, nil
}

func InitSchema(database *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		login TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		full_name TEXT NOT NULL,
		is_admin INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS terminals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		serial_number TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		address TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key_value TEXT NOT NULL UNIQUE,
		key_name TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		card_number TEXT NOT NULL UNIQUE,
		balance REAL NOT NULL DEFAULT 0,
		is_blocked INTEGER NOT NULL DEFAULT 0,
		owner_name TEXT NOT NULL,
		key_id INTEGER NOT NULL,
		FOREIGN KEY (key_id) REFERENCES keys(id)
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		amount REAL NOT NULL,
		card_id INTEGER NOT NULL,
		terminal_id INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY (card_id) REFERENCES cards(id),
		FOREIGN KEY (terminal_id) REFERENCES terminals(id)
	);
	`

	_, err := database.Exec(query)
	return err
}

func GetAllUsers(database *sql.DB) ([]models.User, error) {
	query := `
	SELECT id, login, full_name, is_admin
	FROM users
	ORDER BY id;
	`

	rows, err := database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)

	for rows.Next() {
		var user models.User

		err = rows.Scan(
			&user.ID,
			&user.Login,
			&user.FullName,
			&user.IsAdmin,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func CreateUser(database *sql.DB, req models.CreateUserRequest) (models.User, error) {
	query := `
	INSERT INTO users (login, password, full_name, is_admin)
	VALUES (?, ?, ?, ?);
	`

	result, err := database.Exec(query, req.Login, req.Password, req.FullName, req.IsAdmin)
	if err != nil {
		return models.User{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return models.User{}, err
	}

	user := models.User{
		ID:       id,
		Login:    req.Login,
		FullName: req.FullName,
		IsAdmin:  req.IsAdmin,
	}

	return user, nil
}

func SeedUsers(database *sql.DB) error {
	query := `
	INSERT OR IGNORE INTO users (login, password, full_name, is_admin)
	VALUES
		('admin', 'admin123', 'Admin Cat', 1),
		('onyx', 'meow123', 'Onyx Cat', 0);
	`

	_, err := database.Exec(query)
	return err
}

func GetUserByID(database *sql.DB, id int64) (models.User, error) {
	query := `
	SELECT id, login, full_name, is_admin
	FROM users
	WHERE id = ?;
	`

	var user models.User

	err := database.QueryRow(query, id).Scan(
		&user.ID,
		&user.Login,
		&user.FullName,
		&user.IsAdmin,
	)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func DeleteUserByID(database *sql.DB, id int64) error {
	query := `
	DELETE FROM users
	WHERE id = ?;
	`

	result, err := database.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func UpdateUserByID(database *sql.DB, id int64, req models.UpdateUserRequest) (models.User, error) {
	query := `
	UPDATE users
	SET login = ?, full_name = ?, is_admin = ?
	WHERE id = ?;
	`

	result, err := database.Exec(query, req.Login, req.FullName, req.IsAdmin, id)
	if err != nil {
		return models.User{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.User{}, err
	}

	if rowsAffected == 0 {
		return models.User{}, sql.ErrNoRows
	}

	user := models.User{
		ID:       id,
		Login:    req.Login,
		FullName: req.FullName,
		IsAdmin:  req.IsAdmin,
	}

	return user, nil
}

func GetUserByLoginAndPassword(database *sql.DB, login string, password string) (models.User, error) {
	query := `
	SELECT id, login, full_name, is_admin
	FROM users
	WHERE login = ? AND password = ?;
	`

	var user models.User

	err := database.QueryRow(query, login, password).Scan(
		&user.ID,
		&user.Login,
		&user.FullName,
		&user.IsAdmin,
	)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func SeedData(database *sql.DB) error {
	query := `
	INSERT OR IGNORE INTO users (login, password, full_name, is_admin)
	VALUES
		('admin', 'admin123', 'Admin Cat', 1),
		('user1', 'user123', 'User One', 0);

	INSERT OR IGNORE INTO terminals (serial_number, name, address)
	VALUES
		('TERM-001', 'Terminal One', 'Bauman Street 1');

	INSERT OR IGNORE INTO keys (key_value, key_name)
	VALUES
		('KEY-ABC-123', 'Main transport key');

	INSERT OR IGNORE INTO cards (card_number, balance, is_blocked, owner_name, key_id)
	VALUES
		('CARD-001', 150.0, 0, 'Onyx Cat', 1),
		('CARD-002', 20.0, 1, 'Blocked Cat', 1);
	`

	_, err := database.Exec(query)
	return err
}

func GetTerminalBySerial(database *sql.DB, serial string) (models.Terminal, error) {
	query := `
	SELECT id, serial_number, name, address
	FROM terminals
	WHERE serial_number = ?;
	`

	var terminal models.Terminal

	err := database.QueryRow(query, serial).Scan(
		&terminal.ID,
		&terminal.SerialNumber,
		&terminal.Name,
		&terminal.Address,
	)
	if err != nil {
		return models.Terminal{}, err
	}

	return terminal, nil
}

func GetCardByNumber(database *sql.DB, cardNumber string) (models.Card, error) {
	query := `
	SELECT id, card_number, balance, is_blocked, owner_name, key_id
	FROM cards
	WHERE card_number = ?;
	`

	var card models.Card

	err := database.QueryRow(query, cardNumber).Scan(
		&card.ID,
		&card.CardNumber,
		&card.Balance,
		&card.IsBlocked,
		&card.OwnerName,
		&card.KeyID,
	)
	if err != nil {
		return models.Card{}, err
	}

	return card, nil
}

func AuthorizeTerminalPayment(database *sql.DB, req models.TerminalAuthorizeRequest) (models.TerminalAuthorizeResponse, error) {
	terminal, err := GetTerminalBySerial(database, req.TerminalSerial)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.TerminalAuthorizeResponse{
				Approved: false,
				Message:  "terminal not found",
				Balance:  0,
			}, nil
		}
		return models.TerminalAuthorizeResponse{}, err
	}

	card, err := GetCardByNumber(database, req.CardNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.TerminalAuthorizeResponse{
				Approved: false,
				Message:  "card not found",
				Balance:  0,
			}, nil
		}
		return models.TerminalAuthorizeResponse{}, err
	}

	if card.IsBlocked == 1 {
		return models.TerminalAuthorizeResponse{
			Approved: false,
			Message:  "card is blocked",
			Balance:  card.Balance,
		}, nil
	}

	if card.Balance < req.Amount {
		return models.TerminalAuthorizeResponse{
			Approved: false,
			Message:  "insufficient funds",
			Balance:  card.Balance,
		}, nil
	}

	tx, err := database.Begin()
	if err != nil {
		return models.TerminalAuthorizeResponse{}, err
	}
	defer tx.Rollback()

	newBalance := card.Balance - req.Amount

	_, err = tx.Exec(`
		UPDATE cards
		SET balance = ?
		WHERE id = ?;
	`, newBalance, card.ID)
	if err != nil {
		return models.TerminalAuthorizeResponse{}, err
	}

	_, err = tx.Exec(`
		INSERT INTO transactions (amount, card_id, terminal_id, created_at)
		VALUES (?, ?, ?, ?);
	`, req.Amount, card.ID, terminal.ID, time.Now().Format(time.RFC3339))
	if err != nil {
		return models.TerminalAuthorizeResponse{}, err
	}

	err = tx.Commit()
	if err != nil {
		return models.TerminalAuthorizeResponse{}, err
	}

	return models.TerminalAuthorizeResponse{
		Approved: true,
		Message:  "payment approved",
		Balance:  newBalance,
	}, nil
}

func GetAllKeys(database *sql.DB) ([]models.Key, error) {
	query := `
	SELECT id, key_value, key_name
	FROM keys
	ORDER BY id;
	`

	rows, err := database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]models.Key, 0)

	for rows.Next() {
		var key models.Key

		err = rows.Scan(
			&key.ID,
			&key.KeyValue,
			&key.KeyName,
		)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func LoadKeysForTerminal(database *sql.DB, terminalSerial string) (models.TerminalKeysResponse, error) {
	terminal, err := GetTerminalBySerial(database, terminalSerial)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.TerminalKeysResponse{}, sql.ErrNoRows
		}
		return models.TerminalKeysResponse{}, err
	}

	keys, err := GetAllKeys(database)
	if err != nil {
		return models.TerminalKeysResponse{}, err
	}

	return models.TerminalKeysResponse{
		TerminalSerial: terminal.SerialNumber,
		Keys:           keys,
	}, nil
}

func GetAllCards(database *sql.DB) ([]models.Card, error) {
	query := `
	SELECT id, card_number, balance, is_blocked, owner_name, key_id
	FROM cards
	ORDER BY id;
	`

	rows, err := database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := make([]models.Card, 0)

	for rows.Next() {
		var card models.Card

		err = rows.Scan(
			&card.ID,
			&card.CardNumber,
			&card.Balance,
			&card.IsBlocked,
			&card.OwnerName,
			&card.KeyID,
		)
		if err != nil {
			return nil, err
		}

		cards = append(cards, card)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return cards, nil
}

func CreateCard(database *sql.DB, req models.CreateCardRequest) (models.Card, error) {
	query := `
	INSERT INTO cards (card_number, balance, is_blocked, owner_name, key_id)
	VALUES (?, ?, ?, ?, ?);
	`

	result, err := database.Exec(
		query,
		req.CardNumber,
		req.Balance,
		req.IsBlocked,
		req.OwnerName,
		req.KeyID,
	)
	if err != nil {
		return models.Card{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return models.Card{}, err
	}

	card := models.Card{
		ID:         id,
		CardNumber: req.CardNumber,
		Balance:    req.Balance,
		IsBlocked:  req.IsBlocked,
		OwnerName:  req.OwnerName,
		KeyID:      req.KeyID,
	}

	return card, nil
}
