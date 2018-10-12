package models

import (
	"fmt"
	"github.com/schoolwheels/safestopclient/database"
	"log"
	"strconv"
	"strings"
)

type Alert struct {
	Id int `db:"id" json:"id"`
	Priority string `db:"priority" json:"priority"`
	Text string `db:"text" json:"text"`
}

func AlertsForAdminUser() []Alert{

	r := []Alert{}

	sql := `

select a.id as id,
'All - ' || a.text as text,
a.priority as priority
from safe_stop_broadcast_messages a
where jurisdiction_id is null
and a.bus_route_id is null
and a.bus_route_stop_id is null
and a.bus_id is null
and a.active = true
and a.active = true
and a.start_date::date <= now()::date
and end_date::date >= now()::date

union all

select a.id as id,
b.name || ' - ' || a.text as text,
a.priority as priority
from safe_stop_broadcast_messages a 
join jurisdictions b on b.id = a.jurisdiction_id
where 1 = 1
and a.bus_route_id is null
and a.bus_route_stop_id is null
and a.bus_id is null
and a.active = true
and a.start_date::date <= now()::date
and a.end_date::date >= now()::date

union all 

select a.id as id,
'Route: ' || INITCAP(b.display_name) || ' - ' || a.text as text,
a.priority as priority
from safe_stop_broadcast_messages a 
join bus_routes b on b.id = a.bus_route_id
where 1 = 1
and a.jurisdiction_id is null
and a.bus_route_stop_id is null
and a.bus_id is null
and a.active = true
and a.start_date::date <= now()::date
and a.end_date::date >= now()::date

union all

select a.id as id,
'Stop: ' || INITCAP(b.display_name) || ' - ' || a.text as text,
a.priority as priority
from safe_stop_broadcast_messages a 
join bus_route_stops b on b.id = a.bus_route_stop_id
where 1 = 1
and a.jurisdiction_id is null
and a.bus_route_id is null
and a.bus_id is null
and a.active = true
and a.start_date::date <= now()::date
and a.end_date::date >= now()::date

union all

select a.id as id,
'Bus: ' || INITCAP(b.name) || ' - ' || a.text as text,
a.priority as priority
from safe_stop_broadcast_messages a 
join buses b on b.id = a.bus_id
where 1 = 1
and a.jurisdiction_id is null
and a.bus_route_id is null
and a.bus_route_stop_id is null
and a.active = true
and a.start_date::date <= now()::date
and a.end_date::date >= now()::date

`
	rows, err := database.GetDB().Queryx(sql)
	if err != nil {
		log.Println(err.Error())
		return r
	}

	for i := 0; rows.Next(); i++ {
		a := Alert{}
		err := rows.StructScan(&a)
		if err != nil {
			log.Println(err.Error())
			return r
		}
		r = append(r, a)
	}

	return r
}



func AlertsForSchoolAdminAndRegularUsers(u *User, pg *PermissionGroups) []Alert{

	r := []Alert{}

	sql := `
select a.id as id,
'All - ' || a.text as text,
a.priority as priority
from safe_stop_broadcast_messages a
where jurisdiction_id is null
and a.bus_route_id is null
and a.bus_route_stop_id is null
and a.bus_id is null
and a.active = true
and a.active = true
and a.start_date::date <= now()::date
and end_date::date >= now()::date `


	if UsersClientJurisdictionCount(u,pg) > 0 {

		sql = sql + fmt.Sprintf(`

union all

select a.id as id,
b.name || ' - ' || a.text as text,
a.priority as priority
from safe_stop_broadcast_messages a 
join jurisdictions b on b.id = a.jurisdiction_id
where 1 = 1
and a.bus_route_id is null
and a.bus_route_stop_id is null
and a.bus_id is null
and a.active = true
and a.start_date::date <= now()::date
and a.end_date::date >= now()::date
and b.id in (%s)

union all 

select a.id as id,
'Route: ' || INITCAP(b.display_name) || ' - ' || a.text as text,
a.priority as priority
from safe_stop_broadcast_messages a 
join bus_routes b on b.id = a.bus_route_id
join bus_route_stops c on c.bus_route_id = b.id
where 1 = 1
and a.jurisdiction_id is null
and a.bus_route_stop_id is null
and a.bus_id is null
and a.active = true
and a.start_date::date <= now()::date
and a.end_date::date >= now()::date
and c.id in (select bus_route_stop_id from bus_route_stop_users where user_id = %d)

union all

select a.id as id,
'Stop: ' || INITCAP(b.display_name) || ' - ' || a.text as text,
a.priority as priority
from safe_stop_broadcast_messages a 
join bus_route_stops b on b.id = a.bus_route_stop_id
where 1 = 1
and a.jurisdiction_id is null
and a.bus_route_id is null
and a.bus_id is null
and a.active = true
and a.start_date::date <= now()::date
and a.end_date::date >= now()::date
and b.id in (select bus_route_stop_id from bus_route_stop_users where user_id = %d)

union all

select a.id as id,
'Bus: ' || INITCAP(b.name) || ' - ' || a.text as text,
a.priority as priority
from safe_stop_broadcast_messages a 
join buses b on b.id = a.bus_id
join bus_routes c on c.bus_id = b.id
join bus_route_stops d on d.bus_route_id = c.id
where 1 = 1
and a.jurisdiction_id is null
and a.bus_route_id is null
and a.bus_route_stop_id is null
and a.active = true
and a.start_date::date <= now()::date
and a.end_date::date >= now()::date
and d.id in (select bus_route_stop_id from bus_route_stop_users where user_id = %d)

`, UsersClientJurisdictionIds(u, pg), u.Id, u.Id, u.Id)
	}

	rows, err := database.GetDB().Queryx(sql)
	if err != nil {
		log.Println(err.Error())
		return r
	}

	for i := 0; rows.Next(); i++ {
		a := Alert{}
		err := rows.StructScan(&a)
		if err != nil {
			log.Println(err.Error())
			return r
		}
		r = append(r, a)
	}
	return r
}


func ViewedAlertIdsForUserId(user_id int) []int{
	ids := []int{}

	sql := `select safe_stop_broadcast_message_id from viewed_safe_stop_broadcast_messages where user_id = $1 order by id desc`

	rows, err := database.GetDB().Queryx(sql, user_id)
	if err != nil {
		log.Println(err.Error())
		return ids
	}

	for i := 0; rows.Next(); i++ {
		id := 0
		err := rows.Scan(&id)
		if err != nil {
			log.Println(err.Error())
			return ids
		}
		ids = append(ids, id)
	}
	return ids
}


func SetUsersViewedAlerts(u *User, alert_ids string) {

	if alert_ids != "" {
		ids := strings.Split(alert_ids, ",")
		for x := 0; x < len(ids); x++ {

			id, err := strconv.Atoi(ids[x])
			if err != nil {
				log.Println(err.Error())
				continue
			}

			sql := `
insert into viewed_safe_stop_broadcast_messages (user_id, safe_stop_broadcast_message_id, created_at, updated_at)
select $1, $2, now(), now()
where not exists ( 
	select id from viewed_safe_stop_broadcast_messages where user_id = $1 and safe_stop_broadcast_message_id = $2
)
`
			_, err = database.GetDB().Exec(sql, u.Id, id)
			if err != nil {
				log.Println(err.Error())
				continue
			}

		}
	}
}




type AlertsJurisdiction struct {
	Id int `db:"id"`
	Name string `db:"name"`
}


func AlertsJurisdictionsForUser(user_id int, search string) *[]AlertsJurisdiction{

	where_sql := " and $2 = $2"
	if search != "" {
		where_sql = " and lower(a.name) like '%' || lower($2) || '%'"
	}

	r := []AlertsJurisdiction{}

	sql := `
select a.id, a.name from jurisdictions a 
join jurisdictional_restrictions b on a.id = b.jurisdiction_id
where b.user_id = $1
and a.active = true
` + where_sql + `
order by a.name
`

	rows, err := database.GetDB().Queryx(sql, user_id, search)
	if err != nil {
		log.Println(err.Error())
		return &r
	}

	for i := 0; rows.Next(); i++ {
		a := AlertsJurisdiction{}
		err := rows.StructScan(&a)
		if err != nil {
			log.Println(err.Error())
			return &r
		}
		r = append(r, a)
	}
	return &r
}


type AlertsBus struct {
	Id int `db:"id"`
	Name string `db:"name"`
	ConfigName string `db:"config_name"`
}


func AlertsBusesForUser(user_id int, search string) *[]AlertsBus{

	where_sql := " and $2 = $2"
	if search != "" {
		where_sql = " and lower(d.name) like '%' || lower($2) || '%'"
	}

	r := []AlertsBus{}

	sql := `
select d.id, d.name, c.name as config_name
from jurisdictional_restrictions a 
join gps_configs_jurisdictions b on b.jurisdiction_id = a.jurisdiction_id
join gps_configs c on c.id = b.gps_config_id
join buses d on d.gps_config_id = c.id
where a.user_id = $1
and d.hide = false
` + where_sql + `
order by d.name
`

	rows, err := database.GetDB().Queryx(sql, user_id, search)
	if err != nil {
		log.Println(err.Error())
		return &r
	}

	for i := 0; rows.Next(); i++ {
		a := AlertsBus{}
		err := rows.StructScan(&a)
		if err != nil {
			log.Println(err.Error())
			return &r
		}
		r = append(r, a)
	}
	return &r
}

type AlertsRouteDB struct {
	RouteId int `db:"route_id"`
	RouteName string `db:"route_name"`
	StopId int `db:"stop_id"`
	StopName string `db:"stop_name"`
	StopScheduledTime string `db:"scheduled_time"`
	Bus string `db:"bus"`
}

type AlertsRoute struct {
	RouteId int
	RouteName string
	Bus string
	Stops []AlertsStop
}

type AlertsStop struct {
	StopId int
	StopName string
	StopScheduledTime string
}




func AlertsRoutesForUser(user_id int, search string) *[]AlertsRoute{
	r := []AlertsRoute{}
	rdb := []AlertsRouteDB{}

	where_sql := " and $2 = $2"
	if search != "" {
		where_sql = " and (lower(r.name) like '%' || lower($2) || '%' or lower(b.name) like '%' || lower($2) || '%')"
	}

	sql := `
select r.id as route_id,
r.display_name as route_name,
s.id as stop_id,
s.display_name as stop_name,
to_char(now()::date + (s.scheduled_time_offset + r.start_time) * interval '1 second', 'hh12:mi am') as scheduled_time,
b.name as bus
from bus_routes r 
join bus_route_stops s on s.bus_route_id = r.id
join buses b on b.id = r.bus_id
join gps_configs g on g.id = b.gps_config_id
join jurisdictional_restrictions jr on jr.jurisdiction_id = r.jurisdiction_id
where 1 = 1
and jr.user_id = $1
and r.active = true
and r.deleted = false
and s.active = true
and r.deleted = false
` + where_sql + `
order by r.id, s.scheduled_time_offset
`
	rows, err := database.GetDB().Queryx(sql, user_id, search)
	if err != nil {
		log.Println(err.Error())
		return &r
	}
	for i := 0; rows.Next(); i++ {
		a := AlertsRouteDB{}
		err := rows.StructScan(&a)
		if err != nil {
			log.Println(err.Error())
			return &r
		}
		rdb = append(rdb, a)
	}


	current_route := AlertsRoute{}
	current_stop := AlertsStop{}

	for i := 0; i < len(rdb); i++ {


		if current_route.RouteName == "" {

			current_route.RouteId = rdb[i].RouteId
			current_route.RouteName = rdb[i].RouteName
			current_route.Bus = rdb[i].Bus
			current_route.Stops = []AlertsStop{}

		} else if current_route.RouteName != rdb[i].RouteName {

			r = append(r, current_route)
			current_route = AlertsRoute{}
			current_route.RouteId = rdb[i].RouteId
			current_route.Bus = rdb[i].Bus
			current_route.RouteName = rdb[i].RouteName
			current_route.Stops = []AlertsStop{}

		}

		current_stop = AlertsStop{}
		current_stop.StopId = rdb[i].StopId
		current_stop.StopName = rdb[i].StopName
		current_stop.StopScheduledTime = rdb[i].StopScheduledTime
		current_route.Stops = append(current_route.Stops, current_stop)

	}
	r = append(r, current_route)

	return &r
}


func InsertAlerts(user *User, ids string, priority string, start_date string, end_date string, text string, alert_for string) bool {

	r := false

	if len(priority) == 0 || len(text) == 0 || len(start_date) == 0 || len(end_date) == 0 {
		return r
	}

	id_array := strings.Split(ids, ",")
	for i := 0; i < len(id_array); i++ {

		id, err := strconv.Atoi(id_array[i])
		if err != nil {
			fmt.Println(err)
			continue
		}

		sql := `
insert into safe_stop_broadcast_messages (` + alert_for + `_id,user_id,start_date,end_date,priority,text,created_at,updated_at, active) 
values ($1,$2,$3,$4,$5,$6,now(),now(),true)
`
		_, err = database.GetDB().Exec(
			sql,
			id,
			user.Id,
			start_date,
			end_date,
			priority,
			text)
		if err != nil {
			fmt.Println(err)
			continue
		}
		r = true
	}
	return r
}

