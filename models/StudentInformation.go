package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"log"
	"fmt"
)


type Students struct {
	StudentInformations []StudentInformation
}

type StudentInformation struct {
	Id int `json:"id" db:"id"`
	JurisdictionId int `json:"jurisdiction_id" db:"jurisdiction_id"`
	PersonId int `json:"person_id" db:"person_id"`
	FullName string `json:"full_name" db:"full_name"`
	ScanCode string `json:"scan_code" db:"scan_code"`
}


func StudentsForUser(user_id int, jurisdiction_id int) *Students {
	s := Students{}

	query := `
select a.id,
a.jurisdiction_id,
b.Last_name || ', ' || b.first_name as full_name,
coalesce(scan_code, '') as scan_code,
a.person_id as person_id
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


func StudentScanCode(identifier string, jurisdiction_id int) string {
	scan_code := ""
	query := `
select 
coalesce(scan_code, '') as scan_code
from student_informations 
where sis_identifier = $1
and jurisdiction_id = $2
and deleted = false
`
	row := database.GetDB().QueryRowx(query, identifier, jurisdiction_id)
	if row == nil {
		return ""
	} else {
		err := row.Scan(&scan_code)
		if err != nil {
			return ""
		}
		return scan_code
	}

}

func StudentIdForIdentifier(identifier string, jurisdiction_id int) int {
	student_id := 0
	query := `
select id
from student_informations 
where sis_identifier = $1
and jurisdiction_id = $2
`
	row := database.GetDB().QueryRowx(query, identifier, jurisdiction_id)
	if row == nil {
		return 0
	} else {
		err := row.Scan(&student_id)
		if err != nil {
			return 0
		}
		return student_id
	}
}

func StopIdsForStudentId(student_id int) []int {
	stop_ids := []int{}

	query := `
select id
from bus_route_stops a 
join bus_route_stops_student_informations b
on a.id = b.bus_route_stop_id
where a.deleted = false
and b.student_information_id = $1
`

	rows, err := database.GetDB().Queryx(query, student_id)
	if err != nil {
		log.Println(err.Error())
		return stop_ids
	}

	for rows.Next() {
		si := 0
		err = rows.StructScan(&si)
		if err != nil {
			log.Println(err.Error())
			return stop_ids
		}

		if si > 0 {
			stop_ids = append(stop_ids, si)
		}
	}

	return stop_ids
}


func PersonIdForStudentId(student_id int) int {
	person_id := 0
	query := `select person_id from student_informations where id = $1`
	row := database.GetDB().QueryRowx(query, student_id)
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
