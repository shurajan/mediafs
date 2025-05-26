package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mediafs/internal/service"
	"path/filepath"
)

func NsfwScanHandler(baseDir string, nsfw *service.NsfwService, tasks *service.TaskManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		videoname := filepath.Base(c.Params("videoname"))
		targetDir := filepath.Join(baseDir, videoname)

		taskID := uuid.New().String()
		task := &service.Task{
			ID:     taskID,
			Status: service.StatusPending,
		}
		tasks.AddTask(task)

		go func() {
			tasks.UpdateTask(taskID, service.StatusRunning, "", nil)
			err := nsfw.RunScan(targetDir)
			if err != nil {
				tasks.UpdateTask(taskID, service.StatusFailed, "", err)
				return
			}
			tasks.UpdateTask(taskID, service.StatusCompleted, "scan complete", nil)
		}()

		return c.JSON(fiber.Map{
			"message": "scan started",
			"task_id": taskID,
		})
	}
}

func TaskStatusHandler(tasks *service.TaskManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		taskID := c.Params("taskID")
		task, err := tasks.GetTask(taskID)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "task not found")
		}
		return c.JSON(task)
	}
}
