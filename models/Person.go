package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"fmt"
	"log"
)

func InsertPersonalRelationship(person_id int, person_related_id int) bool {

	if PersonalRelationshipExists(person_id, person_related_id) {
		return false
	}

	relationship_id := 0

	query := `
insert into personal_relationships (
person_id, 
person_related_id, 
personal_relationship_type_id,
created_at, 
updated_at
)
values (
$1,
$2,
1,
now(),
now()
) returning id
`
	row := database.GetDB().QueryRow(query, person_id, person_related_id)
	if row == nil {
		return false
	}

	err := row.Scan(&relationship_id)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}



func DeletePersonalRelationship(person_id int, person_related_id int) bool {
	query := `
	delete from personal_relationships where person_id = $1 and person_related_id = $2
`
	_, err := database.GetDB().Exec(query, person_id, person_related_id)
	if err != nil {
		return false
	}
	return true
}


func PersonalRelationshipExists(person_id int, person_related_id int) bool {
	ct := 0
	query := `select count(*) from personal_relationships where person_id = $1 and person_related_id = $2`
	row := database.GetDB().QueryRowx(query, person_id, person_related_id)
	if row == nil {
		return true
	} else {
		err := row.Scan(&ct)
		if err != nil {
			fmt.Print(err)
			return true
		}
		return (ct > 0)
	}
}

