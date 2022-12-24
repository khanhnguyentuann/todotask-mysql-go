package main

import (
	"time"

	"github.com/astaxie/beego"
)

// Declare a structure to save user information
type User struct {
	ID            int
	Name          string
	MaxTaskPerDay int
}

// Declare a structure to save to-do information
type Task struct {
	ID          int
	UserID      int
	Description string
	CreateAt    time.Time
}

func main() {
	beego.Run()
}
