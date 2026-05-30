package server

import (
	"filler/internal/dao"
	"filler/internal/dto"
	"filler/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SaveResult выгружает состояние 13 таблиц в файл формата Liquibase SQL
// @Summary         Экспортировать результаты в Liquibase скрипт
// @Description     Выгружает данные 13 основных таблиц в структурированный .sql файл в корне проекта с разбиением на changeset-ы
// @Tags            12. Результат: Экспортировать БД в Liquibase скрипт
// @Accept          json
// @Produce         json
// @Param           request body dto.SaveResultRequest true "Параметры экспорта"
// @Success         200  {object}  map[string]string "Сообщение об успешном создании файла"
// @Failure         400  {object}  map[string]string "Ошибка валидации данных"
// @Failure         500  {object}  map[string]string "Внутренняя ошибка при записи файла или чтении БД"
// @Router          /api/v1/save-result [post]
func SaveResult(c *gin.Context) {
	var input dto.SaveResultRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Warn("Ошибка валидации при вызове экспорта БД: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := dao.ExportDatabaseToLiquibase(c.Request.Context(), input.Filename, input.Author, input.StartValue)
	if err != nil {
		logger.Error("Критическая ошибка при генерации Liquibase скрипта: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Файл миграции %s успешно сгенерирован автором %s", input.Filename, input.Author)
	c.JSON(http.StatusOK, gin.H{
		"message": "Скрипт заполнения данных Liquibase успешно сформирован и сохранен в корне приложения",
		"file":    input.Filename,
	})
}
