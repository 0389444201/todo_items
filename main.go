package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type TodoItems struct {
	Id    int    `json:"id" gorm:"column:id;"`
	Title string `json:"title" gorm:"column:title;"`
	//Image       string    `json:"image"`
	Description string     `json:"description" gorm:"column:description;"`
	Status      string     `json:"status" gorm:"column:status;"`
	CreatedAt   *time.Time `json:"create_at" gorm:"column:created_at;"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at;"`
}
type TodoItemCreation struct {
	Id          int    `json:"-" gorm:"column:id;"`
	Title       string `json:"title" gorm:"column:title;"`
	Description string `json:"description" gorm:"column:description;"`
	//Status      string `json:"status" gorm:"column:status"`
}
type TodoItemUpdate struct {
	Title       *string `json:"title" gorm:"column:title;"`
	Description *string `json:"description" gorm:"column:description;"`
	Status      *string `json:"status" gorm:"column:status;"`
}
type Paging struct {
	Page  int   `json:"page" form:"page"`
	Limit int   `json:"limit" form:"limit"`
	Total int64 `json:"total" form:"-"'`
}

func (TodoItems) TableName() string        { return "todo_items" }
func (TodoItemUpdate) TableName() string   { return TodoItems{}.TableName() }
func (TodoItemCreation) TableName() string { return TodoItems{}.TableName() }
func main() {
	dsn := os.Getenv("DB_CONN_STR")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	r := gin.Default()
	v1 := r.Group("/v1")
	items := v1.Group("/items")
	{
		items.POST("", CreateItem(db))
		items.GET("", ListItem(db))
		items.GET("/:id", GetItem(db))
		items.PATCH("/:id", UpdateItem(db))
		items.DELETE("/:id", DeleteItem(db))
	}
	r.Run(":3000")
}
func CreateItem(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var data TodoItemCreation
		if err := c.ShouldBind(&data); err != nil {
			c.JSON(400, gin.H{"error": "fail to created"})
			return
		}
		if err := db.Create(&data).Error; err != nil {
			c.JSON(500, gin.H{"message": err.Error()})
			return
		}
		c.JSON(200, data.Id)
	}
}
func GetItem(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var data TodoItems
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//Cach 1: data.Id = id
		//err = db.First(&data,id).Error
		err = db.Where("id=?", id).First(&data).Error
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, data)
	}
}
func UpdateItem(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var data TodoItemUpdate
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err = c.ShouldBind(&data); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err = db.Where("id=?", id).Updates(&data).Error; err != nil {
			c.JSON(500, gin.H{"error": "cannot updated item"})
			return
		}
		c.JSON(200, true)
	}
}
func ListItem(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var page Paging //Phân trang
		if err := c.ShouldBind(&page); err != nil {
			c.JSON(500, err)
			return
		}

		var data []TodoItems
		result := db.Order("id desc").Find(&data) // Order "id desc" để sắp xếp mảng từ lớn đến nhỏ
		if err := result.Error; err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, data)
	}
}
func DeleteItem(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		err = db.Table(TodoItems{}.TableName()).Where("id=?", id).Updates(map[string]interface{}{
			"status": "Deleted",
		}).Error
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, true)
	}
}
