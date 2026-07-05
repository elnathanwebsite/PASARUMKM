package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var DB *sql.DB

func Connect() error {
	_ = godotenv.Load("../.env") // Try to load from root, ignore if not found

	dbUrl := os.Getenv("TURSO_DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	if dbUrl == "" || authToken == "" {
		return fmt.Errorf("Missing TURSO_DATABASE_URL or TURSO_AUTH_TOKEN in .env")
	}

	// Sanitize URL for WebSockets if needed (though libsql driver usually handles it)
	if strings.HasPrefix(dbUrl, "https://") {
		dbUrl = "libsql://" + strings.TrimPrefix(dbUrl, "https://")
	}

	connStr := fmt.Sprintf("%s?authToken=%s", dbUrl, authToken)

	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	DB = db
	return nil
}
