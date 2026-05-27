package server

import (
	"filler/internal/config"
	"fmt"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router *gin.Engine
}

func NewServer() *Server {
	r := gin.Default()

	r.Static("/docs", "./docs")

	url := ginSwagger.URL("/docs/swagger.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	v1 := r.Group("/api/v1")
	{
		oids := v1.Group("/oids")
		{
			oids.GET("/exact", GetOidsByExactNotation)
			oids.GET("/prefix", GetOidsByPrefixNotation)
			oids.GET("/mib", GetOidsByMib)
			oids.GET("/vendor", GetOidsByVendor)
		}

		components := v1.Group("/components")
		{
			components.POST("", CreateComponent)
			components.GET("", GetAllComponents)
			components.GET("/search", SearchComponents)
			components.GET("/:id", GetComponent)
			components.PUT("/:id", UpdateComponent)
			components.DELETE("/:id", DeleteComponent)
		}

		indicators := v1.Group("/indicators")
		{
			indicators.POST("", CreateIndicator)
			indicators.GET("", GetAllIndicators)
			indicators.GET("/:id", GetIndicator)
			indicators.PUT("/:id", UpdateIndicator)
			indicators.DELETE("/:id", DeleteIndicator)
		}

		params := v1.Group("/params")
		{
			params.POST("", CreateParam)
			params.GET("/unattached", GetUnattachedParams)
			params.GET("/search", SearchParams)
			params.GET("/:id", GetParam)
			params.PUT("/:id", UpdateParam)
			params.DELETE("/:id", DeleteParam)
		}

		relations := v1.Group("/relations")
		{
			relations.POST("", BindParam)
			relations.DELETE("/:componentId/:paramId", UnbindParam)
		}

		paramIndicators := v1.Group("/param-indicators")
		{
			paramIndicators.POST("", CreateParamIndicator)
			paramIndicators.GET("", GetAllParamIndicators)
			paramIndicators.GET("/:id", GetParamIndicator)
			paramIndicators.PUT("/:id", UpdateParamIndicator)
			paramIndicators.DELETE("/:id", DeleteParamIndicator)
		}

		mappings := v1.Group("/mappings")
		{
			mappings.POST("", CreateMapping)
			mappings.GET("", GetAllMappings)
			mappings.GET("/:id", GetMapping)
			mappings.PUT("/:id", UpdateMapping)
			mappings.DELETE("/:id", DeleteMapping)
		}

		deviceComponents := v1.Group("/device-components")
		{
			deviceComponents.POST("", CreateDeviceComponent)
			deviceComponents.GET("", GetAllDeviceComponents)
			deviceComponents.POST("/bind", BindDeviceMapping)
			deviceComponents.GET("/:id", GetDeviceComponent)
			deviceComponents.PUT("/:id", UpdateDeviceComponent)
			deviceComponents.DELETE("/:id", DeleteDeviceComponent)
			deviceComponents.DELETE("/bind/:deviceComponentId/:mappingId", UnbindDeviceMapping)
		}
	}

	return &Server{router: r}
}

func (s *Server) Run() error {
	cfg := config.Get()
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	return s.router.Run(addr)
}
