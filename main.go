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
}

// TableName specifies the name of the table in the database
func (t *TodoTask) TableName() string {
	return "todo_tasks"
}

// User represents a user
type User struct {
	ID             int       `orm:"column(id);pk" json:"id"`
	MaxTasksPerDay int       `orm:"column(max_tasks_per_day)" json:"max_tasks_per_day"`
	CreatedAt      time.Time `orm:"column(created_at);auto_now_add" json:"created_at"`
}

// TableName specifies the name of the table in the database
func (u *User) TableName() string {
	return "users"
}

// TasksController handles the API endpoints for managing todo tasks
type TasksController struct {
	beego.Controller
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
	beego.Router("/users/:user_id/tasks", &TasksController{}, "post:AddTask")

	//Start the server
	beego.Run()
}
