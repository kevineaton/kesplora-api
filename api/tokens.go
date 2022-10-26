package api

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const (
	tokenExpiresMinutesEmail         = 30
	tokenExpiresMinutesPasswordReset = 30
	tokenExpiresMinutesRefresh       = 60 * 24 * 7
	tokenExpiresMinutesAccess        = 10

	tokenTypeEmail         = "email"
	tokenTypePasswordReset = "password_reset"
	tokenTypeRefresh       = "refresh"
	tokenTypeAccess        = "access"
)

// Token is a token struct that holds information about various token needs, including password reset, email verification, and refresh
type Token struct {
	UserID    int64  `json:"userId" db:"userId"`
	TokenType string `json:"tokenType" db:"tokenType"`
	CreatedOn string `json:"createdOn" db:"createdOn"`
	ExpiresOn string `json:"expiresOn" db:"expiresOn"`
	Token     string `json:"token" db:"token"`
}

// saveTokenForUser creates or updates a token for a user
func saveTokenForUser(input *Token) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`INSERT INTO Tokens (userId, tokenType, createdOn, expiresOn, token)
	VALUES
	(:userId, :tokenType, :createdOn, :expiresOn, :token)
	ON DUPLICATE KEY UPDATE
	token = :token,
	expiresOn = :expiresOn
	`, input)
	return err
}

// getTokenForUser gets a single token for a user; it is up to the caller to determine if it is expired
func getTokenForUser(userID int64, tokenType string) (*Token, error) {
	token := &Token{}
	defer token.processForAPI()
	err := config.DBConnection.Get(token, `SELECT * FROM Tokens WHERE userId = ? AND tokenType = ?`, userID, tokenType)
	return token, err
}

// deleteTokenForUser deletes the token for a user
func deleteTokenForUser(userID int64, tokenType string) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Tokens WHERE userId = ? AND tokenType = ?`, userID, tokenType)
	return err
}

// deleteTokensTheExpireBefore deletes ALL tokens before a specific type
func deleteTokensTheExpireBefore(expiresBefore string) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Tokens WHERE expiresOn < ?`, expiresBefore)
	return err
}

// getExpiresTimeForTokenType gets the expiration for a token by the type
func getExpiresTimeForTokenType(tokenType string) (time.Time, error) {
	switch tokenType {
	case tokenTypeEmail:
		return time.Now().Add(tokenExpiresMinutesEmail * time.Minute), nil
	case tokenTypePasswordReset:
		return time.Now().Add(tokenExpiresMinutesPasswordReset * time.Minute), nil
	case tokenTypeRefresh:
		return time.Now().Add(tokenExpiresMinutesRefresh * time.Minute), nil
	case tokenTypeAccess:
		return time.Now().Add(tokenExpiresMinutesAccess * time.Minute), nil
	}
	return time.Now(), errors.New("invalid token type")
}

func generateToken(user *User, tokenType string) (*Token, error) {
	// for email and password, it's pretty straight forward
	// for refresh, there's a bit more
	tokenSize := 8
	if tokenType == tokenTypeRefresh {
		tokenSize = 32
	}

	rand.Seed(time.Now().UnixNano())
	r := rand.Int63n(999999999999)
	str := fmt.Sprintf("%d%s-%d %s", user.ID, randomString(20), r, tokenType)
	hasher := md5.New()
	hasher.Write([]byte(str))
	hash := hex.EncodeToString(hasher.Sum(nil))
	tokenString := hash[0:tokenSize]
	if tokenType == tokenTypeRefresh {
		tokenString = fmt.Sprintf("ker_%s_%s", randomString(12), tokenString)
	}

	token := &Token{
		UserID:    user.ID,
		TokenType: tokenType,
		Token:     tokenString,
	}
	return token, nil
}

func generateCookies(accessToken, refreshToken string) (accessCookie, refreshCookie *http.Cookie) {
	accessCookie = &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  time.Now().Add(time.Minute * tokenExpiresMinutesAccess * 10),
		MaxAge:   tokenExpiresMinutesAccess * 60 * 10,
		Domain:   config.RootAPIDomain,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	refreshCookie = &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(time.Minute * tokenExpiresMinutesRefresh),
		MaxAge:   tokenExpiresMinutesRefresh * 60,
		Domain:   config.RootAPIDomain,
		Path:     "/users/refresh",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	return
}

// randomString takes a length and returns a randomized string of that length with letters, numbers, hyphens, and underscores
func randomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-+=[]{}")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

//
// processors
//

func (input *Token) processForDB() {
	if input.CreatedOn == "" {
		input.CreatedOn = time.Now().Format(timeFormatDB)
	} else {
		input.CreatedOn, _ = parseTimeToTimeFormat(input.CreatedOn, timeFormatDB)
	}

	if input.ExpiresOn == "" {
		expires, _ := getExpiresTimeForTokenType(input.TokenType)
		input.ExpiresOn = expires.Format(timeFormatDB)
	} else {
		input.ExpiresOn, _ = parseTimeToTimeFormat(input.ExpiresOn, timeFormatDB)
	}
}

func (input *Token) processForAPI() {
	input.CreatedOn, _ = parseTimeToTimeFormat(input.CreatedOn, timeFormatAPI)
	input.ExpiresOn, _ = parseTimeToTimeFormat(input.ExpiresOn, timeFormatAPI)
}
