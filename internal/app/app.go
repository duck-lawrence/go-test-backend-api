package app

import (
	"github.com/ducklawrence05/go-test-backend-api/config"
	"github.com/ducklawrence05/go-test-backend-api/internal/infrastructure/wire/managers"
	"github.com/ducklawrence05/go-test-backend-api/internal/initialization"
	"github.com/ducklawrence05/go-test-backend-api/pkg/logger"
)

func Run(cfg *config.Config) {
	// logger
	l := logger.New(cfg.Logger)
	l.Info("Config log successfully")

	// postgres
	pgDb := initialization.NewPostgres(&cfg.Postgres, l)
	l.Info("Init Postgres successfully")

	// redis
	rdb := initialization.NewRedis(&cfg.Redis, l)
	l.Info("Init Redis successfully")

	// ===== usecase =====
	managers, err := managers.InitializeManagers(cfg, pgDb, rdb, l)
	if err != nil {
		l.Fatal(err.Error())
	}

	// init role cache
	go initialization.NewRolesCache(managers.Role, l)

	// ===== router =====
	routerCfg := &initialization.RouterConfig{
		Config: cfg,
		Logger: l,
	}

	router := initialization.InitRouter(routerCfg, managers)

	server := initialization.NewServer(cfg.HTTP.Port, router)
	initialization.RunServer(server, l)
}
