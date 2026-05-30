package dao

import (
	"context"
	"encoding/json"
	"errors"
	"filler/internal/database"
	"filler/internal/dto"
	"filler/internal/model"

	"github.com/jackc/pgx/v5"
)

func CreateMapping(ctx context.Context, d dto.MappingCreate) (*model.Mapping, error) {
	conn := database.Get()
	insertQuery := `
		INSERT INTO public.mapping (indicator, param, frequency, coefficient, enum)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	freqIdx := model.ParsePollingFrequency(d.Frequency)
	enumBytes, err := mapToJSONB(d.Enum)
	if err != nil {
		return nil, err
	}
	var id int64
	err = conn.QueryRow(ctx, insertQuery, d.IndicatorID, d.ParamID, int16(freqIdx), toNullString(d.Coefficient), enumBytes).Scan(&id)
	if err != nil {
		return nil, err
	}
	return GetMappingByID(ctx, id)
}

func GetMappingByID(ctx context.Context, id int64) (*model.Mapping, error) {
	conn := database.Get()
	query := `
		SELECT json_strip_nulls(json_build_object(
			'id', m.id,
			'frequency', pf.value,
			'coefficient', m.coefficient::text,
			'enum', m.enum,
			'param', json_build_object(
				'id', p.id,
				'title', p.title,
				'name_en', p.name_en,
				'name_ru', p.name_ru,
				'type', p_vt.value,
				'value', p.value,
				'description_en', p.description_en,
				'description_ru', p.description_ru,
				'units_en', p.units_en,
				'units_ru', p.units_ru,
				'access', p_ac.value,
				'saved', p.saved,
				'visible', p.visible
			),
			'indicator', json_build_object(
				'id', pi.id,
				'oid_id', pi.oid_id,
				'dotter_notation', pi.dotter_notation,
				'oid', json_build_object(
					'id', o.id,
					'mib_id', o.mib,
					'type', o_at.value,
					'name', o.name,
					'number', o.number,
					'dotter_notation', o.dotter_notation,
					'object_descriptor', o.object_descriptor,
					'syntax', o.syntax,
					'enum', o.enum,
					'status', o_st.value,
					'access', o_oac.value,
					'units', o.units,
					'description', o.description,
					'category', o.category
				)
			)
		))::text
		FROM public.mapping m
		LEFT JOIN public.polling_frequency pf ON m.frequency = pf.id
		LEFT JOIN public.param p ON m.param = p.id
		LEFT JOIN public.var_type p_vt ON p.type = p_vt.id
		LEFT JOIN public.access p_ac ON p.access = p_ac.id
		LEFT JOIN public.param_indicator pi ON m.indicator = pi.id
		LEFT JOIN public.oid o ON pi.oid_id = o.id
		LEFT JOIN public.asn1_type o_at ON o.type = o_at.id
		LEFT JOIN public.oid_status o_st ON o.status = o_st.id
		LEFT JOIN public.oid_access o_oac ON o.access = o_oac.id
		WHERE m.id = $1`
	var jsonStr string
	err := conn.QueryRow(ctx, query, id).Scan(&jsonStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return unmarshalFullMapping([]byte(jsonStr))
}

func GetAllMappings(ctx context.Context) ([]model.Mapping, error) {
	conn := database.Get()
	query := `
		SELECT json_strip_nulls(json_build_object(
			'id', m.id,
			'frequency', pf.value,
			'coefficient', m.coefficient::text,
			'enum', m.enum,
			'param', json_build_object(
				'id', p.id,
				'title', p.title,
				'name_en', p.name_en,
				'name_ru', p.name_ru,
				'type', p_vt.value,
				'value', p.value,
				'description_en', p.description_en,
				'description_ru', p.description_ru,
				'units_en', p.units_en,
				'units_ru', p.units_ru,
				'access', p_ac.value,
				'saved', p.saved,
				'visible', p.visible
			),
			'indicator', json_build_object(
				'id', pi.id,
				'oid_id', pi.oid_id,
				'dotter_notation', pi.dotter_notation,
				'oid', json_build_object(
					'id', o.id,
					'mib_id', o.mib,
					'type', o_at.value,
					'name', o.name,
					'number', o.number,
					'dotter_notation', o.dotter_notation,
					'object_descriptor', o.object_descriptor,
					'syntax', o.syntax,
					'enum', o.enum,
					'status', o_st.value,
					'access', o_oac.value,
					'units', o.units,
					'description', o.description,
					'category', o.category
				)
			)
		))::text
		FROM public.mapping m
		LEFT JOIN public.polling_frequency pf ON m.frequency = pf.id
		LEFT JOIN public.param p ON m.param = p.id
		LEFT JOIN public.var_type p_vt ON p.type = p_vt.id
		LEFT JOIN public.access p_ac ON p.access = p_ac.id
		LEFT JOIN public.param_indicator pi ON m.indicator = pi.id
		LEFT JOIN public.oid o ON pi.oid_id = o.id
		LEFT JOIN public.asn1_type o_at ON o.type = o_at.id
		LEFT JOIN public.oid_status o_st ON o.status = o_st.id
		LEFT JOIN public.oid_access o_oac ON o.access = o_oac.id`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Mapping
	for rows.Next() {
		var jsonStr string
		if err := rows.Scan(&jsonStr); err != nil {
			return nil, err
		}
		m, err := unmarshalFullMapping([]byte(jsonStr))
		if err != nil {
			return nil, err
		}
		list = append(list, *m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func UpdateMapping(ctx context.Context, id int64, d dto.MappingUpdate) (*model.Mapping, error) {
	conn := database.Get()
	query := `
		UPDATE public.mapping 
		SET indicator = $1, param = $2, frequency = $3, coefficient = $4, enum = $5
		WHERE id = $6
		RETURNING id`
	freqIdx := model.ParsePollingFrequency(d.Frequency)
	enumBytes, err := mapToJSONB(d.Enum)
	if err != nil {
		return nil, err
	}
	var updatedID int64
	err = conn.QueryRow(ctx, query, d.IndicatorID, d.ParamID, int16(freqIdx), toNullString(d.Coefficient), enumBytes, id).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return GetMappingByID(ctx, updatedID)
}

func DeleteMapping(ctx context.Context, id int64) (bool, error) {
	conn := database.Get()
	query := `DELETE FROM public.mapping WHERE id = $1`
	commandTag, err := conn.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return commandTag.RowsAffected() > 0, nil
}

func unmarshalFullMapping(jsonBytes []byte) (*model.Mapping, error) {
	var m model.Mapping
	if err := json.Unmarshal(jsonBytes, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
