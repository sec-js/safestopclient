package models

import (
	"fmt"
	"github.com/schoolwheels/safestopclient/database"
	"log"
	"math"
	"strings"
)



type BusRouteSearchResult struct {
	BusRouteId int `json:"bus_route_id" db:"bus_route_id"`
	BusRouteName string `json:"bus_route_name" db:"bus_route_name"`
	BusRouteStartTime string `json:"bus_route_start_time" db:"bus_route_start_time"`
	JurisdictionId int `json:"jurisdiction_id" db:"jurisdiction_id"`
	JurisdictionName string `json:"jurisdiction_name" db:"jurisdiction_name"`
	SearchRadius float64 `json:"search_radius" db:"search_radius"`
}

func BusRoutesForAdminAndSchoolAdminUser(page int, search string, address_1 string, postal_code string, u *User, pg *PermissionGroups) interface{}{

	limit := 20
	offset := (page - 1) * limit
	parameters := []string{}

	r := struct {
		AccurateGeocoding bool `json:"accurate_geocoding"`
		BusRoutes []BusRouteSearchResult `json:"bus_routes"`
		Pages int `json:"pages"`

	} {
		true,
		[]BusRouteSearchResult{},
		1,
	}

	user_coordinates := Geocode(address_1, postal_code)

	geo_condition := ""
	if user_coordinates != nil {

		coordinate_string := fmt.Sprintf("%s %s")
		parameters = append(parameters, coordinate_string)
		param_index := indexOf(coordinate_string, parameters) + 1


		geo_condition = fmt.Sprintf( `
        AND (ST_Distance(
                   ST_Transform(ST_GeomFromText('POINT($%d)',4326),900913),
                   ST_Transform(ST_GeomFromText('POINT(' || CAST(case when e.adjusted_longitude is null then e.longitude else e.adjusted_longitude end as VARCHAR) || ' ' || CAST(case when e.adjusted_latitude is null then e.latitude else e.adjusted_latitude end as VARCHAR) || ')', 4326),900913)
               ) * 0.000621371) < b.search_radius
`, param_index)
	}

	search_condition := ""
		parameters = append(parameters, strings.ToLower("%" + search + "%"))
		param_index := indexOf(strings.ToLower("%" + search + "%"), parameters) + 1

		search_condition = fmt.Sprintf(`
and (lower(a.display_name) like $%d
or lower(b.name) like $%d)
`, param_index, param_index)

	total_sql := `
select count(distinct a.id)
from bus_routes a
join jurisdictions b
on a.jurisdiction_id = b.id
join buses c
on a.bus_id = c.id
join jurisdiction_safe_stop_infos d
on d.jurisdiction_id = b.id
join bus_route_stops e
on e.bus_route_id = a.id
where a.active = true
and a.bus_id is not null
and e.active = true
and e.deleted = false
and b.id in (` + UsersClientJurisdictionIdSQL(u, pg) + `)
` + geo_condition + search_condition


	s := make([]interface{}, len(parameters))
	for i, v := range parameters {
		s[i] = v
	}

	total_routes := 0
	row := database.GetDB().QueryRowx(total_sql, s...)

	if row == nil {

	} else {
		err := row.Scan(&total_routes)
		if err != nil {

		}
	}


	sql := fmt.Sprintf(`
select
a.id as bus_route_id,
case when b.titlecase = true then initcap(a.display_name) when b.titlecase = false then a.display_name end as bus_route_name,
to_char(now()::date + a.start_time * interval '1 second', 'HH12:MI AM')  as bus_route_start_time,
b.id as jurisdiction_id,
initcap(b.name) as jurisdiction_name,
b.search_radius
from bus_routes a
join jurisdictions b
on a.jurisdiction_id = b.id
join buses c
on a.bus_id = c.id
join jurisdiction_safe_stop_infos d
on d.jurisdiction_id = b.id
join bus_route_stops e
on e.bus_route_id = a.id
where a.active = true
and a.bus_id is not null
and e.active = true
and e.deleted = false
and b.id in (` + UsersClientJurisdictionIdSQL(u, pg) + `)
` + geo_condition + search_condition + `
group by a.id, a.display_name, a.start_time, b.id, b.name, b.search_radius
order by jurisdiction_name, bus_route_name
limit %d
offset %d
`, limit, offset)

	rows, err := database.GetDB().Queryx(sql, s...)
	if err != nil {
		log.Println(err.Error())
		return r
	}

	for i := 0; rows.Next(); i++ {
		br := BusRouteSearchResult{}
		err := rows.StructScan(&br)
		if err != nil {
			log.Println(err.Error())
			return r
		}
		r.BusRoutes = append(r.BusRoutes, br)
	}

	if total_routes > 0 {
		r.Pages = int(math.Ceil((float64(total_routes) * 1.0) / (float64(limit) * 1.0)))
	} else {
		r.Pages = 0
	}

	return r
}

func BusRoutesForRegularUsers(page int, address_1 string, postal_code string, u *User, pg *PermissionGroups) interface{}{

	limit := 20
	offset := (page - 1) * limit

	r := struct {
		AccurateGeocoding bool `json:"accurate_geocoding"`
		BusRoutes []BusRouteSearchResult `json:"bus_routes"`
		Pages int `json:"pages"`

	} {
		true,
		[]BusRouteSearchResult{},
		1,
	}

	user_coordinates := Geocode(address_1, postal_code)

	coordinate_string := "0 0"
	if user_coordinates != nil {
		coordinate_string = fmt.Sprintf("%s %s", user_coordinates.Longitude, user_coordinates.Latitude)
	}

	total_sql := `
select count(distinct a.id)
from bus_routes a
join jurisdictions b
on a.jurisdiction_id = b.id
join buses c
on a.bus_id = c.id
join jurisdiction_safe_stop_infos d
on d.jurisdiction_id = b.id
join bus_route_stops e
on e.bus_route_id = a.id
where a.active = true
and a.bus_id is not null
and e.active = true
and e.deleted = false
and b.id in (` + UsersClientJurisdictionIdSQL(u, pg) + `)
and d.status = 'live'
and (ST_Distance(
                   ST_Transform(ST_GeomFromText('POINT(' || $1 || ')',4326),900913),
                   ST_Transform(ST_GeomFromText('POINT(' || CAST(case when e.adjusted_longitude is null then e.longitude else e.adjusted_longitude end as VARCHAR) || ' ' || CAST(case when e.adjusted_latitude is null then e.latitude else e.adjusted_latitude end as VARCHAR) || ')', 4326),900913)
                 ) * 0.000621371) < b.search_radius
`


	total_routes := 0
	row := database.GetDB().QueryRowx(total_sql, coordinate_string)

	if row == nil {

	} else {
		err := row.Scan(&total_routes)
		if err != nil {

		}
	}


	sql := fmt.Sprintf(`
select
a.id as bus_route_id,
case when b.titlecase = true then initcap(a.display_name) when b.titlecase = false then a.display_name end as bus_route_name,
to_char(now()::date + a.start_time * interval '1 second', 'HH12:MI AM')  as bus_route_start_time,
b.id as jurisdiction_id,
initcap(b.name) as jurisdiction_name,
b.search_radius
from bus_routes a
join jurisdictions b
on a.jurisdiction_id = b.id
join buses c
on a.bus_id = c.id
join jurisdiction_safe_stop_infos d
on d.jurisdiction_id = b.id
join bus_route_stops e
on e.bus_route_id = a.id
where a.active = true
and a.bus_id is not null
and e.active = true
and e.deleted = false
and b.id in (` + UsersClientJurisdictionIdSQL(u, pg) + `)
and d.status = 'live'
and (ST_Distance(
                   ST_Transform(ST_GeomFromText('POINT(' || $1 || ')',4326),900913),
                   ST_Transform(ST_GeomFromText('POINT(' || CAST(case when e.adjusted_longitude is null then e.longitude else e.adjusted_longitude end as VARCHAR) || ' ' || CAST(case when e.adjusted_latitude is null then e.latitude else e.adjusted_latitude end as VARCHAR) || ')', 4326),900913)
                 ) * 0.000621371) < b.search_radius
group by a.id, a.display_name, a.start_time, b.id, b.name, b.search_radius
order by jurisdiction_name, bus_route_name
limit %d
offset %d
`, limit, offset)

	rows, err := database.GetDB().Queryx(sql, coordinate_string)
	if err != nil {
		log.Println(err.Error())
		return r
	}

	for i := 0; rows.Next(); i++ {
		br := BusRouteSearchResult{}
		err := rows.StructScan(&br)
		if err != nil {
			log.Println(err.Error())
			return r
		}
		r.BusRoutes = append(r.BusRoutes, br)
	}

	if total_routes > 0 {
		r.Pages = int(math.Ceil((float64(total_routes) * 1.0) / (float64(limit) * 1.0)))
	} else {
		r.Pages = 0
	}

	return r
}


type BusRouteStopSearchResult struct {
	Id int `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	ScheduledTime string `json:"scheduled_time" db:"scheduled_time"`
	RouteName string `json:"route_name" db:"route_name"`
	JurisdictionId int `json:"jurisdiction_id" db:"jurisdiction_id"`
	Selected bool `json:"selected" db:"selected"`
}




func BusRouteStopsForUser(user_id int, bus_route_id int) []BusRouteStopSearchResult{

	r := []BusRouteStopSearchResult{}

	sql := `
SELECT a.id as id,
case when c.titlecase = 't' then initcap(a.display_name) when c.titlecase = 'f' then a.display_name end as name,
to_char(now()::date + (b.start_time + a.scheduled_time_offset) * interval '1 second', 'HH12:MI AM')  as scheduled_time,
case when c.titlecase = 't' then initcap(b.display_name) when c.titlecase = 'f' then b.display_name end as route_name,
c.id as jurisdiction_id,
case when a.id in (select bus_route_stop_id from bus_route_stop_users where user_id = $1) then true else false end as selected
from bus_route_stops a
join bus_routes b
on a.bus_route_id = b.id
join jurisdictions c
on c.id = b.jurisdiction_id
where b.id = $2
and (a.adjusted_active is null and a.active = true or a.adjusted_active = true)
and a.deleted = false
and a.analytic_only = false
order by a.scheduled_time_offset
`

	rows, err := database.GetDB().Queryx(sql, user_id, bus_route_id)
	if err != nil {
		log.Println(err.Error())
	}

	for i := 0; rows.Next(); i++ {
		br := BusRouteStopSearchResult{}
		err := rows.StructScan(&br)
		if err != nil {
			log.Println(err.Error())
		}
		r = append(r, br)
	}

	return r
}