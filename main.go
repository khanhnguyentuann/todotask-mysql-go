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
	ID        int       `orm:"column(id);pk"`
	UserID    int       `orm:"column(user_id)"`
	Task      string    `orm:"column(task)"`
	CreatedAt time.Time `orm:"column(created_at);auto_now_add"`
}

// TableName specifies the name of the table in the database
func (t *TodoTask) TableName() string {
	return "todo_tasks"
}

// User represents a user
type User struct {
	ID             int       `orm:"column(id);pk"`
	MaxTasksPerDay int       `orm:"column(max_tasks_per_day)"`
	CreatedAt      time.Time `orm:"column(created_at);auto_now_add"`
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
	userID, _ := strconv.Atoi(c.Ctx.Input.Param(":user_id"))
	task := c.GetString("task")

	// Check if user has reached their daily task limit
	user := User{ID: userID}
	err := orm.NewOrm().Read(&user)
	if err != nil {
		c.CustomAbort(http.StatusBadRequest, "Invalid user ID")
	}

	today := time.Now().Format("2022-12-12")
	var count int
	orm.NewOrm().Raw("SELECT COUNT(*) FROM todo_tasks WHERE user_id = ? AND DATE(created_at) = ?", userID, today).QueryRow(&count)
	if count >= user.MaxTasksPerDay {
		c.CustomAbort(http.StatusBadRequest, "Daily task limit reached")
	}

	// Add the task to the database
	todoTask := TodoTask{UserID: userID, Task: task}
	_, err = orm.NewOrm().Insert(&todoTask)
	if err != nil {
		c.CustomAbort(http.StatusInternalServerError, "Error adding task")
	}

	c.Data["json"] = map[string]string{"message": "Task added successfully"}
	c.ServeJSON()
}

func main() {
	// Khởi tạo kết nối MySQL
	orm.RegisterDriver("mysql", orm.DRMySQL)                                                   //Đăng kí trình điều khiển mysql
	orm.RegisterDataBase("default", "mysql", "root@tcp(127.0.0.1:3306)/todo_app?charset=utf8") //Khởi động kết nối đến cơ sở dữ liệu

	//Khởi động kết nối với cơ sở dữ diệu
	orm.Debug = true

	//Đăng kí các bảng cần sử dụng
	orm.RegisterModel(new(TodoTask))
	orm.RegisterModel(new(User))

	//Tạo các bảng trong CSDL nếu chúng chưa tồn tại
	orm.RunSyncdb("default", false, true)

	// Khởi tạo Beego router
	beego.Router("/api/tasks", &TasksController{})
	beego.Run()
}

//http://localhost:8080/api/tasks
