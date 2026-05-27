package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"filler/internal/database"
	"filler/internal/dto"
	"filler/internal/model"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type ConfigFlatRow struct {
	ConfigID  int64
	IndID     sql.NullInt64
	IndDesc   sql.NullString
	IndObj    sql.NullString
	IndCont   sql.NullString
	IndName   sql.NullString
	IndLoc    sql.NullString
	IndServ   sql.NullInt16
	DcID      sql.NullInt64
	DcModel   sql.NullInt64
	DcOrder   sql.NullInt32
	DcParent  sql.NullInt64
	DcMapJSON []byte
	CfgThJSON []byte
}

func ScanGenericConfigRows(rows pgx.Rows) ([]ConfigFlatRow, error) {
	var flatRows []ConfigFlatRow
	for rows.Next() {
		var r ConfigFlatRow
		err := rows.Scan(
			&r.ConfigID, &r.IndID, &r.IndDesc, &r.IndObj, &r.IndCont, &r.IndName, &r.IndLoc, &r.IndServ,
			&r.DcID, &r.DcModel, &r.DcOrder, &r.DcParent, &r.DcMapJSON, &r.CfgThJSON,
		)
		if err != nil {
			return nil, err
		}
		flatRows = append(flatRows, r)
	}
	return flatRows, nil
}

func AssembleConfigurations(flatRows []ConfigFlatRow) ([]int64, map[int64]*model.DeviceIndicator, map[int64]*model.DeviceComponent, map[int64][]model.Threshold) {
	configIDs := []int64{}
	seenConfigs := make(map[int64]bool)
	indicatorsMap := make(map[int64]*model.DeviceIndicator)
	dcNodes := make(map[int64]*model.DeviceComponent)
	configThresholdsMap := make(map[int64][]model.Threshold)
	type edge struct{ parent, child int64 }
	var edges []edge
	for _, r := range flatRows {
		if !seenConfigs[r.ConfigID] {
			seenConfigs[r.ConfigID] = true
			configIDs = append(configIDs, r.ConfigID)
			configThresholdsMap[r.ConfigID] = []model.Threshold{}
		}
		if r.IndID.Valid {
			if _, ok := indicatorsMap[r.ConfigID]; !ok {
				ind := mapRowToIndicator(r.IndID.Int64, r.IndDesc, r.IndObj, r.IndCont, r.IndName, r.IndLoc, r.IndServ)
				indicatorsMap[r.ConfigID] = &ind
			}
		}
		if r.DcID.Valid {
			if _, ok := dcNodes[r.DcID.Int64]; !ok {
				node := model.DeviceComponent{
					ID:            r.DcID.Int64,
					ModelID:       r.DcModel.Int64,
					InternalOrder: r.DcOrder.Int32,
					Mappings:      []model.Mapping{},
					Components:    []model.DeviceComponent{},
				}
				if r.DcParent.Valid {
					node.ParentID = &r.DcParent.Int64
					edges = append(edges, edge{parent: r.DcParent.Int64, child: r.DcID.Int64})
				}
				if len(r.DcMapJSON) > 0 && string(r.DcMapJSON) != "[null]" {
					_ = json.Unmarshal(r.DcMapJSON, &node.Mappings)
				}
				dcNodes[r.DcID.Int64] = &node
			}
		}
		if len(r.CfgThJSON) > 0 && string(r.CfgThJSON) != "[null]" && len(configThresholdsMap[r.ConfigID]) == 0 {
			var rawThresholds []model.Threshold
			if err := json.Unmarshal(r.CfgThJSON, &rawThresholds); err == nil {
				thMap := make(map[int64]*model.Threshold)
				for i := range rawThresholds {
					thMap[rawThresholds[i].ID] = &rawThresholds[i]
				}
				for i := range rawThresholds {
					if rawThresholds[i].PreviousID != nil {
						if prev, ok := thMap[*rawThresholds[i].PreviousID]; ok {
							rawThresholds[i].PreviousThreshold = prev
						}
					}
				}
				configThresholdsMap[r.ConfigID] = rawThresholds
			}
		}
	}
	for _, e := range edges {
		pNode, pOk := dcNodes[e.parent]
		cNode, cOk := dcNodes[e.child]
		if pOk && cOk {
			pNode.Components = append(pNode.Components, *cNode)
		}
	}
	configComponentMap := make(map[int64]*model.DeviceComponent)
	for _, r := range flatRows {
		if r.DcID.Valid {
			if node, ok := dcNodes[r.DcID.Int64]; ok {
				configComponentMap[r.ConfigID] = node
			}
		}
	}
	return configIDs, indicatorsMap, configComponentMap, configThresholdsMap
}

func executeGenericConfigSelect(ctx context.Context, table string, idFilter int64) ([]ConfigFlatRow, error) {
	conn := database.Get()
	filterSQL := ""
	if idFilter > 0 {
		filterSQL = fmt.Sprintf("WHERE cfg.id = %d", idFilter)
	}
	thresholdJoinTable := "configuration_threshold"
	thresholdLinkField := "configuration_id"
	if table == "default_configuration" {
		thresholdJoinTable = "default_configuration_threshold"
		thresholdLinkField = "default_configuration_id"
	}
	query := fmt.Sprintf(`
		WITH RECURSIVE target_configs AS (
			SELECT id, indicator, device_component_id FROM public.%s cfg
		),
		device_tree AS (
			SELECT dc.id, dc.model, dc.internal_order, dc.parent, tc.id AS cfg_id
			FROM public.device_component dc
			JOIN target_configs tc ON dc.id = tc.device_component_id
			UNION ALL
			SELECT c.id, c.model, c.internal_order, c.parent, dt.cfg_id
			FROM public.device_component c
			JOIN device_tree dt ON c.parent = dt.id
		),
		aggregated_thresholds AS (
			SELECT ct.%s AS cfg_id,
			       json_strip_nulls(json_agg(json_build_object(
				       'id', t.id, 'source_model', t.source_model, 'source_internal_order', t.source_internal_order,
				       'source_param', t.source_param, 'value', t.value, 'type', t.type, 'operator', t.operator,
				       'enabled', t.enabled, 'target_param', t.target_param, 'target_device', t.target_device,
				       'level', t.level, 'prev_operator', t.prev_operator, 'previous_id', t.previous
			       ))) AS thresholds_json
			FROM public.%s ct
			JOIN public.threshold t ON ct.threshold_id = t.id
			GROUP BY ct.%s
		)
		SELECT 
			cfg.id,
			i.id, i.description, i.object_id, i.contact, i.name, i.location, i.services,
			dt.id, dt.model, dt.internal_order, dt.parent,
			json_strip_nulls(json_agg(json_build_object(
				'id', m.id, 'indicator_id', m.indicator, 'param_id', m.param, 'frequency', m.frequency, 'coefficient', m.coefficient, 'enum', m.enum
			))) FILTER (WHERE m.id IS NOT NULL) AS mappings_json,
			ath.thresholds_json
		FROM public.%s cfg
		LEFT JOIN public.device_indicator i ON cfg.indicator = i.id
		LEFT JOIN device_tree dt ON cfg.id = dt.cfg_id
		LEFT JOIN public.device_component_mapping dcm ON dt.id = dcm.device_component_id
		LEFT JOIN public.mapping m ON dcm.mapping_id = m.id
		LEFT JOIN aggregated_thresholds ath ON cfg.id = ath.cfg_id
		%s
		GROUP BY cfg.id, i.id, i.description, i.object_id, i.contact, i.name, i.location, i.services, dt.id, dt.model, dt.internal_order, dt.parent, ath.thresholds_json
		ORDER BY cfg.id`, table, thresholdLinkField, thresholdJoinTable, thresholdLinkField, table, filterSQL)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return ScanGenericConfigRows(rows)
}

func CreateConfiguration(ctx context.Context, d dto.ConfigurationCreate) (int64, error) {
	conn := database.Get()
	query := `INSERT INTO public.configuration (indicator, device_component_id) VALUES ($1, $2) RETURNING id`
	var dcID sql.NullInt64
	if d.DeviceComponentID != nil {
		dcID.Int64 = *d.DeviceComponentID
		dcID.Valid = true
	}
	var id int64
	err := conn.QueryRow(ctx, query, d.IndicatorID, dcID).Scan(&id)
	return id, err
}

func GetDetailedConfigByID(ctx context.Context, id int64) (*model.DeviceIndicator, *model.DeviceComponent, []model.Threshold, error) {
	flatRows, err := executeGenericConfigSelect(ctx, "configuration", id)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(flatRows) == 0 {
		return nil, nil, nil, nil
	}
	_, indMap, dcMap, thMap := AssembleConfigurations(flatRows)
	return indMap[id], dcMap[id], thMap[id], nil
}

func UpdateConfiguration(ctx context.Context, id int64, d dto.ConfigurationUpdate) (int64, error) {
	conn := database.Get()
	query := `UPDATE public.configuration SET indicator = $1, device_component_id = $2 WHERE id = $3 RETURNING id`
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

func DeleteConfiguration(ctx context.Context, id int64) (bool, error) {
	conn := database.Get()
	query := `DELETE FROM public.configuration WHERE id = $1`
	commandTag, err := conn.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return commandTag.RowsAffected() > 0, nil
}

func GetExpandedConfigurations(ctx context.Context) ([]model.Configuration, error) {
	flatRows, err := executeGenericConfigSelect(ctx, "configuration", 0)
	if err != nil {
		return nil, err
	}

	ids, indMap, dcMap, thMap := AssembleConfigurations(flatRows)
	var list []model.Configuration
	for _, id := range ids {
		list = append(list, model.Configuration{ID: id, Indicator: indMap[id], DeviceComponent: dcMap[id], Thresholds: thMap[id]})
	}
	return list, nil
}

func BindConfigThreshold(ctx context.Context, cfgID, tID int64) error {
	conn := database.Get()
	_, err := conn.Exec(ctx, `INSERT INTO public.configuration_threshold (configuration_id, threshold_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, cfgID, tID)
	return err
}

func UnbindConfigThreshold(ctx context.Context, cfgID, tID int64) (bool, error) {
	conn := database.Get()
	tag, err := conn.Exec(ctx, `DELETE FROM public.configuration_threshold WHERE configuration_id = $1 AND threshold_id = $2`, cfgID, tID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
