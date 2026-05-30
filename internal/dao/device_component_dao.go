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
	if len(nodesMap) > 0 && len(roots) == 0 {
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

// GetDeviceComponentByID рекурсивно собирает всю ветку компонента вниз по иерархии со всеми маппингами для каждого уровня
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
				    'frequency', pf.value,
				    'coefficient', m.coefficient,
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
				)) FILTER (WHERE m.id IS NOT NULL)) AS mappings_json
		FROM device_tree dt
		LEFT JOIN public.device_component_mapping dcm ON dt.id = dcm.device_component_id
		LEFT JOIN public.mapping m ON dcm.mapping_id = m.id
		LEFT JOIN public.polling_frequency pf ON m.frequency = pf.id
		LEFT JOIN public.param p ON m.param = p.id
		LEFT JOIN public.var_type p_vt ON p.type = p_vt.id
		LEFT JOIN public.access p_ac ON p.access = p_ac.id
		LEFT JOIN public.param_indicator pi ON m.indicator = pi.id
		LEFT JOIN public.oid o ON pi.oid_id = o.id
		LEFT JOIN public.asn1_type o_at ON o.type = o_at.id
		LEFT JOIN public.oid_status o_st ON o.status = o_st.id
		LEFT JOIN public.oid_access o_oac ON o.access = o_oac.id
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

// GetAllDeviceComponents возвращает полный массив иерархических деревьев устройств со всеми маппингами
func GetAllDeviceComponents(ctx context.Context) ([]model.DeviceComponent, error) {
	conn := database.Get()
	query := `
		SELECT dc.id, dc.model, dc.internal_order, dc.parent,
		       json_strip_nulls(json_agg(json_build_object(
				    'id', m.id,
				    'frequency', pf.value,
				    'coefficient', m.coefficient,
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
				)) FILTER (WHERE m.id IS NOT NULL)) AS mappings_json
		FROM public.device_component dc
		LEFT JOIN public.device_component_mapping dcm ON dc.id = dcm.device_component_id
		LEFT JOIN public.mapping m ON dcm.mapping_id = m.id
		LEFT JOIN public.polling_frequency pf ON m.frequency = pf.id
		LEFT JOIN public.param p ON m.param = p.id
		LEFT JOIN public.var_type p_vt ON p.type = p_vt.id
		LEFT JOIN public.access p_ac ON p.access = p_ac.id
		LEFT JOIN public.param_indicator pi ON m.indicator = pi.id
		LEFT JOIN public.oid o ON pi.oid_id = o.id
		LEFT JOIN public.asn1_type o_at ON o.type = o_at.id
		LEFT JOIN public.oid_status o_st ON o.status = o_st.id
		LEFT JOIN public.oid_access o_oac ON o.access = o_oac.id
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
