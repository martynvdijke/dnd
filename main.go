package main

import (
	"context"
	"embed"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"villum/db"
	"villum/ent/user"
	"villum/handlers"
)

//go:embed static/*.html static/*.css static/style.css static/js/*.js static/js/chunks/*.js static/sw.js static/manifest.json static/fonts/*.woff2 static/brand/*.svg
var staticFiles embed.FS

var Version = "0.0.0-dev"

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		if os.Getenv("DOCKER") == "true" {
			dbPath = "/db/villum.db"
		} else {
			dbPath = "villum.db"
		}
	}

	if err := db.Init(dbPath); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	mediaPath := initMedia(dbPath)

	db.Seed()
	handlers.SeedCompendiumSchemas()
	handlers.SeedCraftingRecipes()
	db.SeedDefaultEventsSettings()
	handlers.SetAppVersion(Version)
	handlers.SetBaseURL(os.Getenv("BASE_URL"))

	// AUTO_SETUP=true creates the admin user automatically (used for per-worker Playwright isolation).
	if os.Getenv("AUTO_SETUP") == "true" {
		ctx := context.Background()
		count, err := db.Client.User.Query().Where(user.Role("admin")).Count(ctx)
		if err == nil && count == 0 {
			hash, err := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)
			if err == nil {
				_, err = db.Client.User.Create().
					SetUsername("admin").
					SetPassword(string(hash)).
					SetDisplayName("Admin").
					SetRole("admin").
					Save(ctx)
				if err != nil {
					log.Printf("Warning: AUTO_SETUP failed to create admin: %v", err)
				} else {
					log.Println("AUTO_SETUP: admin user created")
				}
			}
		}
	}

	handlers.SetDBPath(dbPath)

	_ = registerSchedulers()

	port := os.Getenv("PORT")
	if port == "" {
		port = "6270"
	}

	r, shutdown := setupRouter(mediaPath)
	defer shutdown()

	log.Printf("villum v%s starting on :%s", Version, port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}
