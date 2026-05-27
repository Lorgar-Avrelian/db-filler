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

func CreateDefaultConfiguration(ctx context.Context, d dto.ConfigurationCreate) (int64, error) {
	conn := database.Get()
	query := `INSERT INTO public.default_configuration (indicator, device_component_id) VALUES ($1, $2) RETURNING id`
	var dcID sql.NullInt64
	if d.DeviceComponentID != nil {
		dcID.Int64 = *d.DeviceComponentID
		dcID.Valid = true
	}
	var id int64
	err := conn.QueryRow(ctx, query, d.IndicatorID, dcID).Scan(&id)
	return id, err
}

func GetDetailedDefaultConfigByID(ctx context.Context, id int64) (*model.DeviceIndicator, *model.DeviceComponent, []model.Threshold, error) {
	flatRows, err := executeGenericConfigSelect(ctx, "default_configuration", id)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(flatRows) == 0 {
		return nil, nil, nil, nil
	}
	_, indMap, dcMap, thMap := AssembleConfigurations(flatRows)
	return indMap[id], dcMap[id], thMap[id], nil
}

func UpdateDefaultConfiguration(ctx context.Context, id int64, d dto.ConfigurationUpdate) (int64, error) {
	conn := database.Get()
	query := `UPDATE public.default_configuration SET indicator = $1, device_component_id = $2 WHERE id = $3 RETURNING id`
	var dcID sql.NullInt64
	if d.DeviceComponentID != nil {
		dcID.Int64 = *d.DeviceComponentID
		dcID.Valid = true
	}
	var updatedID int64
	err := conn.QueryRow(ctx, query, d.IndicatorID, dcID, id).Scan(&updatedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return updatedID, nil
}

func DeleteDefaultConfiguration(ctx context.Context, id int64) (bool, error) {
	conn := database.Get()
	query := `DELETE FROM public.default_configuration WHERE id = $1`
	commandTag, err := conn.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return commandTag.RowsAffected() > 0, nil
}

func GetExpandedDefaultConfigurations(ctx context.Context) ([]model.DefaultConfiguration, error) {
	flatRows, err := executeGenericConfigSelect(ctx, "default_configuration", 0)
	if err != nil {
		return nil, err
	}
	ids, indMap, dcMap, thMap := AssembleConfigurations(flatRows)
	var list []model.DefaultConfiguration
	for _, id := range ids {
		list = append(list, model.DefaultConfiguration{ID: id, Indicator: indMap[id], DeviceComponent: dcMap[id], Thresholds: thMap[id]})
	}
	return list, nil
}

func BindDefaultConfigThreshold(ctx context.Context, defCfgID, tID int64) error {
	conn := database.Get()
	_, err := conn.Exec(ctx, `INSERT INTO public.default_configuration_threshold (default_configuration_id, threshold_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, defCfgID, tID)
	return err
}

func UnbindDefaultConfigThreshold(ctx context.Context, defCfgID, tID int64) (bool, error) {
	conn := database.Get()
	tag, err := conn.Exec(ctx, `DELETE FROM public.default_configuration_threshold WHERE default_configuration_id = $1 AND threshold_id = $2`, defCfgID, tID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
