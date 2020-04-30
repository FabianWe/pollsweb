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

package data

import (
	"fmt"
	"github.com/FabianWe/pollsweb"
	"github.com/nleof/goyesql"
	"io/ioutil"
	"path/filepath"
)

func LoadQueries(files []string) (goyesql.Queries, error) {
	queries := make(goyesql.Queries)

	for _, file := range files {
		fileQueries, err := goyesql.ParseFile(file)
		if err != nil {
			return nil, pollsweb.NewConfigError(fmt.Sprintf("failed to load query file %s", file),
				err)
		}
		for key, query := range fileQueries {
			if _, has := queries[key]; has {
				return nil, pollsweb.NewConfigError(fmt.Sprintf("duplicate query entry for key %s", key), nil)
			}
			queries[key] = query
		}
	}

	return queries, nil
}

func LoadQueriesFromDirectory(directory, fileExtension string) (goyesql.Queries, error) {
	if fileExtension == "" {
		fileExtension = ".sql"
	}
	files, listErr := ioutil.ReadDir(directory)
	if listErr != nil {
		return nil, pollsweb.NewConfigError("unable to read queries", listErr)
	}
	filePaths := make([]string, 0, len(files))
	for _, fileInfo := range files {
		fileName := fileInfo.Name()
		if filepath.Ext(fileName) == fileExtension {
			filePath := filepath.Join(directory, fileName)
			filePaths = append(filePaths, filePath)
		}
	}
	return LoadQueries(filePaths)
}
