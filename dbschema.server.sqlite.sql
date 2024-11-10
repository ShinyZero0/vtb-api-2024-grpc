PRAGMA foreign_keys = ON;

CREATE TABLE clients (
	client_id INTEGER PRIMARY KEY NOT NULL,
	client_secret VARCHAR(128) NOT NULL, -- hexencoded 256byte
	tls_value TEXT NOT NULL,
	tls_type VARCHAR(256) NOT NULL
);
CREATE TABLE tokens (
	expires_at INTEGER NOT NULL,
	value VARCHAR(128) PRIMARY KEY NOT NULL,
	refresh VARCHAR(128) NOT NULL
);
