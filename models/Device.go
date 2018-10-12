package models

import (
	"fmt"
	"github.com/schoolwheels/safestopclient/database"
	"log"
	"strconv"
	"strings"
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






func DevicesForJurisdictions(ids string) *[]string {

	r := []string{}

	id_array := strings.Split(ids, ",")
	safe_id_string := "-1"
	for i := 0; i < len(id_array); i++ {
		n, err := strconv.Atoi(id_array[i])
		if err != nil {
			continue
		}
		if n > 0 {
			safe_id_string = fmt.Sprintf("%s,%d", safe_id_string, n)
		}
	}

	if safe_id_string == "-1" {
		return &r
	}

	sql := `

select sns_arn from (

select j.sns_arn
from users a
join subscriptions b
on a.id = b.user_id
join products c
on c.id = b.product_id

join security_segments i
on i.id = a.security_segment_id
join devices j
on j.user_id = a.id
where c.jurisdiction_id in (` + safe_id_string + `)
and b.active = true
and c.product_type = 'ss'
and now()::date >= b.start_date::date
and now()::date <= b.end_date::date
and i.name = 'SafeStop'
and (a.locked is null or a.locked = false)

union all

select l.sns_arn
from users a
join subscription_sub_accounts b
on b.user_id = a.id
join subscriptions c
on c.id = b.subscription_id
join users d
on d.id = c.user_id
join products e
on e.id = c.product_id

join security_segments k
on k.id = a.security_segment_id
join devices l
on l.user_id = a.id
where e.jurisdiction_id in (` + safe_id_string + `)
and c.active = true
and e.product_type = 'ss'
and now()::date >= c.start_date::date
and now()::date <= c.end_date::date
and k.name = 'SafeStop'
and (a.locked is null or a.locked = false)

union all 

select d.sns_arn 
from devices d
join users u
on u.id = d.user_id
join security_segments ss
on ss.id = u.security_segment_id
join permission_groups_users pguss
on pguss.user_id = u.id
join permission_groups pg
on pg.id = pguss.permission_group_id
where 1=1
and (ss.name = 'SafeStop'
and (u.locked is null or u.locked = false)
and pg.name in ('SafeStop Admin', 'Super Admin'))
or u.super_admin = true

union all 

select d.sns_arn
from devices d
join users u
on u.id = d.user_id
join security_segments ss
on ss.id = u.security_segment_id
join permission_groups_users pguss
on pguss.user_id = u.id
join permission_groups pg
on pg.id = pguss.permission_group_id
join jurisdictional_restrictions jr on jr.user_id = u.id
where 1=1
and jr.jurisdiction_id in (` + safe_id_string + `)
and (ss.name = 'SafeStop'
and (u.locked is null or u.locked = false)
and pg.name in ('License 1 – Transportation Executive', 'License 2 – Transportation Professional'))
) z 
group by z.sns_arn

`
	rows, err := database.GetDB().Query(sql)
	if err != nil {
		return &r
	}

	if rows != nil {
		for rows.Next() {
			d := ""
			err = rows.Scan(&d)
			if err != nil {
				return &r
			}
			if len(d) > 0 {
				r = append(r, d)
			}
		}
	}


	return &r
}


func DevicesForBusIds(ids string) *[]string {

	r := []string{}

	id_array := strings.Split(ids, ",")
	safe_id_string := "-1"
	for i := 0; i < len(id_array); i++ {
		n, err := strconv.Atoi(id_array[i])
		if err != nil {
			continue
		}
		if n > 0 {
			safe_id_string = fmt.Sprintf("%s,%d", safe_id_string, n)
		}
	}

	if safe_id_string == "-1" {
		return &r
	}

	sql := `

select z.sns_arn from (

select j.sns_arn
from users a
join subscriptions b
on a.id = b.user_id
join products c
on c.id = b.product_id
join bus_route_stop_users d
on a.id = d.user_id
join bus_route_stops e
on e.id = d.bus_route_stop_id
join bus_routes f
on f.id = e.bus_route_id
join buses g
on g.id = f.bus_id
join security_segments i
on i.id = a.security_segment_id
join devices j
on j.user_id = a.id
where g.id in (` + safe_id_string + `)
and b.active = true
and now()::date >= b.start_date::date
and now()::date <= b.end_date::date
and i.name = 'SafeStop'
and (a.locked is null or a.locked = false)

union all

select l.sns_arn
from users a
join subscription_sub_accounts b
on b.user_id = a.id
join subscriptions c
on c.id = b.subscription_id
join users d
on d.id = c.user_id
join products e
on e.id = c.product_id
join bus_route_stop_users f
on f.user_id = a.id
join bus_route_stops g
on g.id = f.bus_route_stop_id
join bus_routes h
on h.id = g.bus_route_id
join buses i
on i.id = h.bus_id
join security_segments k
on k.id = a.security_segment_id
join devices l
on l.user_id = a.id
where i.id in (` + safe_id_string + `)
and c.active = true
and now()::date >= c.start_date::date
and now()::date <= c.end_date::date
and k.name = 'SafeStop'
and (a.locked is null or a.locked = false)

union all 

select d.sns_arn 
from devices d
join users u
on u.id = d.user_id
join security_segments ss
on ss.id = u.security_segment_id
join permission_groups_users pguss
on pguss.user_id = u.id
join permission_groups pg
on pg.id = pguss.permission_group_id
where 1=1
and (ss.name = 'SafeStop'
and (u.locked is null or u.locked = false)
and pg.name in ('SafeStop Admin', 'Super Admin'))
or u.super_admin = true

) z
group by sns_arn

`
	rows, err := database.GetDB().Query(sql)
	if err != nil {
		return &r
	}

	if rows != nil {
		for rows.Next() {
			d := ""
			err = rows.Scan(&d)
			if err != nil {
				return &r
			}
			if len(d) > 0 {
				r = append(r, d)
			}
		}
	}

	return &r
}



func DevicesForRouteAndStopIds(route_ids string, stop_ids string) *[]string {

	r := []string{}


	route_id_array := strings.Split(route_ids, ",")
	safe_route_id_string := "-1"
	for i := 0; i < len(route_id_array); i++ {
		n, err := strconv.Atoi(route_id_array[i])
		if err != nil {
			continue
		}
		if n > 0 {
			safe_route_id_string = fmt.Sprintf("%s,%d", safe_route_id_string, n)
		}
	}

	stop_id_array := strings.Split(stop_ids, ",")
	safe_stop_id_string := "-1"
	for i := 0; i < len(stop_id_array); i++ {
		n, err := strconv.Atoi(stop_id_array[i])
		if err != nil {
			continue
		}
		if n > 0 {
			safe_stop_id_string = fmt.Sprintf("%s,%d", safe_stop_id_string, n)
		}
	}

	Jurisdiction_id_string := JurisdictionIdsForRouteAndStopIds(safe_route_id_string, safe_stop_id_string)

	if len(safe_route_id_string) > 0 || len(safe_stop_id_string) > 0 {

		sql := `

select z.sns_arn from (

select j.sns_arn
from users a
join subscriptions b
on a.id = b.user_id
join products c
on c.id = b.product_id
join bus_route_stop_users d
on a.id = d.user_id
join bus_route_stops e
on e.id = d.bus_route_stop_id
join security_segments i
on i.id = a.security_segment_id
join devices j
on j.user_id = a.id
where e.id in (` + safe_stop_id_string + `)
and b.active = true
and now()::date >= b.start_date::date
and now()::date <= b.end_date::date
and i.name = 'SafeStop'
and (a.locked is null or a.locked = false)

union all

select l.sns_arn
from users a
join subscription_sub_accounts b
on b.user_id = a.id
join subscriptions c
on c.id = b.subscription_id
join users d
on d.id = c.user_id
join products e
on e.id = c.product_id
join bus_route_stop_users f
on f.user_id = a.id
join bus_route_stops g
on g.id = f.bus_route_stop_id
join security_segments k
on k.id = a.security_segment_id
join devices l
on l.user_id = a.id
where g.id in (` + safe_stop_id_string + `)
and c.active = true
and now()::date >= c.start_date::date
and now()::date <= c.end_date::date
and k.name = 'SafeStop'
and (a.locked is null or a.locked = false)

union all 

select d.sns_arn 
from devices d
join users u
on u.id = d.user_id
join security_segments ss
on ss.id = u.security_segment_id
join permission_groups_users pguss
on pguss.user_id = u.id
join permission_groups pg
on pg.id = pguss.permission_group_id
where 1=1
and (ss.name = 'SafeStop'
and (u.locked is null or u.locked = false)
and pg.name in ('SafeStop Admin', 'Super Admin'))
or u.super_admin = true

union all 

select d.sns_arn
from devices d
join users u
on u.id = d.user_id
join security_segments ss
on ss.id = u.security_segment_id
join permission_groups_users pguss
on pguss.user_id = u.id
join permission_groups pg
on pg.id = pguss.permission_group_id
join jurisdictional_restrictions jr on jr.user_id = u.id
where 1=1
and jr.jurisdiction_id in (` + Jurisdiction_id_string + `)
and (ss.name = 'SafeStop'
and (u.locked is null or u.locked = false)
and pg.name in ('License 1 – Transportation Executive', 'License 2 – Transportation Professional'))

union all

select j.sns_arn
from users a
join subscriptions b
on a.id = b.user_id
join products c
on c.id = b.product_id
join bus_route_stop_users d
on a.id = d.user_id
join bus_route_stops e
on e.id = d.bus_route_stop_id
join bus_routes f
on f.id = e.bus_route_id

join security_segments i
on i.id = a.security_segment_id
join devices j
on j.user_id = a.id
where f.id in (` + safe_route_id_string + `)
and b.active = true
and now()::date >= b.start_date::date
and now()::date <= b.end_date::date
and i.name = 'SafeStop'
and (a.locked is null or a.locked = false)

union all

select l.sns_arn
from users a
join subscription_sub_accounts b
on b.user_id = a.id
join subscriptions c
on c.id = b.subscription_id
join users d
on d.id = c.user_id
join products e
on e.id = c.product_id
join bus_route_stop_users f
on f.user_id = a.id
join bus_route_stops g
on g.id = f.bus_route_stop_id
join bus_routes h
on h.id = g.bus_route_id

join security_segments k
on k.id = a.security_segment_id
join devices l
on l.user_id = a.id
where h.id in (` + safe_route_id_string + `)
and c.active = true
and now()::date >= c.start_date::date
and now()::date <= c.end_date::date
and k.name = 'SafeStop'
and (a.locked is null or a.locked = false)

) z 
group by z.sns_arn

`
		rows, err := database.GetDB().Query(sql)
		if err != nil {
			return &r
		}

		if rows != nil {
			for rows.Next() {
				d := ""
				err = rows.Scan(&d)
				if err != nil {
					log.Println(err)
					return &r
				}
				if len(d) > 0 {
					r = append(r, d)
				}
			}
		}

	}

	return &r
}

