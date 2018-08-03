package models

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"fmt"
	"github.com/schoolwheels/safestopclient/database"
	"strings"
)

type ClientUser struct {
	*ModelBase
	Error Error `json:"error"`
	Id	int `json:"id" db:"id"`
	Email string `json:"email" db:"email"`
}



type User struct {
	*ModelBase
	Id	int `json:"id" db:"id"`
	Email          string `json:"email" db:"email"`
	PasswordDigest string `json:"password_digest" db:"password_digest"`
	SuperAdmin bool `json:"super_admin" db:"super_admin"`
	PermissionGroups string `json:"permission_groups" db:"permission_groups"`
	//PasswordResetKey string `json:"password_reset_key" db:"password_reset_key"'`
	Locked bool `json:"locked" db:"locked"`
	//CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func UpdateApiToken(user_id int, token string) bool{
	query := "update users set api_token = $1 where id = $2;"
	_, err := database.GetDB().Exec(query, token, user_id)
	if err != nil {
		return false
	}
	return true
}

//func NewUser(email string, password string) *User {
//	hashedPassword, _ := HashPassword(password)
//
//	return &User{
//		Email:          email,
//		PasswordDigest: hashedPassword,
//		Active: true,
//		Superadmin: false,
//	}
//}

func FindUser(id int) *User {

	query := `
select id, 
email, 
password_digest, 
locked,
super_admin,
(select array_to_string(array_agg(a.name), ',') as permission_groups
from permission_groups a 
join permission_groups_users b on b.permission_group_id = a.id 
and b.user_id = id 
and a.security_segment_id = (select id from security_segments where name = 'SafeStop' limit 1))
from users 
where id = $1
limit 1
`
	row := database.GetDB().QueryRowx(query, id)
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
	email = strings.Trim(email, " ")
	email = strings.ToLower(email)

	query := `
select id, 
email, 
password_digest, 
locked,
super_admin,
(select array_to_string(array_agg(a.name), ',') as permission_groups
from permission_groups a 
join permission_groups_users b on b.permission_group_id = a.id 
and b.user_id = id 
and a.security_segment_id = (select id from security_segments where name = 'SafeStop' limit 1))
from users 
where lower(email) = $1
and (security_segment_id = (select id from security_segments where name = 'SafeStop' limit 1)
or super_admin = true)
limit 1
`
	row := database.GetDB().QueryRowx(query, email)
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




func FindUserByToken(token string) *ClientUser {
	if(token == ""){
		return nil
	} else {
		query := "select id, email, password_digest, locked from users where api_token = $1;"
		row := database.GetDB().QueryRowx(query, token)
		if row == nil {
			return nil
		} else {
			u := ClientUser{}
			err := row.StructScan(&u)
			if err != nil {
				fmt.Print(err)
			}
			return &u
		}
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



func HasAnyPermissionGroups(permission_groups []string, user_permission_groups string ) bool {
	has_permission_group := false
	for i := 0; i < len(permission_groups); i++ {
		if(strings.Contains(user_permission_groups, permission_groups[i])){
			has_permission_group = true
			i = len(permission_groups)
		}
	}
	return has_permission_group
}






func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}


//func (u *User) Save() error {
//	var id int
//	var err error
//	if u.Id == 0 {
//		err = database.GetDB().QueryRow(`INSERT INTO users (email, password_digest, password_reset_key, superadmin, active) VALUES ($1,$2,$3,$4,$5) RETURNING id`, u.Email, u.PasswordDigest, u.PasswordResetKey, u.Superadmin, u.Active).Scan(&id)
//	} else {
//		err = database.GetDB().QueryRow(`UPDATE users SET email = $2, password_digest = $3, password_reset_key = $4, superadmin = $5, active = $6 where id = $1 RETURNING id`, u.Id, u.Email, u.PasswordDigest, u.PasswordResetKey, u.Superadmin, u.Active).Scan(&id)
//	}
//	if err != nil {
//		return err
//	}
//	u.Id = int(id)
//	return nil
//}