package models

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"fmt"
	"github.com/schoolwheels/safestopclient/database"
	"strings"
	"github.com/pkg/errors"
)

type ClientUser struct {
	*ModelBase
	//Error Error `json:"error"`
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
	FirstName string `json:"first_name" db:"first_name"`
	LastName string `json:"last_name" db:"last_name"`
	PersonId int `json:"person_id" db:"person_id"`
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

func RegisterUser(email string, password string, first_name string, last_name string) (int, error) {

	person_id := 0
	user_id := 0

	tx, err := database.GetDB().Begin()
	if err != nil {
		return 0, err
	}


	person_query := `
	insert into people (first_name, last_name, created_at, updated_at) values ($1, $2, now(), now()) returning id
`
	row := tx.QueryRow(person_query, first_name, last_name)
	if row == nil {
		tx.Rollback()
		return 0, errors.New("Person id not returned")
	} else {
		err := row.Scan(&person_id)
		if err != nil {
			tx.Rollback()
			return 0, errors.New("Person id scan failed")
		}
	}

	user_query := `
insert into users (
	email, 
	password_digest, 
	person_id,
	source_system, 
	security_segment_id, 
	created_at, 
	updated_at
) values (
$1,
$2,
$3,
'SafeStop',
(select id from security_segments where name = 'SafeStop' limit 1),
now(),
now()
) returning id
`
	row = tx.QueryRow(user_query,
		email,
		password,
		person_id)
	if row == nil {
		tx.Rollback()
		return 0, errors.New("User id not returned")
	} else {
		err := row.Scan(&user_id)
		if err != nil {
			tx.Rollback()
			return 0, errors.New("User id scan failed")
		}
	}

	permission_group_query := `
insert into permission_groups_users (permission_group_id, user_id) values ((select id from permission_groups where name = $1 limit 1),$2)
`
	_, err = tx.Exec(permission_group_query, "License 5 – SafeStop User", user_id)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()
	return user_id, err
}



func ActivateStudentIdentifierSubscription(jurisdiction_id int, product_id int, user *User, student_identifiers []string) (bool, error) {

	one_student_successful := false

	previous_identifier := ""
	for i := 0; i < len(student_identifiers); i++ {

		if len(student_identifiers[i]) > 0 && student_identifiers[i] != previous_identifier {

			tx, err := database.GetDB().Begin()
			if err != nil {
				continue
			}

			student_id := 0
			student_id_query := `
			select id from student_informations where deleted = false and jurisdiction_id = $1 and sis_identifier = $2
`
			row := tx.QueryRow(student_id_query, jurisdiction_id, student_identifiers[i])
			if row == nil {
				tx.Rollback()
				continue
			} else {
				err := row.Scan(&student_id)
				if err != nil {
					tx.Rollback()
					continue
				}

				relationship_query := `
insert into personal_relationships (person_id, person_related_id, personal_relationship_type_id, created_at, updated_at) 
values (
(select person_id from users where id = $1 limit 1), 
(select person_id from student_informations where id = $2 limit 1), 
1, 
now(), 
now())
`
				_, err = tx.Query(relationship_query, user.PersonId, student_id)
				if err != nil {
					tx.Rollback()
					continue
				}


				user_stops_query := `
insert into bus_route_stop_users (user_id, bus_route_stop_id, created_at, updated_at) (
select $1, id, now(), now() from bus_route_stops a 
join bus_route_stops_student_informations b on a.id = b.bus_route_stop_id
where b.student_information_id = $2
and a.deleted = false
and a.id not in (select bus_route_stop_id from bus_route_stop_users where user_id = $1)
)
`
				_, err = tx.Query(user_stops_query, user.Id, student_id)
				if err != nil {
					tx.Rollback()
					continue
				}

				tx.Commit()
				one_student_successful = true

			}
		}

	}

	if one_student_successful == true {

		subscription_query := `
insert into subscriptions (start_date, end_date, user_id, product_id, active, created_at, updated_at) 
values (
(select postgresql_name from time_zones where id = (select time_zone_id from jurisdictions where id = $1)), 
(select effective_end_date from products where id = $2), 
$3, 
now(), 
now()
)`
		_, err := database.GetDB().Query(subscription_query, jurisdiction_id, product_id, user.Id)
		if err != nil {
			return false, errors.New("Subscription creation failed")
		}

		return true, nil

	} else {
		return false, errors.New("No students assigned")
	}


}

func ActivateAccessCodeSubscription(jurisdiction_id int, product_id int, user *User) (bool, error) {

	tx, err := database.GetDB().Begin()
	if err != nil {
		return false, err
	}

	person_id := 0
	person_query := `
	insert into people (first_name, last_name, created_at, updated_at) values ((select to_hex(round(random() * 2^32 - 1)::BIGINT)), (select to_hex(round(random() * 2^32 - 1)::BIGINT)), now(), now()) returning id
`
	row := tx.QueryRow(person_query)
	if row == nil {
		tx.Rollback()
		return false, errors.New("Person id not returned")
	} else {
		err := row.Scan(&person_id)
		if err != nil {
			tx.Rollback()
			return false, errors.New("Person id scan failed")
		}
	}


	student_information_id := 0
	student_query := `
	insert into student_informations (jurisdiction_id, person_id, created_at, updated_at) values ($1, $2, now(), now()) returning id
`
	row = tx.QueryRow(student_query, jurisdiction_id, person_id)
	if row == nil {
		tx.Rollback()
		return false, errors.New("StudentInformation id not returned")
	} else {
		err := row.Scan(&student_information_id)
		if err != nil {
			tx.Rollback()
			return false, errors.New("StudentInformation id scan failed")
		}
	}

	relationship_query := `insert into personal_relationships (person_id, person_related_id, personal_relationship_type_id, created_at, updated_at) values ($1, $2, 1, now(), now())`
	_, err = tx.Exec(relationship_query, user.PersonId, person_id)
	if err != nil {
		tx.Rollback()
		return false, errors.New("StudentInformation id not returned")
	}

	subscription_query := `
insert into subscriptions (start_date, end_date, user_id, product_id, active, created_at, updated_at)
values (
now() at time zone (select postgresql_name from time_zones where id = (select time_zone_id from jurisdictions where id = $1) limit 1),
(select effective_end_date from products where id = $2),
$3,
$2,
true,
now(),
now()
)`
	_, err = tx.Exec(subscription_query, jurisdiction_id, product_id, user.Id)
	if err != nil {
		tx.Rollback()
		return false, errors.New("Subscription creation failed")
	}

	tx.Commit()
	return true, nil
}



























func FindUser(id int) *User {

	query := `
select a.id, 
a.person_id as person_id,
email, 
password_digest, 
coalesce(locked, false) as locked,
coalesce(super_admin, false) as super_admin,
(select array_to_string(array_agg(a.name), ',') as permission_groups
from permission_groups a 
join permission_groups_users b on b.permission_group_id = a.id 
and b.user_id = $1
and a.security_segment_id = (select id from security_segments where name = 'SafeStop' limit 1)),
coalesce(b.first_name, '') as first_name,
coalesce(b.last_name, '') as last_name
from users a 
join people b on b.id = a.person_id
where a.id = $1
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


func EmailExists(email string) bool {
	email_ct := 0
	email = ScrubEmailAddress(email)


	query := `
select count(*)
from users 
where lower(email) = $1
and (security_segment_id = (select id from security_segments where name = 'SafeStop' limit 1) or super_admin = true)
`
	row := database.GetDB().QueryRowx(query, email)
	if row == nil {
		return true
	} else {
		err := row.Scan(&email_ct)
		if err != nil {
			return true
		}
		return (email_ct > 0)
	}

}













func FindUserByEmail(email string) *User {
	email = ScrubEmailAddress(email)

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


func ScrubEmailAddress(email string) string {
	email = strings.Trim(email, " ")
	email = strings.ToLower(email)
	return email
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