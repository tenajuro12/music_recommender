# Spotify Recommender

A personalized music recommendation service that suggests tracks based on mood, weather, and time of day.

## Overview

This application provides tailored music recommendations by analyzing audio features from Spotify and matching them to user context (mood, weather conditions, time of day). Users can save their favorite recommendations as playlists and customize their experience with personal preferences.

## Key Features

- **Contextual Music Recommendations**: Get track suggestions based on your current mood, local weather, and time of day
- **Personalized User Preferences**: Set favorite/disliked genres, tempo ranges, and preferred moods
- **Playlist Management**: Create, save, and share playlists
- **Spotify Integration**: Seamless connection with Spotify API for rich music data

## Project Architecture

The application follows a Clean Architecture approach with:

- **Domain Layer**: Core business entities and interfaces
- **Application Layer**: Use cases implementing business logic
- **Infrastructure Layer**: External services (Spotify API, weather services, database)
- **Interface Layer**: API handlers and middleware

## Tech Stack

- **Backend**: Go (Golang)
- **Database**: PostgreSQL
- **External APIs**: Spotify Web API
- **Authentication**: JWT-based auth
