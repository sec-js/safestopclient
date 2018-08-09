package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"log"
	"fmt"
)

type SubAccountUsers struct {
	Users []SubAccountUser
}

type SubAccountUser struct {
	Id int `json:"id" db:"id"`
	FullName string `json:"full_name" db:"full_name"`
}


func SubAccountUsersForSubscription(subscription_id int) *SubAccountUsers {

	s := SubAccountUsers{}

	query := `
select 
a.id, 
d.last_name || ', ' || d.first_name as full_name
from subscription_sub_accounts a
join subscriptions b on a.subscription_id = b.id
join users c on c.id = a.user_id
join people d on d.id = c.person_id
where b.id = $1
`
	rows, err := database.GetDB().Queryx(query, subscription_id)
	if err != nil {
		log.Println(err)
		return &s
	}

	for rows.Next() {
		sub := SubAccountUser{}
		err = rows.StructScan(&sub)
		if err != nil {
			log.Println(err)
			return &s
		}
		s.Users = append(s.Users, sub)
	}

	return &s
}





func SubAccountUserExists(subscription_id int, user_id int) bool {
	ct := 0
	query := `select count(*) from subscription_sub_accounts where user_id = $1 and subscription_id = $2`
	row := database.GetDB().QueryRowx(query, user_id, subscription_id)
	if row == nil {
		return true
	}

	err := row.Scan(&ct)
	if err != nil {
		fmt.Print(err)
		return true
	}
	return (ct > 0)
}



func InsertSubAccountUser(subscription_id int, user_id int) bool {

	if subscription_id == 0 || user_id == 0 || SubAccountUserExists(subscription_id, user_id) {
		return false
	}

	query := `
insert into subscription_sub_accounts
(
subscription_id,
user_id,
created_at,
updated_at
) values (
$1,
$2,
now(),
now()
)
`
	_, err := database.GetDB().Exec(
		query,
		subscription_id,
		user_id)
	if err != nil {
		return false
	}
	return true
}