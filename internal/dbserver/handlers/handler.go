package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/repository"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/database"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/docker"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

type NotificationConfig struct {
	EmailFrom     string
	EmailPassword string
	BotToken      string
}

type Handler struct {
	db                     *database.Connection
	logger                 logging.Logger
	config                 NotificationConfig
	executor               docker.ExecutorConfig
	jobRepository          repository.JobRepository
	timeJobRepository      repository.TimeJobRepository
	eventJobRepository     repository.EventJobRepository
	conditionJobRepository repository.ConditionJobRepository
	taskRepository         repository.TaskRepository
	userRepository         repository.UserRepository
	keeperRepository       repository.KeeperRepository
	apiKeysRepository      repository.ApiKeysRepository

	scanNowQuery func(*time.Time) error // for testability
}

func NewHandler(db *database.Connection, logger logging.Logger, config NotificationConfig, executor docker.ExecutorConfig) *Handler {
	h := &Handler{
		db:                     db,
		logger:                 logger,
		config:                 config,
		executor:               executor,
		jobRepository:          repository.NewJobRepository(db),
		timeJobRepository:      repository.NewTimeJobRepository(db),
		eventJobRepository:     repository.NewEventJobRepository(db),
		conditionJobRepository: repository.NewConditionJobRepository(db),
		taskRepository:         repository.NewTaskRepository(db),
		userRepository:         repository.NewUserRepository(db),
		keeperRepository:       repository.NewKeeperRepository(db),
		apiKeysRepository:      repository.NewApiKeysRepository(db),
	}
	h.scanNowQuery = h.defaultScanNowQuery
	return h
}

func (h *Handler) defaultScanNowQuery(timestamp *time.Time) error {
	return h.db.Session().Query("SELECT now() FROM system.local").Scan(timestamp)
}

func (h *Handler) getTraceID(c *gin.Context) string {
	traceID, exists := c.Get("trace_id")
	if !exists {
		return ""
	}
	return traceID.(string)
}
