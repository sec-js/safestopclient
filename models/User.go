package models

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/schoolwheels/safestopclient/database"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math/rand"
	"strings"
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

func RegisterUser(email string, password string, first_name string, last_name string, permission_group_name string) (int, error) {

	person_id := 0
	user_id := 0

	password, err := HashPassword(password)
	if err != nil {
		return 0, err
	}

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
	_, err = tx.Exec(permission_group_query, permission_group_name, user_id)
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

func PersonIdForUserId(user_id int) int {
	person_id := 0
	query := `select person_id from users where id = $1`
	row := database.GetDB().QueryRowx(query, user_id)
	if row == nil {
		return 0
	} else {
		err := row.Scan(&person_id)
		if err != nil {
			fmt.Print(err)
			return 0
		}
		return person_id
	}
}




func UserIdForEmail(email string) int {
	email = ScrubEmailAddress(email)
	user_id := 0
	query := `select id from users where lower(email) = lower($1) and security_segment_id = (select id from security_segments where name = 'SafeStop' limit 1)`
	row := database.GetDB().QueryRowx(query, email)
	if row == nil {
		return 0
	}

	err := row.Scan(&user_id)
	if err != nil {
		log.Println(err)
		return 0
	}
	return user_id
}

func UserIdForPasswordResetCode(code string) int {

	user_id := 0
	query := `select id from users where password_reset_code = $1 and security_segment_id = (select id from security_segments where name = 'SafeStop' limit 1)`
	row := database.GetDB().QueryRowx(query, code)
	if row == nil {
		return 0
	}

	err := row.Scan(&user_id)
	if err != nil {
		log.Println(err)
		return 0
	}
	return user_id
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









type UserJurisdictionStopLimit struct {
	JurisdictionId int `json:"jurisdiction_id" db:"jurisdiction_id"`
	Count int `json:"ct" db:"ct"`
	Limit int `json:"limit" db:"limit"`
}




func UsersStopLimits(user_id int) *[]UserJurisdictionStopLimit {
	r := []UserJurisdictionStopLimit{}

	sql := `
select a.jurisdiction_id,
count(*) as ct,
b.selected_stop_limit as limit
from bus_route_stop_users a
join jurisdictions b
on a.jurisdiction_id = b.id
join bus_route_stops c
on c.id = a.bus_route_stop_id
where a.user_id = $1
and c.deleted = false
group by a.jurisdiction_id, b.gate_key, b.selected_stop_limit, a.user_id
`
	rows, err := database.GetDB().Queryx(sql, user_id)
	if err != nil {
		log.Println(err.Error())
		return &r
	}

	for i := 0; rows.Next(); i++ {
		sl := UserJurisdictionStopLimit{}
		err := rows.StructScan(&sl)
		if err != nil {
			log.Println(err.Error())
			return &r
		}
		r = append(r, sl)
	}
	return &r
}


func AddStopToRegularUsers(bus_route_stop_id int, user_id int) bool {
	stop_limit_info := StopLimitInfoForBusRouteStopId(bus_route_stop_id)
	if stop_limit_info != nil {
		if len(stop_limit_info.GateKey) > 0 {
			return InsertBusRouteStopUser(bus_route_stop_id, user_id, stop_limit_info.JurisdictionId)
		} else {
			if UserStopCountForJurisdictionId(stop_limit_info.JurisdictionId, user_id) < stop_limit_info.StopLimit {
				return InsertBusRouteStopUser(bus_route_stop_id, user_id, stop_limit_info.JurisdictionId)
			}
		}
	}
	return false
}

func AddStopToAdminAndSchoolUsers(bus_route_stop_id int, user_id int) bool {
	stop_limit_info := StopLimitInfoForBusRouteStopId(bus_route_stop_id)
	if stop_limit_info != nil {
		return InsertBusRouteStopUser(bus_route_stop_id, user_id, stop_limit_info.JurisdictionId)
	}
	return false
}

func UserStopCountForJurisdictionId(jurisdiction_id int, user_id int) int{
	ct := 0

	sql := `select count(distinct *) from bus_route_stop_users where jurisdiction_id = $1 and user_id = $2`
	row := database.GetDB().QueryRowx(sql, jurisdiction_id, user_id)
	if row != nil {
		err := row.Scan(&ct)
		if err != nil {
			fmt.Print(err)
		}
	}

	return ct
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



func UserHasAnyPermissionGroups(permission_groups []string, user_permission_groups string ) bool {
	has_permission_group := false
	for i := 0; i < len(permission_groups); i++ {
		if(strings.Contains(user_permission_groups, permission_groups[i])){
			has_permission_group = true
			i = len(permission_groups)
		}
	}
	return has_permission_group
}

func UserHasSubscriptionForJurisdiction(user *User, pg *PermissionGroups, jurisdiction_id int) bool {

	if user.SuperAdmin == true {
		return true
	} else if UserHasAnyPermissionGroups([]string{pg.License_1, pg.License_2, pg.License_3, pg.License_4, pg.Admin}, user.PermissionGroups) {
		return true
	} else {

		ct := 0

		query := `select count(*) 
  				  from subscriptions a 
				  join products b on b.id = a.product_id
				  join jurisdictions c on c.id = b.jurisdiction_id
				  where c.id = $1 
			      and a.user_id = $2
				  and a.start_date <= now()::date and a.end_date >= now()::date`

		row := database.GetDB().QueryRowx(query, jurisdiction_id, user.Id)
		err := row.Scan(&ct)
		if err != nil {
			log.Println(err)
			return true
		}

		return (ct > 0)
	}
}

func AddStudentStopsToUser(user_id int, stop_ids []int) bool {
	added_stops := false

	for i := 0; i < len(stop_ids); i++ {

		query := `
insert into bus_route_stops_user (
user_id, 
bus_route_stop_id
) values (
$1,
$2
)`
		_, err := database.GetDB().Exec(query, user_id, stop_ids)
		if err != nil {
			log.Println(err)
			added_stops = false
		}
		added_stops = true
	}

	return added_stops
}




func AddPermissionGroupToUser(user_id int, permission_group string) bool {

	ct := 0

	query := `
select count(*)
from permission_groups_users  
where user_id = $1
and permission_group_id = (select id from permission_groups where name = $2 limit 1)
`
	row := database.GetDB().QueryRowx(query, user_id, permission_group)
	if row == nil {
		return false
	}

	err := row.Scan(&ct)
	if err != nil {
		return false
	}

	if ct > 0 {
		return true
	}


	query = `
insert into permission_groups_users (
user_id, 
permission_group_id
) values (
$1,
(select id from permission_groups where name = $2 limit 1)
)`
	_, err = database.GetDB().Exec(query, user_id, permission_group)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}





//func UsersClientJurisdictionIdSQL(u *User, pg *PermissionGroups) string {
//
//	if((u.SuperAdmin == true) || UserHasAnyPermissionGroups([]string{ pg.Admin}, u.PermissionGroups)) {
//
//		return `
//select array_to_string(array_agg(a.id),',') as ids
//from jurisdictions a
//join jurisdiction_safe_stop_infos b on b.jurisdiction_id = a.id
//where active = true
//and (b.status = 'internal testing' or b.status = 'testing' or b.status = 'live')
//`
//
//	} else if (UserHasAnyPermissionGroups([]string{ pg.License_1, pg.License_2, pg.License_3, pg.License_4}, u.PermissionGroups)){
//
//		return fmt.Sprintf(`
//select array_to_string(array_agg(a.id),',') as ids
//from jurisdictions a
//join jurisdictional_restrictions b on a.id = b.jurisdiction_id
//join jurisdiction_safe_stop_infos c on c.jurisdiction_id = a.id
//where active = true
//and (c.status = 'testing' or c.status = 'live')
//and b.user_id = %d`, u.Id)
//
//	} else {
//
//		return fmt.Sprintf(`
//select array_to_string(array_agg(distinct z.jurisdiction_id),',') as ids from (
//
//select b.jurisdiction_id
//from subscriptions a
//join products b on b.id = a.product_id
//join jurisdictions d on d.id = b.jurisdiction_id
//join time_zones e on e.id = d.time_zone_id
//join jurisdiction_safe_stop_infos f on f.jurisdiction_id = d.id
//where 1 = 1
//and a.active = true
//and b.active = true
//and b.product_type = 'ss'
//and (now() at time zone e.postgresql_name)::date between a.start_date and a.end_date
//and a.user_id = %d
//and f.status = 'live'
//
//union all
//
//select b.jurisdiction_id
//from subscriptions a
//join products b on b.id = a.product_id
//join jurisdictions d on d.id = b.jurisdiction_id
//join time_zones e on e.id = d.time_zone_id
//join subscription_sub_accounts f on f.subscription_id = a.id
//join jurisdiction_safe_stop_infos g on g.jurisdiction_id = d.id
//where 1 = 1
//and a.active = true
//and b.active = true
//and b.product_type = 'ss'
//and (now() at time zone e.postgresql_name)::date between a.start_date and a.end_date
//and f.user_id = %d
//and g.status = 'live'
//) z`, u.Id, u.Id)
//
//	}
//}









func UsersClientJurisdictionIds(u *User, pg *PermissionGroups) string {

	sql := ""

	if((u.SuperAdmin == true) || UserHasAnyPermissionGroups([]string{ pg.Admin}, u.PermissionGroups)) {

		sql = `
select array_to_string(array_agg(a.id),',') as ids
from jurisdictions a
join jurisdiction_safe_stop_infos b on b.jurisdiction_id = a.id
where active = true
and (b.status = 'internal testing' or b.status = 'testing' or b.status = 'live')
`

	} else if (UserHasAnyPermissionGroups([]string{ pg.License_1, pg.License_2, pg.License_3, pg.License_4}, u.PermissionGroups)){

		sql = fmt.Sprintf(`
select array_to_string(array_agg(a.id),',') as ids
from jurisdictions a 
join jurisdictional_restrictions b on a.id = b.jurisdiction_id 
join jurisdiction_safe_stop_infos c on c.jurisdiction_id = a.id
where active = true
and (c.status = 'testing' or c.status = 'live')
and b.user_id = %d`, u.Id)

	} else {

		sql = fmt.Sprintf(`
select array_to_string(array_agg(distinct  z.id),',') as ids from (

select b.jurisdiction_id 
from subscriptions a
join products b on b.id = a.product_id
join jurisdictions d on d.id = b.jurisdiction_id
join time_zones e on e.id = d.time_zone_id
join jurisdiction_safe_stop_infos f on f.jurisdiction_id = d.id
where 1 = 1
and a.active = true
and b.active = true
and b.product_type = 'ss'
and (now() at time zone e.postgresql_name)::date between a.start_date and a.end_date
and a.user_id = %d
and f.status = 'live'

union all

select b.jurisdiction_id 
from subscriptions a
join products b on b.id = a.product_id
join jurisdictions d on d.id = b.jurisdiction_id
join time_zones e on e.id = d.time_zone_id
join subscription_sub_accounts f on f.subscription_id = a.id
join jurisdiction_safe_stop_infos g on g.jurisdiction_id = d.id
where 1 = 1
and a.active = true
and b.active = true
and b.product_type = 'ss'
and (now() at time zone e.postgresql_name)::date between a.start_date and a.end_date
and f.user_id = %d
and g.status = 'live'
) z`, u.Id, u.Id)

	}

	ids := "-1"
	row := database.GetDB().QueryRowx(sql)
	if row == nil {
		return ids
	} else {
		err := row.Scan(&ids)
		if err != nil {
			return ids
		}
	}

	return ids
}

func UsersClientJurisdictionCount(u *User, pg *PermissionGroups) int {

	count := 0
	sql := ""
	if((u.SuperAdmin == true) || UserHasAnyPermissionGroups([]string{ pg.Admin}, u.PermissionGroups)) {

		sql = `
select count(a.id) 
from jurisdictions a
join jurisdiction_safe_stop_infos b on b.jurisdiction_id = a.id
where active = true
and (b.status = 'internal testing' or b.status = 'testing' or b.status = 'live')
`

	} else if (UserHasAnyPermissionGroups([]string{ pg.License_1, pg.License_2, pg.License_3, pg.License_4}, u.PermissionGroups)){

		sql = fmt.Sprintf(`
select count(a.id) 
from jurisdictions a 
join jurisdictional_restrictions b on a.id = b.jurisdiction_id 
join jurisdiction_safe_stop_infos c on c.jurisdiction_id = a.id
where active = true
and (c.status = 'testing' or c.status = 'live')
and b.user_id = %d`, u.Id)

	} else {

		sql = fmt.Sprintf(`
select count(distinct z.jurisdiction_id) from (

select b.jurisdiction_id 
from subscriptions a
join products b on b.id = a.product_id
join jurisdictions d on d.id = b.jurisdiction_id
join time_zones e on e.id = d.time_zone_id
join jurisdiction_safe_stop_infos f on f.jurisdiction_id = d.id
where 1 = 1
and a.active = true
and b.active = true
and b.product_type = 'ss'
and (now() at time zone e.postgresql_name)::date between a.start_date and a.end_date
and a.user_id = %d
and f.status = 'live'

union all

select b.jurisdiction_id 
from subscriptions a
join products b on b.id = a.product_id
join jurisdictions d on d.id = b.jurisdiction_id
join time_zones e on e.id = d.time_zone_id
join subscription_sub_accounts f on f.subscription_id = a.id
join jurisdiction_safe_stop_infos g on g.jurisdiction_id = d.id
where 1 = 1
and a.active = true
and b.active = true
and b.product_type = 'ss'
and (now() at time zone e.postgresql_name)::date between a.start_date and a.end_date
and f.user_id = %d
and g.status = 'live'
) z`, u.Id, u.Id)

	}

	row := database.GetDB().QueryRowx(sql)
	if row == nil {
		return count
	} else {
		err := row.Scan(&count)
		if err != nil {
			return count
		}
	}

	return count
}







type MyStopsDbResult struct {
	StopId int `db:"stop_id"`
	StopName string `db:"stop_name"`
	ScheduledTimeOffset int `db:"scheduled_time_offset"`
	ScheduledTime string `db:"scheduled_time"`
	Audible int `db:"audible"`
	StopLatitude float64 `db:"stop_latitude"`
	StopLongitude float64 `db:"stop_longitude"`
	BusRouteId int `db:"bus_route_id"`
	BusAssigned int `db:"bus_assigned"`
	BusRouteActive bool `db:"bus_route_active"`
	LoopMode string `db:"loop_mode"`
	LoopModeFlag bool `db:"loop_mode_flag"`
	HidePredictions bool `db:"hide_predictions"`
	BusRouteName string `db:"bus_route_name"`
	JurisdictionName string `db:"jurisdiction_name"`
	SkippedAt string `db:"skipped_at"`
	ArrivalTime string `db:"arrival_time"`
	AsOf int `db:"as_of"`
	PredictedTimeOffset int `db:"predicted_time_offset"`
	PredictedTimeString string `db:"predicted_time_string"`
	BusLatitude float64 `db:"bus_latitude"`
	BusLongitude float64 `db:"bus_longitude"`
	ShowBus bool `db:"show_bus"`
}

type MyStopsJurisdiction struct {
	Name string `json:"jn"`
	Routes []MyStopsRoute `json:"routes"`
}

type MyStopsRoute struct {
	Id int `json:"rid"`
	Name string `json:"rn"`
	HidePredictions bool `json:"hp"`
	Stops []MyStopsStop `json:"stops"`
	Active bool `json:"ra"`
	Audible int `json:"a"`
	BusAssigned bool `json:"ba"`
	Errors bool `json:"e"`
	Shuttle bool `json:"sh"`
}

type MyStopsStop struct {
	Id int `json:"sid"`
	BusRouteId int `json:"rid"`
	Name string `json:"sn"`
	ScheduledTime string `json:"sst"`
	AsOf string `json:"ao"`
	Latitude float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	TimeTitle string `json:"tt"`
	Time string `json:"t"`
	TimeClass string `json:"tc"`
}

type MapViewStop struct {
	StopId int `json:"sid"`
	StopName string `json:"sn"`
	StopLatitude float64 `json:"slat"`
	StopLongitude float64 `json:"slng"`
	StopScheduledTime string `json:"sst"`
	BusLatitude float64 `json:"blat"`
	BusLongitude float64 `json:"blng"`
	Time string `json:"t"`
	TimeTitle string `json:"tt"`
	TimeClass string `json:"tc"`
	AsOf string `json:"ao"`
	BusRouteId int `json:"rid"`
	BusRouteName string `json:"rn"`
	HidePredictions bool `json:"hp"`
	Shuttle bool `json:"sh"`
	Audible int `json:"a"`
	ShowBus bool `json:"sb"`
}



func UsersMyStops(u *User, pg *PermissionGroups) []MyStopsDbResult {

	jurisdiction_ids := UsersClientJurisdictionIds(u, pg)

	dbr := []MyStopsDbResult{}

sql := fmt.Sprintf(`

select a.id as stop_id,
case when c.titlecase = true
then initcap(a.display_name) 
when c.titlecase = false 
then a.display_name end as stop_name,

case when (extract(second from now() at time zone e.postgresql_name) + 
           extract(minute from now() at time zone e.postgresql_name) * 60 + 
           extract(hour from now() at time zone e.postgresql_name) * 60 * 60) 
between (b.start_time + (b.adjusted_start * 60)) and (b.end_time + (b.adjusted_end * 60)) then true else false end as show_bus,



to_char(now()::date + (a.scheduled_time_offset + b.start_time) * interval '1 second', 'hh12:mi am') as scheduled_time,

a.scheduled_time_offset as scheduled_time_offset,

case when f.audible_type is not null then
                      case when f.audible_type = 'ampm' then 
                        '1'
                        when f.audible_type = 'am' then 
                          case when (b.start_time) < 36000 then
                            '1'
                          else
                            '0'
                          end       
                        when f.audible_type = 'pm' then
                          case when (b.start_time) > 46800 then
                            '1'
                          else
                            '0'
                          end
                      end
                    else
			                '0'
                    end as audible,

case when a.adjusted_latitude is not null then a.adjusted_latitude else a.latitude end as stop_latitude,

case when a.adjusted_longitude is not null then a.adjusted_longitude else a.longitude end as stop_longitude,

b.id as bus_route_id,

coalesce(b.bus_id, -1) as bus_assigned,

coalesce(b.active, true) as bus_route_active,

coalesce(b.loop_mode, '') as loop_mode,

case when coalesce(b.loop_mode, '') = 'off' then false else true end as loop_mode_flag,

coalesce(b.hide_predictions, false) as hide_predictions,
                   
case when c.titlecase = 't' then coalesce(initcap(b.display_name), '') when c.titlecase = 'f' then coalesce(b.display_name, '') end as bus_route_name,
                   
initcap(coalesce(c.name, '')) as jurisdiction_name,
                   
coalesce((select to_char(skipped_at at time zone 'utc' at time zone e.postgresql_name, 'hh12:mi am') from bus_route_stop_activity_logs where bus_route_stop_id = a.id and (created_at at time zone 'utc' at time zone e.postgresql_name)::date = (now() at time zone e.postgresql_name)::date and skipped_at is not null order by created_at desc limit 1), '') as skipped_at,
                   
coalesce((select to_char(arrived_time at time zone 'utc' at time zone e.postgresql_name, 'hh12:mi am') from bus_route_stop_activity_logs where bus_route_stop_id = a.id and (created_at at time zone 'utc' at time zone e.postgresql_name)::date = (now() at time zone e.postgresql_name)::date and arrived_time is not null order by id desc limit 1), '') as arrival_time,

coalesce(case when (select predicted_time_offset
		    from bus_route_stop_activity_logs
			where bus_route_stop_id = a.id
				and created_at::date = now()::date order by id desc limit 1) is not null
then (select extract(minutes from (now() - created_at)) + extract(hours from (now() - created_at)) * 60
		   from bus_route_stop_activity_logs
		   where bus_route_stop_id = a.id
			and created_at::date = now()::date order by id desc limit 1)
end, -1) as as_of,



--case when (select predicted_time_offset
--		    from bus_route_stop_activity_logs
--			where bus_route_stop_id = a.id
--				and (created_at at time zone 'utc' at time zone e.postgresql_name)::date = (now() at time zone e.postgresql_name)::date order by id desc limit 1) is not null
--
--then (select to_char((created_at + (predicted_time_offset * interval '1 second')) at time zone 'utc' at time zone e.postgresql_name, 'hh12:mi am')
--		   from bus_route_stop_activity_logs
--		   where bus_route_stop_id = a.id
--			and (created_at at time zone 'utc' at time zone e.postgresql_name)::date = (now() at time zone e.postgresql_name)::date order by id desc limit 1)
--else ''
--end as predicted_time_offset,



case when (select predicted_time_offset from bus_route_stop_activity_logs where bus_route_stop_id = a.id and (created_at at time zone 'utc' at time zone e.postgresql_name)::date = (now() at time zone e.postgresql_name)::date order by id desc limit 1) is not null
then (select predicted_time_offset from bus_route_stop_activity_logs where bus_route_stop_id = a.id and (created_at at time zone 'utc' at time zone e.postgresql_name)::date = (now() at time zone e.postgresql_name)::date order by id desc limit 1)
else -1
end as predicted_time_offset,





case when (select predicted_time_offset
		    from bus_route_stop_activity_logs
			where bus_route_stop_id = a.id
				and (created_at at time zone 'utc' at time zone e.postgresql_name)::date = (now() at time zone e.postgresql_name)::date order by id desc limit 1) is not null

then (select predicted_time_string from bus_route_stop_activity_logs where bus_route_stop_id = a.id
			and (created_at at time zone 'utc' at time zone e.postgresql_name)::date = (now() at time zone e.postgresql_name)::date order by id desc limit 1)
else ''
end as predicted_time_string,

coalesce(d.latitude, -1) as bus_latitude,

coalesce(d.longitude, -1) as bus_longitude

from bus_route_stops a
join bus_routes b on a.bus_route_id = b.id
join jurisdictions c on b.jurisdiction_id = c.id
left join buses d on b.bus_id = d.id
join time_zones e on e.id = c.time_zone_id
left join safe_stop_audibles f on f.jurisdiction_id = c.id
              and (f.created_at at time zone 'utc' at time zone e.postgresql_name)::date = (now() at time zone e.postgresql_name)::date
              and f.active = true
where 1=1
and c.id in (%s)
and a.id in (select bus_route_stop_id from bus_route_stop_users where user_id = %d)
and c.active = 't'
and a.deleted = 'f'
and a.active = 't'
order by c.name, b.display_name, a.scheduled_time_offset

`, jurisdiction_ids, u.Id)

	rows, err := database.GetDB().Queryx(sql)
	if err != nil {
		log.Println(err.Error())
		return dbr
	}

	for i := 0; rows.Next(); i++ {
		s := MyStopsDbResult{}
		err := rows.StructScan(&s)
		if err != nil {
			log.Println(err.Error())
			return dbr
		}
		dbr = append(dbr, s)
	}

	return dbr
}






func UserScanNotifications(u *User) []ScanNotification {

	r := []ScanNotification{}

	sql := `
select
a.id as id, 
to_char(b.date_occurred at time zone 'UTC' at time zone e.postgresql_name, 'HH12:MI AM') as date_occurred,
f.name as name
from bus_rider_scan_notifications a 
join bus_rider_scans b on b.id = a.bus_rider_scan_id
join buses c on c.id = b.bus_id
join gps_configs d on d.id = c.gps_config_id
join time_zones e on e.id = d.time_zone_id
join bus_rider_scan_notification_subscriptions f on f.id = a.bus_rider_scan_notification_subscription_id
where a.dismissed_at is null
and (now() at time zone e.postgresql_name)::date = (b.date_occurred at time zone 'UTC' at time zone e.postgresql_name)::date
and f.user_id = $1
order by date_occurred
`
	rows, err := database.GetDB().Queryx(sql, u.Id)
	if err != nil {
		log.Println(err.Error())
		return r
	}

	for i := 0; rows.Next(); i++ {
		s := ScanNotification{}
		err := rows.StructScan(&s)
		if err != nil {
			log.Println(err.Error())
		}
		r = append(r, s)
	}
	return r
}


func GenerateUserPasswordResetCode(user_id int) string {
	code := RandStringBytes(25)
	query := "update users set password_reset_code = $1 where id = $2;"
	_, err := database.GetDB().Exec(query, code, user_id)
	if err != nil {
		return code
	}
	return code
}

func ClearUserPasswordResetCode(user_id int) bool {
	query := "update users set password_reset_code = null where id = $1;"
	_, err := database.GetDB().Exec(query, user_id)
	if err != nil {
		return false
	}
	return true
}



const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}


func UpdateUserPassword(user_id int, password string) bool{
	hashed_password, err := HashPassword(password)
	if err != nil {
		return false
	}

	query := "update users set password_digest = $1 where id = $2;"

	_, err = database.GetDB().Exec(query, hashed_password, user_id)
	if err != nil {
		return false
	}
	return true
}

















func UsersAvailableBusRoutes(user *User, page int, search_condition string, address_1 string, postal_code string){
//	limit := 20
//	offset := (page - 1) * limit
//	parameters := []string{}
//
//	user_coordinates := Geocode(address_1, postal_code)
//
//	geo_sql := ""
//	if user_coordinates.Latitude != "" {
//
//		coordinate_string := fmt.Sprintf("%s %s")
//		parameters = append(parameters, coordinate_string)
//		param_index := indexOf(coordinate_string, parameters) + 1
//
//
//		geo_sql = fmt.Sprintf( `
//        AND (ST_Distance(
//                   ST_Transform(ST_GeomFromText('POINT($%d)',4326),900913),
//                   ST_Transform(ST_GeomFromText('POINT(' || CAST(case when e.adjusted_longitude is null then e.longitude else e.adjusted_longitude end as VARCHAR) || ' ' || CAST(case when e.adjusted_latitude is null then e.latitude else e.adjusted_latitude end as VARCHAR) || ')', 4326),900913)
//               ) * 0.000621371) < b.search_radius
//`, param_index)
//	}
//
//	search_sql := ""
//	if search_condition != "" {
//
//		parameters = append(parameters, strings.ToLower("%" + search_condition + "%"))
//		param_index := indexOf(strings.ToLower(search_condition), parameters) + 1
//
//
//		search_sql = fmt.Sprint(`
//and (lower(a.display_name) like $%d
//or lower(b.name) like $%d)"
//`, param_index, param_index)
//	}
//
//	total_sql := `select count(distinct a.id) `
//
//	select_sql := `
//select
//a.id as bus_route_id,
//case when b.titlecase = true then initcap(a.display_name) when b.titlecase = false then a.display_name end as bus_route_name,
//to_char(now()::date + a.start_time * interval '1 second', 'HH12:MI AM')  as bus_route_start_time,
//b.id as jurisdiction_id,
//initcap(b.name) as jurisdiction_name,
//b.search_radius
//`
//
//	from_sql := `
//from bus_routes a
//join jurisdictions b
//on a.jurisdiction_id = b.id
//join buses c
//on a.bus_id = c.id
//join jurisdiction_safe_stop_infos d
//on d.jurisdiction_id = b.id
//join bus_route_stops e
//on e.bus_route_id = a.id
//`
//
//	where_sql := `
//where a.active = true
//and a.bus_id is not null
//and e.active = true
//and e.deleted = false
//and b.id in (7) --(#{ client_jurisdictions.map(&:id).join(',') })
//`
//
//	group_by_sql := `
//group by a.id, a.display_name, a.start_time, b.id, b.name, b.search_radius
//`
//
//	order_by_sql := fmt.Sprintf(`
//order by jurisdiction_name, bus_route_name
//limit %d
//offset %d
//`, limit, offset)
//
//
//
//if()
















}







func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func indexOf(element string, data []string) (int) {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1    //not found.
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