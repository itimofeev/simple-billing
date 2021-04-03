package main

import (
	"github.com/itimofeev/simple-billing/internal/app/repository"
	"github.com/itimofeev/simple-billing/internal/app/service"
)

func main() {
	repo := repository.New("postgresql://postgres:password@localhost:5432/postgres?sslmode=disable")
	srv := service.New(repo)
	_ = srv
}
