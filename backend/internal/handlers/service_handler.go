// ============================================================
// FILE: internal/handlers/service_handler.go
// WHAT IT IS:     HTTP handlers for the services/offerings API
// ENDPOINTS:      GET    /api/v1/services          (public)
//                 GET    /api/v1/services/admin     (protected — includes inactive)
//                 POST   /api/v1/services           (protected)
//                 PUT    /api/v1/services/:id        (protected)
//                 DELETE /api/v1/services/:id        (protected)
// LAST UPDATED:   2026-05-19 — initial creation
// ============================================================

package handlers

import (
	"imtiaz-portfolio/internal/models"
	"imtiaz-portfolio/internal/services"
	"imtiaz-portfolio/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	svc *services.ServiceService
}

func NewServiceHandler(svc *services.ServiceService) *ServiceHandler {
	return &ServiceHandler{svc: svc}
}

func (h *ServiceHandler) GetAll(c *gin.Context) {
	items, err := h.svc.GetAll()
	if err != nil {
		utils.InternalError(c, "Failed to fetch services")
		return
	}
	utils.Success(c, "Services fetched successfully", items)
}

func (h *ServiceHandler) GetAllAdmin(c *gin.Context) {
	items, err := h.svc.GetAllAdmin()
	if err != nil {
		utils.InternalError(c, "Failed to fetch services")
		return
	}
	utils.Success(c, "Services fetched successfully", items)
}

func (h *ServiceHandler) Create(c *gin.Context) {
	var service models.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		utils.BadRequest(c, "Invalid service data", err.Error())
		return
	}
	if utils.IsEmpty(service.Title) {
		utils.BadRequest(c, "Service title is required", nil)
		return
	}
	if err := h.svc.Create(&service); err != nil {
		utils.InternalError(c, "Failed to create service")
		return
	}
	utils.Created(c, "Service created successfully", service)
}

func (h *ServiceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid service ID", nil)
		return
	}
	var updates models.Service
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.BadRequest(c, "Invalid service data", err.Error())
		return
	}
	item, err := h.svc.Update(uint(id), &updates)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Service updated successfully", item)
}

func (h *ServiceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid service ID", nil)
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Service deleted successfully", nil)
}
