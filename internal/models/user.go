package models

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var userCache = UserCache{}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)

	uuid := uuid.New().String()
	tx.Statement.SetColumn("ID", uuid)
	return nil
}

func (u *User) Create(user User) (string, error) {
	result := db.Create(&user)

	if result.Error != nil {
		return "", result.Error
	}
	return user.ID, nil
}

func (u *User) FindOne(id string) (User, error) {
	var user User
	var err error

	if user, err = userCache.Read(id); err != nil {
		return user, err
	}

	if user.ID != "" {
		return user, nil
	}
	db.First(&user, "id = ?", id)

	if err = userCache.Write(user); err != nil {
		return user, err
	}

	return user, nil
}

func (u *User) FindByEMail(email string) (User, error) {
	var user User
	var err error

	if user, err = userCache.Read(email); err != nil {
		return user, err
	}

	if user.ID != "" {
		return user, nil
	}
	db.First(&user, "email = ?", email)

	if err = userCache.Write(user); err != nil {
		return user, err
	}

	return user, nil
}

func (u *User) FindAll() ([]User, error) {
	var users []User
	db.Find(&users)

	return users, nil
}

// Update updates one user in the database, using the information
// stored in the receiver u
func (u *User) Update() error {
	db.Save(&u)

	if err := userCache.Write(*u); err != nil {
		return err
	}

	return nil
}

func (u *User) Delete(id string) error {
	db.Delete(&User{}, id)
	if err := userCache.Delete(id); err != nil {
		return err
	}

	return nil
}

// ResetPassword is the method we will use to change a user's password.
func (u *User) ResetPassword(password string) error {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	db.Model(&User{}).Where("id = ?", u.ID).Update("password", hashedPassword)
	db.Model(&User{}).Where("id = ?", u.ID).Update("\"passwordResetExpiresAt\"", time.Now())

	return nil
}

func (u *User) PasswordMatches(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainTextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (u *User) ValidRole(role string) bool {
	roles := []string{"admin", "client", "staff"}

	for _, r := range roles {
		if r == role {
			return true
		}
	}

	return false
}

func (u *User) SetRole(role string) error {
	isValidRole := u.ValidRole(role)

	if !isValidRole {
		return errors.New("invalid user role")
	}

	u.Role = role
	return nil
}

func (u *User) FindByPasswordResetToken(resetToken string) (User, error) {
	var user User
	hashedToken := sha256.New()
	hashedToken.Write([]byte(resetToken))
	hashedTokenByteSlice := hashedToken.Sum(nil)
	hashedTokenString := hex.EncodeToString(hashedTokenByteSlice)

	result := db.Where("\"passwordResetToken\" = ? AND \"passwordResetExpiresAt\" > ?", hashedTokenString, time.Now()).Find(&user)
	if result.Error != nil {
		return user, result.Error
	}

	return user, nil
}

func (u *User) CreatePasswordResetToken() (string, error) {
	resetToken := uuid.NewString()

	hashedToken := sha256.New()
	hashedToken.Write([]byte(resetToken))
	hashedTokenByteSlice := hashedToken.Sum(nil)
	hashedTokenString := hex.EncodeToString(hashedTokenByteSlice)

	u.PasswordResetToken = hashedTokenString
	u.PasswordResetExpiresAt = time.Now().Add(20 * time.Minute)

	db.Save(&u)

	return resetToken, nil
}
