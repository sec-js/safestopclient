package models

import (
	"fmt"
	"github.com/schoolwheels/safestopclient/database"
)

func InsertDevice(device_platform string, device_token string, user_id int) bool {

	if device_platform == "" || device_token == "" {
		return false
	}

	if DeviceExists(device_platform, device_token) == true {
		return true
	}

	query := `
insert into devices
(
device_platform,
notification_token,
user_id,
created_at,
updated_at
) values (
$1,
$2,
$3,
now(),
now()
)
`
	_, err := database.GetDB().Exec(query, device_platform, device_token, user_id)
	if err != nil {
		return false
	}
	return true
}

func DeviceExists(device_platform string, device_token string) bool {
	ct := 0
	query := `select count(*) from devices where device_platform_id = (select id from device_platforms where name = $1 limit 1) and notification_token = $2`
	row := database.GetDB().QueryRowx(query, device_platform, device_token)
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



func UpdateDeviceARN(device_platform string, device_token string, arn string) bool {

	if device_platform == "" || device_token == "" || arn == "" {
		return false
	}

	query := `update devices set sns_arn = $1 where notification_token = $2 and device_platform = $3`

	_, err := database.GetDB().Exec(query, arn, device_token, device_platform)
	if err != nil {
		return false
	}
	return true
}

func UpdateDeviceARNAndUser(device_platform string, device_token string, arn string, user_id int) bool {

	if device_platform == "" || device_token == "" || arn == "" || user_id <= 0 {
		return false
	}

	query := `update devices set sns_arn = $1, user_id = $2 where notification_token = $3 and device_platform = $4`

	_, err := database.GetDB().Exec(query, arn, user_id, device_token, device_platform)
	if err != nil {
		return false
	}
	return true
}
