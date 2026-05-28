package dao

import (
	"context"
	"database/sql"
	"filler/internal/database"
	"filler/internal/logger"
	"filler/internal/model"
)

func LoadEnumsFromDB(ctx context.Context) error {
	conn := database.Get()
	accessMap := make(map[int16]string)
	aRows, err := conn.Query(ctx, `SELECT id, value FROM public.access`)
	if err == nil {
		defer aRows.Close()
		for aRows.Next() {
			var id int16
			var val string
			_ = aRows.Scan(&id, &val)
			accessMap[id] = val
		}
	}
	varTypeMap := make(map[int16]string)
	vRows, err := conn.Query(ctx, `SELECT id, value FROM public.var_type`)
	if err == nil {
		defer vRows.Close()
		for vRows.Next() {
			var id int16
			var val string
			_ = vRows.Scan(&id, &val)
			varTypeMap[id] = val
		}
	}
	pollMap := make(map[int16]string)
	pRows, err := conn.Query(ctx, `SELECT id, value FROM public.polling_frequency`)
	if err == nil {
		defer pRows.Close()
		for pRows.Next() {
			var id int16
			var val string
			_ = pRows.Scan(&id, &val)
			pollMap[id] = val
		}
	}
	asn1Map := make(map[int16]string)
	asRows, err := conn.Query(ctx, `SELECT id, value FROM public.asn1_type`)
	if err == nil {
		defer asRows.Close()
		for asRows.Next() {
			var id int16
			var val string
			_ = asRows.Scan(&id, &val)
			asn1Map[id] = val
		}
	}
	statusMap := make(map[int16]string)
	stRows, err := conn.Query(ctx, `SELECT id, value FROM public.oid_status`)
	if err == nil {
		defer stRows.Close()
		for stRows.Next() {
			var id int16
			var val string
			_ = stRows.Scan(&id, &val)
			statusMap[id] = val
		}
	}
	oidAccessMap := make(map[int16]string)
	oaRows, err := conn.Query(ctx, `SELECT id, value FROM public.oid_access`)
	if err == nil {
		defer oaRows.Close()
		for oaRows.Next() {
			var id int16
			var val string
			_ = oaRows.Scan(&id, &val)
			oidAccessMap[id] = val
		}
	}
	logicMap := make(map[int16]string)
	loRows, err := conn.Query(ctx, `SELECT id, value FROM public.logic_operator`)
	if err == nil {
		defer loRows.Close()
		for loRows.Next() {
			var id int16
			var val string
			_ = loRows.Scan(&id, &val)
			logicMap[id] = val
		}
	}
	alarmMap := make(map[int16]string)
	alRows, err := conn.Query(ctx, `SELECT id, value FROM public.alarm_level`)
	if err == nil {
		defer alRows.Close()
		for alRows.Next() {
			var id int16
			var val string
			_ = alRows.Scan(&id, &val)
			alarmMap[id] = val
		}
	}
	var vendors []map[string]interface{}
	vdRows, err := conn.Query(ctx, `SELECT id, name, directory FROM public.vendor`)
	if err == nil {
		defer vdRows.Close()
		for vdRows.Next() {
			var id int64
			var name string
			var dir sql.NullString
			if err := vdRows.Scan(&id, &name, &dir); err == nil {
				vendors = append(vendors, map[string]interface{}{
					"id":        id,
					"name":      name,
					"directory": dir.String,
				})
			}
		}
	}
	model.LoadRegistries(
		accessMap,
		varTypeMap,
		pollMap,
		asn1Map,
		statusMap,
		oidAccessMap,
		logicMap,
		alarmMap,
		vendors,
	)
	logger.Info("Все системные справочники успешно загружены из БД в память приложения (%d access, %d types, %d frequencies, %d asn1, %d statuses, %d oid_access, %d logic_ops, %d alarms, %d vendors)",
		len(accessMap), len(varTypeMap), len(pollMap), len(asn1Map), len(statusMap), len(oidAccessMap), len(logicMap), len(alarmMap), len(vendors))
	return nil
}

func stringToNull(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}
