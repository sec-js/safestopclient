package models

import "github.com/schoolwheels/safestopclient/database"


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
