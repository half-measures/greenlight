package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"greenlight.alexedwards.net/internal/validator"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication" //scope to allow token Auth
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	//create token inst with userID, expire, and scope info
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}
	//init 16bit slice
	randomBytes := make([]byte, 16)
	//use read func from rand package to fill slice with random bytes
	//from OS CSPRNG, will err if not returned
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	//encode slice to base32 string and put it in token plaintxt field
	//note that end of string can have a = at the end, we dont want that
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	//Gen a SHA-256 hash of plaintext token string
	//sha256 func returns a array wiht len 32, convert to slice with :
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]
	return token, nil
}

// func to check plaintext token provided is 52 bytes long exactly
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

// Define TokenModel type
type TokenModel struct {
	DB *sql.DB
}

// This new method is creating a new token struct and inserts data into tokens table in sql
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = m.Insert(token)
	return token, err
}

// insert func looks for certian token in the tokens table
func (m TokenModel) Insert(token *Token) error {
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)`
	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// deleteall kills all tokens for a user and scope
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
	DELETE FROM tokens
	WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
