package server

import (
	"filler/internal/dao"
	"filler/internal/dto"
	"filler/internal/logger"
	"filler/internal/model"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetOidsByExactNotation возвращает OID по точному совпадению dotter_notation
// @Summary         Полноразмерный OID по dotter_notation
// @Tags            4. Парсер: OID
// @Produce         json
// @Param           notation query    string  true  "Точная dotter_notation"
// @Success         200      {array}  model.Oid
// @Failure         400      {object} map[string]string
// @Failure         500      {object} map[string]string
// @Router          /api/v1/oids/exact [get]
func GetOidsByExactNotation(c *gin.Context) {
	notation := c.Query("notation")
	if notation == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр 'notation' не может быть пустым"})
		return
	}
	res, err := dao.GetOidsByExactDotter(c.Request.Context(), notation)
	if err != nil {
		logger.Error("Ошибка DAO при выборке точного OID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetOidsByPrefixNotation возвращает OID по префиксу с пагинацией и сортировкой в Go
// @Summary         Поиск OID по префиксу с пагинацией
// @Description     Возвращает отсортированный по dotter_notation список OID (по 100 на страницу). Сортировка на стороне приложения.
// @Tags            4. Парсер: OID
// @Produce         json
// @Param           prefix   query    string  true  "Префикс dotter_notation"
// @Param           page     query    int     false "Номер страницы (дефолт: 1)"
// @Success         200      {object} dto.OidPageResponse
// @Failure         400      {object} map[string]string
// @Failure         500      {object} map[string]string
// @Router          /api/v1/oids/prefix [get]
func GetOidsByPrefixNotation(c *gin.Context) {
	prefix := c.Query("prefix")
	if prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр 'prefix' не может быть пустым"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	res, err := dao.GetOidsByDotterPrefix(c.Request.Context(), prefix)
	if err != nil {
		logger.Error("Ошибка DAO при выборке OID по префиксу: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(res, func(i, j int) bool {
		return model.CompareOids(res[i].DotterNotation, res[j].DotterNotation)
	})

	perPage := 100
	totalItems := len(res)
	start := (page - 1) * perPage
	end := start + perPage

	var pagedItems []model.Oid
	if start < totalItems {
		if end > totalItems {
			end = totalItems
		}
		pagedItems = res[start:end]
	} else {
		pagedItems = []model.Oid{}
	}

	c.JSON(http.StatusOK, dto.OidPageResponse{
		Page:       page,
		PerPage:    perPage,
		TotalItems: totalItems,
		Items:      pagedItems,
	})
}

// GetOidsByMib возвращает OID по названию MIB
// @Summary         Получить OID по имени MIB
// @Tags            4. Парсер: OID
// @Produce         json
// @Param           name     query    string  true  "Название MIB"
// @Success         200      {array}  model.Oid
// @Failure         400      {object} map[string]string
// @Failure         500      {object} map[string]string
// @Router          /api/v1/oids/mib [get]
func GetOidsByMib(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр 'name' не может быть пустым"})
		return
	}
	res, err := dao.GetOidsByMibName(c.Request.Context(), name)
	if err != nil {
		logger.Error("Ошибка DAO при выборке OID по MIB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetOidsByVendor возвращает OID вендора с пагинацией и сортировкой в Go
// @Summary         Получить OID по вендору с пагинацией
// @Description     Находит вендора в кэше памяти, выгружает OID, сортирует в Go и отдает по 100 штук на страницу.
// @Tags            4. Парсер: OID
// @Produce         json
// @Param           identity query    string  true  "Имя вендора или его директория"
// @Param           page     query    int     false "Номер страницы (дефолт: 1)"
// @Success         200      {object} dto.OidPageResponse
// @Failure         400      {object} map[string]string
// @Failure         404      {object} map[string]string
// @Failure         500      {object} map[string]string
// @Router          /api/v1/oids/vendor [get]
func GetOidsByVendor(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр 'identity' не может быть пустым"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if 1 > page {
		page = 1
	}
	vID, found := model.LookupVendorID(identity)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Вендор не найден в памяти приложения"})
		return
	}
	res, err := dao.GetOidsByVendorID(c.Request.Context(), vID)
	if err != nil {
		logger.Error("Ошибка DAO при выборке OID по вендору: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sort.Slice(res, func(i, j int) bool {
		return model.CompareOids(res[i].DotterNotation, res[j].DotterNotation)
	})
	perPage := 100
	totalItems := len(res)
	start := (page - 1) * perPage
	end := start + perPage
	var pagedItems []model.Oid
	if totalItems > start {
		if end > totalItems {
			end = totalItems
		}
		pagedItems = res[start:end]
	} else {
		pagedItems = []model.Oid{}
	}
	c.JSON(http.StatusOK, dto.OidPageResponse{
		Page:       page,
		PerPage:    perPage,
		TotalItems: totalItems,
		Items:      pagedItems,
	})
}

// GetOidsByDotterAndMib возвращает OID по точному совпадению dotter_notation и имени MIB
// @Summary         Поиск OID по dotter_notation и имени MIB
// @Tags            4. Парсер: OID
// @Produce         json
// @Param           notation query    string  true  "Точная dotter_notation"
// @Param           mib      query    string  true  "Точное название MIB"
// @Success         200      {array}  model.Oid
// @Failure         400      {object} map[string]string
// @Failure         500      {object} map[string]string
// @Router          /api/v1/oids/exact-with-mib [get]
func GetOidsByDotterAndMib(c *gin.Context) {
	notation := c.Query("notation")
	mibName := c.Query("mib")
	if notation == "" || mibName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметры 'notation' и 'mib' не могут быть пустыми"})
		return
	}
	res, err := dao.GetOidsByDotterAndMibName(c.Request.Context(), notation, mibName)
	if err != nil {
		logger.Error("Ошибка DAO при выборке OID по dotter и MIB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
