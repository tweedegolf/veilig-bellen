CREATE TABLE sessions (
	secret text,
	dtmf text UNIQUE NOT NULL,
	purpose text,
	disclosed text,
	irma_status text,
	created timestamp NOT NULL DEFAULT now());
