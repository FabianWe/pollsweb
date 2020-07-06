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
	"fmt"
	"html/template"
	"path/filepath"
)

const (
	BaseTemplatePath = "base.gohtml"
)

func GetDefaultFuncMap() template.FuncMap {
	return template.FuncMap{}
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

func (provider *TemplateProvider) InitBase(additionalFiles ...string) error {
	paths := make([]string, len(additionalFiles)+1)
	paths[0] = filepath.Join(provider.RootPath, BaseTemplatePath)
	for i, file := range additionalFiles {
		paths[i+1] = filepath.Join(provider.RootPath, file)
	}
	base := template.New("").Funcs(provider.FuncMap)
	var err error
	base, err = template.ParseFiles(paths...)
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

func (provider *TemplateProvider) RegisterDefaults() (int, error) {
	// all functions have the same form, store them in a slice and apply them
	generators := []func() error{
		provider.registerHomeTemplate,
		provider.registerPeriodsListTemplate,
	}
	numTemplates := len(generators)
	for _, generator := range generators {
		if err := generator(); err != nil {
			return -1, err
		}
	}
	return numTemplates, nil
}
