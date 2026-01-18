package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

func Load() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system address")
	flag.Parse()

	if envAddr := os.Getenv("RUN_ADDRESS"); envAddr != "" {
		cfg.RunAddress = envAddr
	}
	if envDB := os.Getenv("DATABASE_URI"); envDB != "" {
		cfg.DatabaseURI = envDB
	}
	if envAccrual := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrual != "" {
		cfg.AccrualSystemAddress = envAccrual
	}

	return cfg
}
