package models

import (
	"fmt"
	"github.com/schoolwheels/safestopclient/database"
)



func InsertBusRouteStopUser(bus_route_stop_id int, user_id int, jurisdiction_id int) bool {

	if bus_route_stop_id == 0 || user_id == 0 || jurisdiction_id == 0 || SubAccountUserExists(bus_route_stop_id, user_id) {
		return false
	}

	query := `
insert into bus_route_stop_users
(
bus_route_stop_id,
user_id,
jurisdiction_id,
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
	_, err := database.GetDB().Exec(
		query,
		bus_route_stop_id,
		user_id,
		jurisdiction_id)
	if err != nil {
		return false
	}
	return true
}

func BusRouteStopUserExists(bus_route_stop_id int, user_id int) bool {
	ct := 0
	query := `select count(*) from bus_route_stop_users where user_id = $1 and bus_route_stop_id = $2`
	row := database.GetDB().QueryRowx(query, user_id, bus_route_stop_id)
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


func DeleteBusRouteStopUser(bus_route_stop_id int, user_id int) bool {
	query := `delete from bus_route_stop_users where user_id = $1 and bus_route_stop_id = $2`
	_, err := database.GetDB().Exec(query, user_id, bus_route_stop_id)
	if err != nil {
		return false
	}
	return true
}