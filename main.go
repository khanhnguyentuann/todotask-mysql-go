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

// UsersController handles the API endpoints for managing users
type UsersController struct {
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

// GetUsers retrieves a list of all users
func (c *UsersController) GetUsers() {
	// Create a new ORM object
	o := orm.NewOrm()
	// Get the list of users from the database
	var users []User
	_, err := o.QueryTable("users").All(&users)
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error retrieving users")
	}

	// Return the list of users in the response body
	c.Data["json"] = &users
	c.ServeJSON()
}

// AddUser adds a new user
func (c *UsersController) AddUser() {
	// Get the user name and maximum tasks per day from the request body
	name := c.GetString("name")
	if name == "" {
		c.CustomAbort(http.StatusBadRequest, "Name cannot be empty")
	}
	maxTasksPerDayString := c.GetString("maxtasks")
	if maxTasksPerDayString == "" {
		c.CustomAbort(http.StatusBadRequest, "Max tasks per day cannot be empty")
	}
	maxTasksPerDay, err := strconv.Atoi(maxTasksPerDayString)
	if err != nil {
		c.CustomAbort(http.StatusBadRequest, "Invalid max tasks per day")
	}

	// Create a new ORM object
	o := orm.NewOrm()

	// Add the user to the database
	user := User{Name: name, MaxTasksPerDay: maxTasksPerDay}
	_, err = o.Insert(&user)
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error inserting user")
	}

	// Return the newly created user in the response body
	c.Data["json"] = map[string]string{"message": "User added successfully"}
	c.ServeJSON()
}

// UpdateUser updates an existing user
func (c *UsersController) UpdateUser() {
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
	// Get the updated user name and maximum tasks per day from the request body
	name := c.GetString("name")
	if name == "" {
		c.CustomAbort(http.StatusBadRequest, "Name cannot be empty")
	}

	maxTasksPerDayString := c.GetString("maxtasks")
	if maxTasksPerDayString == "" {
		c.CustomAbort(http.StatusBadRequest, "Max tasks per day cannot be empty")
	}
	maxTasksPerDay, err := strconv.Atoi(maxTasksPerDayString)
	if err != nil {
		c.CustomAbort(http.StatusBadRequest, "Invalid max tasks per day")
	}

	// Create a new ORM object
	o := orm.NewOrm()

	// Get the user from the database
	user := User{ID: userID}
	err = o.Read(&user)
	if err == orm.ErrNoRows {
		c.CustomAbort(http.StatusBadRequest, "User not found")
	} else if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error reading user")
	}

	// Update the user in the database
	user.Name = name
	user.MaxTasksPerDay = maxTasksPerDay
	_, err = o.Update(&user)
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error updating user")
	}

	// Return the updated user in the response body
	c.Data["json"] = map[string]string{"message": "User updated successfully"}
	c.ServeJSON()
}

// DeleteUser deletes an existing user
func (c *UsersController) DeleteUser() {
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

	// Get the user from the database
	user := User{ID: userID}
	err = o.Read(&user)
	if err == orm.ErrNoRows {
		c.CustomAbort(http.StatusBadRequest, "User not found")
	} else if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error reading user")
	}

	// Delete the user from the database
	_, err = o.Delete(&user)
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error deleting user")
	}

	// Return a success message in the response body
	c.Data["json"] = "User deleted successfully"
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

	//Set up the router for task
	beego.Router("/:user_id", &TasksController{}, "get:GetTasks")
	beego.Router("/:user_id/:task_id", &TasksController{}, "get:GetTaskByID")
	beego.Router("/:user_id", &TasksController{}, "post:AddTask")
	beego.Router("/:user_id/:task_id", &TasksController{}, "put:UpdateTask")
	beego.Router("/:user_id/:task_id", &TasksController{}, "delete:DeleteTask")

	//Set up the router for user
	beego.Router("/", &UsersController{}, "get:GetUsers")
	beego.Router("/", &UsersController{}, "post:AddUser")
	beego.Router("/:user_id", &UsersController{}, "put:UpdateUser")
	beego.Router("/:user_id", &UsersController{}, "delete:DeleteUser")

	//Start the server
	beego.Run()
}
