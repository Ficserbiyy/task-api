package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Ficserbiyy/task-api/internal/auth"
	"github.com/Ficserbiyy/task-api/internal/config"
	"github.com/Ficserbiyy/task-api/internal/models"
)

// Create method creates a new task.
//
// @Summary 	Create a task
// @Description Create a new task
// @Tags 		tasks
// @Accept 		json
// @Produce 	json
// @Param 		task body models.CreateTaskRequest true "Task"
// @Success 	201 {object} models.TaskResponse
// @Failure		401 {string} string "Unauthorized"
// @Failure 	400 {string} string "Invalid request body"
// @Failure		500 {string} string "Unable to create the task"
// @Router 		/tasks [post]
func (s *TaskRepository) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.GetCurrentUser(r.Context())
		if !ok {
			config.ErrUnauthorized.Raise(w)
			return
		}

		var req models.CreateTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			config.ErrInvalidRequest.Raise(w)
			return
		}

		if req.Title == "" {
			http.Error(w, "title field cannot be empty", http.StatusBadRequest)
			return
		}

		dbTask := models.Task{
			OwnerID:     userID,
			Title:       req.Title,
			Description: req.Description,
		}

		if err := s.DB.WithContext(r.Context()).Create(&dbTask).Error; err != nil {
			http.Error(w, "task creation failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(dbTask.ResponseModel())
	}
}
