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

package pollsweb

import (
	"go.uber.org/zap"
)

// InitLogger creates a new zap SugaredLogger.
// Debug can be used to define if debug messages should be logged as well.
func InitLogger(debug bool) (*zap.SugaredLogger, error) {
	var raw *zap.Logger
	var initErr error
	if debug {
		raw, initErr = zap.NewDevelopment()
	} else {
		raw, initErr = zap.NewProduction()
	}
	if initErr != nil {
		return nil, initErr
	}
	return raw.Sugar(), nil
}
