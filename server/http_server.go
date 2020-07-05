// Copyright 2020 Fabian Wenzelmann <fabianwen@posteo.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"github.com/FabianWe/pollsweb"
	"github.com/FabianWe/pollsweb/pollsdata"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"log"
	"time"
)

type AppContext struct {
	Logger      *zap.SugaredLogger
	DataHandler pollsdata.DataHandler
}

func NewAppContext(logger *zap.SugaredLogger, dataHandler pollsdata.DataHandler) *AppContext {
	return &AppContext{
		Logger:      logger,
		DataHandler: dataHandler,
	}
}

func NewAppContextMongo(ctx context.Context, logger *zap.SugaredLogger, uri, databaseName string) (*AppContext, error) {
	res := NewAppContext(logger, nil)
	logger.Info("connecting to mongodb")
	mongoClient, connectErr := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if connectErr != nil {
		return res, connectErr
	}
	pingErr := mongoClient.Ping(ctx, nil)
	if pingErr != nil {
		return res, pingErr
	}
	mongoHandler := pollsdata.NewMongoDataHandler(mongoClient, databaseName)
	res.DataHandler = mongoHandler
	return res, nil
}

// TODO defer call to close, defer call to logger.sync
func (appContext *AppContext) Close(ctx context.Context) error {
	appContext.Logger.Info("closing app context")
	if appContext.DataHandler == nil {
		appContext.Logger.Info("no database connection was established, nothing to close")
		return nil
	}
	appContext.Logger.Info("closing database connection")
	closeErr := appContext.DataHandler.Close(ctx)
	if closeErr != nil {
		appContext.Logger.Errorw("error closing data handler",
			"error", closeErr)
		return closeErr
	}
	return nil
}

// TODO document: always close context
func initWithMongo(uri, databaseName string, startTimeout time.Duration, logger *zap.SugaredLogger) (*AppContext, error) {
	ctx, startCtxCancel := context.WithTimeout(context.Background(), startTimeout)
	defer startCtxCancel()
	return NewAppContextMongo(ctx, logger, uri, databaseName)
}

// TODO likely to change, find a nicer way for options
func RunServerMongo(uri, databaseName string, startTimeout time.Duration, debug bool) {
	start := time.Now()
	logger, loggerErr := pollsweb.InitLogger(debug)
	if loggerErr != nil {
		log.Fatalln("unable to init logging system, exiting")
	}
	logger.Infow("starting application",
		"start", start)
	appContext, initErr := initWithMongo(uri, databaseName, startTimeout, logger)
	defer func() {
		runtime := time.Since(start)
		logger.Infow("stopping application",
			"app-runtime", runtime)
		// TODO add a timeout here?
		if closeErr := appContext.Close(context.Background()); closeErr != nil {
			logger.Errorw("shutting down application caused an error",
				"error", closeErr)
		}
		_ = logger.Sync()
	}()
	if initErr != nil {
		logger.Errorw("error while setting up mongodb connection, exiting",
			"error", initErr)
		return
	}
}
