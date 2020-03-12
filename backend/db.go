package main

import "crypto/rand"
import "database/sql"
import "encoding/base64"
import "fmt"
import "log"
import "math/big"

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
	return dtmf.String(), base64.URLEncoding.EncodeToString(secret), nil
}

func (db Database) NewSession() (DTMF, Secret, error) {
	for {
		dtmf, secret, err := generateSecrets()
		if err != nil {
			err = fmt.Errorf("failed to generate secrets: %w", err)
			log.Print(err)
			return "", "", err
		}

		_, err = db.db.Exec("INSERT INTO sessions VALUES (?, ?, DEFAULT, DEFAULT)", dtmf, secret)
		// TODO: Continue retrying if error is unique violation.
		if err != nil {
			err = fmt.Errorf("failed to insert session in database: %w", err)
			log.Print(err)
			return "", "", err
		}

		return dtmf, secret, nil
	}
}

func (db Database) secretFromDTMF(dtmf string) (string, error) {
	var secret string
	row := db.db.QueryRow("SELECT secret FROM sessions WHERE dtmf = ?", dtmf)
	err := row.Scan(&secret)
	return secret, err
}

func (db Database) storeDisclosed(secret string, disclosed string) error {
	_, err := db.db.Exec("UPATE sessions SET disclosed = ? WHERE secret = ?", disclosed, secret)
	return err
}

func (db Database) getDisclosed(secret string) (string, error) {
	var disclosed string
	row := db.db.QueryRow("SELECT disclosed FROM sessions WHERE secret = ?", secret)
	err := row.Scan(&disclosed)
	return disclosed, err
}

func (db Database) expire() error {
	_, err := db.db.Exec("DELETE FROM sessions WHERE created < now() - '1 hour'::inerval")
	return err
}
