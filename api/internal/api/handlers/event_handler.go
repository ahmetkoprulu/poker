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

// @Summary Create a new event
// @Description Event olusturmak icin kullanilir.
// @Tags events
// @Accept json
// @Produce json
// @Param event body models.Event true "Event object to create"
// @Success 200 {object} models.Event
// @Failure 400 {object} ErrorResponse
// @Router /events [post]
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

// @Summary Update an existing event
// @Description Event bilgilerini guncellemek icin kullanilir.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param event body models.Event true "Event object to update"
// @Success 200 {object} models.Event
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /events/{id} [put]
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

// @Summary Get an event by ID
// @Description Event bilgilerini almak icin kullanilir.
// @Tags events
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} models.Event
// @Failure 404 {object} ErrorResponse
// @Router /events/{id} [get]
func (h *EventHandler) GetEvent(c *gin.Context) {
	id := c.Param("id")
	event, err := h.service.GetEvent(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "Event not found")
		return
	}

	Ok(c, event)
}

// @Summary List all events
// @Description Tum eventleri listelemek icin kullanilir.
// @Tags events
// @Produce json
// @Success 200 {array} models.Event
// @Failure 400 {object} ErrorResponse
// @Router /events [get]
func (h *EventHandler) ListEvents(c *gin.Context) {
	events, err := h.service.ListEvents(c.Request.Context())
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, events)
}

// @Summary Create a new event schedule
// @Description Event icin yeni bir Schedule olusturmak icin kullanilir.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param schedule body models.EventSchedule true "Schedule object to create"
// @Success 200 {object} models.EventSchedule
// @Failure 400 {object} ErrorResponse
// @Router /events/{id}/schedules [post]
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

// @Summary Update an event schedule
// @Description Schedule bilgilerini guncellemek icin kullanilir.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID"
// @Param schedule body models.EventSchedule true "Schedule object to update"
// @Success 200 {object} models.EventSchedule
// @Failure 400 {object} ErrorResponse
// @Router /events/schedules/{id} [put]
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

// @Summary Get a schedule by ID
// @Description Schedule bilgilerini almak icin kullanilir.
// @Tags events
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} models.EventSchedule
// @Failure 404 {object} ErrorResponse
// @Router /events/schedules/{id} [get]
func (h *EventHandler) GetSchedule(c *gin.Context) {
	id := c.Param("id")
	schedule, err := h.service.GetSchedule(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "Schedule not found")
		return
	}

	Ok(c, schedule)
}

// @Summary List active schedules
// @Description Tum aktif event Schedule'lari listelemek icin kullanilir.
// @Tags events
// @Produce json
// @Success 200 {array} models.ActiveEventSchedule
// @Failure 400 {object} ErrorResponse
// @Router /events/schedules/actives [get]
func (h *EventHandler) ListActiveSchedules(c *gin.Context) {
	schedules, err := h.service.ListActiveSchedules(c.Request.Context())
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, schedules)
}

// @Summary Get schedules by event
// @Description Event icin tum Schedule'lari listelemek icin kullanilir.
// @Tags events
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {array} models.EventSchedule
// @Failure 400 {object} ErrorResponse
// @Router /events/{id}/schedules [get]
func (h *EventHandler) GetSchedulesByEvent(c *gin.Context) {
	eventID := c.Param("id")
	schedules, err := h.service.GetSchedulesByEvent(c.Request.Context(), eventID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Ok(c, schedules)
}

// @Summary Play an event
// @Description Event Schedule icin oyun oynamak icin kullanilir.
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID"
// @Param play_data body models.PlayEventRequest true "Play data"
// @Security Bearer
// @Success 200 {object} models.EventPlayResult
// @Failure 400 {object} ErrorResponse
// @Router /events/schedules/{id}/play [post]
func (h *EventHandler) PlayEvent(c *gin.Context) {
	scheduleId := c.Param("id")
	playerId := c.GetString("playerID")
	if playerId == "" {
		BadRequest(c, "Player ID is required")
		return
	}

	model := BindModel[models.PlayEventRequest](c)
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

// @Summary Get player event
// @Description Player'in Event oyun State'ini cekmek icin kullanilir. Eger Player ilk kez geliyorsa State'i olusturulur, varsa mevcut State bilgisi dondurulur. Event oynanmadan once bir kez ugranmasi gerekir.
// @Tags events
// @Produce json
// @Param id path string true "Schedule ID"
// @Security Bearer
// @Success 200 {object} models.PlayerEvent
// @Failure 400 {object} ErrorResponse
// @Router /events/schedules/{id}/player [get]
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
