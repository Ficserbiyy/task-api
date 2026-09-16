package handlers

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/Ficserbiyy/task-api/internal/auth"
	"github.com/Ficserbiyy/task-api/internal/config"
	"github.com/Ficserbiyy/task-api/internal/models"
	"gorm.io/gorm"
)

func validatePaginationInput(p, l string) (int, int) {
	// Clean and validate pagination inputs
	page, _ := strconv.Atoi(p)
	limit, _ := strconv.Atoi(l)

	if page <= 0 {
		page = 1
	}

	switch {
	case limit > 100:
		limit = 100
	case limit <= 0:
		limit = 10
	}
	return page, limit
}

// paginate handles limit, offset, and
// optional sorting logic dynamically.
func paginate(page, limit int, sort string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Apply sorting if provided
		switch sort {
		case "id_desc":
			db = db.Order("id DESC")
		default:
			db = db.Order("id ASC")
		}

		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}

// Create method creates a new task
// in the database and returns TaskResponse.
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

// ListTasks method retrieves from the database all tasks
// owned by the current user and returns []TaskResponse.
//
// @Summary     Get Tasks
// @Description Get a paginated and sorted list of your tasks
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       page  query    int    false  "Page number (default: 1)"
// @Param       limit query    int    false  "Number of items (default: 10, max: 100)"
// @Param       sort  query    string false  "Sort field and order (e.g., 'id_desc')"
// @Success     200   {object} models.Pagination
// @Failure		401   {string} string "Unauthorized"
// @Failure 	500   {string} string "Inernal Server Error"
// @Router      /tasks [get]
func (s *TaskRepository) ListTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.GetCurrentUser(r.Context())
		if !ok {
			config.ErrUnauthorized.Raise(w)
			return
		}

		var totalRows int64
		var dbRows []models.Task
		query := r.URL.Query()

		// Read and validate raw query parameters
		page, limit := validatePaginationInput(
			query.Get("page"),
			query.Get("limit"),
		)
		sort := query.Get("sort")

		s.DB.Model(&models.Task{}).
			Where("owner_id = ?", userID).
			Count(&totalRows)

		// Fetch records using the dynamic scope
		err := s.DB.Where("owner_id = ?", userID).
			Scopes(paginate(page, limit, sort)).
			Find(&dbRows).Error

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		totalPages := int(math.Ceil(float64(totalRows) / float64(limit)))

		// Generate a response from slice of tasks
		var tasks []models.TaskResponse
		for i := range dbRows {
			tasks = append(tasks, dbRows[i].ResponseModel())
		}
		response := models.Pagination{
			Limit:      limit,
			Page:       page,
			TotalRows:  totalRows,
			TotalPages: totalPages,
			Rows:       tasks,
		}

		// Encode directly to writer
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("error encoding tasks response: %v", err)
		}
	}
}
