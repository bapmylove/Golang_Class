package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite" // 匯入 SQLite 驅動
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var db *sql.DB

func initDatabase() {
	var err error
	db, err = sql.Open("sqlite", "todo.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 創建表（如果不存在）
	createTableSQL := `CREATE TABLE IF NOT EXISTS ToDo_List (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        done INTEGER
    );`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

func main() {
	initDatabase()

	router := gin.Default()
	router.GET("/todos", getTodos)
	router.POST("/todos", createTodo)
	router.PUT("/todos/:id", updateTodo)
	router.DELETE("/todos/:id", deleteTodo)

	router.Run(":8080")
}

func getTodos(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, done FROM ToDo_List")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Done); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		todos = append(todos, todo)
	}

	c.JSON(http.StatusOK, todos)
}

func createTodo(c *gin.Context) {
	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 將 done 轉換為 INTEGER
	done := 0
	if todo.Done {
		done = 1
	}

	result, err := db.Exec("INSERT INTO ToDo_List (title, done) VALUES (?, ?)", todo.Title, done)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todo.ID = int(id)
	c.JSON(http.StatusCreated, todo)
}

func updateTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 將 done 轉換為 INTEGER
	done := 0
	if todo.Done {
		done = 1
	}

	_, err = db.Exec("UPDATE ToDo_List SET title = ?, done = ? WHERE id = ?", todo.Title, done, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todo.ID = id
	c.JSON(http.StatusOK, todo)
}

func deleteTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	_, err = db.Exec("DELETE FROM ToDo_List WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted"})
}
