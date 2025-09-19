package main

import (
	"log"
	"nordik-drive-api/config"
	"nordik-drive-api/internal/auth"
	"nordik-drive-api/internal/chat"
	"nordik-drive-api/internal/file"
	"nordik-drive-api/internal/logs"
	"nordik-drive-api/internal/role"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	dsn := "host=" + cfg.DBHost +
		" user=" + cfg.DBUser +
		" password=" + cfg.DBPassword +
		" dbname=" + cfg.DBName +
		" port=" + cfg.DBPort +
		" sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://34.145.18.109/", "https://nordik-drive-react-724838782318.us-west1.run.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	logService := &logs.LogService{DB: db}
	userService := &auth.AuthService{DB: db, CFG: &cfg}
	auth.RegisterRoutes(r, userService, logService)

	fileService := &file.FileService{DB: db}
	file.RegisterRoutes(r, fileService, logService)

	roleService := &role.RoleService{DB: db}
	role.RegisterRoutes(r, roleService)

	logs.RegisterRoutes(r, logService)

	chatService := &chat.ChatService{DB: db}
	chat.RegisterRoutes(r, chatService)

	// --- Cloud Run expects plain HTTP, on $PORT, bind to 0.0.0.0 ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on 0.0.0.0:%s ...", port)
	log.Fatal(r.Run("0.0.0.0:" + port))
}
