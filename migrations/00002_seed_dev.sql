-- +goose up
INSERT INTO users (login, password, full_name, is_admin)
VALUES
    ('admin', 'admin123', 'Admin Cat', 1),
    ('onyx', 'meow123', 'Onyx Cat', 0),
    ('user1', 'user123', 'User One', 0);

INSERT INTO terminals (serial_number, name, address)
VALUES
    ('TERM-001', 'Terminal One', 'Bauman Street 1');

INSERT INTO keys (key_value, key_name)
VALUES
    ('KEY-ABC-123', 'Main transport key');

INSERT INTO cards (card_number, balance, is_blocked, owner_name, key_id)
VALUES
    ('CARD-001', 150.0, 0, 'Onyx Cat', 1),
    ('CARD-002', 20.0, 1, 'Blocked Cat', 1);

-- +goose down
DELETE FROM cards WHERE card_number IN ('CARD-001', 'CARD-002');
DELETE FROM keys WHERE key_value = 'KEY-ABC-123';
DELETE FROM terminals WHERE serial_number = 'TERM-001';
DELETE FROM users WHERE login IN ('admin', 'onyx', 'user1');