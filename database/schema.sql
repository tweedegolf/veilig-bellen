-- Each session corresponds to a single IRMA session and is identified by the
-- session secret, or by the DTMF code before the session secret is available.
CREATE TABLE sessions (
	-- Secret that grants permission to retrieve the disclosed attributes.
	-- This is momentarily null while it is retrieved from the IRMA server.
	secret text,
	-- DTMF code used to connect the corresponding call.
	dtmf text UNIQUE NOT NULL,
	-- Identifier for the call purpose of this session.
	purpose text,
	-- The IRMA attributes that were disclosed.
	disclosed text,
	-- The status of the session.
	status text,
	-- The moment this session was created.
	created timestamp NOT NULL DEFAULT now());

-- Each row defines the backend responsible for polling a single feed.
CREATE TABLE feeds (
	-- Identity of the feed to be polled and the name of the corresponding
	-- postgres channel: "kcc" or a session secret.
	feed_id text PRIMARY KEY,
	-- Identity of the backend responsible for polling.
	backend_id text,
	-- When the last poll was started.
	last_polled timestamp NOT NULL DEFAULT now());
