// ============================================================
// FILE: internal/handlers/architecture_handler.go
// ENDPOINTS:      GET    /api/v1/architecture          (public)
//                 GET    /api/v1/architecture/admin     (protected)
//                 POST   /api/v1/architecture           (protected)
//                 PUT    /api/v1/architecture/:id        (protected)
//                 DELETE /api/v1/architecture/:id        (protected)
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

type ArchitectureHandler struct {
	svc *services.ArchitectureService
}

func NewArchitectureHandler(svc *services.ArchitectureService) *ArchitectureHandler {
	return &ArchitectureHandler{svc: svc}
}

func (h *ArchitectureHandler) GetAll(c *gin.Context) {
	items, err := h.svc.GetAll()
	if err != nil {
		utils.InternalError(c, "Failed to fetch architecture items")
		return
	}
	utils.Success(c, "Architecture fetched successfully", items)
}

func (h *ArchitectureHandler) GetAllAdmin(c *gin.Context) {
	items, err := h.svc.GetAllAdmin()
	if err != nil {
		utils.InternalError(c, "Failed to fetch architecture items")
		return
	}
	utils.Success(c, "Architecture fetched successfully", items)
}

func (h *ArchitectureHandler) Create(c *gin.Context) {
	var arch models.Architecture
	if err := c.ShouldBindJSON(&arch); err != nil {
		utils.BadRequest(c, "Invalid data", err.Error())
		return
	}
	if utils.IsEmpty(arch.Name) {
		utils.BadRequest(c, "Name is required", nil)
		return
	}
	if err := h.svc.Create(&arch); err != nil {
		utils.InternalError(c, "Failed to create architecture item")
		return
	}
	utils.Created(c, "Architecture item created", arch)
}

func (h *ArchitectureHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid ID", nil)
		return
	}
	var updates models.Architecture
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.BadRequest(c, "Invalid data", err.Error())
		return
	}
	item, err := h.svc.Update(uint(id), &updates)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Architecture item updated", item)
}

func (h *ArchitectureHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid ID", nil)
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Architecture item deleted", nil)
}
