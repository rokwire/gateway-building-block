// Copyright 2021 Board of Trustees of the University of Illinois
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"fmt"
	"strings"

	"github.com/rokwire/logging-library-go/logutils"
)

//IsError returns true if the provided error interface is an Error
func IsError(err error) bool {
	_, ok := err.(*Error)
	return ok
}

//IsError returns the provided error interface as an Error
//	Returns nil if the provided error is not an Error
func AsError(err error) *Error {
	errOut, _ := err.(*Error)
	return errOut
}

//New returns an Error containing the provided message
func New(message string) *Error {
	message = strings.ToLower(message)
	return &Error{root: &ErrorContext{message: message, function: getErrorPrevFuncName()}}
}

//Newf returns an Error containing the formatted message
func Newf(message string, args ...interface{}) *Error {
	message = strings.ToLower(message)
	message = fmt.Sprintf(message, args...)
	return &Error{root: &ErrorContext{message: message, function: getErrorPrevFuncName()}}
}

//Wrap returns an Error containing the provided message and error
func Wrap(message string, err error) *Error {
	message = strings.ToLower(message)
	context := ErrorContext{message: message, function: getErrorPrevFuncName()}
	if e, ok := err.(*Error); ok {
		return e.wrap(&context)
	}
	return &Error{root: &context, internal: err}
}

//Wrapf returns an Error containing the formatted message and provided error
func Wrapf(format string, err error, args ...interface{}) *Error {
	format = strings.ToLower(format)
	message := fmt.Sprintf(format, args...)
	context := ErrorContext{message: message, function: getErrorPrevFuncName()}
	if e, ok := err.(*Error); ok {
		return e.wrap(&context)
	}
	return &Error{root: &context, internal: err}
}

//ErrorData generates an error for a data element
//	status: The status of the data
//	dataType: The data type that the error is occurring on
//	args: Any args that should be included in the message (nil if none)
func ErrorData(status logutils.MessageDataStatus, dataType logutils.MessageDataType, args logutils.MessageArgs) *Error {
	message := logutils.MessageData(status, dataType, args)
	message = strings.ToLower(message)
	return &Error{root: &ErrorContext{message: message, function: getErrorPrevFuncName()}}
}

//WrapErrorData wraps an error for a data element
//	status: The status of the data
//	dataType: The data type that the error is occurring on
//	args: Any args that should be included in the message (nil if none)
//  err: Error to wrap
func WrapErrorData(status logutils.MessageDataStatus, dataType logutils.MessageDataType, args logutils.MessageArgs, err error) *Error {
	message := logutils.MessageData(status, dataType, args)
	message = strings.ToLower(message)
	context := ErrorContext{message: message, function: getErrorPrevFuncName()}
	if e, ok := err.(*Error); ok {
		return e.wrap(&context)
	}
	return &Error{root: &context, internal: err}
}

//ErrorAction generates an error for an action
//	action: The action that is occurring
//	dataType: The data type that the action is occurring on
//	args: Any args that should be included in the message (nil if none)
func ErrorAction(action logutils.MessageActionType, dataType logutils.MessageDataType, args logutils.MessageArgs) *Error {
	message := logutils.MessageAction(logutils.StatusError, action, dataType, args)
	message = strings.ToLower(message)
	return &Error{root: &ErrorContext{message: message, function: getErrorPrevFuncName()}}
}

//WrapErrorAction wraps an error for an action
//	action: The action that is occurring
//	dataType: The data type that the action is occurring on
//	args: Any args that should be included in the message (nil if none)
//	err: Error to wrap
func WrapErrorAction(action logutils.MessageActionType, dataType logutils.MessageDataType, args logutils.MessageArgs, err error) *Error {
	message := logutils.MessageAction(logutils.StatusError, action, dataType, args)
	message = strings.ToLower(message)
	context := ErrorContext{message: message, function: getErrorPrevFuncName()}
	if e, ok := err.(*Error); ok {
		return e.wrap(&context)
	}
	return &Error{root: &context, internal: err}
}

//Root returns the root message of an Error
//	If not an Error, returns err.Error()
func Root(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(*Error); ok {
		return e.Root()
	}
	return err.Error()
}

//RootContext returns the root context of an Error
//	If not an Error, returns err.Error()
func RootContext(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(*Error); ok {
		return e.RootContext()
	}
	return err.Error()
}

//Trace returns the trace messages of an Error
//	If not an Error, returns err.Error()
func Trace(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(*Error); ok {
		return e.Trace()
	}
	return err.Error()
}

//TraceContext returns the trace context of an Error
//	If not an Error, returns err.Error()
func TraceContext(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(*Error); ok {
		return e.TraceContext()
	}
	return err.Error()
}

//RootErr returns the root message of an Error as an error
//	If not an Error, returns err
func RootErr(err error) error {
	if e, ok := err.(*Error); ok {
		return e.RootErr()
	}
	return err
}

//RootContextErr returns the root context of an Error as an error
//	If not an Error, returns err
func RootContextErr(err error) error {
	if e, ok := err.(*Error); ok {
		return e.RootContextErr()
	}
	return err
}

//TraceErr returns the trace messages of an Error as an error
//	If not an Error, returns err
func TraceErr(err error) error {
	if e, ok := err.(*Error); ok {
		return e.TraceErr()
	}
	return err
}

//TraceContextErr returns the trace context of an Error as an error
//	If not an Error, returns err
func TraceContextErr(err error) error {
	if e, ok := err.(*Error); ok {
		return e.TraceContextErr()
	}
	return err
}

//Status returns the status
//	If not an Error, returns ""
func Status(err error) string {
	if e, ok := err.(*Error); ok {
		return e.Status()
	}
	return ""
}

//SetStatus sets the status and returns the result
//	If not an Error, returns a new Error with the status
func SetStatus(err error, status string) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e.SetStatus(status)
	}
	errOut := Error{internal: err}
	return errOut.SetStatus(status)
}

//Tags returns the tags an Error
//	If not an Error, returns an empty list
func Tags(err error) []string {
	if e, ok := err.(*Error); ok {
		return e.Tags()
	}
	return []string{}
}

//AddTag adds the provided tag to err and returns the result
//	If not an Error, returns a new Error with the tag
func AddTag(err error, tag string) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e.AddTag(tag)
	}
	errOut := Error{internal: err}
	return errOut.AddTag(tag)
}

//HasTag returns true if err has the provided tag
//	If not an Error, returns false
func HasTag(err error, tag string) bool {
	if e, ok := err.(*Error); ok {
		return e.HasTag(tag)
	}
	return false
}

//getErrorPrevFuncName - fetches the previous function name for error functions
func getErrorPrevFuncName() string {
	return logutils.GetFuncName(4)
}
