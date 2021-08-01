package models

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/hidracloud/hidra/database"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const PASSWORD_COST int = 4

// Represent user
type User struct {
	gorm.Model
	ID       uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Email    string    `gorm:"unique;primaryKey"`
	Password []byte    `json:"-"`
}

// Get one user given an email
func GetUserByEmail(email string) *User {
	var user User
	database.ORM.First(&user, "email = ?", email)

	return &user
}

// Get user by id
func GetUserById(id string) *User {
	var user User
	database.ORM.First(&user, "id = ?", id)
	return &user
}

// Create a new username
func CreateUser(email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), PASSWORD_COST)

	if err != nil {
		return nil, err
	}

	newUser := User{ID: uuid.NewV4(), Email: email, Password: hashedPassword}

	if result := database.ORM.Create(&newUser); result.Error != nil {
		return nil, result.Error
	}

	return &newUser, nil
}

// Generate a new login token
func CreateUserToken(user *User) (string, error) {
	var err error

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = user.ID
	atClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("JWT_SECRET_TOKEN")))
	if err != nil {
		return "", err
	}

	return token, nil
}

// Return user by http header
func GetLoggedUser(r *http.Request) *User {
	return GetUserById(r.Header.Get("user_id"))
}

// Verify if token is correct
func VerifyUserToken(tokenString string) (jwt.Claims, error) {
	signingKey := []byte(os.Getenv("JWT_SECRET_TOKEN"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims, err
}

// Verify if user is logged
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.Header.Get("Authorization")

		if len(tokenString) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Missing Authorization Header"))
			return
		}
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
		claims, err := VerifyUserToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Error verifying JWT token: " + err.Error()))
			return
		}

		account_id := claims.(jwt.MapClaims)["user_id"].(string)
		exp := claims.(jwt.MapClaims)["exp"].(float64)

		if exp < float64(time.Now().Unix()) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Error verifying JWT token: " + err.Error()))
			return
		}

		r.Header.Set("user_id", account_id)

		next.ServeHTTP(w, r)
	})
}
