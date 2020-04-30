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

// internalErrorSentinelType is used only for the constant "ErrPollWeb", this way we have one sentinel value
// to expose.
// The type PollWebError tests for this constant in its Is(error) method.
type internalErrorSentinelType struct{}

// The type must implement the error interface.
func (err internalErrorSentinelType) Error() string {
	return "gopolls web error"
}

// ErrPollWeb is a constant that can be used with a type check.
// All internal errors can be used in a statement like errors.Is(err, ErrPollWeb)
// and return true.
// This can be useful when you want to distinguish between an error from gopollsweb and an "outside" error.
// If you want to dig deeper, for example find out if an error is of a special internal type, you should use
// errors.As(err, *ERROR_TYPE).
var ErrPollWeb = internalErrorSentinelType{}

// PollWebError is an error used for errors that should be considered a library internal error, such as verification
//errors.
// The type itself does not implement the error interface, but only the method Is(err error) from the error
// package.
// This way you can just embed this type in your own error type and Is(err, ErrPollWeb) will return true.
type PollWebError struct{}

// Is returns true if err == ErrPollWeb.
func (pollErr PollWebError) Is(err error) bool {
	return err == ErrPollWeb
}

type UUIDGenError struct {
	PollWebError
	Wrapped error
}

func NewUUIDGenError(err error) UUIDGenError {
	return UUIDGenError{
		PollWebError: PollWebError{},
		Wrapped:      err,
	}
}

func (err UUIDGenError) Error() string {
	return "can't generate UUID: " + err.Wrapped.Error()
}

func (err UUIDGenError) Unwrap() error {
	return err.Wrapped
}
