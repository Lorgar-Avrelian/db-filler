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

// Вспомогательный метод выгрузки плоского списка и сборки в дерево
func buildTreeFromRows(rows pgx.Rows) ([]model.DeviceComponent, error) {
	nodesMap := make(map[int64]*model.DeviceComponent)
	var roots []model.DeviceComponent
	type edge struct {
		parentID int64
		childID  int64
	}
	var edges []edge
	for rows.Next() {
		var dc model.DeviceComponent
		var parent sql.NullInt64
		var jsonBytes []byte
		if err := rows.Scan(&dc.ID, &dc.ModelID, &dc.InternalOrder, &parent, &jsonBytes); err != nil {
			return nil, err
		}
		if parent.Valid {
			dc.ParentID = &parent.Int64
		}
		dc.Mappings = []model.Mapping{}
		dc.Components = []model.DeviceComponent{}
		if len(jsonBytes) > 0 && string(jsonBytes) != "[null]" {
			_ = json.Unmarshal(jsonBytes, &dc.Mappings)
		}
		nodesMap[dc.ID] = &dc
		if parent.Valid {
			edges = append(edges, edge{parentID: parent.Int64, childID: dc.ID})
		}
	}
	for _, e := range edges {
		parentBox, parentExists := nodesMap[e.parentID]
		childBox, childExists := nodesMap[e.childID]

		if parentExists && childExists {
			parentBox.Components = append(parentBox.Components, *childBox)
		}
	}
	for _, node := range nodesMap {
		if node.ParentID == nil {
			roots = append(roots, *node)
		} else {
			if _, ok := nodesMap[*node.ParentID]; !ok {
				roots = append(roots, *node)
			}
		}
	}
	if len(roots) == 0 && len(nodesMap) > 0 {
		for _, node := range nodesMap {
			roots = append(roots, *node)
			break
		}
	}
	return roots, nil
}

func CreateDeviceComponent(ctx context.Context, d dto.DeviceComponentCreate) (*model.DeviceComponent, error) {
	conn := database.Get()
	query := `
		INSERT INTO public.device_component (model, internal_order, parent)
		VALUES ($1, $2, $3)
		RETURNING id, model, internal_order, parent`
	var parent sql.NullInt64
	if d.ParentID != nil {
		parent.Int64 = *d.ParentID
		parent.Valid = true
	}
	var dc model.DeviceComponent
	err := conn.QueryRow(ctx, query, d.ModelID, d.InternalOrder, parent).
		Scan(&dc.ID, &dc.ModelID, &dc.InternalOrder, &parent)
	if err != nil {
		return nil, err
	}

	if parent.Valid {
		dc.ParentID = &parent.Int64
	}
	dc.Mappings = []model.Mapping{}
	dc.Components = []model.DeviceComponent{}
	return &dc, nil
}

// GetDeviceComponentByID рекурсивно собирает всю ветку компонента вниз по иерархии
func GetDeviceComponentByID(ctx context.Context, id int64) (*model.DeviceComponent, error) {
	conn := database.Get()
	query := `
		WITH RECURSIVE device_tree AS (
			SELECT id, model, internal_order, parent FROM public.device_component WHERE id = $1
			UNION ALL
			SELECT c.id, c.model, c.internal_order, c.parent
			FROM public.device_component c
			JOIN device_tree dt ON c.parent = dt.id
		)
		SELECT dt.id, dt.model, dt.internal_order, dt.parent,
		       json_strip_nulls(json_agg(json_build_object(
			       'id', m.id,
			       'indicator_id', m.indicator,
			       'param_id', m.param,
			       'frequency', m.frequency,
			       'coefficient', m.coefficient,
			       'enum', m.enum
		       ))) FILTER (WHERE m.id IS NOT NULL) AS mappings_json
		FROM device_tree dt
		LEFT JOIN public.device_component_mapping dcm ON dt.id = dcm.device_component_id
		LEFT JOIN public.mapping m ON dcm.mapping_id = m.id
		GROUP BY dt.id, dt.model, dt.internal_order, dt.parent`
	rows, err := conn.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tree, err := buildTreeFromRows(rows)
	if err != nil {
		return nil, err
	}
	if len(tree) == 0 {
		return nil, nil
	}
	for _, node := range tree {
		if node.ID == id {
			return &node, nil
		}
	}
	return &tree[0], nil
}

// GetAllDeviceComponents возвращает полный массив иерархических деревьев устройств
func GetAllDeviceComponents(ctx context.Context) ([]model.DeviceComponent, error) {
	conn := database.Get()
	query := `
		SELECT dc.id, dc.model, dc.internal_order, dc.parent,
		       json_strip_nulls(json_agg(json_build_object(
			       'id', m.id,
			       'indicator_id', m.indicator,
			       'param_id', m.param,
			       'frequency', m.frequency,
			       'coefficient', m.coefficient,
			       'enum', m.enum
		       ))) FILTER (WHERE m.id IS NOT NULL) AS mappings_json
		FROM public.device_component dc
		LEFT JOIN public.device_component_mapping dcm ON dc.id = dcm.device_component_id
		LEFT JOIN public.mapping m ON dcm.mapping_id = m.id
		GROUP BY dc.id, dc.model, dc.internal_order, dc.parent`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return buildTreeFromRows(rows)
}

func UpdateDeviceComponent(ctx context.Context, id int64, d dto.DeviceComponentUpdate) (*model.DeviceComponent, error) {
	conn := database.Get()
	query := `
		UPDATE public.device_component 
		SET model = $1, internal_order = $2, parent = $3
		WHERE id = $4
		RETURNING id, model, internal_order, parent`
	var parent sql.NullInt64
	if d.ParentID != nil {
		parent.Int64 = *d.ParentID
		parent.Valid = true
	}
	var dc model.DeviceComponent
	err := conn.QueryRow(ctx, query, d.ModelID, d.InternalOrder, parent, id).
		Scan(&dc.ID, &dc.ModelID, &dc.InternalOrder, &parent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if parent.Valid {
		dc.ParentID = &parent.Int64
	}
	return &dc, nil
}

func DeleteDeviceComponent(ctx context.Context, id int64) (bool, error) {
	conn := database.Get()
	query := `DELETE FROM public.device_component WHERE id = $1`
	commandTag, err := conn.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return commandTag.RowsAffected() > 0, nil
}

func BindDeviceMapping(ctx context.Context, d dto.BindDeviceMappingRequest) error {
	conn := database.Get()
	query := `INSERT INTO public.device_component_mapping (device_component_id, mapping_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := conn.Exec(ctx, query, d.DeviceComponentID, d.MappingID)
	return err
}

func UnbindDeviceMapping(ctx context.Context, dcID, mID int64) (bool, error) {
	conn := database.Get()
	query := `DELETE FROM public.device_component_mapping WHERE device_component_id = $1 AND mapping_id = $2`
	commandTag, err := conn.Exec(ctx, query, dcID, mID)
	if err != nil {
		return false, err
	}
	return commandTag.RowsAffected() > 0, nil
}
