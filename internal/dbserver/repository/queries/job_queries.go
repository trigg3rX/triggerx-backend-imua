package queries

// Create Queries
const (
	GetMaxJobIDQuery = `SELECT MAX(job_id) FROM triggerx.job_data`

	CreateJobDataQuery = `
			INSERT INTO triggerx.job_data (
				job_id, job_title, task_definition_id, user_id, link_job_id, chain_status,
				custom, time_frame, recurring, status, job_cost_prediction,
				created_at, updated_at, timezone, is_imua
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	// 15 values to be inserted, so 15 ?s
)

// Write Queries
const (
	UpdateJobDataFromUserQuery = `
			UPDATE triggerx.job_data 
			SET job_title = ?, time_frame = ?, recurring = ?, status = ?,
			job_cost_prediction = ?, updated_at = ?
			WHERE job_id = ?`

	UpdateJobDataLastExecutedAtQuery = `
			UPDATE triggerx.job_data 
			SET task_ids = ?, job_cost_actual = ?, last_executed_at = ?
			WHERE job_id = ?`

	UpdateJobDataStatusQuery = `
			UPDATE triggerx.job_data 
			SET status = ?, updated_at = ?
			WHERE job_id = ?`

	UpdateTimeJobIntervalQuery = `
		UPDATE triggerx.time_job_data
		SET time_interval = ?
		WHERE job_id = ?`
)

// Read Queries
const (
	GetJobDataByJobIDQuery = `
			SELECT job_id, job_title, task_definition_id, user_id, link_job_id, chain_status,
				custom, time_frame, recurring, status, job_cost_prediction, job_cost_actual,
				task_ids, created_at, updated_at, last_executed_at, timezone, is_imua
			FROM triggerx.job_data 
			WHERE job_id = ?`

	GetTaskDefinitionIDByJobIDQuery = `
			SELECT task_definition_id FROM triggerx.job_data 
			WHERE job_id = ?`

	GetTaskIDsByJobIDQuery = `
			SELECT task_ids FROM triggerx.job_data 
			WHERE job_id = ?`

	// New query to get task_id and fee for all tasks of a job
	GetTaskFeesByJobIDQuery = `
			SELECT task_id, task_opx_cost FROM triggerx.task_data
			WHERE job_id = ? ALLOW FILTERING`
)
