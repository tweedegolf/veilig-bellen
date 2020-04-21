CREATE TABLE sessions (
	secret text,
	dtmf text UNIQUE NOT NULL,
	purpose text,
	disclosed text,
	status text,
	created timestamp NOT NULL DEFAULT now());
