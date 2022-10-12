package api

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenCRD(t *testing.T) {
	SetupConfig()
	userID := rand.Int63n(99999999)
	input, err := generateToken(&User{ID: userID}, tokenTypeEmail)
	assert.Nil(t, err)
	assert.Equal(t, tokenTypeEmail, input.TokenType)
	assert.NotEqual(t, "", input.Token)

	err = saveTokenForUser(input)
	assert.Nil(t, err)

	found, err := getTokenForUser(userID, tokenTypeEmail)
	assert.Nil(t, err)
	assert.Equal(t, tokenTypeEmail, found.TokenType)
	assert.NotEqual(t, "", found.CreatedOn)
	assert.NotEqual(t, "", found.ExpiresOn)
	assert.NotEqual(t, "", found.Token)

	err = deleteTokenForUser(userID, tokenTypeRefresh)
	assert.Nil(t, err)

	err = deleteTokenForUser(userID, tokenTypeEmail)
	assert.Nil(t, err)

	// create a new one then test the expires on
	older := &Token{
		UserID:    userID,
		TokenType: tokenTypeRefresh,
		ExpiresOn: time.Now().AddDate(0, 0, -3).Format(timeFormatDB),
	}
	err = saveTokenForUser(older)
	assert.Nil(t, err)

	err = deleteTokensTheExpireBefore(time.Now().AddDate(0, 0, -1).Format(timeFormatDB))
	assert.Nil(t, err)

	_, err = getTokenForUser(userID, tokenTypeRefresh)
	assert.NotNil(t, err)

}
