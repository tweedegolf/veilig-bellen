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

// Check the database for any orphaned feeds (caretaker not making progress in
// the last three seconds) and adopt some of them. Return the feed_ids
// identifying the adopted feeds.
func (db Database) AdoptOrphans() ([]string, error) {
	// The current implementation adopts all orphaned feeds.
	rows, err := db.db.Query(`
		UPDATE feeds
		SET backend_id = $1, last_polled = now()
		WHERE backend_id != $1
		AND last_polled < (now() - interval '3 seconds')
		RETURNING feed_id
		`, db.backendIdentity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []string
	for rows.Next() {
		var feed string
		err := rows.Scan(&feed)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}

	return feeds, nil
}

func (db Database) Notify(channel, key, value string) error {
	message := key + " " + value
	row := db.db.QueryRow("SELECT pg_notify($1, $2)", channel, message)
	var ignored string
	err := row.Scan(&ignored)
	return err
}

func (db Database) NewFeed(feed_id string) error {
	_, err := db.db.Exec("INSERT INTO feeds VALUES ($1, $2, now())", feed_id, db.backendIdentity)
	return err
}

// Tells the database a feed is still being polled. Returns true if the current
// backend is still responsible for polling the feed.
func (db Database) TouchFeed(feed_id string) (bool, error) {
	result, err := db.db.Exec(`
		UPDATE feeds
		SET last_polled = now()
		WHERE feed_id = $1
		AND backend_id = $2`, feed_id, db.backendIdentity)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
