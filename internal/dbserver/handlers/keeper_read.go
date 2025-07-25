package handlers

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/metrics"
)

func (h *Handler) GetPerformers(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetPerformers] trace_id=%s - Retrieving performers", traceID)
	trackDBOp := metrics.TrackDBOperation("read", "keepers")
	performers, err := h.keeperRepository.GetKeeperAsPerformer()
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetPerformers] Error retrieving performers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sort.Slice(performers, func(i, j int) bool {
		return performers[i].KeeperID < performers[j].KeeperID
	})

	h.logger.Infof("[GetPerformers] Successfully retrieved %d performers", len(performers))
	c.JSON(http.StatusOK, performers)
}

func (h *Handler) GetKeeperData(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetKeeperData] trace_id=%s - Retrieving keeper data", traceID)
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperData] Retrieving keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperData] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid keeper ID format",
			"code":  "INVALID_KEEPER_ID",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_data")
	keeperData, err := h.keeperRepository.GetKeeperDataByID(keeperIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetKeeperData] Error retrieving keeper data: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Keeper not found",
			"code":  "KEEPER_NOT_FOUND",
		})
		return
	}

	h.logger.Infof("[GetKeeperData] Successfully retrieved keeper with ID: %s", keeperID)
	c.JSON(http.StatusOK, keeperData)
}

func (h *Handler) GetKeeperTaskCount(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetKeeperTaskCount] trace_id=%s - Retrieving keeper task count", traceID)
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperTaskCount] Retrieving task count for keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperTaskCount] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid keeper ID format",
			"code":  "INVALID_KEEPER_ID",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_tasks")
	taskCount, err := h.keeperRepository.GetKeeperTaskCount(keeperIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetKeeperTaskCount] Error retrieving task count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperTaskCount] Successfully retrieved task count %d for keeper ID: %s", taskCount, keeperID)
	c.JSON(http.StatusOK, gin.H{"no_executed_tasks": taskCount})
}

func (h *Handler) GetKeeperPoints(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetKeeperPoints] trace_id=%s - Retrieving points for keeper", traceID)
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperPoints] Retrieving points for keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperPoints] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_points")
	points, err := h.keeperRepository.GetKeeperPointsByIDInDB(keeperIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetKeeperPoints] Error retrieving keeper points: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperPoints] Successfully retrieved points %f for keeper ID: %s", points, keeperID)
	c.JSON(http.StatusOK, gin.H{"keeper_points": points})
}

func (h *Handler) GetKeeperCommunicationInfo(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetKeeperCommunicationInfo] trace_id=%s - Retrieving communication info for keeper", traceID)
	keeperID := c.Param("id")
	h.logger.Infof("[GetKeeperChatInfo] Retrieving chat ID, keeper name, and email for keeper with ID: %s", keeperID)

	keeperIDInt, err := strconv.ParseInt(keeperID, 10, 64)
	if err != nil {
		h.logger.Errorf("[GetKeeperChatInfo] Error parsing keeper ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_communication")
	keeperData, err := h.keeperRepository.GetKeeperCommunicationInfo(keeperIDInt)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[GetKeeperChatInfo] Error retrieving chat ID, keeper name, and email for ID %s: %v", keeperID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("[GetKeeperChatInfo] Successfully retrieved chat ID, keeper name, and email for ID: %s", keeperID)
	c.JSON(http.StatusOK, keeperData)
}
