package postgres

import (
	"context"
	"fmt"
	"sync"

	"github.com/1412335/moneyforward-go-coding-challenge/pkg/configs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//Used to execute client creation procedure only once.
var once sync.Once

// errors

type DataAccessLayer struct {
	dbConfig *configs.Database
	//Used during creation of singleton client object in GetMongoClient().
	dbInstance *gorm.DB
}

func NewDataAccessLayer(ctx context.Context, cfg *configs.Database) (*DataAccessLayer, error) {
	dal := &DataAccessLayer{
		dbConfig: cfg,
	}
	if _, err := dal.Connect(ctx); err != nil {
		return nil, err
	}
	return dal, nil
}

// Build connection string
func (dal *DataAccessLayer) buildConnectionDSN() string {
	cfg := dal.dbConfig
	return fmt.Sprintf("host=%s port=%v user=%s dbname=%s sslmode=disable password=%s", cfg.Host, cfg.Port, cfg.User, cfg.Scheme, cfg.Password)
}

// Connect
func (dal *DataAccessLayer) Connect(ctx context.Context) (*gorm.DB, error) {
	//Perform connection creation operation only once.
	var err error
	once.Do(func() {
		// build connection string
		dsn := dal.buildConnectionDSN()
		// connect db
		db, e := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if e != nil {
			err = e
			return
		}

		// debug
		if dal.dbConfig.Debug {
			db = db.Debug()
		}

		sqlDB, e := db.DB()
		if e != nil {
			err = e
			return
		}
		// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		sqlDB.SetMaxIdleConns(dal.dbConfig.MaxIdleConns)

		// SetMaxOpenConns sets the maximum number of open connections to the database.
		sqlDB.SetMaxOpenConns(dal.dbConfig.MaxOpenConns)

		// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
		sqlDB.SetConnMaxLifetime(dal.dbConfig.ConnectTimeout)

		dal.dbInstance = db
	})
	return dal.dbInstance, err
}

func (dal *DataAccessLayer) Disconnect() error {
	sqlDB, err := dal.dbInstance.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (dal *DataAccessLayer) GetDatabase() *gorm.DB {
	return dal.dbInstance
}

func (dal *DataAccessLayer) Transaction(ctx context.Context, trans func(tx *gorm.DB) error) error {
	return dal.dbInstance.WithContext(ctx).Transaction(trans)
}
