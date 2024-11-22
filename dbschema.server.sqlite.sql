PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL2;
PRAGMA temp_store = MEMORY;
PRAGMA synchronous = NORMAL;

CREATE TABLE users (
	-- ub-common-name INTEGER ::= 64 https://www.rfc-editor.org/rfc/rfc5280
	user_id VARCHAR(64) PRIMARY KEY NOT NULL
);
CREATE TABLE roles (
	role_id INTEGER PRIMARY KEY NOT NULL,
	role_name VARCHAR(16) NOT NULL
);
CREATE TABLE user_roles (
	user_id VARCHAR(64) NOT NULL REFERENCES users(user_id),
	role_id INTEGER NOT NULL REFERENCES roles(role_id),
	PRIMARY KEY(user_id, role_id)
);
CREATE TABLE messages (
	message_id INTEGER PRIMARY KEY NOT NULL,
	user_id VARCHAR(64) NOT NULL REFERENCES users(user_id),
	content TEXT NOT NULL,
	created_at INTEGER NOT NULL
);
INSERT INTO roles VALUES(0, 'read');
INSERT INTO roles VALUES(1, 'write');
INSERT INTO roles VALUES(2, 'admin');
