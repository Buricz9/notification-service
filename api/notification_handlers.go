package main

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	redisHigh       = "notifications:pending_high_priority"
	statusPending   = "Pending"
	statusSent      = "Sent"
	statusDelivered = "Delivered"
	statusFailed    = "Failed"
	statusCancelled = "Cancelled"
)

// wszystkie ścieżki /api/notifications/**
func registerRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client) {
	g := r.Group("/api/notifications")

	g.POST("/", func(c *gin.Context) {
		var dto CreateNotificationDTO
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		n := Notification{
			Recipient:   dto.Recipient,
			Message:     dto.Message,
			CreatedAt:   nowUnix(),
			ScheduledAt: dto.ScheduledAt,
			Priority:    dto.Priority,
			Status:      statusPending,
			Channel:     dto.Channel,
			TimeZone:    dto.TimeZone,
			RetryCnt:    0,
		}

		if err := db.Create(&n).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Header("Location", "/api/notifications/"+intToStr(n.ID))
		c.JSON(http.StatusCreated, n)
	})

	g.GET("/:id", func(c *gin.Context) {
		var n Notification
		if err := db.First(&n, c.Param("id")).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.Status(http.StatusNotFound)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, n)
	})

	g.GET("/", func(c *gin.Context) {
		var list []Notification
		if err := db.Find(&list).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, list)
	})

	g.PUT("/:id", func(c *gin.Context) {
		var dto ModifyNotificationDTO
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var n Notification
		if err := db.First(&n, c.Param("id")).Error; err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// modyfikacje warunkowe
		if dto.Recipient != nil {
			n.Recipient = *dto.Recipient
		}
		if dto.Message != nil {
			n.Message = *dto.Message
		}
		if dto.ScheduledAt != nil {
			if *dto.ScheduledAt < nowUnix() {
				c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled time cannot be in the past"})
				return
			}
			n.ScheduledAt = *dto.ScheduledAt
		}
		if dto.Priority != nil {
			n.Priority = *dto.Priority
		}
		if dto.Channel != nil {
			n.Channel = *dto.Channel
		}
		if dto.TimeZone != nil {
			n.TimeZone = *dto.TimeZone
		}

		if err := db.Save(&n).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	g.DELETE("/:id", func(c *gin.Context) {
		if err := db.Delete(&Notification{}, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// ---------------------------------------------
	//  POST /api/notifications/:id/send-now
	// ---------------------------------------------
	g.POST("/:id/send-now", func(c *gin.Context) {
		var n Notification
		if err := db.First(&n, c.Param("id")).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.Status(http.StatusNotFound)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		// jeśli już zakończone – nie wysyłamy ponownie
		if n.Status == statusDelivered || n.Status == statusFailed || n.Status == statusCancelled {
			c.JSON(http.StatusConflict, gin.H{"error": "notification already finalized"})
			return
		}

		// publish na kolejkę High (zawsze!)
		raw, _ := json.Marshal(n)
		if err := rdb.Publish(context.Background(), redisHigh, raw).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// status -> Sent, RetryCnt bez zmian
		n.Status = statusSent

		if err := db.Save(&n).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusAccepted) // 202
	})

	// ----------------------------------------------------
	//  POST /api/notifications/:id/cancel
	// ----------------------------------------------------
	g.POST("/:id/cancel", func(c *gin.Context) {
		var n Notification
		if err := db.First(&n, c.Param("id")).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.Status(http.StatusNotFound)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		// jeśli już “finalized”, nie wolno anulować
		if n.Status == statusDelivered || n.Status == statusFailed || n.Status == statusCancelled {
			c.JSON(http.StatusConflict, gin.H{"error": "notification already finalized"})
			return
		}

		n.Status = statusCancelled
		if err := db.Save(&n).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusAccepted) // 202
	})

}
