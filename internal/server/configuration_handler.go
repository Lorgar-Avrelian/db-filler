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

// CreateConfiguration создает рабочую конфигурацию
// @Summary         Создать рабочую конфигурацию
// @Tags            10. Конфигурация: Конфигурации устройств
// @Accept          json
// @Produce         json
// @Param           request body dto.ConfigurationCreate true "Данные конфигурации"
// @Success         201  {object}  model.Configuration "Возвращает полностью раскрытое содержание созданной конфигурации"
// @Failure         400  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/configurations [post]
func CreateConfiguration(c *gin.Context) {
	var input dto.ConfigurationCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := dao.CreateConfiguration(c.Request.Context(), input)
	if err != nil {
		logger.Error("Ошибка DAO при создании конфигурации: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ind, dc, th, _ := dao.GetDetailedConfigByID(c.Request.Context(), id)
	c.JSON(http.StatusCreated, model.Configuration{ID: id, Indicator: ind, DeviceComponent: dc, Thresholds: th})
}

// GetConfiguration возвращает конфигурацию по ID
// @Summary         Получить рабочую конфигурацию по ID
// @Tags            10. Конфигурация: Конфигурации устройств
// @Produce         json
// @Param           id   path      int  true  "ID Конфигурации"
// @Success         200  {object}  model.Configuration "Возвращает полностью раскрытое содержание"
// @Failure         404  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/configurations/{id} [get]
func GetConfiguration(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ind, dc, th, err := dao.GetDetailedConfigByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ind == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Конфигурация не найдена"})
		return
	}
	c.JSON(http.StatusOK, model.Configuration{ID: id, Indicator: ind, DeviceComponent: dc, Thresholds: th})
}

// GetAllConfigurations возвращает все конфигурации
// @Summary         Получить все рабочие конфигурации
// @Tags            10. Конфигурация: Конфигурации устройств
// @Produce         json
// @Success         200  {array}   model.Configuration
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/configurations [get]
func GetAllConfigurations(c *gin.Context) {
	res, err := dao.GetExpandedConfigurations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// UpdateConfiguration обновляет конфигурацию по ID
// @Summary         Обновить рабочую конфигурацию по ID
// @Tags            10. Конфигурация: Конфигурации устройств
// @Accept          json
// @Produce         json
// @Param           id      path      int  true  "ID Конфигурации"
// @Param           request body dto.ConfigurationUpdate true "Новые данные"
// @Success         200  {object}  model.Configuration "Возвращает полностью раскрытое обновленное содержание"
// @Failure         400  {object}  map[string]string
// @Failure         404  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/configurations/{id} [put]
func UpdateConfiguration(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input dto.ConfigurationUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedID, err := dao.UpdateConfiguration(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if updatedID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Конфигурация не найдена для обновления"})
		return
	}
	ind, dc, th, _ := dao.GetDetailedConfigByID(c.Request.Context(), updatedID)
	c.JSON(http.StatusOK, model.Configuration{ID: updatedID, Indicator: ind, DeviceComponent: dc, Thresholds: th})
}

// DeleteConfiguration удаляет конфигурацию по ID
// @Summary         Удалить рабочую конфигурацию по ID
// @Tags            10. Конфигурация: Конфигурации устройств
// @Param           id   path      int  true  "ID Конфигурации"
// @Success         204  "No Content"
// @Failure         404  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/configurations/{id} [delete]
func DeleteConfiguration(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	found, err := dao.DeleteConfiguration(c.Request.Context(), id)
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

// BindConfigThreshold связывает рабочую конфигурацию с порогом
// @Summary         Привязать порог к рабочей конфигурации
// @Tags            10. Конфигурация: Конфигурации устройств
// @Accept          json
// @Produce         json
// @Param           request body dto.BindParamRequest true "ID конфигурации и ID порога (используем BindParamRequest для совместимости структуры)"
// @Success         200  {object}  map[string]string
// @Failure         400  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/configurations/bind [post]
func BindConfigThreshold(c *gin.Context) {
	var input struct {
		ConfigurationID int64 `json:"configuration_id" binding:"required"`
		ThresholdID     int64 `json:"threshold_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := dao.BindConfigThreshold(c.Request.Context(), input.ConfigurationID, input.ThresholdID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Порог успешно привязан к рабочей конфигурации"})
}

// UnbindConfigThreshold разрывает связь рабочей конфигурации с порогом
// @Summary         Удалить связь рабочей конфигурации с порогом
// @Tags            10. Конфигурация: Конфигурации устройств
// @Param           configurationId path      int  true  "ID Рабочей конфигурации"
// @Param           thresholdId     path      int  true  "ID Порога"
// @Success         204  "No Content"
// @Failure         404  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/configurations/bind/{configurationId}/{thresholdId} [delete]
func UnbindConfigThreshold(c *gin.Context) {
	cfgID, _ := strconv.ParseInt(c.Param("configurationId"), 10, 64)
	tID, _ := strconv.ParseInt(c.Param("thresholdId"), 10, 64)
	found, err := dao.UnbindConfigThreshold(c.Request.Context(), cfgID, tID)
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
