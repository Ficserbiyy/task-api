package handlers

import (
	"context"
	"encoding/json"
	"errors"
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

// This function returns Task
// if found in the database,
// otherwise gorm.ErrRecordNotFound.
func getTaskByID(
	id, ownerID uint,
	db *gorm.DB,
	ctx context.Context,
) (models.Task, error) {
	var task models.Task

	err := db.WithContext(ctx).
		Where("id = ? AND owner_id = ?", id, ownerID).
		First(&task).Error

	return task, err
}

// This function removes Task
// from the database.
func deleteTask(task *models.Task, db *gorm.DB, ctx context.Context) error {
	return db.WithContext(ctx).
		Delete(task).Error
}

// This function updates an existing
// Task in the database.
func updateTask(
	task models.Task,
	data models.CreateTaskRequest,
	db *gorm.DB,
	ctx context.Context,
) error {
	if data.Title == "" && data.Description == "" {
		return nil
	}

	newTask := models.Task{
		Title:       data.Title,
		Description: data.Description,
	}

	return db.WithContext(ctx).
		Model(&task).
		Updates(newTask).Error
}

// @Summary 	Get a user
//
// @Description Receive the current user metadata
//
// @Tags 		protected
//
// @Produce 	json
//
// @Success 	200 {object} models.UserResponse
//
// @Failure		401 {string} string "Unauthorized"
//
// @Router 		/me [get]
func (s *TaskRepository) Me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := auth.GetCurrentUser(ctx)
		if !ok {
			config.ErrUnauthorized.Raise(w)
			return
		}

		var user models.User
		err := s.DB.WithContext(ctx).
			Where("id = ?", userID).
			First(&user).Error

		if err != nil {
			config.ErrInternal.Raise(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user.ResponseModel()); err != nil {
			log.Printf("error encoding user response: %v", err)
		}

	}
}

// Create method creates a new task
// in the database and returns TaskResponse.
//
// @Summary 	Create a task
//
// @Description Create a new task
//
// @Tags 		protected
//
// @Accept 		json
//
// @Produce 	json
//
// @Param 		task body models.CreateTaskRequest true "Task"
//
// @Success 	201 {object} models.TaskResponse
//
// @Failure		401 {string} string "Unauthorized"
//
// @Failure 	400 {string} string "Invalid request body"
//
// @Failure		500 {string} string "Unable to create the task"
//
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
//
// @Description Get a paginated and sorted list of your tasks
//
// @Tags        protected
//
// @Accept      json
//
// @Produce     json
//
// @Param       page   query    int    false  "Page number (default: 1)"
//
// @Param       limit  query    int    false  "Number of items (default: 10, max: 100)"
//
// @Param       sort   query    string false  "Sort field and order (e.g., 'id_desc')"
//
// @Success     200    {object} models.Pagination
//
// @Failure		401    {string} string "Unauthorized"
//
// @Failure 	500    {string} string "Inernal Server Error"
//
// @Router      /tasks [get]
func (s *TaskRepository) List() http.HandlerFunc {
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

// GetTask method retrieves single task from the database
// owned by the current user and returns TaskResponse.
//
// @Summary     Get a task
//
// @Description Receive an existing, single task record
//
// @Tags 		protected
//
// @Produce 	json
//
// @Param       id   path      int true "Task ID"
//
// @Success     200  {object}  models.TaskResponse
//
// @Failure		400  {string}  string "Invalid task ID"
//
// @Failure		401  {string}  string "Unauthorized"
//
// @Failure     404  {string}  string "Task not found"
//
// @Failure     500  {string}  string "Internal Server Error"
//
// @Router 		/tasks/{id}    [get]
func (s *TaskRepository) GetOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.GetCurrentUser(r.Context())
		if !ok {
			config.ErrUnauthorized.Raise(w)
			return
		}

		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)

		if err != nil || id <= 0 {
			config.ErrInvalidTaskID.Raise(w)
			return
		}

		task, err := getTaskByID(uint(id), userID, s.DB, r.Context())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				config.ErrTaskNotFound.Raise(w)
				return
			}
			config.ErrInternal.Raise(w)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(task.ResponseModel()); err != nil {
			log.Printf("error encoding tasks response: %v", err)
		}
	}
}

// Delete method removes a task from the database by its id.
//
// @Summary     Delete a task
//
// @Description Delete an existing task record
//
// @Tags 		protected
//
// @Produce 	json
//
// @Param       id   path 	   int true "Task ID"
//
// @Success     200  "Task successfully deleted"
//
// @Failure		400  {string}  string "Invalid task ID"
//
// @Failure		401  {string}  string "Unauthorized"
//
// @Failure     404  {string}  string "Task not found"
//
// @Failure     500  {string}  string "Internal Server Error"
//
// @Router 		/tasks/{id}    [delete]
func (s *TaskRepository) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := auth.GetCurrentUser(ctx)
		if !ok {
			config.ErrUnauthorized.Raise(w)
			return
		}

		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)

		if err != nil || id <= 0 {
			config.ErrInvalidTaskID.Raise(w)
			return
		}

		task, err := getTaskByID(uint(id), userID, s.DB, ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				config.ErrTaskNotFound.Raise(w)
				return
			}
			config.ErrInternal.Raise(w)
			log.Println(err)
			return
		}

		if err := deleteTask(&task, s.DB, ctx); err != nil {
			config.ErrInternal.Raise(w)
			log.Println(err)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Task successfully deleted",
		})
	}
}

// Update method updates a task in the database by its ID.
//
// @Summary 	 Update a task
//
// @Description  Update task details by ID
//
// @Tags 		 protected
//
// @Accept 		 json
//
// @Produce 	 json
//
// @Param        id    path      int  true  "Task ID"
//
// @Param        body  body models.CreateTaskRequest true "Task update payload"
//
// @Success 	 200   "Task successfully updated"
//
// @Failure		 400   {string}  string "Invalid request body"
//
// @Failure		 401   {string}  string "Unauthorized"
//
// @Failure      404   {string}  string "Task not found"
//
// @Failure      500   {string}  string "Internal Server Error"
//
// @Router 		 /tasks/{id} 	 [patch]
func (s *TaskRepository) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, ok := auth.GetCurrentUser(ctx)
		if !ok {
			config.ErrUnauthorized.Raise(w)
			return
		}

		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)

		if err != nil || id <= 0 {
			config.ErrInvalidTaskID.Raise(w)
			return
		}

		var req models.CreateTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			config.ErrInvalidRequest.Raise(w)
			return
		}

		task, err := getTaskByID(uint(id), userID, s.DB, ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				config.ErrTaskNotFound.Raise(w)
				return
			}
			config.ErrInternal.Raise(w)
			log.Println(err)
			return
		}

		if err := updateTask(task, req, s.DB, ctx); err != nil {
			config.ErrInternal.Raise(w)
			log.Println(err)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Task successfully updated",
		})
	}
}
