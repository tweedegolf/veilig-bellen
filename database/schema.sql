CREATE TABLE sessions (
	secret text UNIQUE NOT NULL,
	dtmf text UNIQUE NOT NULL,
	disclosed text,
	created timestamp NOT NULL DEFAULT now(),
)
