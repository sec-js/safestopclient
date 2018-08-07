package models

import "github.com/schoolwheels/safestopclient/database"

type FailedRegistrationAttempt struct {
	JurisdictionId string
	FirstName string
	LastName string
	Email string
	StudentFirstName string
	StudentLastName string
	IdOrCodeAttempted string
	PostalCode string
}


func InsertFailedRegistrationAttempt(fra *FailedRegistrationAttempt) bool {
	query := `
insert into safe_stop_failed_registration_attempts
(
jurisdiction_id,
first_name,
last_name,
email,
student_first_name,
student_last_name,
id_or_code_attempted,
created_at,
updated_at
) values (
$1,
$2,
$3,
$4,
$5,
$6,
$7,
now(),
now()
)
`
	_, err := database.GetDB().Exec(
		query,
		fra.JurisdictionId,
		fra.FirstName,
		fra.LastName,
		fra.Email,
		fra.StudentFirstName,
		fra.StudentLastName,
		fra.IdOrCodeAttempted)
	if err != nil {
		return false
	}
	return true
}
