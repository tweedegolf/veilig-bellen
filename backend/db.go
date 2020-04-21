package main

import "crypto/rand"
import "database/sql"
import "fmt"
import "math/big"
import "time"
import "github.com/lib/pq"


var ErrNoRows = sql.ErrNoRows

type Database struct {
	db *sql.DB
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

		_, err = db.db.Exec("INSERT INTO sessions VALUES (NULL, $1, $2, DEFAULT, DEFAULT)", dtmf, purpose)
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

func (db Database) getDisclosed(secret string) (purpose string, disclosed string, err error) {
	row := db.db.QueryRow("SELECT purpose, disclosed FROM sessions WHERE secret = $1", secret)
	err = row.Scan(&purpose, &disclosed)
	return purpose, disclosed, err
}

func (db Database) updateSessionStatus(secret string, status string) error {
	_, err := db.db.Exec("UPDATE sessions SET status = $1 WHERE secret = $2", status, secret)
	return err
}

func (db Database) activeSessionCount() (int, error) {
	var res int
	row := db.db.QueryRow(`
		SELECT COUNT(*) AS count 
		FROM sessions 
		WHERE status NOT IN ('UNREACHABLE', 'TIMEOUT', 'DONE', 'CANCELLED') 
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
