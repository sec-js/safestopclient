package models

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"fmt"
	"github.com/schoolwheels/safestopclient/database"
	"time"
)

type User struct {
	*ModelBase
	Id	int `json:"id" db:"id"`
	Email          string `json:"email" db:"email"`
	PasswordDigest string `json:"password_digest" db:"password_digest"`
	PasswordResetKey string `json:"password_reset_key" db:"password_reset_key"'`
	Active bool `json:"active" db:"active"`
	Superadmin bool `json:"superadmin" db:"superadmin"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func NewUser(email string, password string) *User {
	hashedPassword, _ := HashPassword(password)

	return &User{
		Email:          email,
		PasswordDigest: hashedPassword,
		Active: true,
		Superadmin: false,
	}
}

func FindUser(id int) *User {

	queryFindUser := "select * from users where id = $1;"
	row := database.GetDB().QueryRowx(queryFindUser, id)
	if row == nil {
		return nil
	} else {
		u := User{}
		err := row.StructScan(&u)
		if err != nil {
			fmt.Print(err)
		}
		return &u
	}
}

func FindUserByEmail(email string) *User {
	queryFindUser := "select * from users where email = $1;"
	row := database.GetDB().QueryRowx(queryFindUser, email)
	if row == nil {
		return nil
	} else {
		u := User{}
		err := row.StructScan(&u)
		if err != nil {
			fmt.Print(err)
		}
		return &u
	}
}

func AuthenticateUser(email string, password string) *User {

	var passwordDigest string
	err := database.GetDB().QueryRow("SELECT password_digest FROM users WHERE email=$1;", email).Scan(&passwordDigest)
	if err != nil {
		log.Fatal(err)
	}

	//hashedPassword, _ := HashPassword(password)

	if CheckPasswordHash(password, passwordDigest) {
		return &User{
			Email: email,
		}
	}

	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}


func (u *User) Save() error {
	var id int
	var err error
	if u.Id == 0 {
		err = database.GetDB().QueryRow(`INSERT INTO users (email, password_digest, password_reset_key, superadmin, active) VALUES ($1,$2,$3,$4,$5) RETURNING id`, u.Email, u.PasswordDigest, u.PasswordResetKey, u.Superadmin, u.Active).Scan(&id)
	} else {
		err = database.GetDB().QueryRow(`UPDATE users SET email = $2, password_digest = $3, password_reset_key = $4, superadmin = $5, active = $6 where id = $1 RETURNING id`, u.Id, u.Email, u.PasswordDigest, u.PasswordResetKey, u.Superadmin, u.Active).Scan(&id)
	}
	if err != nil {
		return err
	}
	u.Id = int(id)
	return nil
}