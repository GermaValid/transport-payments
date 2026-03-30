-- +goose up
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    login TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    full_name TEXT NOT NULL,
    is_admin INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE terminals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    serial_number TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    address TEXT NOT NULL
);

CREATE TABLE keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_value TEXT NOT NULL UNIQUE,
    key_name TEXT NOT NULL
);

CREATE TABLE cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_number TEXT NOT NULL UNIQUE,
    balance REAL NOT NULL DEFAULT 0,
    is_blocked INTEGER NOT NULL DEFAULT 0,
    owner_name TEXT NOT NULL,
    key_id INTEGER NOT NULL,
    FOREIGN KEY (key_id) REFERENCES keys(id)
);

CREATE TABLE transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount REAL NOT NULL,
    card_id INTEGER NOT NULL,
    terminal_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (card_id) REFERENCES cards(id),
    FOREIGN KEY (terminal_id) REFERENCES terminals(id)
);

-- +goose down
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS keys;
DROP TABLE IF EXISTS terminals;
DROP TABLE IF EXISTS users;