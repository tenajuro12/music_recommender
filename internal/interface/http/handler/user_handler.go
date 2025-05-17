package handler

import (
	"spotify_recommender/internal/app/usecase"
	"spotify_recommender/internal/interface/http/middleware"
)

type AuthHandler struct {
	userUseCase   *usecase.UserManagementUseCase
	jwtMiddleware *middleware.JWTMiddleware
}
