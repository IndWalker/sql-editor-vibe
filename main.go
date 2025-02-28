package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"example/user/playground/dbmanager"
	"example/user/playground/sqlvalidator"
)

type SQLValidationRequest struct {
	SQL     string `json:"sql" binding:"required"`
	Dialect string `json:"dialect" binding:"required"`
}

type QueryResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
}

func main() {
	fmt.Println("Starting SQL Playground server...")

	// Initialize database connections
	err := dbmanager.InitDatabases()
	if err != nil {
		fmt.Printf("Error initializing database connections: %v\n", err)
	}

	// Initialize gin router
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve static files
	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// Route for the main page
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Health check endpoint
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "ok",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Group API routes
	api := r.Group("/api")
	{
		api.POST("/validate-sql", validateAndExecuteSQL)
		api.GET("/db-status", getDatabaseStatus)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start the server in a goroutine
	go func() {
		fmt.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	fmt.Println("Server exited properly")
}

func validateAndExecuteSQL(c *gin.Context) {
	var req SQLValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// First run safety checks
	safetyCheck := sqlvalidator.IsSafeDDLOperation(req.SQL, req.Dialect)
	if !safetyCheck.Safe {
		c.JSON(http.StatusOK, gin.H{
			"valid": false,
			"error": safetyCheck.Error,
		})
		return
	}

	// Then validate the SQL
	valid, err := sqlvalidator.Validate(req.SQL, req.Dialect)
	if !valid {
		c.JSON(http.StatusOK, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	// If validation succeeds, execute the query
	db, err := dbmanager.GetDatabaseConnection(req.Dialect)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":  true,
			"error":  "Database connection error: " + err.Error(),
			"result": nil,
		})
		return
	}

	// Execute the SQL query and get results
	result, err := executeQuery(db, req.SQL, req.Dialect)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":  true,
			"error":  "Query execution error: " + err.Error(),
			"result": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":  true,
		"result": result,
	})
}

// executeQuery executes the SQL query and returns results
func executeQuery(db *sql.DB, query string, dialect string) (*QueryResult, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Prepare result container
	result := &QueryResult{
		Columns: columns,
		Rows:    [][]interface{}{},
	}

	// Prepare value holders
	count := 0
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	// Iterate through rows
	for rows.Next() {
		if count >= 10 { // Limit to 10 rows
			break
		}

		err = rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		// Convert values to strings or appropriate type for JSON
		row := make([]interface{}, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = nil
			} else {
				switch v := val.(type) {
				case []byte:
					row[i] = string(v)
				default:
					row[i] = v
				}
			}
		}

		result.Rows = append(result.Rows, row)
		count++
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// getDatabaseStatus returns the status of all database connections
func getDatabaseStatus(c *gin.Context) {
	statuses := dbmanager.GetConnectionStatuses()
	c.JSON(http.StatusOK, statuses)
}
