package api

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

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
	Password        string `json:"password,omitempty" db:"password"`
	DateOfBirth     string `json:"dateOfBirth" db:"dateOfBirth"`
	ParticipantCode string `json:"participantCode" db:"participantCode"`
	Status          string `json:"status" db:"status"`
	SystemRole      string `json:"systemRole" db:"systemRole"`
	CreatedOn       string `json:"createdOn" db:"createdOn"`
	LastLoginOn     string `json:"lastLoginOn" db:"lastLoginOn"`
	Access          string `json:"access"`
	Refresh         string `json:"refresh"` // web clients should not store this in local storage and should instead use the cookies!
	Expires         string `json:"expires"`
}

// CreateUser creates a new user in the db
func CreateUser(input *User) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO Users (title, firstName, lastName, pronouns, email, password, dateOfBirth, participantCode, status, systemRole, createdOn, lastLoginOn)
	VALUES
	(:title, :firstName, :lastName, :pronouns, :email, :password, :dateOfBirth, :participantCode, :status, :systemRole, :createdOn, :lastLoginOn)`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdateUser updates a user
func UpdateUser(input *User) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`UPDATE Users SET
	title = :title,
	firstName = :firstName,
	lastName = :lastName,
	pronouns = :pronouns,
	email = :email,
	password = :password,
	dateOfBirth = :dateOfBirth,
	participantCode = :participantCode,
	status = :status,
	systemRole = :systemRole,
	createdOn = :createdOn,
	lastLoginOn = :lastLoginOn
	WHERE id = :id`, input)
	return err
}

// DeleteUser completely deletes a user, and should really only be used in tests
func DeleteUser(userID int64) error {
	// TODO: as we add other user entries, we should delete them here (things like progress, etc)
	_, err := config.DBConnection.Exec("DELETE FROM Users WHERE id = ?", userID)
	if err != nil {
		return err
	}
	_, err = config.DBConnection.Exec("DELETE FROM Tokens WHERE userId = ?", userID)
	return err
}

// GetUserByID gets a user by the id
func GetUserByID(userID int64) (*User, error) {
	user := &User{}
	defer user.processForAPI()
	err := config.DBConnection.Get(user, `SELECT * FROM Users WHERE id = ?`, userID)
	return user, err
}

// GetUserByEmail gets a user by an email
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	defer user.processForAPI()
	err := config.DBConnection.Get(user, `SELECT * FROM Users WHERE email = ?`, email)
	return user, err
}

// GetUserByParticipantCode gets a user by the participant code
func GetUserByParticipantCode(participantCode string) (*User, error) {
	user := &User{}
	defer user.processForAPI()
	err := config.DBConnection.Get(user, `SELECT * FROM Users WHERE participantCode = ?`, participantCode)
	return user, err
}

// GetAllUsersOnPlatform gets all the users on the platform
func GetAllUsersOnPlatform() ([]User, error) {
	users := []User{}
	err := config.DBConnection.Select(&users, `SELECT * FROM Users`)
	for i := range users {
		users[i].processForAPI()
	}
	return users, err
}

func AttemptLoginForUser(emailOrCode, password string) (*User, error) {
	// if the user value contains an @ we assume and email, otherwise, we assume it's
	// a participant
	user := &User{}
	var err error
	if strings.Contains(emailOrCode, "@") {
		err = config.DBConnection.Get(user, `SELECT * FROM Users WHERE email = ?`, emailOrCode)
	} else {
		err = config.DBConnection.Get(user, `SELECT * FROM Users WHERE participantCode = ?`, emailOrCode)
	}
	if err != nil {
		return user, err
	}
	isValid := checkEncryptedPassword(password, user.Password)
	if !isValid {
		return user, errors.New("password did not match")
	}
	user.processForAPI()
	return user, nil
}

func LogOutUser(userID int64) error {
	// delete the refresh token so that when the access expires, it won't work
	// for now, that's it
	return deleteTokenForUser(userID, tokenTypeRefresh)
}

func userGenerateTokens(user *User, generateRefreshToken bool) (accessToken *Token, accessExpires string, refreshToken *Token, err error) {
	accessTokenString, accessExpires, err := generateJWT(user)
	if err != nil {
		return
	}
	accessToken = &Token{}
	accessToken.CreatedOn = time.Now().Format(timeFormatAPI)
	accessToken.ExpiresOn = accessExpires
	accessToken.TokenType = tokenTypeAccess
	accessToken.UserID = user.ID
	accessToken.Token = accessTokenString

	if generateRefreshToken {
		refreshToken, err = generateToken(user, tokenTypeRefresh)
		if err != nil {
			return
		}
		err = saveTokenForUser(refreshToken)
		if err != nil {
			return
		}
	}
	return
}

func createTestUser(defaults *User) error {
	if defaults.Password == "" {
		defaults.Password = fmt.Sprintf("test_P@%d!!", rand.Int63n(99999999999999))
	}
	if defaults.FirstName == "" {
		defaults.FirstName = "Admin"
	}
	if defaults.LastName == "" {
		defaults.LastName = "Admin"
	}
	if defaults.Email == "" {
		defaults.Email = fmt.Sprintf("test_%d@kesplora.com", rand.Int63n(99999999999999))
	}
	if defaults.Status == "" {
		defaults.Status = UserStatusActive
	}
	if defaults.SystemRole == "" {
		defaults.SystemRole = UserSystemRoleUser
	}
	err := CreateUser(defaults)
	if err != nil {
		return err
	}
	access, expires, refresh, err := userGenerateTokens(defaults, true)
	if err != nil {
		return err
	}
	err = saveTokenForUser(refresh)
	if err != nil {
		return err
	}
	defaults.Access = access.Token
	defaults.Expires = expires
	defaults.Refresh = refresh.Token
	return err
}

// jwtUser is a stripped down user for encoding into a jwt
type jwtUser struct {
	ID              int64  `json:"id" `
	Title           string `json:"title" `
	FirstName       string `json:"firstName" `
	LastName        string `json:"lastName"`
	Pronouns        string `json:"pronouns" `
	Email           string `json:"email" `
	DateOfBirth     string `json:"dateOfBirth" `
	ParticipantCode string `json:"participantCode" `
	Status          string `json:"status" `
	SystemRole      string `json:"systemRole"`
	Expires         string `json:"expires"`
}

type jwtClaims struct {
	User    jwtUser `json:"user"`
	Expires string  `json:"exp"`
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
	fmt.Printf("\n\t\t2: %+v\n", err)
	return jwtUser{}, errors.New("could not parse jwt")
}

func encryptPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkEncryptedPassword(plainPassword, encrypted string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encrypted), []byte(plainPassword))
	return err == nil
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
	// check if we need to change the password
	if input.Password != "" && !strings.HasPrefix(input.Password, "$2a$") {
		// we have a plaintext password, so hash it
		hashed, err := encryptPassword(input.Password)
		if err == nil {
			input.Password = hashed
		}
	}
}

func (input *User) processForAPI() {
	input.CreatedOn, _ = parseTimeToTimeFormat(input.CreatedOn, timeFormatAPI)
	input.LastLoginOn, _ = parseTimeToTimeFormat(input.LastLoginOn, timeFormatAPI)
	input.Password = ""
}

// Bind binds the data for the HTTP
func (data *User) Bind(r *http.Request) error {
	return nil
}
