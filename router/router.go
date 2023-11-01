package router

import (
	"alb-manager/apisix/route"
	"alb-manager/apisix/service"
	"alb-manager/conf"
	"github.com/gin-gonic/gin"
	"net/http"
)

func InitRouter() *gin.Engine {
	router := gin.Default()

	// 添加 Token 验证中间件
	router.Use(tokenMiddleware())

	// 测试路由
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test ok!")
	})

	// 服务器组相关
	v1 := router.Group("/api/v1/service")
	{
		// 节点上下线
		v1.POST("/updown", service.NodeUpDown)

		// 创建服务器组
		v1.POST("/create", service.CreateServerGroup)

		// 编辑服务器组
		v1.POST("/update", service.UpdateServerGroup)

		// 删除服务器组
		v1.POST("/delete", service.DeleteServerGroup)

		// 查询服务器组
		v1.POST("/get", service.GetServerGroup)

		// 反查服务器组
		v1.POST("/get_by_ip", service.GetServerGroupByIp)

		// 查询节点数量
		v1.POST("/get_node_num", service.GetServerGroupNodes)
	}

	// 路由规则相关
	v2 := router.Group("/api/v1/route")
	{
		v2.Any("/test", route.Test)
		// 创建路由规则
		v2.POST("/create", route.CreateRoutes)

		// 编辑路由规则
		v2.POST("/update", route.UpdateRoutes)

		// 删除路由规则
		v2.POST("/delete", route.DeleteRoutes)

		// 查询路由规则
		v2.POST("/get", route.GetRoutes)

		// 启用/禁用路由规则
		v2.POST("/updown", route.UpdownRoutes)
	}

	return router
}

// Token 验证中间件
func tokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("OPS-TOKEN")
		if token != conf.ViperConfig.GetString("ops-token") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}
		c.Next()
	}
}
