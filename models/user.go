package models

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/golang-jwt/jwt"
	"github.com/hidracloud/hidra/database"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const passwordCost int = 4

// User Represent user
type User struct {
	gorm.Model
	ID             uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Email          string    `gorm:"unique;primaryKey"`
	Password       []byte    `json:"-"`
	TwoFactorToken string    `json:"-" gorm:"type:char(32)"`
}

// GetUserByEmail Get one user given an email
func GetUserByEmail(email string) *User {
	var user User

	orm, err := database.GetORM(true)
	if err != nil {
		return nil
	}

	db, err := orm.DB()
	if err != nil {
		return nil
	}
	defer db.Close()

	orm.First(&user, "email = ?", email)
	return &user
}

// GetUserCount Get count of users
func GetUserCount() int64 {
	var count int64
	orm, err := database.GetORM(true)
	if err != nil {
		return -1
	}

	db, err := orm.DB()
	if err != nil {
		return -1
	}
	defer db.Close()
	orm.Model(&User{}).Count(&count)
	return count
}

// GetUserByID Get user by id
func GetUserByID(id string) *User {
	var user User
	orm, err := database.GetORM(true)
	if err != nil {
		return nil
	}

	db, err := orm.DB()
	if err != nil {
		return nil
	}
	defer db.Close()
	orm.First(&user, "id = ?", id)
	return &user
}

// CreateUser Create a new username
func CreateUser(email, password string) (*User, error) {
	err := checkmail.ValidateFormat(email)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)

	if err != nil {
		return nil, err
	}

	newUser := User{ID: uuid.NewV4(), Email: email, Password: hashedPassword}

	orm, err := database.GetORM(false)
	if err != nil {
		return nil, err
	}

	db, err := orm.DB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if result := orm.Create(&newUser); result.Error != nil {
		return nil, result.Error
	}

	return &newUser, nil
}

// Update2FAToken Update 2FA token
func (user *User) Update2FAToken(token string) error {
	user.TwoFactorToken = token
	orm, err := database.GetORM(false)
	if err != nil {
		return err
	}

	db, err := orm.DB()
	if err != nil {
		return err
	}
	defer db.Close()

	return orm.Save(user).Error
}

// UpdatePassword Update user password
func (user *User) UpdatePassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)

	if err != nil {
		return err
	}

	user.Password = hashedPassword

	orm, err := database.GetORM(false)
	if err != nil {
		return err
	}

	db, err := orm.DB()
	if err != nil {
		return err
	}
	defer db.Close()

	if result := orm.Save(user); result.Error != nil {
		return result.Error
	}

	return nil
}

// CreateUserToken Generate a new login token
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

// GetLoggedUser Return user by http header
func GetLoggedUser(r *http.Request) *User {
	return GetUserByID(r.Header.Get("user_id"))
}

// VerifyUserToken Verify if token is correct
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

// AuthMiddleware Verify if user is logged
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

		accountID := claims.(jwt.MapClaims)["user_id"].(string)
		exp := claims.(jwt.MapClaims)["exp"].(float64)

		if exp < float64(time.Now().Unix()) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Error verifying JWT token: " + err.Error()))
			return
		}

		r.Header.Set("user_id", accountID)

		next.ServeHTTP(w, r)
	})
}
