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
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"reflect"
)

func GetDefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"dict": func(pairs ...interface{}) (map[string]interface{}, error) {
			if len(pairs)%2 != 0 {
				return nil, errors.New("\"dict\" must be given an equal number of elements")
			}
			res := make(map[string]interface{}, len(pairs)/2)
			for i := 0; i < len(pairs); i += 2 {
				key, value := pairs[i], pairs[i+1]
				if keyString, okay := key.(string); okay {
					res[keyString] = value
				} else {
					return nil, fmt.Errorf("invalid key in \"dict\", keys must be strings, got type %v", reflect.TypeOf(key))
				}
			}
			return res, nil
		},
		"safe_js": func(s string) template.JS {
			return template.JS(s)
		},
		"safe_js_string": func(s string) template.JSStr {
			return template.JSStr(s)
		},
	}
}

type TemplateProvider struct {
	RootPath     string
	BaseTemplate *template.Template
	FuncMap      template.FuncMap
	TemplateMap  map[string]*template.Template
}

func NewTemplateProvider(root string) *TemplateProvider {
	return &TemplateProvider{
		RootPath:     root,
		BaseTemplate: nil,
		FuncMap:      GetDefaultFuncMap(),
		TemplateMap:  make(map[string]*template.Template),
	}
}

func (provider *TemplateProvider) InitBase() error {
	paths := []string{"base.gohtml",
		filepath.Join("voters", "voters_table.gohtml"),
		filepath.Join("periods", "period_form.gohtml"),
	}
	for i, file := range paths {
		paths[i] = filepath.Join(provider.RootPath, file)
	}
	base := template.New("base.gohtml").Funcs(provider.FuncMap)
	var err error
	base, err = base.ParseFiles(paths...)
	if err != nil {
		return err
	}
	provider.BaseTemplate = base
	provider.TemplateMap["base"] = base
	return nil
}

func (provider *TemplateProvider) RegisterTemplate(name string, paths ...string) (*template.Template, error) {
	clone, cloneErr := provider.BaseTemplate.Clone()
	if cloneErr != nil {
		return nil, cloneErr
	}
	fullPaths := make([]string, len(paths))
	for i, path := range paths {
		fullPaths[i] = filepath.Join(provider.RootPath, path)
	}
	newTemplate, templateErr := clone.ParseFiles(fullPaths...)
	if templateErr != nil {
		templateErr = fmt.Errorf("can't load template with name \"%s\": %w", name, templateErr)
		return nil, templateErr
	}
	provider.TemplateMap[name] = newTemplate
	return newTemplate, nil
}

func (provider *TemplateProvider) registerHomeTemplate() error {
	_, err := provider.RegisterTemplate("home", "home.gohtml")
	return err
}

func (provider *TemplateProvider) registerPeriodsListTemplate() error {
	_, err := provider.RegisterTemplate("periods-list", filepath.Join("periods", "periods_list.gohtml"))
	return err
}

func (provider *TemplateProvider) registerPeriodsDetailTemplate() error {
	_, err := provider.RegisterTemplate("periods-detail", filepath.Join("periods", "periods_detail.gohtml"))
	return err
}

func (provider *TemplateProvider) registerNewPeriodTemplate() error {
	_, err := provider.RegisterTemplate("periods-new", filepath.Join("periods", "periods_new.gohtml"))
	return err
}

func (provider *TemplateProvider) RegisterDefaults() (int, error) {
	// all functions have the same form, store them in a slice and apply them
	generators := []func() error{
		provider.registerHomeTemplate,
		provider.registerPeriodsListTemplate,
		provider.registerPeriodsDetailTemplate,
		provider.registerNewPeriodTemplate,
	}
	numTemplates := len(generators)
	for _, generator := range generators {
		if err := generator(); err != nil {
			return -1, err
		}
	}
	return numTemplates, nil
}
