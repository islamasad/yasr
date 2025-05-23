package demo

import (
	"yasr/internal/api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterDemoRoutes(r *gin.RouterGroup, db *gorm.DB) {
	demoGroup := r.Group("/demo",
		middleware.RateLimiter(),
		middleware.DatabaseMiddleware(db),
	)
	{
		demoGroup.GET("/dashboard", DemoDashboardHandler)
		demoGroup.GET("/dashboard/qr", DemoQRHandler)
		demoGroup.GET("/cashier", DemoCashierHandler)
		demoGroup.GET("/api/orders", GetOrdersHandler)
		demoGroup.GET("/order", DemoOrderHandler)
		demoGroup.GET("/menu", MenuHandler)
		demoGroup.POST("/order", SubmitOrderHandler)
		demoGroup.PUT("/orders/:id/status", UpdateOrderStatus)
		demoGroup.GET("/orders/stream", OrderStreamHandler)
	}
}
