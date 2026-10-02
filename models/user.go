package models

import (
	"database/sql"
	"errors"

	"example.com/rest-api/db"
	"example.com/rest-api/utils"
	"github.com/mattn/go-sqlite3"
)

// ErrInvalidCredentials is returned when the email is unknown or the password
// does not match the stored hash.
var ErrInvalidCredentials = errors.New("credentials invalid")

// ErrEmailTaken is returned when a user with the same email already exists.
var ErrEmailTaken = errors.New("email already in use")

type User struct {
	ID       int64
	Email    string `binding:"required"`
	Password string `binding:"required"`
}

func (u *User) Save() error {
	query := "INSERT INTO users(email, password) VALUES (?, ?)"
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}

	result, err := stmt.Exec(u.Email, hashedPassword)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return ErrEmailTaken
		}
		return err
	}

	userId, err := result.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = userId
	return nil
}

func (u *User) ValidateCredentials() error {
	query := "SELECT id, password FROM users WHERE email = ?"
	row := db.DB.QueryRow(query, u.Email)

	var retrievedPassword string
	err := row.Scan(&u.ID, &retrievedPassword)
	if err != nil {
		// An unknown email is a failed login, not a server error.
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return err
	}

	if !utils.CheckPasswordHash(u.Password, retrievedPassword) {
		return ErrInvalidCredentials
	}

	return nil
}
