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

type ConfigError struct {
	PollWebError
	ErrMessage   string
	WrappedError error
}

func NewConfigError(message string, wrapped error) ConfigError {
	return ConfigError{
		ErrMessage:   message,
		WrappedError: wrapped,
	}
}

func (err ConfigError) Error() string {
	message := "config error: " + err.ErrMessage
	if err.WrappedError != nil {
		message += ". caused by: " + err.WrappedError.Error()
	}
	return message
}

func (err ConfigError) Unwrap() error {
	return err.WrappedError
}
