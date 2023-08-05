package db

import (
	"access_governance_system/configs"
	"context"
	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"
	"go.uber.org/zap"
)

type dbLogger struct {
	logger *zap.SugaredLogger
}

func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	query, err := q.FormattedQuery()
	if err != nil {
		return c, nil
	}

	d.logger.Debug(string(query))
	return c, nil
}

func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	return nil
}

func StartDB(config configs.DB, logger *zap.SugaredLogger) (*pg.DB, error) {
	options, err := pg.ParseURL(config.URL)
	if err != nil {
		logger.Errorw("failed to parse db url", "error", err)
		return nil, err
	}

	db := pg.Connect(options)
	db.AddQueryHook(dbLogger{logger})

	collection := migrations.NewCollection()

	err = collection.DiscoverSQLMigrations("migrations")
	if err != nil {
		logger.Errorw("failed to discover migrations", "error", err)
		return nil, err
	}
	logger.Info("migrations discovered")

	_, _, err = collection.Run(db, "init")
	if err != nil {
		logger.Errorw("failed to init migrations", "error", err)
		return nil, err
	}
	logger.Info("migrations initialized")

	oldVersion, newVersion, err := collection.Run(db, "up")
	if err != nil {
		logger.Errorw("failed to run migrations", "error", err)
		return nil, err
	}

	if newVersion != oldVersion {
		logger.Infof("migrated from version %d to %d\n", oldVersion, newVersion)
	} else {
		logger.Infof("version is %d\n", oldVersion)
	}

	return db, nil
}
