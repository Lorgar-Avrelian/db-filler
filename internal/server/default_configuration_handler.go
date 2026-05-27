package server

import (
	"filler/internal/dao"
	"filler/internal/dto"
	"filler/internal/logger"
	"filler/internal/model"
	"net/http"
	"strconv"

	_ "filler/internal/model"

	"github.com/gin-gonic/gin"
)

// CreateDefaultConfiguration создает конфигурацию по умолчанию
// @Summary         Создать конфигурацию по умолчанию
// @Tags            9. Конфигурация: Конфигурации по-умолчанию
// @Accept          json
// @Produce         json
// @Param           request body dto.ConfigurationCreate true "Данные конфигурации"
// @Success         201  {object}  model.DefaultConfiguration "Возвращает полностью раскрытое содержание созданной конфигурации"
// @Failure         400  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/default-configurations [post]
func CreateDefaultConfiguration(c *gin.Context) {
	var input dto.ConfigurationCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := dao.CreateDefaultConfiguration(c.Request.Context(), input)
	if err != nil {
		logger.Error("Ошибка DAO при создании дефолтной конфигурации: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ind, dc, th, _ := dao.GetDetailedDefaultConfigByID(c.Request.Context(), id)
	c.JSON(http.StatusCreated, model.DefaultConfiguration{ID: id, Indicator: ind, DeviceComponent: dc, Thresholds: th})
}

// GetDefaultConfiguration возвращает конфигурацию по умолчанию по ID
// @Summary         Получить дефолтную конфигурацию по ID
// @Tags            9. Конфигурация: Конфигурации по-умолчанию
// @Produce         json
// @Param           id   path      int  true  "ID Конфигурации"
// @Success         200  {object}  model.DefaultConfiguration "Возвращает полностью раскрытое содержание"
// @Failure         404  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/default-configurations/{id} [get]
func GetDefaultConfiguration(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ind, dc, th, err := dao.GetDetailedDefaultConfigByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ind == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Конфигурация по умолчанию не найдена"})
		return
	}
	c.JSON(http.StatusOK, model.DefaultConfiguration{ID: id, Indicator: ind, DeviceComponent: dc, Thresholds: th})
}

// GetAllDefaultConfigurations возвращает все конфигурации по умолчанию
// @Summary         Получить все конфигурации по умолчанию
// @Tags            9. Конфигурация: Конфигурации по-умолчанию
// @Produce         json
// @Success         200  {array}   model.DefaultConfiguration
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/default-configurations [get]
func GetAllDefaultConfigurations(c *gin.Context) {
	res, err := dao.GetExpandedDefaultConfigurations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// UpdateDefaultConfiguration обновляет дефолтную конфигурацию по ID
// @Summary         Обновить конфигурацию по умолчанию по ID
// @Tags            9. Конфигурация: Конфигурации по-умолчанию
// @Accept          json
// @Produce         json
// @Param           id      path      int  true  "ID Конфигурации"
// @Param           request body dto.ConfigurationUpdate true "Новые данные"
// @Success         200  {object}  model.DefaultConfiguration "Возвращает полностью раскрытое обновленное содержание"
// @Failure         400  {object}  map[string]string
// @Failure         404  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/default-configurations/{id} [put]
func UpdateDefaultConfiguration(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input dto.ConfigurationUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedID, err := dao.UpdateDefaultConfiguration(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if updatedID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Конфигурация по умолчанию не найдена для обновления"})
		return
	}
	ind, dc, th, _ := dao.GetDetailedDefaultConfigByID(c.Request.Context(), updatedID)
	c.JSON(http.StatusOK, model.DefaultConfiguration{ID: updatedID, Indicator: ind, DeviceComponent: dc, Thresholds: th})
}

// DeleteDefaultConfiguration удаляет дефолтную конфигурацию по ID
// @Summary         Удалить дефолтную конфигурацию по ID
// @Tags            9. Конфигурация: Конфигурации по-умолчанию
// @Param           id   path      int  true  "ID Конфигурации"
// @Success         204  "No Content"
// @Failure         404  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/default-configurations/{id} [delete]
func DeleteDefaultConfiguration(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	found, err := dao.DeleteDefaultConfiguration(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Конфигурация не найдена"})
		return
	}
	c.Status(http.StatusNoContent)
}

// BindDefaultConfigThreshold связывает дефолтную конфигурацию с порогом
// @Summary         Привязать порог к дефолтной конфигурации
// @Tags            9. Конфигурация: Конфигурации по-умолчанию
// @Accept          json
// @Produce         json
// @Param           request body dto.BindParamRequest true "ID конфигурации и ID порога"
// @Success         200  {object}  map[string]string
// @Failure         400  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/default-configurations/bind [post]
func BindDefaultConfigThreshold(c *gin.Context) {
	var input struct {
		DefaultConfigurationID int64 `json:"default_configuration_id" binding:"required"`
		ThresholdID            int64 `json:"threshold_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := dao.BindDefaultConfigThreshold(c.Request.Context(), input.DefaultConfigurationID, input.ThresholdID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Порог успешно привязан к дефолтной конфигурации"})
}

// UnbindDefaultConfigThreshold разрывает связь дефолтной конфигурации с порогом
// @Summary         Удалить связь дефолтной конфигурации с порогом
// @Tags            9. Конфигурация: Конфигурации по-умолчанию
// @Param           defaultConfigurationId path      int  true  "ID Дефолтной конфигурации"
// @Param           thresholdId            path      int  true  "ID Порога"
// @Success         204  "No Content"
// @Failure         404  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/default-configurations/bind/{defaultConfigurationId}/{thresholdId} [delete]
func UnbindDefaultConfigThreshold(c *gin.Context) {
	defCfgID, _ := strconv.ParseInt(c.Param("defaultConfigurationId"), 10, 64)
	tID, _ := strconv.ParseInt(c.Param("thresholdId"), 10, 64)
	found, err := dao.UnbindDefaultConfigThreshold(c.Request.Context(), defCfgID, tID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Связь не найдена"})
		return
	}
	c.Status(http.StatusNoContent)
}
