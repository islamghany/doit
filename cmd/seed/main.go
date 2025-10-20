package main

import (
	"context"
	"flag"
	"log"
	"time"

	"doit/internal/config"
	"doit/internal/data/seeder"
	"doit/pkg/database"
)

func main() {
	env := flag.String("env", "dev", "Environment (dev, test, prod)")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := database.New(context.Background(), database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Database:        cfg.Database.Name,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		MaxConns:        cfg.Database.MaxOpenConns,
		MinConns:        5,
		MaxConnLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Second,
		MaxConnIdleTime: 30 * time.Minute,
		DisableTLS:      cfg.Database.DisableTLS,
		LogLevel:        cfg.App.LogLevel,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	seeder := seeder.New(pool.Pool)
	if err := seeder.Run(context.Background(), *env); err != nil {
		log.Fatal(err)
	}

	log.Println("Seeding completed successfully")
}
