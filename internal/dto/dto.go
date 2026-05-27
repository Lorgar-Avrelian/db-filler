package dto

import "filler/internal/model"

type ComponentCreate struct {
	Title         string `json:"title" binding:"required" example:"Модуль Питания"`
	NameEn        string `json:"name_en" binding:"required" example:"power_module"`
	NameRu        string `json:"name_ru" binding:"required" example:"модуль_питания"`
	BaseComponent *int64 `json:"base_component,omitempty" example:"1"`
	DescriptionEn string `json:"description_en,omitempty" example:"Main power controller unit"`
	DescriptionRu string `json:"description_ru,omitempty" example:"Главный контроллер питания"`
	Access        string `json:"access" binding:"required" example:"ADMIN"`
}

type ComponentUpdate struct {
	Title         string `json:"title" binding:"required" example:"Обновленный Модуль"`
	NameEn        string `json:"name_en" binding:"required" example:"power_module_v2"`
	NameRu        string `json:"name_ru" binding:"required" example:"модуль_питания_в2"`
	BaseComponent *int64 `json:"base_component,omitempty" example:"1"`
	DescriptionEn string `json:"description_en,omitempty" example:"Updated main power controller unit"`
	DescriptionRu string `json:"description_ru,omitempty" example:"Обновленный главный контроллер питания"`
	Access        string `json:"access" binding:"required" example:"OWNER"`
}

type ParamCreate struct {
	Title         string `json:"title" binding:"required" example:"Макс. Напряжение"`
	NameEn        string `json:"name_en" binding:"required" example:"max_voltage"`
	NameRu        string `json:"name_ru" binding:"required" example:"макс_напряжение"`
	Type          string `json:"type" binding:"required" example:"INT"`
	Value         string `json:"value,omitempty" example:"220"`
	DescriptionEn string `json:"description_en,omitempty" example:"Maximum operational voltage"`
	DescriptionRu string `json:"description_ru,omitempty" example:"Максимальное рабочее напряжение"`
	UnitsEn       string `json:"units_en,omitempty" example:"V"`
	UnitsRu       string `json:"units_ru,omitempty" example:"В"`
	Access        string `json:"access" binding:"required" example:"ADMIN"`
	Saved         bool   `json:"saved" example:"true"`
	Visible       bool   `json:"visible" example:"true"`
}

type ParamUpdate struct {
	Title         string `json:"title" binding:"required" example:"Макс. Напряжение"`
	NameEn        string `json:"name_en" binding:"required" example:"max_voltage"`
	NameRu        string `json:"name_ru" binding:"required" example:"макс_напряжение"`
	Type          string `json:"type" binding:"required" example:"INT"`
	Value         string `json:"value,omitempty" example:"240"`
	DescriptionEn string `json:"description_en,omitempty" example:"Maximum operational voltage"`
	DescriptionRu string `json:"description_ru,omitempty" example:"Максимальное рабочее напряжение"`
	UnitsEn       string `json:"units_en,omitempty" example:"V"`
	UnitsRu       string `json:"units_ru,omitempty" example:"В"`
	Access        string `json:"access" binding:"required" example:"ADMIN"`
	Saved         bool   `json:"saved" example:"true"`
	Visible       bool   `json:"visible" example:"true"`
}

type BindParamRequest struct {
	ComponentID int64 `json:"component_id" binding:"required" example:"10"`
	ParamID     int64 `json:"param_id" binding:"required" example:"5"`
}

type OidPageResponse struct {
	Page       int         `json:"page" example:"1"`
	PerPage    int         `json:"per_page" example:"100"`
	TotalItems int         `json:"total_items" example:"450"`
	Items      []model.Oid `json:"items"`
}

type DeviceIndicatorCreate struct {
	Description *string `json:"description" example:"Hardware: x86 family"`
	ObjectID    *string `json:"object_id" example:"1.3.6.1.4.1.9.1.516"`
	Contact     *string `json:"contact" example:"sysadmin@company.com"`
	Name        *string `json:"name" example:"Core-Switch-01"`
	Location    *string `json:"location" example:"Server Room, Rack 3"`
	Services    *int16  `json:"services" example:"72"`
}

type DeviceIndicatorUpdate struct {
	Description *string `json:"description" example:"Updated Hardware info"`
	ObjectID    *string `json:"object_id" example:"1.3.6.1.4.1.9.1.516"`
	Contact     *string `json:"contact" example:"sysadmin_new@company.com"`
	Name        *string `json:"name" example:"Core-Switch-01-Updated"`
	Location    *string `json:"location" example:"Server Room, Rack 4"`
	Services    *int16  `json:"services" example:"74"`
}

type ParamIndicatorCreate struct {
	OidID          string  `json:"oid_id" binding:"required" example:"00000000-0000-0000-0000-000000000000"`
	DotterNotation *string `json:"dotter_notation" example:"1.3.6.1.2.1.1.1.0"`
}

type ParamIndicatorUpdate struct {
	OidID          string  `json:"oid_id" binding:"required" example:"00000000-0000-0000-0000-000000000000"`
	DotterNotation *string `json:"dotter_notation" example:"1.3.6.1.2.1.1.3.0"`
}

type MappingCreate struct {
	IndicatorID int64   `json:"indicator_id" binding:"required" example:"2"`
	ParamID     int64   `json:"param_id" binding:"required" example:"5"`
	Frequency   string  `json:"frequency" binding:"required" example:"MEDIUM"` // Строковый энум
	Coefficient *string `json:"coefficient,omitempty" example:"1.5"`
}

type MappingUpdate struct {
	IndicatorID int64   `json:"indicator_id" binding:"required" example:"2"`
	ParamID     int64   `json:"param_id" binding:"required" example:"5"`
	Frequency   string  `json:"frequency" binding:"required" example:"HIGH"`
	Coefficient *string `json:"coefficient,omitempty" example:"2.0"`
}

type DeviceComponentCreate struct {
	ModelID       int64  `json:"model_id" binding:"required" example:"10"`
	InternalOrder int32  `json:"internal_order" example:"1"`
	ParentID      *int64 `json:"parent_id,omitempty" example:"3"`
}

type DeviceComponentUpdate struct {
	ModelID       int64  `json:"model_id" binding:"required" example:"10"`
	InternalOrder int32  `json:"internal_order" example:"2"`
	ParentID      *int64 `json:"parent_id,omitempty" example:"3"`
}

type BindDeviceMappingRequest struct {
	DeviceComponentID int64 `json:"device_component_id" binding:"required" example:"1"`
	MappingID         int64 `json:"mapping_id" binding:"required" example:"5"`
}
