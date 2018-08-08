package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"log"
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