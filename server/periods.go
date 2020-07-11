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
	"fmt"
	"github.com/FabianWe/pollsweb/pollsdata"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

// TODO return not founds
// TODO dates / times: what is the meaning of Start for example? is this some UTC time? or always that day in local?

func ShowPeriodSettingsListHandleFunc(ctx context.Context, requestContext *RequestContext, w http.ResponseWriter, r *http.Request) error {
	periods, periodsGetErr := requestContext.DataHandler.GetLatestPeriods(ctx, -1, time.Time{})
	if periodsGetErr != nil {
		return periodsGetErr
	}
	data := requestContext.PrepareTemplateRenderData()
	data["periods_list"] = periods
	return executeBuffered(requestContext.Templates.TemplateMap["periods-list"], data, w)
}

func PeriodDetailsHandleFunc(ctx context.Context, requestContext *RequestContext, w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	slug := vars["slug"]
	queryArgs := pollsdata.NewPeriodSettingsQueryArgs().
		SetSlug(&slug)
	period, getErr := requestContext.DataHandler.GetPeriod(ctx, queryArgs)
	if getErr != nil {
		return getErr
	}
	data := requestContext.PrepareTemplateRenderData()
	data["period"] = period
	return executeBuffered(requestContext.Templates.TemplateMap["periods-detail"], data, w)
}

func getEditPeriodDetailsHandleFunc(ctx context.Context, requestContext *RequestContext, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func postEditPeriodDetailsHandleFunc(ctx context.Context, requestContext *RequestContext, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func EditPeriodDetailsHandleFunc(ctx context.Context, requestContext *RequestContext, w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return getEditPeriodDetailsHandleFunc(ctx, requestContext, w, r)
	}
	return postEditPeriodDetailsHandleFunc(ctx, requestContext, w, r)
}

func getNewPeriodHandleFunc(ctx context.Context, requestContext *RequestContext, w http.ResponseWriter, r *http.Request) error {
	data := requestContext.PrepareTemplateRenderData()
	data["period"] = pollsdata.EmptyPeriodSettingsModel()
	return executeBuffered(requestContext.Templates.TemplateMap["periods-new"], data, w)
}

func postNewPeriodHandleFunc(ctx context.Context, requestContext *RequestContext, w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	// TODO deal with multierror etc here?
	form, formErr := DecodePeriodForm(r.Form)
	if formErr != nil {
		return formErr
	}
	fmt.Println(form)
	return nil
}

func NewPeriodHandleFunc(ctx context.Context, requestContext *RequestContext, w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return getNewPeriodHandleFunc(ctx, requestContext, w, r)
	}
	return postNewPeriodHandleFunc(ctx, requestContext, w, r)
}
