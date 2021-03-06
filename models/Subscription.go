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

func RemovedStudentStopsToBeDeleted(person_id int) []int{

	r := []int{}

	query := `
select a.bus_route_stop_id
from bus_route_stops_student_informations a
join student_informations b
on a.student_information_id = b.id
where b.person_id = $1
`

	rows, err := database.GetDB().Queryx(query, person_id)
	if rows == nil {
		return r
	}

	for rows.Next() {
		pid := 0
		err = rows.Scan(&pid)
		if err != nil {
			log.Println(err.Error())
			return r
		}
		r = append(r, pid)
	}

	return r

}

func DeleteStopsFromUser(student_information_id int, user_id int) bool {

	query := `
delete
from bus_route_stop_users brsu
where brsu.user_id = $1
and brsu.bus_route_stop_id in (select bus_route_stop_id from bus_route_stops_student_informations where student_information_id = $2)


`
	log.Println(query)

	_, err := database.GetDB().Exec(query, user_id, student_information_id)

	if err != nil {
		return false
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

func StudentPersonIdsForSubscription(subscription_id int) []int{

	r := []int{}

	query := `
select f.person_id
from subscriptions a
join users b on a.user_id = b.id
join people c on c.id = b.person_id
join personal_relationships d on d.person_id = c.id
join products e on e.id = a.product_id
join student_informations f on f.person_id = d.person_related_id and f.jurisdiction_id = e.jurisdiction_id
where a.id = $1
`

	rows, err := database.GetDB().Queryx(query, subscription_id)
	if rows == nil {
		return r
	}

	for rows.Next() {
		sid := 0
		err = rows.Scan(&sid)
		if err != nil {
			log.Println(err.Error())
			return r
		}
		r = append(r, sid)
	}

	return r
}


