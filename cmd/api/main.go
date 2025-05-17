// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log"
	https "net/http"
	"os"
	"os/signal"
	"spotify_recommender/internal/app/usecase"
	"spotify_recommender/internal/domain/service"
	"spotify_recommender/internal/infrastructure/cache"
	"spotify_recommender/internal/infrastructure/database/postgres"
	"spotify_recommender/internal/infrastructure/external/spotify"
	"spotify_recommender/internal/infrastructure/external/weather"
	"spotify_recommender/internal/interface/http/handler"
	"spotify_recommender/internal/interface/http/middleware"
	http "spotify_recommender/internal/interface/http/router"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	spotifyClient := setupSpotifyClient()
	weatherClient := setupWeatherClient()

	userRepo := postgres.NewUserRepository(db)
	trackRepo := postgres.NewTrackRepository(db)
	playlistRepo := postgres.NewPlaylistRepository(db)
	recommendationRepo := postgres.NewRecommendationRepository(db)

	recommendationService := service.NewRecommendationService(userRepo, trackRepo, recommendationRepo)
	playlistService := service.NewPlaylistService(playlistRepo, trackRepo, userRepo, recommendationRepo)

	userManagementUseCase := usecase.NewUserManagementUseCase(userRepo)
	getRecommendationsUseCase := usecase.NewGetRecommendationsUseCase(recommendationService, trackRepo, weatherClient)
	savePlaylistUseCase := usecase.NewSavePlaylistUseCase(playlistService)
	savePlaylistFromRecommendationUseCase := usecase.NewSavePlaylistFromRecommendationUseCase(playlistService)

	jwtMiddleware := setupJWTMiddleware()

	userHandler := handler.NewUserHandler(userManagementUseCase, jwtMiddleware)
	recommendationHandler := handler.NewRecommendationHandler(getRecommendationsUseCase)
	playlistHandler := handler.NewPlaylistHandler(
		savePlaylistUseCase,
		savePlaylistFromRecommendationUseCase,
		playlistService,
	)

	r := http.router.Setup(userHandler, recommendationHandler, playlistHandler, jwtMiddleware)

	server := &https.Server{
		Addr:         fmt.Sprintf(":%s", getEnv("PORT", "8080")),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
		serverStopCtx()
	}()

	log.Printf("Server is running on port %s", getEnv("PORT", "8080"))
	err = server.ListenAndServe()
	if err != nil && err != https.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v", server.Addr, err)
	}

	<-serverCtx.Done()
	log.Println("Server stopped")
}

func setupDatabase() (*sqlx.DB, error) {
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/spotify_recommender?sslmode=disable")

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func setupRedisCache() (*cache.RedisCache, error) {
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")
	return cache.NewRedisCache(redisURL)
}

func setupSpotifyClient() *spotify.Client {
	config := spotify.Config{
		ClientID:     getEnv("SPOTIFY_CLIENT_ID", ""),
		ClientSecret: getEnv("SPOTIFY_CLIENT_SECRET", ""),
		RedirectURI:  getEnv("SPOTIFY_REDIRECT_URI", "http://localhost:8080/api/auth/spotify/callback"),
	}

	return spotify.NewClient(config)
}

func setupWeatherClient() *openweathermap.Client {
	config := openweathermap.Config{
		APIKey: getEnv("OPENWEATHERMAP_API_KEY", ""),
		Units:  "metric",
	}

	return openweathermap.NewClient(config)
}

func setupJWTMiddleware() *middleware.JWTMiddleware {
	config := middleware.JWTConfig{
		SecretKey:     getEnv("JWT_SECRET_KEY", "supersecretkey"),
		TokenDuration: 24 * time.Hour, // 24 hours
	}

	return middleware.NewJWTMiddleware(config)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
