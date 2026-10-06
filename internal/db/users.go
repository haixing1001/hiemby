package db

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// NewID generates a random hex ID.
func NewID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func digest(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// CreateUser inserts a user. Returns id.
func (d *DB) CreateUser(name, password string, admin bool) (string, error) {
	if len(password) < 3 || len(password) > 72 {
		return "", errPasswordLength
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	id := NewID()
	adm := 0
	if admin {
		adm = 1
	}
	_, err = d.Exec(`INSERT INTO users(id,name,hash,admin) VALUES(?,?,?,?)`,
		id, name, string(hash), adm)
	return id, err
}

var errPasswordLength = passwordError("password must be 3-72 bytes")

type passwordError string

func (e passwordError) Error() string { return string(e) }

// VerifyUser checks credentials, returns user id.
func (d *DB) VerifyUser(name, password string) (string, bool) {
	var id, hash string
	var admin int
	err := d.QueryRow(`SELECT id,hash,admin FROM users WHERE name=?`, name).Scan(&id, &hash, &admin)
	if err != nil {
		return "", false
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", false
	}
	return id, true
}

// UserCount returns number of users.
func (d *DB) UserCount() int {
	var n int
	d.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n
}

// IssueToken creates a login token, enforcing max devices.
func (d *DB) IssueToken(userID, device string, maxDevices int) (string, error) {
	tok := NewID()
	// enforce device limit: delete oldest if over
	var count int
	d.QueryRow(`SELECT COUNT(*) FROM tokens WHERE user_id=?`, userID).Scan(&count)
	if count >= maxDevices && maxDevices > 0 {
		d.Exec(`DELETE FROM tokens WHERE hash IN (
			SELECT hash FROM tokens WHERE user_id=? ORDER BY expires ASC LIMIT ?)`,
			userID, count-maxDevices+1)
	}
	_, err := d.Exec(`INSERT INTO tokens(hash,user_id,device,expires) VALUES(?,?,?,?)`,
		digest(tok), userID, device, time.Now().Add(30*24*time.Hour).Unix())
	return tok, err
}

// UserByToken resolves a token to user id + admin flag.
func (d *DB) UserByToken(tok string) (id, name string, admin bool, ok bool) {
	var adm int
	err := d.QueryRow(`SELECT u.id,u.name,u.admin FROM users u
		JOIN tokens t ON t.user_id=u.id
		WHERE t.hash=? AND t.expires>?`,
		digest(tok), time.Now().Unix()).Scan(&id, &name, &adm)
	if err != nil {
		return "", "", false, false
	}
	return id, name, adm == 1, true
}

// RevokeTokens deletes all tokens for a user.
func (d *DB) RevokeTokens(userID string) {
	d.Exec(`DELETE FROM tokens WHERE user_id=?`, userID)
}

// SetPassword updates a user's password.
func (d *DB) SetPassword(userID, password string) error {
	if len(password) < 3 || len(password) > 72 {
		return errPasswordLength
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = d.Exec(`UPDATE users SET hash=? WHERE id=?`, string(hash), userID)
	if err == nil {
		d.RevokeTokens(userID)
	}
	return err
}

// MaxDevices returns the device limit for a user.
func (d *DB) MaxDevices(userID string) int {
	var m int
	d.QueryRow(`SELECT max_devices FROM users WHERE id=?`, userID).Scan(&m)
	if m <= 0 {
		m = 2
	}
	return m
}
