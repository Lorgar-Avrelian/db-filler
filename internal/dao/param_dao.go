package dao

import (
	"context"
	"database/sql"
	"errors"
	"filler/internal/database"
	"filler/internal/dto"
	"filler/internal/model"

	"github.com/jackc/pgx/v5"
)

func CreateParam(ctx context.Context, d dto.ParamCreate) (*model.Param, error) {
	conn := database.Get()
	query := `
		INSERT INTO public.param (title, name_en, name_ru, type, value, description_en, description_ru, units_en, units_ru, access, saved, visible)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, title, name_en, name_ru, type, value, description_en, description_ru, units_en, units_ru, access, saved, visible`
	typeIdx := model.ParseVarType(d.Type)
	accessIdx := model.ParseAccess(d.Access)
	var p model.Param
	var valNull, descEn, descRu, unEn, unRu sql.NullString
	var tRaw, aRaw int16
	err := conn.QueryRow(ctx, query, d.Title, d.NameEn, d.NameRu, int16(typeIdx), stringToNull(d.Value), d.DescriptionEn, d.DescriptionRu, stringToNull(d.UnitsEn), stringToNull(d.UnitsRu), int16(accessIdx), d.Saved, d.Visible).
		Scan(&p.ID, &p.Title, &p.NameEn, &p.NameRu, &tRaw, &valNull, &descEn, &descRu, &unEn, &unRu, &aRaw, &p.Saved, &p.Visible)
	if err != nil {
		return nil, err
	}
	p.Type = model.VarType(tRaw)
	p.Access = model.Access(aRaw)
	p.DescriptionEn = descEn.String
	p.DescriptionRu = descRu.String
	p.UnitsEn = unEn.String
	p.UnitsRu = unRu.String
	return &p, nil
}

func GetParamByID(ctx context.Context, id int64) (*model.Param, error) {
	conn := database.Get()
	query := `SELECT id, title, name_en, name_ru, type, value, description_en, description_ru, units_en, units_ru, access, saved, visible FROM public.param WHERE id = $1`
	var p model.Param
	var valNull, descEn, descRu, unEn, unRu sql.NullString
	var tRaw, aRaw int16
	err := conn.QueryRow(ctx, query, id).
		Scan(&p.ID, &p.Title, &p.NameEn, &p.NameRu, &tRaw, &valNull, &descEn, &descRu, &unEn, &unRu, &aRaw, &p.Saved, &p.Visible)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	p.Type = model.VarType(tRaw)
	p.Access = model.Access(aRaw)
	p.DescriptionEn = descEn.String
	p.DescriptionRu = descRu.String
	p.UnitsEn = unEn.String
	p.UnitsRu = unRu.String
	return &p, nil
}

func UpdateParam(ctx context.Context, id int64, d dto.ParamUpdate) (*model.Param, error) {
	conn := database.Get()
	query := `
		UPDATE public.param 
		SET title = $1, name_en = $2, name_ru = $3, type = $4, value = $5, description_en = $6, description_ru = $7, units_en = $8, units_ru = $9, access = $10, saved = $11, visible = $12
		WHERE id = $13
		RETURNING id, title, name_en, name_ru, type, value, description_en, description_ru, units_en, units_ru, access, saved, visible`
	typeIdx := model.ParseVarType(d.Type)
	accessIdx := model.ParseAccess(d.Access)
	var p model.Param
	var valNull, descEn, descRu, unEn, unRu sql.NullString
	var tRaw, aRaw int16
	err := conn.QueryRow(ctx, query, d.Title, d.NameEn, d.NameRu, int16(typeIdx), stringToNull(d.Value), stringToNull(d.DescriptionEn), stringToNull(d.DescriptionRu), stringToNull(d.UnitsEn), stringToNull(d.UnitsRu), int16(accessIdx), d.Saved, d.Visible, id).
		Scan(&p.ID, &p.Title, &p.NameEn, &p.NameRu, &tRaw, &valNull, &descEn, &descRu, &unEn, &unRu, &aRaw, &p.Saved, &p.Visible)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	p.Type = model.VarType(tRaw)
	p.Access = model.Access(aRaw)
	p.DescriptionEn = descEn.String
	p.DescriptionRu = descRu.String
	p.UnitsEn = unEn.String
	p.UnitsRu = unRu.String
	return &p, nil
}

func GetUnattachedParams(ctx context.Context) ([]model.Param, error) {
	conn := database.Get()
	query := `
		SELECT p.id, p.title, p.name_en, p.name_ru, p.type, p.value, p.description_en, p.description_ru, p.units_en, p.units_ru, p.access, p.saved, p.visible
		FROM public.param p
		LEFT JOIN public.component_param cp ON p.id = cp.param_id
		WHERE cp.param_id IS NULL`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var params []model.Param
	for rows.Next() {
		var p model.Param
		var descEn, descRu, unEn, unRu sql.NullString
		var tRaw, aRaw int16
		err := rows.Scan(&p.ID, &p.Title, &p.NameEn, &p.NameRu, &tRaw, &p.Value, &descEn, &descRu, &unEn, &unRu, &aRaw, &p.Saved, &p.Visible)
		if err != nil {
			return nil, err
		}
		p.Type = model.VarType(tRaw)
		p.Access = model.Access(aRaw)
		p.DescriptionEn = descEn.String
		p.DescriptionRu = descRu.String
		p.UnitsEn = unEn.String
		p.UnitsRu = unRu.String
		params = append(params, p)
	}
	return params, nil
}

func SearchParams(ctx context.Context, q string) ([]model.Param, error) {
	conn := database.Get()
	query := `
		SELECT id, title, name_en, name_ru, type, value, description_en, description_ru, units_en, units_ru, access, saved, visible
		FROM public.param
		WHERE title ILIKE $1 
		   OR name_en ILIKE $1 
		   OR name_ru ILIKE $1 
		   OR description_en ILIKE $1 
		   OR description_ru ILIKE $1`
	likeQuery := "%" + q + "%"
	rows, err := conn.Query(ctx, query, likeQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var params []model.Param
	for rows.Next() {
		var p model.Param
		var descEn, descRu, unEn, unRu sql.NullString
		var tRaw, aRaw int16
		err := rows.Scan(&p.ID, &p.Title, &p.NameEn, &p.NameRu, &tRaw, &p.Value, &descEn, &descRu, &unEn, &unRu, &aRaw, &p.Saved, &p.Visible)
		if err != nil {
			return nil, err
		}
		p.Type = model.VarType(tRaw)
		p.Access = model.Access(aRaw)
		p.DescriptionEn = descEn.String
		p.DescriptionRu = descRu.String
		p.UnitsEn = unEn.String
		p.UnitsRu = unRu.String
		params = append(params, p)
	}
	return params, nil
}

func DeleteParam(ctx context.Context, id int64) (bool, error) {
	conn := database.Get()
	query := `DELETE FROM public.param WHERE id = $1`
	commandTag, err := conn.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return commandTag.RowsAffected() > 0, nil
}
