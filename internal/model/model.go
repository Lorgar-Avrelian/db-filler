package model

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type Param struct {
	ID            int64   `json:"id" example:"1"`
	Title         string  `json:"title" example:"Макс. Напряжение"`
	NameEn        string  `json:"name_en" example:"max_voltage"`
	NameRu        string  `json:"name_ru" example:"макс_напряжение"`
	Type          VarType `json:"type" swaggertype:"primitive,string" example:"INT"`
	Value         string  `json:"value" example:"220"`
	DescriptionEn string  `json:"description_en,omitempty" example:"Maximum operational voltage"`
	DescriptionRu string  `json:"description_ru,omitempty" example:"Максимальное рабочее напряжение"`
	UnitsEn       string  `json:"units_en,omitempty" example:"V"`
	UnitsRu       string  `json:"units_ru,omitempty" example:"В"`
	Access        Access  `json:"access" swaggertype:"primitive,string" example:"ADMIN"`
	Saved         bool    `json:"saved" example:"true"`
	Visible       bool    `json:"visible" example:"true"`
}

type Component struct {
	ID            int64   `json:"id" example:"10"`
	Title         string  `json:"title" example:"Модуль Питания"`
	NameEn        string  `json:"name_en" example:"power_module"`
	NameRu        string  `json:"name_ru" example:"модуль_питания"`
	BaseComponent *int64  `json:"base_component,omitempty" example:"1"`
	DescriptionEn string  `json:"description_en,omitempty" example:"Main power controller unit"`
	DescriptionRu string  `json:"description_ru,omitempty" example:"Главный контроллер питания"`
	Access        Access  `json:"access" swaggertype:"primitive,string" example:"USER"`
	Params        []Param `json:"params,omitempty"`
}

type Oid struct {
	ID               uuid.UUID       `json:"id" example:"00000000-0000-0000-0000-000000000000"`
	MibID            *int64          `json:"mib_id,omitempty" example:"1"`
	Type             Asn1Type        `json:"type" swaggertype:"primitive,string" example:"OBJECT-TYPE"`
	Name             string          `json:"name" example:"sysDescr"`
	Number           *int32          `json:"number,omitempty" example:"1"`
	DotterNotation   string          `json:"dotter_notation" example:"1.3.6.1.2.1.1.1"`
	ObjectDescriptor string          `json:"object_descriptor" example:"1.3.6.1.2.1.1.1"`
	Syntax           string          `json:"syntax,omitempty" example:"DisplayString"`
	Enum             json.RawMessage `json:"enum,omitempty" swaggertype:"primitive,object"`
	Status           *OidStatus      `json:"status,omitempty" swaggertype:"primitive,string" example:"current"`
	Access           *OidAccess      `json:"access,omitempty" swaggertype:"primitive,string" example:"read-only"`
	Units            string          `json:"units,omitempty" example:"seconds"`
	Description      string          `json:"description,omitempty" example:"A textual description of the entity."`
	Category         string          `json:"category,omitempty" example:"system"`
}

// CompareOids Сравнивает два OID по числовым сегментам dotter_notation.
// Возвращает true, если o1 должен идти раньше o2.
func CompareOids(o1, o2 string) bool {
	parts1 := strings.Split(o1, ".")
	parts2 := strings.Split(o2, ".")
	len1 := len(parts1)
	len2 := len(parts2)
	minLen := len1
	if len2 < minLen {
		minLen = len2
	}
	for i := 0; i < minLen; i++ {
		n1, err1 := strconv.Atoi(parts1[i])
		n2, err2 := strconv.Atoi(parts2[i])
		if err1 != nil || err2 != nil {
			if parts1[i] != parts2[i] {
				return parts1[i] < parts2[i]
			}
			continue
		}
		if n1 != n2 {
			return n1 < n2
		}
	}
	return len1 < len2
}

type DeviceIndicator struct {
	ID          int64   `json:"id" example:"1"`
	Description *string `json:"description,omitempty" example:"Hardware: x86 family"`
	ObjectID    *string `json:"object_id,omitempty" example:"1.3.6.1.4.1.9.1.516"`
	Contact     *string `json:"contact,omitempty" example:"sysadmin@company.com"`
	Name        *string `json:"name,omitempty" example:"Core-Switch-01"`
	Location    *string `json:"location,omitempty" example:"Server Room, Rack 3"`
	Services    *int16  `json:"services,omitempty" example:"12"`
}

type ParamIndicator struct {
	ID             int64   `json:"id" example:"1"`
	OidID          string  `json:"oid_id" example:"00000000-0000-0000-0000-000000000000"`
	DotterNotation *string `json:"dotter_notation,omitempty" example:"1.3.6.1.2.1.1.1.0"`
	Oid            *Oid    `json:"oid,omitempty"`
}

type Mapping struct {
	ID          int64            `json:"id" example:"1"`
	IndicatorID int64            `json:"indicator_id" example:"2"`
	ParamID     int64            `json:"param_id" example:"5"`
	Frequency   PollingFrequency `json:"frequency" swaggertype:"primitive,string" example:"MEDIUM"`
	Coefficient *string          `json:"coefficient,omitempty" example:"1.500000000000"`
	Enum        json.RawMessage  `json:"enum,omitempty" swaggertype:"primitive,object"`
}

type DeviceComponent struct {
	ID            int64             `json:"id" example:"1"`
	ModelID       int64             `json:"model_id" example:"10"`
	InternalOrder int32             `json:"internal_order" example:"1"`
	ParentID      *int64            `json:"parent_id,omitempty" example:"3"`
	Mappings      []Mapping         `json:"mappings"`
	Components    []DeviceComponent `json:"components"`
}

type Configuration struct {
	ID              int64            `json:"id" example:"1"`
	Indicator       *DeviceIndicator `json:"indicator"`
	DeviceComponent *DeviceComponent `json:"device_component,omitempty"`
	Thresholds      []Threshold      `json:"thresholds"`
}

type DefaultConfiguration struct {
	ID              int64            `json:"id" example:"1"`
	Indicator       *DeviceIndicator `json:"indicator"`
	DeviceComponent *DeviceComponent `json:"device_component,omitempty"`
	Thresholds      []Threshold      `json:"thresholds"`
}

type Threshold struct {
	ID                  int64         `json:"id" example:"1"`
	SourceModel         int64         `json:"source_model" example:"10"`
	SourceInternalOrder int64         `json:"source_internal_order" example:"1"`
	SourceParam         string        `json:"source_param" example:"temperature"`
	Value               string        `json:"value" example:"85"`
	Type                VarType       `json:"type" swaggertype:"primitive,string" example:"INT"`
	Operator            LogicOperator `json:"operator" swaggertype:"primitive,string" example:"=="`
	Enabled             bool          `json:"enabled" example:"true"`
	TargetParam         *string       `json:"target_param,omitempty" example:"alarm_state"`
	TargetDevice        *int64        `json:"target_device,omitempty" example:"5"`
	Level               AlarmLevel    `json:"level" swaggertype:"primitive,string" example:"WARNING"`
	PrevOperator        LogicOperator `json:"prev_operator" swaggertype:"primitive,string" example:"&&"`
	PreviousID          *int64        `json:"previous_id,omitempty" example:"2"`
	PreviousThreshold   *Threshold    `json:"previous_threshold,omitempty"`
}
