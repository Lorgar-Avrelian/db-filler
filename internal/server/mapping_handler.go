package server

import (
	"filler/internal/dao"
	"filler/internal/dto"
	"filler/internal/logger"
	"net/http"
	"strconv"

	_ "filler/internal/model"

	"github.com/gin-gonic/gin"
)

// CreateMapping создает новый маппинг
// @Summary         Создать маппинг
// @Tags            7. Конфигурация: Сопоставления параметров
// @Accept          json
// @Produce         json
// @Param           request body dto.MappingCreate true "Данные маппинга"
// @Success         201  {object}  model.Mapping
// @Failure         400  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/mappings [post]
func CreateMapping(c *gin.Context) {
	var input dto.MappingCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Warn("Ошибка валидации при создании маппинга: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := dao.CreateMapping(c.Request.Context(), input)
	if err != nil {
		logger.Error("Ошибка DAO при создании маппинга: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

// GetMapping возвращает маппинг по ID
// @Summary         Получить маппинг по ID
// @Tags            7. Конфигурация: Сопоставления параметров
// @Produce         json
// @Param           id   path      int  true  "ID Маппинга"
// @Success         200  {object}  model.Mapping
// @Failure         400  {object}  map[string]string
// @Failure         404  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/mappings/{id} [get]
func GetMapping(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID маппинга"})
		return
	}
	res, err := dao.GetMappingByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("Ошибка DAO при получении маппинга %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Маппинг не найден"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetAllMappings возвращает все маппинги
// @Summary         Получить все маппинги
// @Tags            7. Конфигурация: Сопоставления параметров
// @Produce         json
// @Success         200  {array}   model.Mapping
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/mappings [get]
func GetAllMappings(c *gin.Context) {
	res, err := dao.GetAllMappings(c.Request.Context())
	if err != nil {
		logger.Error("Ошибка DAO при выгрузке всех маппингов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// UpdateMapping обновляет маппинг по ID
// @Summary         Обновить маппинг по ID
// @Tags            7. Конфигурация: Сопоставления параметров
// @Accept          json
// @Produce         json
// @Param           id      path      int  true  "ID Маппинга"
// @Param           request body dto.MappingUpdate true "Новые данные"
// @Success         200  {object}  model.Mapping
// @Failure         400  {object}  map[string]string
// @Failure         404  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/mappings/{id} [put]
func UpdateMapping(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID маппинга"})
		return
	}
	var input dto.MappingUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Warn("Ошибка валидации при обновлении маппинга %d: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := dao.UpdateMapping(c.Request.Context(), id, input)
	if err != nil {
		logger.Error("Ошибка DAO при обновлении маппинга %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Маппинг не найден для обновления"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// DeleteMapping удаляет маппинг по ID
// @Summary         Удалить маппинг по ID
// @Tags            7. Конфигурация: Сопоставления параметров
// @Param           id   path      int  true  "ID Маппинга"
// @Success         204  "No Content"
// @Failure         400  {object}  map[string]string
// @Failure         404  {object}  map[string]string
// @Failure         500  {object}  map[string]string
// @Router          /api/v1/mappings/{id} [delete]
func DeleteMapping(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID маппинга"})
		return
	}
	found, err := dao.DeleteMapping(c.Request.Context(), id)
	if err != nil {
		logger.Error("Ошибка DAO при удалении маппинга %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Маппинг не найден в системе"})
		return
	}
	c.Status(http.StatusNoContent)
}
