package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"filler/internal/database"
	"filler/internal/dto"
	"filler/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func CreateParamIndicator(ctx context.Context, d dto.ParamIndicatorCreate) (*model.ParamIndicator, error) {
	conn := database.Get()
	query := `
		INSERT INTO public.param_indicator (oid_id, dotter_notation)
		VALUES ($1, $2)
		RETURNING id, oid_id, dotter_notation`
	var pi model.ParamIndicator
	var dotter sql.NullString
	var oID sql.NullString
	var oMib sql.NullInt64
	var oNum sql.NullInt32
	var oType sql.NullInt16
	var oName, oDotter, oDesc, oSyn, oUnits, oCat, oObj sql.NullString
	var oStRaw, oAcRaw sql.NullInt16
	var oEnumBytes []byte
	err := conn.QueryRow(ctx, query, d.OidID, toNullString(d.DotterNotation)).
		Scan(&pi.ID, &pi.OidID, &dotter,
			&oID, &oMib, &oType, &oName, &oNum, &oDotter, &oObj, &oSyn, &oEnumBytes, &oStRaw, &oAcRaw, &oUnits, &oDesc, &oCat)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if dotter.Valid {
		pi.DotterNotation = &dotter.String
	}
	if oID.Valid && oID.String != "" {
		parsedUUID, err := uuid.Parse(oID.String)
		if err == nil {
			o := model.Oid{ID: parsedUUID}
			o.Type = model.Asn1Type(oType.Int16)
			if oMib.Valid {
				o.MibID = &oMib.Int64
			}
			if oNum.Valid {
				o.Number = &oNum.Int32
			}
			o.Name = oName.String
			o.DotterNotation = oDotter.String
			o.ObjectDescriptor = oObj.String
			o.Syntax = oSyn.String
			if len(oEnumBytes) > 0 {
				o.Enum = json.RawMessage(oEnumBytes)
			}
			o.Units = oUnits.String
			o.Description = oDesc.String
			o.Category = oCat.String
			if oStRaw.Valid {
				st := model.OidStatus(oStRaw.Int16)
				o.Status = &st
			}
			if oAcRaw.Valid {
				ac := model.OidAccess(oAcRaw.Int16)
				o.Access = &ac
			}
			pi.Oid = &o
		}
	}
	return &pi, nil
}

func GetParamIndicatorByID(ctx context.Context, id int64) (*model.ParamIndicator, error) {
	conn := database.Get()
	query := `
		SELECT pi.id, pi.oid_id, pi.dotter_notation,
		       o.id, o.mib, o.type, o.name, o.number, o.dotter_notation, o.object_descriptor, o.syntax, o.enum, o.status, o.access, o.units, o.description, o.category
		FROM public.param_indicator pi
		LEFT JOIN public.oid o ON pi.oid_id = o.id
		WHERE pi.id = $1`
	var pi model.ParamIndicator
	var dotter sql.NullString
	var oID sql.NullString
	var oMib sql.NullInt64
	var oNum sql.NullInt32
	var oType sql.NullInt16
	var oName, oDotter, oDesc, oSyn, oUnits, oCat, oObj sql.NullString
	var oStRaw, oAcRaw sql.NullInt16
	var oEnumBytes []byte
	err := conn.QueryRow(ctx, query, id).Scan(
		&pi.ID, &pi.OidID, &dotter,
		&oID, &oMib, &oType, &oName, &oNum, &oDotter, &oObj, &oSyn, &oEnumBytes, &oStRaw, &oAcRaw, &oUnits, &oDesc, &oCat,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if dotter.Valid {
		pi.DotterNotation = &dotter.String
	}
	if oID.Valid && oID.String != "" {
		parsedUUID, err := uuid.Parse(oID.String)
		if err == nil {
			o := model.Oid{ID: parsedUUID}
			o.Type = model.Asn1Type(oType.Int16)
			if oMib.Valid {
				o.MibID = &oMib.Int64
			}
			if oNum.Valid {
				o.Number = &oNum.Int32
			}
			o.Name = oName.String
			o.DotterNotation = oDotter.String
			o.ObjectDescriptor = oObj.String
			o.Syntax = oSyn.String
			if len(oEnumBytes) > 0 {
				o.Enum = json.RawMessage(oEnumBytes)
			}
			o.Units = oUnits.String
			o.Description = oDesc.String
			o.Category = oCat.String
			if oStRaw.Valid {
				st := model.OidStatus(oStRaw.Int16)
				o.Status = &st
			}
			if oAcRaw.Valid {
				ac := model.OidAccess(oAcRaw.Int16)
				o.Access = &ac
			}
			pi.Oid = &o
		}
	}
	return &pi, nil
}

func GetAllParamIndicators(ctx context.Context) ([]model.ParamIndicator, error) {
	conn := database.Get()
	query := `
		SELECT pi.id, pi.oid_id, pi.dotter_notation,
		       o.id, o.mib, o.type, o.name, o.number, o.dotter_notation, o.object_descriptor, o.syntax, o.enum, o.status, o.access, o.units, o.description, o.category
		FROM public.param_indicator pi
		LEFT JOIN public.oid o ON pi.oid_id = o.id`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.ParamIndicator
	for rows.Next() {
		var pi model.ParamIndicator
		var dotter sql.NullString
		var oID sql.NullString
		var oMib sql.NullInt64
		var oNum sql.NullInt32
		var oType sql.NullInt16
		var oName, oDotter, oDesc, oSyn, oUnits, oCat, oObj sql.NullString
		var oStRaw, oAcRaw sql.NullInt16
		var oEnumBytes []byte
		err := rows.Scan(
			&pi.ID, &pi.OidID, &dotter,
			&oID, &oMib, &oType, &oName, &oNum, &oDotter, &oObj, &oSyn, &oEnumBytes, &oStRaw, &oAcRaw, &oUnits, &oDesc, &oCat,
		)
		if err != nil {
			return nil, err
		}
		if dotter.Valid {
			pi.DotterNotation = &dotter.String
		}
		if oID.Valid && oID.String != "" {
			parsedUUID, err := uuid.Parse(oID.String)
			if err == nil {
				o := model.Oid{ID: parsedUUID}
				o.Type = model.Asn1Type(oType.Int16)
				if oMib.Valid {
					o.MibID = &oMib.Int64
				}
				if oNum.Valid {
					o.Number = &oNum.Int32
				}
				o.Name = oName.String
				o.DotterNotation = oDotter.String
				o.ObjectDescriptor = oObj.String
				o.Syntax = oSyn.String
				if len(oEnumBytes) > 0 {
					o.Enum = json.RawMessage(oEnumBytes)
				}
				o.Units = oUnits.String
				o.Description = oDesc.String
				o.Category = oCat.String
				if oStRaw.Valid {
					st := model.OidStatus(oStRaw.Int16)
					o.Status = &st
				}
				if oAcRaw.Valid {
					ac := model.OidAccess(oAcRaw.Int16)
					o.Access = &ac
				}
				pi.Oid = &o
			}
		}
		list = append(list, pi)
	}
	return list, nil
}

func UpdateParamIndicator(ctx context.Context, id int64, d dto.ParamIndicatorUpdate) (*model.ParamIndicator, error) {
	conn := database.Get()
	query := `
		UPDATE public.param_indicator 
		SET oid_id = $1, dotter_notation = $2
		WHERE id = $3
		RETURNING id, oid_id, dotter_notation`
	var pi model.ParamIndicator
	var dotter sql.NullString
	err := conn.QueryRow(ctx, query, d.OidID, toNullString(d.DotterNotation), id).
		Scan(&pi.ID, &pi.OidID, &dotter)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if dotter.Valid {
		pi.DotterNotation = &dotter.String
	}
	return &pi, nil
}

func DeleteParamIndicator(ctx context.Context, id int64) (bool, error) {
	conn := database.Get()
	query := `DELETE FROM public.param_indicator WHERE id = $1`
	commandTag, err := conn.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return commandTag.RowsAffected() > 0, nil
}
