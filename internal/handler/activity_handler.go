package handler

import (
	"context"
	"net/http"

	"github.com/fiverr/go_platform/core/logging"
	"github.com/gin-gonic/gin"
)

// ActivityService is the part of the activity use case the handler depends on.
type ActivityService interface {
	GetActions(ctx context.Context, username string) ([]string, error)
}

// ActivityHandler handles GET /activity.
type ActivityHandler struct {
	svc    ActivityService
	logger logging.Logger
}

// NewActivityHandler builds an ActivityHandler backed by the given
// ActivityService.
func NewActivityHandler(svc ActivityService, logger logging.Logger) *ActivityHandler {
	return &ActivityHandler{svc: svc, logger: logger}
}

func (ctrl *ActivityHandler) Get(c *gin.Context) {
	var req usernameRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondError(c, http.StatusBadRequest, bindErrorText(err))
		return
	}

	actions, err := ctrl.svc.GetActions(c.Request.Context(), req.Username)
	if err != nil {
		respondFromError(c, ctrl.logger, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"actions": actions})
}
