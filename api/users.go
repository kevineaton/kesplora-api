package api

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	UserStatusActive   = "active"
	UserStatusPending  = "pending"
	UserStatusLocked   = "locked"
	UserStatusDisabled = "disabled"

	UserSystemRoleUser        = "user"
	UserSystemRoleAdmin       = "admin"
	UserSystemRoleParticipant = "participant"
)

// User is a person with a login that has permission to "do stuff". This is for researchers, site admins, and participants
type User struct {
	ID              int64  `json:"id" db:"id"`
	Title           string `json:"title" db:"title"`
	FirstName       string `json:"firstName" db:"firstName"`
	LastName        string `json:"lastName" db:"lastName"`
	Pronouns        string `json:"pronouns" db:"pronouns"`
	Email           string `json:"email" db:"email"`
	Password        string `json:"password" db:"password"`
	DateOfBirth     string `json:"dateOfBirth" db:"dateOfBirth"`
	ParticipantCode string `json:"participantCode" db:"participantCode"`
	Status          string `json:"status" db:"status"`
	SystemRole      string `json:"systemRole" db:"systemRole"`
	CreatedOn       string `json:"createdOn" db:"createdOn"`
	LastLoginOn     string `json:"lastLoginOn" db:"ilastLoginOnd"`
}

func CreateUser(input *User) error {

}

func UpdateUser(input *User) error {

}

func DeleteUser(userID int64) error {

}

func GetUserByID(userID int64) (*User, error) {

}

func GetUserByEmail(email string) (*User, error) {

}

func GetAllUsersOnPlatform() ([]User, error) {

}

func AttemptLoginForUser(*User) error {

}

func LogOutUser(userID int64) error {

}

// jwtUser is a stripped down user for encoding into a jwt
type jwtUser struct {
	ID              int64  `json:"id" `
	Title           string `json:"title" `
	FirstName       string `json:"firstName" `
	LastName        string `json:"lastName""`
	Pronouns        string `json:"pronouns" `
	Email           string `json:"email" `
	DateOfBirth     string `json:"dateOfBirth" `
	ParticipantCode string `json:"participantCode" `
	Status          string `json:"status" `
	SystemRole      string `json:"systemRole"`
	Expires         string `json:"expires"`
}

type jwtClaims struct {
	User jwtUser `json:"user"`
	jwt.StandardClaims
}

func generateJWT(input *User) (string, string, error) {
	expires := time.Now().Add(tokenExpiresMinutesAccess * time.Minute).Format(timeFormatAPI)
	user := jwtUser{
		ID:              input.ID,
		Title:           input.Title,
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		Pronouns:        input.Pronouns,
		Email:           input.Email,
		DateOfBirth:     input.DateOfBirth,
		ParticipantCode: input.ParticipantCode,
		Status:          input.Status,
		SystemRole:      input.SystemRole,
		Expires:         expires,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
		"exp":  expires,
	})
	tokenString, err := token.SignedString([]byte(config.JWTSigningString))

	return tokenString, expires, err
}

func parseJWT(input string) (jwtUser, error) {
	token, err := jwt.ParseWithClaims(input, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, nil
		}
		return []byte(config.JWTSigningString), nil
	})
	if err != nil {
		return jwtUser{}, errors.New("could not parse jwt")
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		u := claims.User
		return u, nil
	}
	return jwtUser{}, errors.New("could not parse jwt")
}

func (input *User) processForDB() {
	if input.Status == "" {
		input.Status = UserStatusPending
	}
	if input.SystemRole == "" {
		input.SystemRole = UserSystemRoleUser
	}
	if input.DateOfBirth == "" {
		input.DateOfBirth = "1970-01-01"
	} else {
		input.DateOfBirth, _ = parseTimeToTimeFormat(input.DateOfBirth, dateFormat)
	}
	if input.CreatedOn == "" {
		input.CreatedOn = time.Now().Format(timeFormatDB)
	} else {
		input.CreatedOn, _ = parseTimeToTimeFormat(input.CreatedOn, timeFormatDB)
	}
	if input.LastLoginOn == "" {
		input.LastLoginOn = time.Now().Format(timeFormatDB)
	} else {
		input.LastLoginOn, _ = parseTimeToTimeFormat(input.LastLoginOn, timeFormatDB)
	}
}

func (input *User) processForAPI() {
	input.CreatedOn, _ = parseTimeToTimeFormat(input.CreatedOn, timeFormatAPI)
	input.LastLoginOn, _ = parseTimeToTimeFormat(input.LastLoginOn, timeFormatAPI)
}
