package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"filler/internal/database"
	"filler/internal/model"

	"github.com/jackc/pgx/v5"
)

func scanOidRows(pgxRows pgx.Rows) ([]model.Oid, error) {
	var oids []model.Oid
	for pgxRows.Next() {
		var o model.Oid
		var mib sql.NullInt64
		var num sql.NullInt32
		var syn, units, desc, cat sql.NullString
		var stRaw, acRaw sql.NullInt16
		var tRaw int16
		var enumBytes []byte
		err := pgxRows.Scan(
			&o.ID, &mib, &tRaw, &o.Name, &num, &o.DotterNotation, &o.ObjectDescriptor,
			&syn, &enumBytes, &stRaw, &acRaw, &units, &desc, &cat,
		)
		if err != nil {
			return nil, err
		}
		o.Type = model.Asn1Type(tRaw)
		if mib.Valid {
			o.MibID = &mib.Int64
		}
		if num.Valid {
			o.Number = &num.Int32
		}
		o.Syntax = syn.String
		if len(enumBytes) > 0 {
			o.Enum = json.RawMessage(enumBytes)
		}
		o.Units = units.String
		o.Description = desc.String
		o.Category = cat.String

		if stRaw.Valid {
			st := model.OidStatus(stRaw.Int16)
			o.Status = &st
		}
		if acRaw.Valid {
			ac := model.OidAccess(acRaw.Int16)
			o.Access = &ac
		}
		oids = append(oids, o)
	}
	return oids, nil
}

func GetOidsByExactDotter(ctx context.Context, dotter string) ([]model.Oid, error) {
	conn := database.Get()
	query := `SELECT id, mib, type, name, number, dotter_notation, object_descriptor, syntax, enum, status, access, units, description, category FROM public.oid WHERE dotter_notation = $1`
	rows, err := conn.Query(ctx, query, dotter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOidRows(rows)
}

func GetOidsByDotterPrefix(ctx context.Context, prefix string) ([]model.Oid, error) {
	conn := database.Get()
	query := `SELECT id, mib, type, name, number, dotter_notation, object_descriptor, syntax, enum, status, access, units, description, category FROM public.oid WHERE dotter_notation LIKE $1`
	rows, err := conn.Query(ctx, query, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOidRows(rows)
}

func GetOidsByMibName(ctx context.Context, mibName string) ([]model.Oid, error) {
	conn := database.Get()
	query := `
		SELECT o.id, o.mib, o.type, o.name, o.number, o.dotter_notation, o.object_descriptor, o.syntax, o.enum, o.status, o.access, o.units, o.description, o.category 
		FROM public.oid o
		JOIN public.mib m ON o.mib = m.id
		WHERE m.name = $1`
	rows, err := conn.Query(ctx, query, mibName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOidRows(rows)
}

func GetOidsByVendorID(ctx context.Context, vID model.VendorID) ([]model.Oid, error) {
	conn := database.Get()
	query := `
		SELECT o.id, o.mib, o.type, o.name, o.number, o.dotter_notation, o.object_descriptor, o.syntax, o.enum, o.status, o.access, o.units, o.description, o.category 
		FROM public.oid o
		JOIN public.mib m ON o.mib = m.id
		WHERE m.vendor = $1`
	rows, err := conn.Query(ctx, query, int64(vID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOidRows(rows)
}

// GetOidsByDotterAndMibName возвращает OID по точному совпадению dotter_notation и имени MIB
func GetOidsByDotterAndMibName(ctx context.Context, dotter string, mibName string) ([]model.Oid, error) {
	conn := database.Get()
	query := `
		SELECT o.id, o.mib, o.type, o.name, o.number, o.dotter_notation, o.object_descriptor, o.syntax, o.enum, o.status, o.access, o.units, o.description, o.category 
		FROM public.oid o
		JOIN public.mib m ON o.mib = m.id
		WHERE o.dotter_notation = $1 AND m.name = $2`
	rows, err := conn.Query(ctx, query, dotter, mibName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOidRows(rows)
}

// GetOidsByDotterMibAndVendor возвращает OID по точному совпадению dotter_notation, имени MIB и имени/директории вендора (поддерживает указатель на string для NULL)
func GetOidsByDotterMibAndVendor(ctx context.Context, dotter string, mibName string, vendorIdentity *string) ([]model.Oid, error) {
	conn := database.Get()
	var rows pgx.Rows
	var err error
	if vendorIdentity == nil {
		query := `
			SELECT o.id, o.mib, o.type, o.name, o.number, o.dotter_notation, o.object_descriptor, o.syntax, o.enum, o.status, o.access, o.units, o.description, o.category 
			FROM public.oid o
			JOIN public.mib m ON o.mib = m.id
			WHERE o.dotter_notation = $1 AND m.name = $2 AND m.vendor IS NULL`
		rows, err = conn.Query(ctx, query, dotter, mibName)
	} else {
		query := `
			SELECT o.id, o.mib, o.type, o.name, o.number, o.dotter_notation, o.object_descriptor, o.syntax, o.enum, o.status, o.access, o.units, o.description, o.category 
			FROM public.oid o
			JOIN public.mib m ON o.mib = m.id
			JOIN public.vendor v ON m.vendor = v.id
			WHERE o.dotter_notation = $1 AND m.name = $2 AND (v.name = $3 OR v.directory = $3)`
		rows, err = conn.Query(ctx, query, dotter, mibName, *vendorIdentity)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOidRows(rows)
}
