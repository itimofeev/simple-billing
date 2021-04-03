package main

import "github.com/itimofeev/simple-billing/internal/app/repository"

func main() {
	store := repository.New("postgresql://postgres:password@localhost:5432/postgres?sslmode=disable")
	_ = store
}
