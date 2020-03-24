package main

import "crypto/rand"
import "database/sql"
import "encoding/base64"
import "fmt"
import "math/big"
import "time"

import "github.com/lib/pq"

var ErrNoRows = sql.ErrNoRows

type Database struct {
	db *sql.DB
}

// Generate the secrets for a new session.
func generateSecrets() (DTMF, Secret, error) {
	dtmf, err := rand.Int(rand.Reader, big.NewInt(10_000_000_000))
	if err != nil {
		return "", "", err
	}
	secret := make([]byte, 24)
	_, err = rand.Read(secret)
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("%010d", dtmf), base64.URLEncoding.EncodeToString(secret), nil
}

func (db Database) NewSession() (DTMF, Secret, error) {
	var err error
	for attempt := 0; attempt < 10; attempt++ {
		var dtmf string
		var secret string
		dtmf, secret, err = generateSecrets()
		if err != nil {
			err = fmt.Errorf("failed to generate secrets: %w", err)
			return "", "", err
		}

		_, err = db.db.Exec("INSERT INTO sessions VALUES ($1, $2, DEFAULT, DEFAULT)", secret, dtmf)
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code.Name() == "unique_violation" {
			time.Sleep(100 * time.Millisecond)
			continue
		} else if err != nil {
			err = fmt.Errorf("failed to store session: %w", err)
			return "", "", err
		}

		return dtmf, secret, nil
	}
	err = fmt.Errorf("failed to find unique secrets: %w", err)
	return "", "", err
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

func (db Database) getDisclosed(secret string) (string, error) {
	var disclosed string
	row := db.db.QueryRow("SELECT disclosed FROM sessions WHERE secret = $1", secret)
	err := row.Scan(&disclosed)
	return disclosed, err
}

func (db Database) expire() error {
	_, err := db.db.Exec("DELETE FROM sessions WHERE created < now() - '1 hour'::interval")
	return err
}
