package main

import (
	"ea-job-executor/executor"
	"ea-job-executor/logger"
)

func main() {
	// Set up the logger
	logger.Slog.Info("Starting the application")

	filePath := "agentjob.json"
	executor.ExecuteAgentJob(filePath)

}
