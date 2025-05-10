package httpdelivery

import (
	"net/http"
	"strconv"

	"github.com/Buricz9/notification-service/internal/domain"
	"github.com/Buricz9/notification-service/internal/dto"
	"github.com/Buricz9/notification-service/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up HTTP endpoints for notifications
func RegisterRoutes(r *gin.Engine, svc *service.NotificationService) {
	g := r.Group("/api/notifications")

	g.POST("/", createHandler(svc))
	g.GET("/:id", getHandler(svc))
	g.GET("/", listHandler(svc))
	g.PUT("/:id", updateHandler(svc))
	g.DELETE("/:id", deleteHandler(svc))
	g.POST("/:id/send-now", sendNowHandler(svc))
	g.POST("/:id/cancel", cancelHandler(svc))
}

func createHandler(svc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.Create
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		n := domain.Notification{
			Recipient:   req.Recipient,
			Message:     req.Message,
			ScheduledAt: req.ScheduledAt,
			Priority:    domain.Priority(req.Priority),
			Channel:     req.Channel,
			TimeZone:    req.TimeZone,
		}
		n, err := svc.Create(c.Request.Context(), n)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Header("Location", "/api/notifications/"+strconv.FormatUint(uint64(n.ID), 10))
		c.JSON(http.StatusCreated, n)
	}
}

func getHandler(svc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		n, err := svc.FindByID(c.Request.Context(), uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, n)
	}
}

func listHandler(svc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		list, err := svc.List(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, list)
	}
}

func updateHandler(svc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var req dto.ModifyNotification
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		n := domain.Notification{ID: uint(id)}
		if req.Recipient != nil {
			n.Recipient = *req.Recipient
		}
		if req.Message != nil {
			n.Message = *req.Message
		}
		if req.ScheduledAt != nil {
			n.ScheduledAt = *req.ScheduledAt
		}
		if req.Priority != nil {
			n.Priority = domain.Priority(*req.Priority)
		}
		if req.Channel != nil {
			n.Channel = *req.Channel
		}
		if req.TimeZone != nil {
			n.TimeZone = *req.TimeZone
		}

		if err := svc.Modify(c.Request.Context(), n); err != nil {
			switch err {
			case service.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			case service.ErrFinalized:
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func deleteHandler(svc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		if err := svc.Delete(c.Request.Context(), uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func sendNowHandler(svc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		if err := svc.SendNow(c.Request.Context(), uint(id)); err != nil {
			switch err {
			case service.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			case service.ErrFinalized:
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.Status(http.StatusAccepted)
	}
}

func cancelHandler(svc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		if err := svc.Cancel(c.Request.Context(), uint(id)); err != nil {
			switch err {
			case service.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			case service.ErrFinalized:
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.Status(http.StatusAccepted)
	}
}
