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
	"bytes"
	"context"
	"fmt"
	"github.com/FabianWe/pollsweb"
	"github.com/FabianWe/pollsweb/pollsdata"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"html/template"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type MongoConfig struct {
	UserName       string `mapstructure:"username"`
	Password       string
	Host           string
	Port           int
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
	Database       string
}

func NewMongoConfig() *MongoConfig {
	return &MongoConfig{
		UserName:       "",
		Password:       "",
		Host:           "localhost",
		Port:           27017,
		ConnectTimeout: time.Second * 10,
		Database:       "gopolls",
	}
}

type AppConfig struct {
	Mongodb *MongoConfig
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Mongodb: NewMongoConfig(),
	}
}

type AppContext struct {
	Logger         *zap.SugaredLogger
	DataHandler    pollsdata.DataHandler
	Templates      *TemplateProvider
	LogRemoteAddr  bool
	HandlerTimeout time.Duration
}

func NewAppContext(logger *zap.SugaredLogger, dataHandler pollsdata.DataHandler, templateRoot string) *AppContext {
	return &AppContext{
		Logger:         logger,
		DataHandler:    dataHandler,
		Templates:      NewTemplateProvider(templateRoot),
		LogRemoteAddr:  true,
		HandlerTimeout: time.Second * 30,
	}
}

func NewAppContextMongo(ctx context.Context, logger *zap.SugaredLogger, uri, databaseName, templateRoot string) (*AppContext, error) {
	res := NewAppContext(logger, nil, templateRoot)
	logger.Info("connecting to mongodb")
	mongoClient, connectErr := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if connectErr != nil {
		return res, connectErr
	}
	pingErr := mongoClient.Ping(ctx, nil)
	if pingErr != nil {
		return res, pingErr
	}
	logger.Info("connection to mongodb established")
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
func initWithMongo(uri, databaseName string, startTimeout time.Duration, logger *zap.SugaredLogger, templateRoot string) (*AppContext, error) {
	ctx, startCtxCancel := context.WithTimeout(context.Background(), startTimeout)
	defer startCtxCancel()
	return NewAppContextMongo(ctx, logger, uri, databaseName, templateRoot)
}

// TODO likely to change, find a nicer way for options
func RunServerMongo(config *AppConfig, templateRoot string, debug bool) {
	uri := GetMongoURI(config.Mongodb.UserName,
		config.Mongodb.Password,
		config.Mongodb.Host,
		config.Mongodb.Port)
	start := time.Now()
	logger, loggerErr := pollsweb.InitLogger(debug)
	if loggerErr != nil {
		log.Fatalln("unable to init logging system, exiting")
	}
	logger.Info("starting application")
	appContext, initErr := initWithMongo(uri, config.Mongodb.Database, config.Mongodb.ConnectTimeout, logger, templateRoot)
	defer func() {
		runtime := time.Since(start)
		logger.Infow("stopping application",
			"app-runtime", runtime)
		closeCtx, closeDeferFunc := context.WithTimeout(context.Background(), config.Mongodb.ConnectTimeout)
		defer closeDeferFunc()
		if closeErr := appContext.Close(closeCtx); closeErr != nil {
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

	logger.Infow("loading templates",
		"template-root", templateRoot)
	if templateInitErr := appContext.Templates.InitBase(); templateInitErr != nil {
		logger.Errorw("can't load template base file, exiting",
			"error", templateInitErr)
		return
	}
	if numTemplates, loadErr := appContext.Templates.RegisterDefaults(); loadErr != nil {
		logger.Errorw("can't load templates, exiting",
			"error", loadErr)
		return
	} else {
		logger.Infof("loaded %d templates", numTemplates)
	}

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	homeHandler := Handler{
		AppContext: appContext,
		HandleFunc: HomeHandleFunc,
	}
	r.Handle("/", &homeHandler)
	// TODO test if shutdown later works correctly (closing mongodb)
	http.Handle("/", r)
	if httpServeErr := http.ListenAndServe("localhost:8080", nil); httpServeErr != nil {
		logger.Infow("server shut down: listen error",
			"error", httpServeErr)
	}

}

type HandlerError interface {
	error
	HttpCode() int
}

type Error struct {
	Err  error
	Code int
}

func NewError(err error, code int) Error {
	return Error{
		Err:  err,
		Code: code,
	}
}

func (e Error) Error() string {
	return e.Err.Error()
}

func (e Error) Unwrap() error {
	return e.Err
}

func (e Error) HttpCode() int {
	return e.Code
}

type HandleFunc func(ctx context.Context, appContext *AppContext, w http.ResponseWriter, r *http.Request) error

func ExecSecure(f HandleFunc, ctx context.Context, appContext *AppContext, w http.ResponseWriter, r *http.Request) (err error) {
	defer func() {
		if r := recover(); r != nil {
			appContext.Logger.Errorw("recovered from HandleFunc, returning it as error",
				"recover", r)
			// should always be nil in case of panic
			if err == nil {
				err = fmt.Errorf("recovered from a handler panic: %v", r)
			}
		}
	}()
	err = f(ctx, appContext, w, r)
	return
}

type Handler struct {
	*AppContext
	HandleFunc HandleFunc
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.Logger.Debugw("request done",
			"duration", time.Since(start))
	}()
	if h.LogRemoteAddr {
		h.Logger.Infow("handling request",
			"remote-addr", r.RemoteAddr,
			"request-url", r.URL.String())
	} else {
		h.Logger.Infow("handling request",
			"request-url", r.URL.String())
	}
	ctx, cancel := context.WithTimeout(context.Background(), h.HandlerTimeout)
	defer cancel()
	err := ExecSecure(h.HandleFunc, ctx, h.AppContext, w, r)
	if err == nil {
		return
	}
	h.Logger.Errorw("error handling request",
		"error", err)
	switch e := err.(type) {
	case HandlerError:
		http.Error(w, e.Error(), e.HttpCode())
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

var byteBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func getByteBuffer() *bytes.Buffer {
	return byteBufferPool.Get().(*bytes.Buffer)
}

func releaseBytesBuffer(b *bytes.Buffer) {
	b.Reset()
	byteBufferPool.Put(b)
}

func executeBuffered(t *template.Template, data interface{}, w http.ResponseWriter) error {
	buff := getByteBuffer()
	defer releaseBytesBuffer(buff)
	// execute the template to the buffer, on error return that error
	templateErr := t.Execute(buff, data)
	if templateErr != nil {
		return templateErr
	}
	// still capture errors from w, but at least we got all template errors first
	_, copyErr := io.Copy(w, buff)
	return copyErr
}

func executeTemplateBuffered(t *template.Template, name string, data interface{}, w http.ResponseWriter) error {
	buff := getByteBuffer()
	defer releaseBytesBuffer(buff)
	templateErr := t.ExecuteTemplate(w, name, data)
	if templateErr != nil {
		return templateErr
	}
	_, copyErr := io.Copy(w, buff)
	return copyErr
}

func HomeHandleFunc(ctx context.Context, appContext *AppContext, w http.ResponseWriter, r *http.Request) error {
	return executeBuffered(appContext.Templates.TemplateMap["home"], nil, w)
}
