package repository

import (
	"errors"
	"time"

	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/database"
)

type ConditionJobRepository interface {
	CreateConditionJob(conditionJob *types.ConditionJobData) error
	GetConditionJobByJobID(jobID int64) (types.ConditionJobData, error)
	CompleteConditionJob(jobID int64) error
	UpdateConditionJobStatus(jobID int64, isActive bool) error
}

type conditionJobRepository struct {
	db *database.Connection
}

func NewConditionJobRepository(db *database.Connection) ConditionJobRepository {
	return &conditionJobRepository{
		db: db,
	}
}

func (r *conditionJobRepository) CreateConditionJob(conditionJob *types.ConditionJobData) error {
	err := r.db.Session().Query(queries.CreateConditionJobDataQuery,
		conditionJob.JobID, conditionJob.TaskDefinitionID, conditionJob.ExpirationTime, conditionJob.Recurring,
		conditionJob.ConditionType, conditionJob.UpperLimit, conditionJob.LowerLimit,
		conditionJob.ValueSourceType, conditionJob.ValueSourceUrl, conditionJob.TargetChainID,
		conditionJob.TargetContractAddress, conditionJob.TargetFunction,
		conditionJob.ABI, conditionJob.ArgType, conditionJob.Arguments,
		conditionJob.DynamicArgumentsScriptUrl, conditionJob.IsCompleted, conditionJob.IsActive,
		time.Now(), time.Now()).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *conditionJobRepository) GetConditionJobByJobID(jobID int64) (types.ConditionJobData, error) {
	var conditionJob types.ConditionJobData
	err := r.db.Session().Query(queries.GetConditionJobDataByJobIDQuery, jobID).Scan(
		&conditionJob.JobID, &conditionJob.ExpirationTime, &conditionJob.Recurring, &conditionJob.ConditionType,
		&conditionJob.UpperLimit, &conditionJob.LowerLimit, &conditionJob.ValueSourceType,
		&conditionJob.ValueSourceUrl, &conditionJob.TargetChainID, &conditionJob.TargetContractAddress,
		&conditionJob.TargetFunction, &conditionJob.ABI, &conditionJob.ArgType, &conditionJob.Arguments,
		&conditionJob.DynamicArgumentsScriptUrl, &conditionJob.IsCompleted, &conditionJob.IsActive,
	)
	if err != nil {
		return types.ConditionJobData{}, errors.New("failed to get condition job by job ID")
	}

	return conditionJob, nil
}

func (r *conditionJobRepository) CompleteConditionJob(jobID int64) error {
	err := r.db.Session().Query(queries.CompleteConditionJobStatusQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to complete condition job")
	}

	err = r.db.Session().Query(queries.UpdateJobDataToCompletedQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to update job_data status to completed")
	}

	return nil
}

func (r *conditionJobRepository) UpdateConditionJobStatus(jobID int64, isActive bool) error {
	err := r.db.Session().Query(queries.UpdateConditionJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update condition job status")
	}

	return nil
}
