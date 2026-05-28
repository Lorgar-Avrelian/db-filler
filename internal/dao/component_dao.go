package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"filler/internal/database"
	"filler/internal/dto"
	"filler/internal/model"

	"github.com/jackc/pgx/v5"
)

func CreateComponent(ctx context.Context, d dto.ComponentCreate) (*model.Component, error) {
	conn := database.Get()
	query := `
		INSERT INTO public.component (title, name_en, name_ru, base_component, description_en, description_ru, access)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, name_en, name_ru, access`
	var baseComp sql.NullInt64
	if d.BaseComponent != nil {
		baseComp.Int64 = *d.BaseComponent
		baseComp.Valid = true
	}
	accessIdx := model.ParseAccess(d.Access)
	var c model.Component
	var aRaw int16
	err := conn.QueryRow(ctx, query, d.Title, d.NameEn, d.NameRu, baseComp, stringToNull(d.DescriptionEn), stringToNull(d.DescriptionRu), int16(accessIdx)).
		Scan(&c.ID, &c.Title, &c.NameEn, &c.NameRu, &aRaw)
	if err != nil {
		return nil, err
	}
	c.Access = model.Access(aRaw)
	return &c, nil
}

func GetComponentByID(ctx context.Context, id int64) (*model.Component, error) {
	conn := database.Get()
	compQuery := `SELECT id, title, name_en, name_ru, access FROM public.component WHERE id = $1`
	var c model.Component
	var aRaw int16
	err := conn.QueryRow(ctx, compQuery, id).Scan(&c.ID, &c.Title, &c.NameEn, &c.NameRu, &aRaw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	c.Access = model.Access(aRaw)
	paramQuery := `
		WITH RECURSIVE component_hierarchy AS (
			SELECT id, base_component FROM public.component WHERE id = $1
			UNION ALL
			SELECT c.id, c.base_component 
			FROM public.component c
			JOIN component_hierarchy ch ON c.id = ch.base_component
		)
		SELECT p.id, p.title, p.name_en, p.name_ru, p.type, p.value, p.description_en, p.description_ru, p.units_en, p.units_ru, p.access, p.saved, p.visible
		FROM public.param p
		JOIN public.component_param cp ON p.id = cp.param_id
		JOIN component_hierarchy ch ON cp.component_id = ch.id`
	rows, err := conn.Query(ctx, paramQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var params []model.Param
	for rows.Next() {
		var p model.Param
		var valNull, descEn, descRu, unEn, unRu sql.NullString
		var tRaw, paRaw int16
		err := rows.Scan(&p.ID, &p.Title, &p.NameEn, &p.NameRu, &tRaw, &valNull, &descEn, &descRu, &unEn, &unRu, &paRaw, &p.Saved, &p.Visible)
		if err != nil {
			return nil, err
		}
		p.Type = model.VarType(tRaw)
		p.Access = model.Access(paRaw)
		p.DescriptionEn = descEn.String
		p.DescriptionRu = descRu.String
		p.UnitsEn = unEn.String
		p.UnitsRu = unRu.String
		p.Value = valNull.String
		params = append(params, p)
	}
	c.Params = params
	return &c, nil
}

func UpdateComponent(ctx context.Context, id int64, d dto.ComponentUpdate) (*model.Component, error) {
	conn := database.Get()
	query := `
		UPDATE public.component 
		SET title = $1, name_en = $2, name_ru = $3, base_component = $4, description_en = $5, description_ru = $6, access = $7
		WHERE id = $8
		RETURNING id, title, name_en, name_ru, access`
	var baseComp sql.NullInt64
	if d.BaseComponent != nil {
		baseComp.Int64 = *d.BaseComponent
		baseComp.Valid = true
	}
	accessIdx := model.ParseAccess(d.Access)
	var c model.Component
	var aRaw int16
	err := conn.QueryRow(ctx, query, d.Title, d.NameEn, d.NameRu, baseComp, stringToNull(d.DescriptionEn), stringToNull(d.DescriptionRu), int16(accessIdx), id).
		Scan(&c.ID, &c.Title, &c.NameEn, &c.NameRu, &aRaw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	c.Access = model.Access(aRaw)
	return &c, nil
}

func GetAllComponentsWithParams(ctx context.Context) ([]model.Component, error) {
	conn := database.Get()
	query := `
		WITH RECURSIVE component_hierarchy AS (
			SELECT id AS root_id, id AS current_id, base_component FROM public.component
			UNION ALL
			SELECT ch.root_id, c.id, c.base_component
			FROM public.component c
			JOIN component_hierarchy ch ON c.id = ch.base_component
		),
		aggregated_params AS (
			SELECT ch.root_id,
			       json_strip_nulls(json_agg(json_build_object(
				       'id', p.id,
				       'title', p.title,
				       'name_en', p.name_en,
				       'name_ru', p.name_ru,
				       'type', p.type,
				       'value', p.value,
				       'description_en', p.description_en,
				       'description_ru', p.description_ru,
				       'units_en', p.units_en,
				       'units_ru', p.units_ru,
				       'access', p.access,
				       'saved', p.saved,
				       'visible', p.visible
			       ))) AS params_json
			FROM component_hierarchy ch
			JOIN public.component_param cp ON ch.current_id = cp.component_id
			JOIN public.param p ON cp.param_id = p.id
			GROUP BY ch.root_id
		)
		SELECT c.id, c.title, c.name_en, c.name_ru, c.access, ap.params_json
		FROM public.component c
		LEFT JOIN aggregated_params ap ON c.id = ap.root_id`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var components []model.Component
	for rows.Next() {
		var c model.Component
		var paramsJSON []byte
		var aRaw int16
		err := rows.Scan(&c.ID, &c.Title, &c.NameEn, &c.NameRu, &aRaw, &paramsJSON)
		if err != nil {
			return nil, err
		}
		c.Access = model.Access(aRaw)
		if len(paramsJSON) > 0 && string(paramsJSON) != "[null]" {
			if err := json.Unmarshal(paramsJSON, &c.Params); err != nil {
				return nil, err
			}
		} else {
			c.Params = []model.Param{}
		}

		components = append(components, c)
	}
	return components, nil
}

func SearchComponents(ctx context.Context, q string) ([]model.Component, error) {
	conn := database.Get()
	query := `
		SELECT id, title, name_en, name_ru, access
		FROM public.component
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
	var components []model.Component
	for rows.Next() {
		var c model.Component
		var aRaw int16
		err := rows.Scan(&c.ID, &c.Title, &c.NameEn, &c.NameRu, &aRaw)
		if err != nil {
			return nil, err
		}
		c.Access = model.Access(aRaw)
		c.Params = []model.Param{}
		components = append(components, c)
	}
	return components, nil
}

func DeleteComponent(ctx context.Context, id int64) (bool, error) {
	conn := database.Get()
	query := `
		WITH RECURSIVE targets AS (
			SELECT id FROM public.component WHERE id = $1
			UNION ALL
			SELECT c.id FROM public.component c
			JOIN targets t ON c.base_component = t.id
		),
		delete_relations AS (
			DELETE FROM public.component_param 
			WHERE component_id IN (SELECT id FROM targets)
			RETURNING component_id
		),
		pre_delete_components AS (
			UPDATE public.component 
			SET base_component = NULL 
			WHERE base_component IN (SELECT id FROM targets)
		)
		DELETE FROM public.component 
		WHERE id IN (SELECT id FROM targets)`
	commandTag, err := conn.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return commandTag.RowsAffected() > 0, nil
}
