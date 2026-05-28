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

func CreateMapping(ctx context.Context, d dto.MappingCreate) (*model.Mapping, error) {
	conn := database.Get()
	query := `
		INSERT INTO public.mapping (indicator, param, frequency, coefficient, enum)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, indicator, param, frequency, coefficient, enum`
	freqIdx := model.ParsePollingFrequency(d.Frequency)
	enumBytes, err := mapToJSONB(d.Enum)
	if err != nil {
		return nil, err
	}
	var m model.Mapping
	var freqRaw int16
	var coeff sql.NullString
	var resEnumBytes []byte
	err = conn.QueryRow(ctx, query, d.IndicatorID, d.ParamID, int16(freqIdx), toNullString(d.Coefficient), enumBytes).
		Scan(&m.ID, &m.IndicatorID, &m.ParamID, &freqRaw, &coeff, &resEnumBytes)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	m.Frequency = model.PollingFrequency(freqRaw)
	if coeff.Valid {
		m.Coefficient = &coeff.String
	}
	if len(resEnumBytes) > 0 {
		_ = json.Unmarshal(resEnumBytes, &m.Enum)
	}
	return &m, nil
}

func GetMappingByID(ctx context.Context, id int64) (*model.Mapping, error) {
	conn := database.Get()
	query := `SELECT id, indicator, param, frequency, coefficient, enum FROM public.mapping WHERE id = $1`
	var m model.Mapping
	var freqRaw int16
	var coeff sql.NullString
	var enumBytes []byte
	err := conn.QueryRow(ctx, query, id).Scan(&m.ID, &m.IndicatorID, &m.ParamID, &freqRaw, &coeff, &enumBytes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	m.Frequency = model.PollingFrequency(freqRaw)
	if coeff.Valid {
		m.Coefficient = &coeff.String
	}
	if len(enumBytes) > 0 {
		if err := json.Unmarshal(enumBytes, &m.Enum); err != nil {
			return nil, err
		}
	}
	return &m, nil
}

func GetAllMappings(ctx context.Context) ([]model.Mapping, error) {
	conn := database.Get()
	query := `SELECT id, indicator, param, frequency, coefficient, enum FROM public.mapping`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Mapping
	for rows.Next() {
		var m model.Mapping
		var freqRaw int16
		var coeff sql.NullString
		var enumBytes []byte
		err := rows.Scan(&m.ID, &m.IndicatorID, &m.ParamID, &freqRaw, &coeff, &enumBytes)
		if err != nil {
			return nil, err
		}
		m.Frequency = model.PollingFrequency(freqRaw)
		if coeff.Valid {
			m.Coefficient = &coeff.String
		}
		if len(enumBytes) > 0 {
			if err := json.Unmarshal(enumBytes, &m.Enum); err != nil {
				return nil, err
			}
		}
		list = append(list, m)
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
		SET indicator = $1, param = $2, frequency = $3, coefficient = $4
		WHERE id = $5
		RETURNING id, indicator, param, frequency, coefficient`
	freqIdx := model.ParsePollingFrequency(d.Frequency)
	enumBytes, err := mapToJSONB(d.Enum)
	if err != nil {
		return nil, err
	}
	var m model.Mapping
	var freqRaw int16
	var coeff sql.NullString
	var resEnumBytes []byte
	err = conn.QueryRow(ctx, query, d.IndicatorID, d.ParamID, int16(freqIdx), toNullString(d.Coefficient), enumBytes, id).
		Scan(&m.ID, &m.IndicatorID, &m.ParamID, &freqRaw, &coeff, &resEnumBytes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	m.Frequency = model.PollingFrequency(freqRaw)
	if coeff.Valid {
		m.Coefficient = &coeff.String
	}
	if len(resEnumBytes) > 0 {
		_ = json.Unmarshal(resEnumBytes, &m.Enum)
	}
	return &m, nil
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
