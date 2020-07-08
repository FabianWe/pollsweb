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
	"net/url"
	"sync"
	"time"
)

const slugRegexString = `[a-zA-Z0-9_-]+`

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

type LocalizationConfig struct {
	DefaultTimezoneName string `mapstructure:"time_zone"`
	DefaultDateFormat   string `mapstructure:"date_format"`
	DefaultTimeFormat   string `mapstructure:"time_format"`
	DefaultLanguage     string `mapstructure:"language"`
}

func NewLocalizationConfig() *LocalizationConfig {
	return &LocalizationConfig{
		DefaultTimezoneName: "Europe/Berlin",
		DefaultDateFormat:   "02.01.2006",
		DefaultTimeFormat:   "02.01.2006 15:04",
		DefaultLanguage:     "de-DE",
	}
}

type AppConfig struct {
	Mongodb      *MongoConfig
	Localization *LocalizationConfig
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Mongodb:      NewMongoConfig(),
		Localization: NewLocalizationConfig(),
	}
}

type AppContext struct {
	*AppConfig
	Logger         *zap.SugaredLogger
	DataHandler    pollsdata.DataHandler
	Templates      *TemplateProvider
	LogRemoteAddr  bool
	HandlerTimeout time.Duration
	// used to generate URLs
	// must be set by hand, the NewAppContext... methods don't do this
	Router *mux.Router
	// date / datetime formats: converted automatically from the options, maybe we can do better by allowing an
	// overwrite
	// they must be set by hand, the NewAppContext... methods don't do this. You can use SetTimeFormats.
	DefaultMomentJSDateFormat     string
	DefaultMomentJSDateTimeFormat string
	DefaultGijgoDateFormat        string
	DefaultGijgoDateTimeFormat    string
}

func NewAppContext(config *AppConfig, logger *zap.SugaredLogger, dataHandler pollsdata.DataHandler, templateRoot string) *AppContext {
	return &AppContext{
		AppConfig:                     config,
		Logger:                        logger,
		DataHandler:                   dataHandler,
		Templates:                     NewTemplateProvider(templateRoot),
		LogRemoteAddr:                 true,
		HandlerTimeout:                time.Second * 30,
		Router:                        nil,
		DefaultMomentJSDateFormat:     "",
		DefaultMomentJSDateTimeFormat: "",
		DefaultGijgoDateFormat:        "",
		DefaultGijgoDateTimeFormat:    "",
	}
}

func NewAppContextMongo(ctx context.Context, config *AppConfig, logger *zap.SugaredLogger, templateRoot string) (*AppContext, error) {
	uri := GetMongoURI(config.Mongodb.UserName,
		config.Mongodb.Password,
		config.Mongodb.Host,
		config.Mongodb.Port)
	res := NewAppContext(config, logger, nil, templateRoot)
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
	mongoHandler := pollsdata.NewMongoDataHandler(mongoClient, config.Mongodb.Database)
	res.DataHandler = mongoHandler
	return res, nil
}

func (appContext *AppContext) SetTimeFormats() {
	goDateFormat, goDateTimeFormat := appContext.Localization.DefaultDateFormat, appContext.Localization.DefaultTimeFormat
	momentDateFormat, momentDateTimeFormat := pollsweb.MomentJSDateFormatter.ConvertFormat(goDateFormat),
		pollsweb.MomentJSDateFormatter.ConvertFormat(goDateTimeFormat)
	gijgoDateFormat, gijgoDateTimeFormat := pollsweb.GijgoDateFormatter.ConvertFormat(goDateFormat),
		pollsweb.GijgoDateFormatter.ConvertFormat(goDateTimeFormat)
	appContext.DefaultMomentJSDateFormat = momentDateFormat
	appContext.DefaultMomentJSDateTimeFormat = momentDateTimeFormat
	appContext.DefaultGijgoDateFormat = gijgoDateFormat
	appContext.DefaultGijgoDateTimeFormat = gijgoDateTimeFormat
	appContext.Logger.Debugw("automatically transformed time formats for support libraries",
		"go-date-format", goDateTimeFormat,
		"moment-js-date-format", momentDateFormat,
		"gijgo-date-format", gijgoDateFormat,
		"go-date-time-format", goDateTimeFormat,
		"moment-js-date-time-format", momentDateTimeFormat,
		"gijgo-date-time-format", gijgoDateTimeFormat)
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

type RequestContext struct {
	*AppContext
}

func NewRequestContext(appContext *AppContext) *RequestContext {
	return &RequestContext{
		AppContext: appContext,
	}
}

func (requestContext *RequestContext) PrepareTemplateRenderData() map[string]interface{} {
	res := make(map[string]interface{}, 10)
	res["request_context"] = requestContext
	return res
}

func (requestContext *RequestContext) GetDateTimeFormat() string {
	return requestContext.Localization.DefaultTimeFormat
}

func (requestContext *RequestContext) GetDateFormat() string {
	return requestContext.Localization.DefaultDateFormat
}

func (requestContext *RequestContext) FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(requestContext.GetDateTimeFormat())
}

func (requestContext *RequestContext) FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(requestContext.GetDateFormat())
}

func (requestContext *RequestContext) GetTimezoneName() string {
	return requestContext.Localization.DefaultTimezoneName
}

func (requestContext *RequestContext) GetMomentJSDateFormat() string {
	return pollsweb.MomentJSDateFormatter.ConvertFormat(requestContext.GetDateFormat())
}

func (requestContext *RequestContext) GetMomentJSDateTimeFormat() string {
	return pollsweb.MomentJSDateFormatter.ConvertFormat(requestContext.GetDateTimeFormat())
}

func (requestContext *RequestContext) GetGijgoDateFormat() string {
	return pollsweb.GijgoDateFormatter.ConvertFormat(requestContext.GetDateFormat())
}

func (requestContext *RequestContext) GetGijgoDateTimeFormat() string {
	return pollsweb.GijgoDateFormatter.ConvertFormat(requestContext.GetDateTimeFormat())
}

func (requestContext *RequestContext) FormatMeetingTime(meetingTime *pollsdata.MeetingTimeTemplateModel) string {
	// TODO use a user-specific format
	weekdayString := meetingTime.Weekday.String()
	return fmt.Sprintf("%s, %2d:%2d", weekdayString, meetingTime.Hour, meetingTime.Minute)
}

func (requestContext *RequestContext) URL(name string, pairs ...string) (*url.URL, error) {
	return requestContext.Router.Get(name).URL(pairs...)
}

func (requestContext *RequestContext) URLString(name string, pairs ...string) (string, error) {
	u, err := requestContext.URL(name, pairs...)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// TODO document: always close context
func initWithMongo(config *AppConfig, logger *zap.SugaredLogger, templateRoot string) (*AppContext, error) {
	ctx, startCtxCancel := context.WithTimeout(context.Background(), config.Mongodb.ConnectTimeout)
	defer startCtxCancel()
	return NewAppContextMongo(ctx, config, logger, templateRoot)
}

func RunServerMongo(config *AppConfig, templateRoot string, debug bool) {
	start := time.Now()
	logger, loggerErr := pollsweb.InitLogger(debug)
	if loggerErr != nil {
		log.Fatalln("unable to init logging system, exiting")
	}
	logger.Info("starting application")
	appContext, initErr := initWithMongo(config, logger, templateRoot)
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

	appContext.SetTimeFormats()
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
	// set router in context
	appContext.Router = r
	homeHandler := Handler{
		AppContext: appContext,
		HandleFunc: HomeHandleFunc,
	}
	listPeriodsHandler := Handler{
		AppContext: appContext,
		HandleFunc: ShowPeriodSettingsListHandleFunc,
	}
	periodDetailHandler := Handler{
		AppContext: appContext,
		HandleFunc: PeriodDetailsHandleFunc,
	}
	newPeriodHandler := Handler{
		AppContext: appContext,
		HandleFunc: NewPeriodHandleFunc,
	}
	editPeriodHandler := Handler{
		AppContext: appContext,
		HandleFunc: EditPeriodDetailsHandleFunc,
	}
	r.PathPrefix("/static/{file}").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static")))).
		Methods(http.MethodGet).
		Name("static")
	r.Handle("/", &homeHandler).
		Methods(http.MethodGet).
		Name("home")
	r.Handle("/periods/", &listPeriodsHandler).
		Methods(http.MethodGet).
		Name("periods-list")
	r.Handle("/periods/new/", &newPeriodHandler).
		Methods(http.MethodGet, http.MethodPost).
		Name("periods-new")
	r.Handle(fmt.Sprintf("/period/{slug:%s}/", slugRegexString), &periodDetailHandler).
		Methods(http.MethodGet).
		Name("periods-detail")
	r.Handle(fmt.Sprintf("/period/{slug:%s}/edit/", slugRegexString), &editPeriodHandler).
		Methods(http.MethodGet, http.MethodPost).
		Name("periods-edit")

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

type HandleFunc func(ctx context.Context, requestContext *RequestContext, w http.ResponseWriter, r *http.Request) error

func ExecSecure(f HandleFunc, ctx context.Context, requestContext *RequestContext, w http.ResponseWriter, r *http.Request) (err error) {
	defer func() {
		if r := recover(); r != nil {
			requestContext.Logger.Errorw("recovered panic from HandleFunc, returning it as error",
				"recover", r)
			// should always be nil in case of panic
			if err == nil {
				err = fmt.Errorf("recovered from a handler panic: %v", r)
			}
		}
	}()
	err = f(ctx, requestContext, w, r)
	return
}

type Handler struct {
	*AppContext
	HandleFunc HandleFunc
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestContext := NewRequestContext(h.AppContext)
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
	err := ExecSecure(h.HandleFunc, ctx, requestContext, w, r)
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
