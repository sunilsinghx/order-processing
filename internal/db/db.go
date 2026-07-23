package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/sunilsinghx/order-processing/internal/config"
)

func New(cfg *config.Config) (*pgx.Conn, error) {

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Println("DB ERROR: ", err)
		return nil, err
	}

	log.Println("Connected to DB ✅")

	return conn, nil
}
