package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

// TodoTask represents a todo task
type TodoTask struct {
	ID        int       `orm:"column(id);pk" json:"id"`
	UserID    int       `orm:"column(user_id)" json:"user_id"`
	Task      string    `orm:"column(task)" json:"task"`
	CreatedAt time.Time `orm:"column(created_at);auto_now_add" json:"created_at"`
	UpdatedAt time.Time `orm:"column(updated_at);auto_now_add" json:"updated_at"`
}

// TableName specifies the name of the table in the database
func (t *TodoTask) TableName() string {
	return "todo_tasks"
}

// User represents a user
type User struct {
	ID             int       `orm:"column(id);pk" json:"id"`
	Name           string    `orm:"column(name)" json:"name"`
	MaxTasksPerDay int       `orm:"column(max_tasks_per_day)" json:"max_tasks_per_day"`
	CreatedAt      time.Time `orm:"column(created_at);auto_now_add" json:"created_at"`
	UpdatedAt      time.Time `orm:"column(updated_at);auto_now_add" json:"updated_at"`
}

// TableName specifies the name of the table in the database
func (u *User) TableName() string {
	return "users"
}

// TasksController handles the API endpoints for managing todo tasks
type TasksController struct {
	beego.Controller
}

// GetTasks retrieves a list of todo tasks for a specific user
func (c *TasksController) GetTasks() {
	// Extract the user ID from the URL parameters
	var userID int
	// Get the user ID string from the request parameters
	userIDString := c.Ctx.Input.Param(":user_id")
	if userIDString == "" {
		c.CustomAbort(http.StatusBadRequest, "User id cannot be empty")
	}
	// Convert the user ID string to an integer
	userID, err := strconv.Atoi(userIDString)
	if err != nil {
		// If the conversion fails, return a custom error response to the client
		c.CustomAbort(http.StatusBadRequest, "Invalid user id")
	}

	// Create a new ORM object
	o := orm.NewOrm()

	// Get the list of tasks from the database
	var tasks []TodoTask
	_, err = o.QueryTable("todo_tasks").Filter("user_id", userID).All(&tasks)
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error retrieving tasks")
	}

	// Return the list of tasks in the response body
	c.Data["json"] = &tasks
	c.ServeJSON()
}

// GetTaskByID retrieves a specific todo task by its ID and checks if the current user has that task
func (c *TasksController) GetTaskByID() {
	// Extract the task ID and user ID from the URL parameters
	var taskID, userID int
	// Get the task ID and user ID strings from the request parameters
	userIDString := c.Ctx.Input.Param(":user_id")
	taskIDString := c.Ctx.Input.Param(":task_id")

	if userIDString == "" {
		c.CustomAbort(http.StatusBadRequest, "User id cannot be empty")
	}
	if taskIDString == "" {
		c.CustomAbort(http.StatusBadRequest, "Task id cannot be empty")
	}

	// Convert the task ID and user ID strings to integers
	userID, err := strconv.Atoi(userIDString)
	taskID, err = strconv.Atoi(taskIDString)

	if err != nil {
		// If the conversion fails, return a custom error response to the client
		c.CustomAbort(http.StatusBadRequest, "Invalid user id")
	}
	if err != nil {
		// If the conversion fails, return a custom error response to the client
		c.CustomAbort(http.StatusBadRequest, "Invalid task id")
	}

	// Get the task from the database
	o := orm.NewOrm()
	task := TodoTask{ID: taskID}
	err = o.Read(&task)
	if err == orm.ErrNoRows {
		c.CustomAbort(http.StatusBadRequest, "Task not found")
	} else if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error reading task")
	}

	// Check if the current user has the task
	if task.UserID != userID {
		c.CustomAbort(http.StatusForbidden, "You do not have permission to access this task")
	}

	// Return the task in the response body
	c.Data["json"] = &task
	c.ServeJSON()
}

// AddTask adds a new todo task
func (c *TasksController) AddTask() {
	// Extract the user ID from the URL parameters
	var userID int
	// Get the user ID string from the request parameters
	userIDString := c.Ctx.Input.Param(":user_id")
	if userIDString == "" {
		c.CustomAbort(http.StatusBadRequest, "User id cannot be empty")
	}
	// Convert the user ID string to an integer
	userID, err := strconv.Atoi(userIDString)
	if err != nil {
		// If the conversion fails, return a custom error response to the client
		c.CustomAbort(http.StatusBadRequest, "Invalid user id")
	}

	// Get the task from the request body
	task := c.GetString("task")
	if task == "" {
		c.CustomAbort(http.StatusBadRequest, "Task cannot be empty")
	}

	// Create a new ORM object
	o := orm.NewOrm()

	// Check if the user exists
	user := User{ID: userID}
	err = o.Read(&user)
	if err == orm.ErrNoRows {
		c.CustomAbort(http.StatusBadRequest, "User ID not found")
	} else if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error reading user")
	}

	// Check if the user has reached the task limit per day
	today := time.Now().Format("2006-01-02")
	var count int
	err = o.Raw("SELECT COUNT(*) FROM todo_tasks WHERE user_id = ? AND DATE(created_at) = ?", userID, today).QueryRow(&count)
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error counting tasks")
	}
	if count >= user.MaxTasksPerDay {
		c.CustomAbort(http.StatusBadRequest, "Daily task limit reached")
	}

	// Add the task to the database
	todoTask := TodoTask{UserID: userID, Task: task}
	_, err = o.Insert(&todoTask)
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error adding task")
	}

	c.Data["json"] = map[string]string{"message": "Task added successfully"}
	c.ServeJSON()
}

// UpdateTask updates a specific todo task by its ID and checks if the current user has that task
func (c *TasksController) UpdateTask() {
	// Extract the task ID and user ID from the URL parameters
	var taskID, userID int
	// Get the task ID and user ID strings from the request parameters
	userIDString := c.Ctx.Input.Param(":user_id")
	taskIDString := c.Ctx.Input.Param(":task_id")
	if userIDString == "" {
		c.CustomAbort(http.StatusBadRequest, "User id cannot be empty")
	}
	if taskIDString == "" {
		c.CustomAbort(http.StatusBadRequest, "Task id cannot be empty")
	}

	// Convert the task ID and user ID strings to integers
	userID, err := strconv.Atoi(userIDString)
	taskID, err = strconv.Atoi(taskIDString)
	if err != nil {
		// If the conversion fails, return a custom error response to the client
		c.CustomAbort(http.StatusBadRequest, "Invalid user id")
	}
	if err != nil {
		// If the conversion fails, return a custom error response to the client
		c.CustomAbort(http.StatusBadRequest, "Invalid task id")
	}

	// Get the task from the database
	o := orm.NewOrm()
	task := TodoTask{ID: taskID}
	err = o.Read(&task)
	if err == orm.ErrNoRows {
		c.CustomAbort(http.StatusBadRequest, "Task not found")
	} else if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error reading task")
	}

	// Check if the current user has the task
	if task.UserID != userID {
		c.CustomAbort(http.StatusForbidden, "You do not have permission to access this task")
	}

	// Get the updated task from the request body
	updatedTask := c.GetString("task")
	if updatedTask == "" {
		c.CustomAbort(http.StatusBadRequest, "Task cannot be empty")
	}

	// Update the task in the database
	task.Task = updatedTask
	task.UpdatedAt = time.Now()
	_, err = o.Update(&task)
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error updating task")
	}

	// Return the updated task in the response body
	c.Data["json"] = map[string]string{"message": "Task updated successfully"}
	c.ServeJSON()
}

// DeleteTask deletes a specific todo task by its ID and checks if the current user has that task
func (c *TasksController) DeleteTask() {
	// Extract the task ID and user ID from the URL parameters
	var taskID, userID int
	// Get the task ID and user ID strings from the request parameters
	taskIDString := c.Ctx.Input.Param(":task_id")
	userIDString := c.Ctx.Input.Param(":user_id")
	if taskIDString == "" {
		c.CustomAbort(http.StatusBadRequest, "Task id cannot be empty")
	}
	if userIDString == "" {
		c.CustomAbort(http.StatusBadRequest, "User id cannot be empty")
	}
	// Convert the task ID and user ID strings to integers
	userID, err := strconv.Atoi(userIDString)
	taskID, err = strconv.Atoi(taskIDString)
	if err != nil {
		// If the conversion fails, return a custom error response to the client
		c.CustomAbort(http.StatusBadRequest, "Invalid task id")
	}
	if err != nil {
		// If the conversion fails, return a custom error response to the client
		c.CustomAbort(http.StatusBadRequest, "Invalid user id")
	}

	// Get the task from the database
	o := orm.NewOrm()
	task := TodoTask{ID: taskID}
	err = o.Read(&task)
	if err == orm.ErrNoRows {
		c.CustomAbort(http.StatusBadRequest, "Task not found")
	} else if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error reading task")
	}

	// Check if the current user has the task
	if task.UserID != userID {
		c.CustomAbort(http.StatusForbidden, "You do not have permission to delete this task")
	}

	// Delete the task from the database
	_, err = o.Delete(&task)
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error deleting task")
	}

	// Return a success response to the client
	c.Data["json"] = map[string]string{"message": "Task deleted successfully"}
	c.ServeJSON()
}

// DeleteAllTasks deletes all tasks for a specific user
func (c *TasksController) DeleteAllTasks() {
	// Extract the user ID from the URL parameters
	var userID int
	// Get the user ID string from the request parameters
	userIDString := c.Ctx.Input.Param(":user_id")
	if userIDString == "" {
		c.CustomAbort(http.StatusBadRequest, "User id cannot be empty")
	}
	// Convert the user ID string to an integer
	userID, err := strconv.Atoi(userIDString)
	if err != nil {
		// If the conversion fails, return a custom error response to the client
		c.CustomAbort(http.StatusBadRequest, "Invalid user id")
	}
	// Create a new ORM object
	o := orm.NewOrm()

	// Delete all tasks for the user from the database
	_, err = o.QueryTable("todo_tasks").Filter("user_id", userID).Delete()
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error deleting tasks")
	}

	// Return a success message in the response body
	c.Data["json"] = map[string]string{"message": "All tasks deleted"}
	c.ServeJSON()
}

func main() {
	//Register the MySQL driver with ORM.
	orm.RegisterDriver("mysql", orm.DRMySQL)
	//Register the default database with ORM.
	orm.RegisterDataBase("default", "mysql", "root@tcp(127.0.0.1:3306)/todo_app?charset=utf8")

	// Enable debugging for ORM. This will print SQL queries and other debug information to the console.
	// Use this for debugging.
	orm.Debug = true

	//Register the tables to use
	orm.RegisterModel(&User{})
	orm.RegisterModel(&TodoTask{})

	//Set up the router
	beego.Router("/:user_id", &TasksController{}, "get:GetTasks")
	beego.Router("/:user_id/:task_id", &TasksController{}, "get:GetTaskByID")
	beego.Router("/:user_id", &TasksController{}, "post:AddTask")
	beego.Router("/:user_id/:task_id", &TasksController{}, "put:UpdateTask")
	beego.Router("/:user_id/:task_id", &TasksController{}, "delete:DeleteTask")
	beego.Router("/:user_id", &TasksController{}, "delete:DeleteAllTasks")

	//Start the server
	beego.Run()
}
