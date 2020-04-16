CREATE TABLE backends (
	-- Identifies the backend node.
	backend_id text PRIMARY KEY,
	-- The last moment the backend reported having made progress.
	last_seen timestamp NOT NULL DEFAULT now());

-- Each session corresponds to a single IRMA session and is identified by the
-- session secret, or by the DTMF code before the session secret is available.
CREATE TABLE sessions (
	-- Secret that grants permission to retrieve the disclosed attributes.
	-- This is momentarily null while it is retrieved from the IRMA server.
	secret text,
	-- DTMF code used to connect the corresponding call.
	dtmf text UNIQUE NOT NULL,
	-- Identity of the backend responsible for any polling needs.
	backend_id text REFERENCES backends,
	-- Identifier for the call purpose of this session.
	purpose text,
	-- The IRMA attributes that were disclosed.
	disclosed text,
	-- The status of the session.
	status text,
	-- The moment this session was created.
	created timestamp NOT NULL DEFAULT now());
