package handlers

import (
	"context"

	"github.com/ahmetkoprulu/rtrp/internal/services"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	service *services.EventService
}

func NewEventHandler(service *services.EventService) *EventHandler {
	return &EventHandler{service: service}
}

func (h *EventHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	events := router.Group("/events")
	{
		events.POST("/", h.CreateEvent)
		events.GET("/", h.ListEvents)
		events.PUT("/:id", h.UpdateEvent)
		events.GET("/:id", h.GetEvent)
		events.POST("/:id/schedules", h.CreateSchedule)
		events.GET("/:id/schedules", h.GetSchedulesByEvent)

		events.POST("/schedules/:id/play", authMiddleware, h.PlayEvent)
		events.PUT("/schedules/:id", h.UpdateSchedule)
		events.GET("/schedules/actives", h.ListActiveSchedules)
		events.GET("/schedules/:id", h.GetSchedule)
		events.GET("/schedules/:id/player", authMiddleware, h.GetPlayerEvent)
	}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	model := BindModel[models.Event](c)
	if model == nil {
		return
	}

	if err := h.service.CreateEvent(c.Request.Context(), model); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, model)
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	model := BindModel[models.Event](c)
	if model == nil {
		return
	}

	model.ID = c.Param("id")
	if err := h.service.UpdateEvent(c.Request.Context(), model); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, model)
}

func (h *EventHandler) GetEvent(c *gin.Context) {
	id := c.Param("id")
	event, err := h.service.GetEvent(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "Event not found")
		return
	}

	Ok(c, event)
}

func (h *EventHandler) ListEvents(c *gin.Context) {
	events, err := h.service.ListEvents(c.Request.Context())
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, events)
}

func (h *EventHandler) CreateSchedule(c *gin.Context) {
	model := BindModel[models.EventSchedule](c)
	if model == nil {
		return
	}

	model.EventID = c.Param("id")
	if err := h.service.CreateSchedule(c.Request.Context(), model); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, model)
}

func (h *EventHandler) UpdateSchedule(c *gin.Context) {
	model := BindModel[models.EventSchedule](c)
	if model == nil {
		return
	}

	model.ID = c.Param("id")
	if err := h.service.UpdateSchedule(c.Request.Context(), model); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, model)
}

func (h *EventHandler) GetSchedule(c *gin.Context) {
	id := c.Param("id")
	schedule, err := h.service.GetSchedule(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "Schedule not found")
		return
	}

	Ok(c, schedule)
}

func (h *EventHandler) ListActiveSchedules(c *gin.Context) {
	schedules, err := h.service.ListActiveSchedules(c.Request.Context())
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, schedules)
}

func (h *EventHandler) GetSchedulesByEvent(c *gin.Context) {
	eventID := c.Param("id")
	schedules, err := h.service.GetSchedulesByEvent(c.Request.Context(), eventID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, schedules)
}

func (h *EventHandler) PlayEvent(c *gin.Context) {
	scheduleId := c.Param("id")
	playerId := c.GetString("playerID")
	if playerId == "" {
		BadRequest(c, "Player ID is required")
		return
	}

	model := BindModel[PlayEventRequest](c)
	if model == nil {
		return
	}

	result, err := h.service.PlayEvent(context.Background(), playerId, scheduleId, model.Data)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, result)
}

func (h *EventHandler) GetPlayerEvent(c *gin.Context) {
	playerId := c.GetString("playerID")
	scheduleId := c.Param("id")
	events, err := h.service.GetOrCreatePlayerEvent(context.Background(), playerId, scheduleId)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, events)
}

type PlayEventRequest struct {
	Data map[string]interface{} `json:"play_data"`
}
