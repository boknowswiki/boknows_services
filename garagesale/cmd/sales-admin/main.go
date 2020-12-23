package main

import (
	"fmt"
	"log"
	"os"

	"github.com/boknowswiki/boknows_services/garagesale/internal/platform/conf"
	"github.com/boknowswiki/boknows_services/garagesale/internal/platform/database"
	"github.com/boknowswiki/boknows_services/garagesale/internal/schema"
	"github.com/pkg/errors"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("run failed %v", err)
	}
}

func run() error {
	// =========================================================================
	// Configuration

	var cfg struct {
		DB struct {
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:postgres,noprint"`
			Host       string `conf:"default:localhost"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:false"`
		}
		Args conf.Args
	}

	if err := conf.Parse(os.Args[1:], "SALES", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("SALES", &cfg)
			if err != nil {
				return errors.Wrap(err, "main : generating usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	cfg.DB.DisableTLS = true

	// Initialize dependencies.
	db, err := database.Open(database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	})
	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}
	defer db.Close()

	switch cfg.Args.Num(0) {
	case "migrate":
		if err := schema.Migrate(db); err != nil {
			log.Println("error applying migrations", err)
			os.Exit(1)
		}
		fmt.Println("Migrations complete")
		return nil

	case "seed":
		if err := schema.Seed(db); err != nil {
			log.Println("error seeding database", err)
			os.Exit(1)
		}
		fmt.Println("Seed data complete")
		return nil
	}

	return nil
}
