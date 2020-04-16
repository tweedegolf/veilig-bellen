package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/lib/pq"
)

var ErrNoRows = sql.ErrNoRows

type Database struct {
	db              *sql.DB
	listener        *pq.Listener
	backendIdentity string
}

func (db Database) NewSession(purpose string) (DTMF, error) {
	var err error
	for attempt := 0; attempt < 10; attempt++ {
		var n *big.Int
		n, err = rand.Int(rand.Reader, big.NewInt(10_000_000_000))
		if err != nil {
			return "", err
		}

		dtmf := fmt.Sprintf("%010d", n)
		if err != nil {
			err = fmt.Errorf("failed to generate secrets: %w", err)
			return "", err
		}

		_, err = db.db.Exec("INSERT INTO sessions VALUES (NULL, $1, $2, $3, DEFAULT, DEFAULT, DEFAULT)", dtmf, db.backendIdentity, purpose)
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code.Name() == "unique_violation" {
			time.Sleep(100 * time.Millisecond)
			continue
		} else if err != nil {
			err = fmt.Errorf("failed to store session: %w", err)
			return "", err
		}

		return dtmf, nil
	}
	err = fmt.Errorf("failed to find unique secrets: %w", err)
	return "", err
}

func (db Database) storeSecret(dtmf string, secret string) error {
	_, err := db.db.Exec("UPDATE sessions SET secret = $1 WHERE dtmf = $2", secret, dtmf)
	return err
}

func (db Database) secretFromDTMF(dtmf string) (string, error) {
	var secret string
	row := db.db.QueryRow("SELECT secret FROM sessions WHERE dtmf = $1", dtmf)
	err := row.Scan(&secret)
	return secret, err
}

func (db Database) storeDisclosed(secret string, disclosed string) error {
	_, err := db.db.Exec("UPDATE sessions SET disclosed = $1 WHERE secret = $2", disclosed, secret)
	return err
}

func (db Database) setStatus(secret string, status string) error {
	_, err := db.db.Exec("UPDATE sessions SET status = $1 WHERE secret = $2", status, secret)
	db.Notify(secret, "status", status)
	return err
}

func (db Database) getStatus(secret string) (string, error) {
	var status string
	row := db.db.QueryRow("SELECT status FROM sessions WHERE secret = $1", secret)
	err := row.Scan(&status)
	return status, err
}

func (db Database) getDisclosed(secret string) (purpose string, disclosed string, err error) {
	row := db.db.QueryRow("SELECT purpose, disclosed FROM sessions WHERE secret = $1", secret)
	err = row.Scan(&purpose, &disclosed)
	return purpose, disclosed, err
}

func (db Database) activeSessionCount() (int, error) {
	var res int
	row := db.db.QueryRow(`
		SELECT COUNT(*) AS count
		FROM sessions
		WHERE status NOT IN ('IRMA-UNREACHABLE', 'IRMA-TIMEOUT', 'IRMA-CANCELLED', 'DONE')
		AND status IS NOT NULL
	`)
	err := row.Scan(&res)
	if err != nil {
		return -1, err
	}
	return res, nil
}

func (db Database) expire() error {
	_, err := db.db.Exec("DELETE FROM sessions WHERE created < now() - '1 hour'::interval")
	return err
}

// Check the database for any orphaned sessions (caretaker not making progress
// in the last three seconds) and adopt some of them. Return the secrets
// identifying the adopted sessions. Also send a notification for each secret
// with an updated value for its backend_id. Should be called immediately after
// Keepalive to prevent it from adopting orphans from itself.
func (db Database) AdoptOrphans() ([]string, error) {
	// The current implementation adopts all orphaned sessions.
	rows, err := db.db.Query(`
		UPDATE sessions
		SET backend_id = $1
		FROM backends
		WHERE sessions.backend_id = backends.backend_id
		AND backends.last_seen < (now() - interval '3 seconds')
		RETURNING sessions.secret`,
		db.backendIdentity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []string
	for rows.Next() {
		var secret string
		err := rows.Scan(&secret)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, secret)
		db.Notify(secret, "backend_id", db.backendIdentity)
	}

	return secrets, nil
}

func (db Database) Notify(channel, key, value string) error {
	message := key + " " + value
	row := db.db.QueryRow("SELECT pg_notify($1, $2)", channel, message)
	var ignored string
	err := row.Scan(&ignored)
	return err
}

// Inform cluster that this node is still up. Should be called from the main
// action loop such that it is not called when no progress is being made.
func (db Database) Keepalive() error {
	_, err := db.db.Exec(`
		INSERT INTO backends (backend_id, last_seen)
		VALUES ($1, now())
		ON CONFLICT (backend_id) DO UPDATE
			SET last_seen = now()`, db.backendIdentity)
	return err
}
