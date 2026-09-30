package handler

import (
	"net/http"

	"github.com/fiverr/book_service_/internal/service"
	"github.com/gin-gonic/gin"
)

// ActivityHandler handles GET /activity.
type ActivityHandler struct {
	svc service.ActivityService
}

// NewActivityHandler builds an ActivityHandler backed by the given
// ActivityService.
func NewActivityHandler(svc service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

func (ctrl *ActivityHandler) Get(c *gin.Context) {
	username := c.Query("username")
	if !requireFields(c, requiredField{"username", username}) {
		return
	}

	actions, err := ctrl.svc.GetActions(c.Request.Context(), username)
	if err != nil {
		respondFromError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"actions": actions})
}
