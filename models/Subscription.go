package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"log"
	)

type Subscriptions struct {
	Subscriptions []Subscription
}

type Subscription struct {
	Id int `json:"id" db:"id"`
	ProductId int `json:"product_id" db:"product_id"`
	ProductName string `json:"product_name" db:"product_name"`
	JurisdictionId int `json:"jurisdiction_id" db:"jurisdiction_id"`
	JurisdictionName string `json:"jurisdiction_name" db:"jurisdiction_name"`
	UserId int `json:"user_id" db:"user_id"`
	StartDate string `json:"start_date" db:"start_date"`
	EndDate string `json:"end_date" db:"end_date"`
}


func SubscriptionsForUser(user *User) *Subscriptions {
	subs := Subscriptions{}

	query := `
select 
c.id as id,
a.id as jurisdiction_id, 
a.name as jurisdiction_name, 
b.id as product_id,
b.name as product_name
from jurisdictions a
join products b on b.jurisdiction_id = a.id
join subscriptions c on b.id = c.product_id
join users d on c.user_id = d.id
join time_zones e on e.id = a.time_zone_id
where c.start_date <= (now() at time zone e.postgresql_name)::date
and c.end_date >= (now() at time zone e.postgresql_name)::date
and c.active = 't'
and b.product_type = 'ss'
and d.id = $1
order by jurisdiction_name
`
	rows, err := database.GetDB().Queryx(query, user.Id)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	for rows.Next() {
		sub := Subscription{}
		err = rows.StructScan(&sub)
		if err != nil {
			log.Println(err.Error())
			return nil
		}
		subs.Subscriptions = append(subs.Subscriptions, sub)
	}

	if len(subs.Subscriptions) > 0 {
		return &subs
	} else {
		return nil
	}
}


func FindSubscription(id int) *Subscription {

	query := `
select 
c.id as id,
a.id as jurisdiction_id, 
a.name as jurisdiction_name, 
b.id as product_id,
b.name as product_name,
c.user_id as user_id
from jurisdictions a
join products b on b.jurisdiction_id = a.id
join subscriptions c on b.id = c.product_id
join users d on c.user_id = d.id
join time_zones e on e.id = a.time_zone_id
where c.id = $1
`

	row := database.GetDB().QueryRowx(query, id)
	if row == nil {
		return nil
	}

	s := Subscription{}
	err := row.StructScan(&s)
	if err != nil {
		log.Println(err)
		return nil
	}

	return &s

}


func AddStudentToSubscription(subscription_id int, student_identifier string) bool {

	subscription := FindSubscription(subscription_id)
	if subscription == nil {
		return false
	}

	student_id := StudentIdForIdentifier(student_identifier, subscription.JurisdictionId)
	if student_id == 0 {
		return false
	}

	person_id := PersonIdForUserId(subscription.UserId)
	person_related_id := PersonIdForStudentId(student_id)
	if person_id == 0 || person_related_id == 0 {
		return false
	}

	relationship_added := InsertPersonalRelationship(person_id, person_related_id)
	if !relationship_added {
		return false
	}

	stop_ids := StopIdsForStudentId(student_id)
	AddStudentStopsToUser(subscription.UserId, stop_ids)

	sub_account_users := SubAccountUsersForSubscription(subscription_id)
	for i := 0; i < len(sub_account_users.Users); i++ {
		person_id = 0
		person_id = PersonIdForUserId(sub_account_users.Users[i].Id)

		relationship_added = InsertPersonalRelationship(person_id, person_related_id)
		if !relationship_added {
			continue
		}
		AddStudentStopsToUser(sub_account_users.Users[i].Id, stop_ids)
	}

	return true
}


func RemoveStudentFromSubscription(subscription_id int, student_id int) bool {

	subscription := FindSubscription(subscription_id)
	if subscription == nil {
		return false
	}

	person_id := PersonIdForUserId(subscription.UserId)
	person_related_id := PersonIdForStudentId(student_id)

	if person_id == 0 || person_related_id == 0 {
		return false
	}

	relationship_deleted := DeletePersonalRelationship(person_id, person_related_id)
	if !relationship_deleted {
		return false
	}

	sub_account_users := SubAccountUsersForSubscription(subscription_id)
	for i := 0; i < len(sub_account_users.Users); i++ {
		person_id = 0
		person_id = PersonIdForUserId(sub_account_users.Users[i].Id)

		relationship_deleted = DeletePersonalRelationship(person_id, person_related_id)
		if !relationship_deleted {
			continue
		}
	}

	return true
}


func AddSubAccountUserToSubscription(subscription *Subscription, sub_account_user_id int) bool {

	sub_account_added := InsertSubAccountUser(subscription.Id, sub_account_user_id)
	if sub_account_added == false {
		return false
	}

	students := StudentsForUser(subscription.UserId, subscription.JurisdictionId)
	person_id := PersonIdForUserId(sub_account_user_id)

	for i := 0; i < len(students.StudentInformations); i++ {

		person_related_id := PersonIdForStudentId(students.StudentInformations[i].Id)
		if person_id == 0 || person_related_id == 0 {
			continue
		}

		relationship_added := InsertPersonalRelationship(person_id, person_related_id)
		if !relationship_added {
			continue
		}

		stop_ids := StopIdsForStudentId(students.StudentInformations[i].Id)
		AddStudentStopsToUser(sub_account_user_id, stop_ids)

	}

	return true
}