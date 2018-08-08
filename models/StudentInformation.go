package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"log"
)


type Students struct {
	StudentInformations []StudentInformation
}

type StudentInformation struct {
	Id int `json:"id" db:"id"`
	JurisdictionId int `json:"jurisdiction_id" db:"jurisdiction_id"`
	FullName string `json:"full_name" db:"full_name"`
	ScanCode string `json:"scan_code" db:"scan_code"`
}


func StudentsForUser(user_id int, jurisdiction_id int) *Students {
	s := Students{}

	query := `
select a.id,
a.jurisdiction_id,
b.Last_name || ', ' || b.first_name as full_name,
coalesce(scan_code, '') as scan_code
from student_informations a 
join people b on a.person_id = b.id
join personal_relationships c on c.person_related_id = b.id
join people d on c.person_id = d.id
join users e on e.person_id = d.id
where e.id = $1
and a.jurisdiction_id = $2
and deleted = false
`
	rows, err := database.GetDB().Queryx(query, user_id, jurisdiction_id)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	for rows.Next() {
		si := StudentInformation{}
		err = rows.StructScan(&si)
		if err != nil {
			log.Println(err.Error())
			return nil
		}
		s.StudentInformations = append(s.StudentInformations, si)
	}

	return &s
}





func StudentIdentifierExists(identifier string, jurisdiction_id int) bool {
	student_ct := 0
	query := `
select count(*)
from student_informations 
where sis_identifier = $1
and jurisdiction_id = $2
and deleted = false
`
	row := database.GetDB().QueryRowx(query, identifier, jurisdiction_id)
	if row == nil {
		return true
	} else {
		err := row.Scan(&student_ct)
		if err != nil {
			return true
		}
		return (student_ct > 0)
	}

}
