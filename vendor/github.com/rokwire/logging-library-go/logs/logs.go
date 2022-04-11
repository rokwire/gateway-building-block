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

package logs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/errors"
	"github.com/rokwire/logging-library-go/logutils"
	"github.com/sirupsen/logrus"
)

//HttpResponse is an entity which contains the data to be sent in an HTTP response
type HttpResponse struct {
	ResponseCode int
	Headers      map[string][]string
	Body         []byte
}

//NewHttpResponse generates an HttpResponse with the provided data
func NewHttpResponse(body []byte, headers map[string]string, code int) HttpResponse {
	preparedHeaders := make(map[string][]string, len(headers))
	for key, value := range headers {
		preparedHeaders[key] = []string{value}
	}

	return HttpResponse{ResponseCode: code, Headers: preparedHeaders, Body: body}
}

//NewErrorHttpResponse generates an HttpResponse with the correct headers for an error string
func NewErrorHttpResponse(body string, code int) HttpResponse {
	headers := map[string][]string{}
	headers["Content-Type"] = []string{"text/plain; charset=utf-8"}
	headers["X-Content-Type-Options"] = []string{"nosniff"}

	return HttpResponse{ResponseCode: code, Headers: headers, Body: []byte(body)}
}

//NewErrorJsonHttpResponse generates an HttpResponse with the correct headers for a JSON encoded error
func NewErrorJsonHttpResponse(body string, code int) HttpResponse {
	headers := map[string][]string{}
	headers["Content-Type"] = []string{"application/json; charset=utf-8"}
	headers["X-Content-Type-Options"] = []string{"nosniff"}

	return HttpResponse{ResponseCode: code, Headers: headers, Body: []byte(body)}
}

//HttpRequestProperties is an entity which contains the properties of an HTTP request
type HttpRequestProperties struct {
	Method     string
	Path       string
	RemoteAddr string
	UserAgent  string
}

func (h HttpRequestProperties) Match(r *http.Request) bool {
	if h.Method != "" && h.Method != r.Method {
		return false
	}

	if h.Path != "" && h.Path != r.URL.Path {
		return false
	}

	if h.RemoteAddr != "" && h.RemoteAddr != r.RemoteAddr {
		return false
	}

	if h.UserAgent != "" && h.UserAgent != r.UserAgent() {
		return false
	}

	return true
}

//NewAwsHealthCheckHttpRequestProperties creates an HttpRequestProperties object for a standard AWS ELB health checker
//	Path: The path that the health checks are performed on. If empty, "/version" is used as the default value.
func NewAwsHealthCheckHttpRequestProperties(path string) HttpRequestProperties {
	if path == "" {
		path = "/version"
	}
	return HttpRequestProperties{Method: "GET", Path: path, UserAgent: "ELB-HealthChecker/2.0"}
}

//Logger struct defines a wrapper for a logger object
type Logger struct {
	entry            *logrus.Entry
	sensitiveHeaders []string
	suppressRequests []HttpRequestProperties
}

//LoggerOpts provides configuration options for the Logger type
type LoggerOpts struct {
	//JsonFmt: When true, logs will be output in JSON format. Otherwise logs will be in logfmt
	JsonFmt bool
	//SensitiveHeaders: A list of any headers that contain sensitive information and should not be logged
	//				    Defaults: Authorization, Csrf
	SensitiveHeaders []string
	//SuppressRequests: A list of HttpRequestProperties of requests that should not be logged
	//					Any "Warn" or higher severity logs will still be logged.
	//					This is useful to prevent info logs from health checks and similar requests from
	//					flooding the logs
	//					All specified fields in the provided HttpRequestProperties must match for the logs
	//					to be suppressed. Empty fields will be ignored.
	SuppressRequests []HttpRequestProperties
}

//NewLogger is constructor for a logger object with initial configuration at the service level
// Params:
//		serviceName: A meaningful service name to be associated with all logs
//		opts: Configuration options for the Logger
func NewLogger(serviceName string, opts *LoggerOpts) *Logger {
	var baseLogger = logrus.New()
	sensitiveHeaders := []string{"Authorization", "Csrf"}
	var suppressRequests []HttpRequestProperties

	if opts != nil {
		if opts.JsonFmt {
			baseLogger.Formatter = &logrus.JSONFormatter{}
		} else {
			baseLogger.Formatter = &logrus.TextFormatter{}
		}

		sensitiveHeaders = append(sensitiveHeaders, opts.SensitiveHeaders...)
		suppressRequests = opts.SuppressRequests
	}

	standardFields := logrus.Fields{"service_name": serviceName} //All common fields for logs of a given service
	contextLogger := &Logger{entry: baseLogger.WithFields(standardFields), sensitiveHeaders: sensitiveHeaders, suppressRequests: suppressRequests}
	return contextLogger
}

func (l *Logger) SetLevel(level logLevel) {
	switch level {
	case Debug:
		l.entry.Logger.SetLevel(logrus.DebugLevel)
	case Info:
		l.entry.Logger.SetLevel(logrus.InfoLevel)
	case Warn:
		l.entry.Logger.SetLevel(logrus.WarnLevel)
	case Error:
		l.entry.Logger.SetLevel(logrus.ErrorLevel)
	default:
	}
}

func (l *Logger) withFields(fields logutils.Fields) *Logger {
	return &Logger{entry: l.entry.WithFields(fields.ToMap())}
}

//Fatal prints the log with a fatal error message and stops the service instance
//WARNING: Please only use for critical error messages that should prevent the service from running
func (l *Logger) Fatal(message string) {
	l.entry.Fatal(message)
}

//Fatalf prints the log with a fatal format error message and stops the service instance
//WARNING: Please only use for critical error messages that should prevent the service from running
func (l *Logger) Fatalf(message string, args ...interface{}) {
	l.entry.Fatalf(message, args...)
}

//Error prints the log at error level with given message
func (l *Logger) Error(message string) {
	l.entry.Error(message)
}

//ErrorWithFields prints the log at error level with given fields and message
func (l *Logger) ErrorWithFields(message string, fields logutils.Fields) {
	l.entry.WithFields(fields.ToMap()).Error(message)
}

//Errorf prints the log at error level with given formatted string
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

//Info prints the log at info level with given message
func (l *Logger) Info(message string) {
	l.entry.Info(message)
}

//InfoWithFields prints the log at info level with given fields and message
func (l *Logger) InfoWithFields(message string, fields logutils.Fields) {
	l.entry.WithFields(fields.ToMap()).Info(message)
}

//Infof prints the log at info level with given formatted string
func (l *Logger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

//Debug prints the log at debug level with given message
func (l *Logger) Debug(message string) {
	l.entry.Debug(message)
}

//DebugWithFields prints the log at debug level with given fields and message
func (l *Logger) DebugWithFields(message string, fields logutils.Fields) {
	l.entry.WithFields(fields.ToMap()).Debug(message)
}

//Debugf prints the log at debug level with given formatted string
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

//Warn prints the log at warn level with given message
func (l *Logger) Warn(message string) {
	l.entry.Warn(message)
}

//WarnWithFields prints the log at warn level with given fields and message
func (l *Logger) WarnWithFields(message string, fields logutils.Fields) {
	l.entry.WithFields(fields.ToMap()).Warn(message)
}

//Warnf prints the log at warn level with given formatted string
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

type RequestContext struct {
	Method     string
	Path       string
	Headers    map[string][]string
	PrevSpanID string
}

func (r RequestContext) String() string {
	return fmt.Sprintf("%s %s prev_span_id: %s headers: %v", r.Method, r.Path, r.PrevSpanID, r.Headers)
}

//Log struct defines a log object of a request
type Log struct {
	logger    *Logger
	traceID   string
	spanID    string
	request   RequestContext
	context   logutils.Fields
	layer     int
	suppress  bool
	hasLogged bool
}

//NewLog is a constructor for a log object
func (l *Logger) NewLog(traceID string, request RequestContext) *Log {
	if traceID == "" {
		traceID = uuid.New().String()
	}
	spanID := uuid.New().String()
	log := &Log{l, traceID, spanID, request, logutils.Fields{}, 0, false, false}
	return log
}

//NewRequestLog is a constructor for a log object for a request
func (l *Logger) NewRequestLog(r *http.Request) *Log {
	if r == nil {
		return &Log{logger: l}
	}

	traceID := r.Header.Get("trace-id")
	if traceID == "" {
		traceID = uuid.New().String()
	}

	prevSpanID := r.Header.Get("span-id")
	spanID := uuid.New().String()

	method := r.Method
	path := r.URL.Path

	headers := make(map[string][]string)
	for key, value := range r.Header {
		var logValue []string
		//do not log sensitive information
		if logutils.ContainsString(l.sensitiveHeaders, key) {
			logValue = append(logValue, "---")
		} else {
			logValue = value
		}
		headers[key] = logValue
	}

	request := RequestContext{Method: method, Path: path, Headers: headers, PrevSpanID: prevSpanID}

	suppress := false
	for _, props := range l.suppressRequests {
		if props.Match(r) {
			suppress = true
			break
		}
	}

	log := &Log{l, traceID, spanID, request, logutils.Fields{}, 0, suppress, false}
	return log
}

func (l *Log) resetLayer() {
	l.layer = 0
}

func (l *Log) addLayer(layer int) {
	l.layer += layer
}

//getRequestFields() populates a map with all the fields of a request
//	layer: Number of function calls between caller and getRequestFields()
func (l *Log) getRequestFields() logutils.Fields {
	if l == nil {
		return logutils.Fields{}
	}

	l.hasLogged = true
	fields := logutils.Fields{"trace_id": l.traceID, "span_id": l.spanID, "function_name": getLogPrevFuncName(l.layer)}
	if l.suppress {
		fields["suppress"] = true
	}
	l.resetLayer()

	return fields
}

//SetHeaders sets the trace and span id headers for a request to another service
//	This function should always be called when making a request to another rokwire service
func (l *Log) SetHeaders(r *http.Request) {
	if l == nil {
		return
	}

	r.Header.Set("trace-id", l.traceID)
	r.Header.Set("span-id", l.spanID)
}

//LogData logs and returns a data message at the designated level
//	level: The log level (Info, Debug, Warn, Error)
//	status: The status of the data
//	dataType: The data type
//	args: Any args that should be included in the message (nil if none)
func (l *Log) LogData(level logLevel, status logutils.MessageDataStatus, dataType logutils.MessageDataType, args logutils.MessageArgs) string {
	msg := logutils.MessageData(status, dataType, args)
	l.addLayer(1)

	switch level {
	case Debug:
		l.Debug(msg)
	case Info:
		l.Info(msg)
	case Warn:
		l.Warn(msg)
	case Error:
		l.Error(msg)
	default:
		l.resetLayer()
	}

	return msg
}

//WarnData logs and returns a data message for the given error at the warn level
//	status: The status of the data
//	dataType: The data type
//	err: Error message
func (l *Log) WarnData(status logutils.MessageDataStatus, dataType logutils.MessageDataType, err error) string {
	message := logutils.MessageData(status, dataType, nil)

	l.addLayer(1)
	defer l.resetLayer()

	return l.WarnError(message, err)
}

//ErrorData logs and returns a data message for the given error at the error level
//	status: The status of the data
//	dataType: The data type
//	err: Error message
func (l *Log) ErrorData(status logutils.MessageDataStatus, dataType logutils.MessageDataType, err error) string {
	message := logutils.MessageData(status, dataType, nil)

	l.addLayer(1)
	defer l.resetLayer()

	return l.LogError(message, err)
}

//RequestErrorData logs a data message and error and sets it as the HTTP response
//	w: The http response writer for the active request
//	status: The status of the data
//	dataType: The data type
//	args: Any args that should be included in the message (nil if none)
//	err: The error received from the application
//	code: The HTTP response code to be set
//	showDetails: Only provide 'msg' not 'err' in HTTP response when false
func (l *Log) RequestErrorData(w http.ResponseWriter, status logutils.MessageDataStatus, dataType logutils.MessageDataType, args logutils.MessageArgs, err error, code int, showDetails bool) {
	message := logutils.MessageData(status, dataType, args)

	l.addLayer(1)
	defer l.resetLayer()

	l.RequestError(w, message, err, code, showDetails)
}

//HttpResponseErrorData logs a data message and error and generates an HttpResponse
//	status: The status of the data
//	dataType: The data type
//	args: Any args that should be included in the message (nil if none)
//	err: The error received from the application
//	code: The HTTP response code to be set
//	showDetails: Only provide 'msg' not 'err' in HTTP response when false
func (l *Log) HttpResponseErrorData(status logutils.MessageDataStatus, dataType logutils.MessageDataType, args logutils.MessageArgs, err error, code int, showDetails bool) HttpResponse {
	message := logutils.MessageData(status, dataType, args)

	l.addLayer(1)
	defer l.resetLayer()

	return l.HttpResponseError(message, err, code, showDetails)
}

//LogAction logs and returns an action message at the designated level
//	level: The log level (Info, Debug, Warn, Error)
//	status: The status of the action
//	action: The action that is occurring
//	dataType: The data type that the action is occurring on
//	args: Any args that should be included in the message (nil if none)
func (l *Log) LogAction(level logLevel, status logutils.MessageActionStatus, action logutils.MessageActionType, dataType logutils.MessageDataType, args logutils.MessageArgs) string {
	msg := logutils.MessageAction(status, action, dataType, args)
	l.addLayer(1)

	switch level {
	case Debug:
		l.Debug(msg)
	case Info:
		l.Info(msg)
	case Warn:
		l.Warn(msg)
	case Error:
		l.Error(msg)
	default:
		l.resetLayer()
	}

	return msg
}

//WarnAction logs and returns an action message for the given error at the warn level
//	action: The action that is occurring
//	dataType: The data type that the action is occurring on
//	err: Error message
func (l *Log) WarnAction(action logutils.MessageActionType, dataType logutils.MessageDataType, err error) string {
	message := logutils.MessageAction(logutils.StatusError, action, dataType, nil)

	l.addLayer(1)
	defer l.resetLayer()

	return l.WarnError(message, err)
}

//ErrorAction logs and returns an action message for the given error at the error level
//	action: The action that is occurring
//	dataType: The data type that the action is occurring on
//	err: Error message
func (l *Log) ErrorAction(action logutils.MessageActionType, dataType logutils.MessageDataType, err error) string {
	message := logutils.MessageAction(logutils.StatusError, action, dataType, nil)

	l.addLayer(1)
	defer l.resetLayer()

	return l.LogError(message, err)
}

//RequestSuccessAction sets the provided success action message as the HTTP response, sets standard headers, and stores the message
// 	to the log context
//	Params:
//		w: The http response writer for the active request
//		action: The action that is occurring
//		dataType: The data type that the action is occurring on
//		args: Any args that should be included in the message (nil if none)
func (l *Log) RequestSuccessAction(w http.ResponseWriter, action logutils.MessageActionType, dataType logutils.MessageDataType, args logutils.MessageArgs) {
	message := logutils.MessageAction(logutils.StatusSuccess, action, dataType, args)
	l.RequestSuccessMessage(w, message)
}

//RequestErrorAction logs an action message and error and sets it as the HTTP response
//	w: The http response writer for the active request
//	action: The action that is occurring
//	dataType: The data type
//	args: Any args that should be included in the message (nil if none)
//	err: The error received from the application
//	code: The HTTP response code to be set
//	showDetails: Only generated message not 'err' in HTTP response when false
func (l *Log) RequestErrorAction(w http.ResponseWriter, action logutils.MessageActionType, dataType logutils.MessageDataType, args logutils.MessageArgs, err error, code int, showDetails bool) {
	message := logutils.MessageAction(logutils.StatusError, action, dataType, args)

	l.addLayer(1)
	defer l.resetLayer()

	l.RequestError(w, message, err, code, showDetails)
}

//HttpResponseSuccessAction generates an HttpResponse with the provided success action message, sets standard headers,
// 	and stores the message to the log context
//	Params:
//		action: The action that is occurring
//		dataType: The data type that the action is occurring on
//		args: Any args that should be included in the message (nil if none)
func (l *Log) HttpResponseSuccessAction(action logutils.MessageActionType, dataType logutils.MessageDataType, args logutils.MessageArgs) HttpResponse {
	message := logutils.MessageAction(logutils.StatusSuccess, action, dataType, args)
	return l.HttpResponseSuccessMessage(message)
}

//HttpResponseErrorAction logs an action message and error and generates an HttpResponse
//	action: The action that is occurring
//	dataType: The data type
//	args: Any args that should be included in the message (nil if none)
//	err: The error received from the application
//	code: The HTTP response code to be set
//	showDetails: Only generated message not 'err' in HTTP response when false
func (l *Log) HttpResponseErrorAction(action logutils.MessageActionType, dataType logutils.MessageDataType, args logutils.MessageArgs, err error, code int, showDetails bool) HttpResponse {
	message := logutils.MessageAction(logutils.StatusError, action, dataType, args)

	l.addLayer(1)
	defer l.resetLayer()

	return l.HttpResponseError(message, err, code, showDetails)
}

//Info prints the log at info level with given message
func (l *Log) Info(message string) {
	if l == nil || l.logger == nil || l.suppress {
		return
	}

	requestFields := l.getRequestFields()
	l.logger.withFields(requestFields).Info(message)
}

//InfoWithDetails prints the log at info level with given fields and message
func (l *Log) InfoWithDetails(message string, details logutils.Fields) {
	if l == nil || l.logger == nil || l.suppress {
		return
	}

	requestFields := l.getRequestFields()
	requestFields["details"] = details
	l.logger.withFields(requestFields).Info(message)
}

//Infof prints the log at info level with given formatted string
func (l *Log) Infof(format string, args ...interface{}) {
	if l == nil || l.logger == nil || l.suppress {
		return
	}

	requestFields := l.getRequestFields()
	l.logger.withFields(requestFields).Infof(format, args...)
}

//Debug prints the log at debug level with given message
func (l *Log) Debug(message string) {
	if l == nil || l.logger == nil || l.suppress {
		return
	}

	requestFields := l.getRequestFields()
	l.logger.withFields(requestFields).Debug(message)
}

//DebugWithDetails prints the log at debug level with given fields and message
func (l *Log) DebugWithDetails(message string, details logutils.Fields) {
	if l == nil || l.logger == nil || l.suppress {
		return
	}

	requestFields := l.getRequestFields()
	requestFields["details"] = details
	l.logger.withFields(requestFields).Debug(message)
}

//Debugf prints the log at debug level with given formatted string
func (l *Log) Debugf(format string, args ...interface{}) {
	if l == nil || l.logger == nil || l.suppress {
		return
	}

	requestFields := l.getRequestFields()
	l.logger.withFields(requestFields).Debugf(format, args...)
}

//Warn prints the log at warn level with given message
func (l *Log) Warn(message string) {
	if l == nil || l.logger == nil {
		return
	}

	requestFields := l.getRequestFields()
	l.logger.withFields(requestFields).Warn(message)
}

//WarnWithDetails prints the log at warn level with given details and message
func (l *Log) WarnWithDetails(message string, details logutils.Fields) {
	if l == nil || l.logger == nil {
		return
	}

	requestFields := l.getRequestFields()
	requestFields["details"] = details
	l.logger.withFields(requestFields).Warn(message)
}

//Warnf prints the log at warn level with given formatted string
func (l *Log) Warnf(format string, args ...interface{}) {
	if l == nil || l.logger == nil {
		return
	}

	requestFields := l.getRequestFields()
	l.logger.withFields(requestFields).Warnf(format, args...)
}

//WarnError prints the log at warn level with given message and error
//	Returns error message as string
func (l *Log) WarnError(message string, err error) string {
	msg := fmt.Sprintf("%s: %s", message, errors.Root(err))
	if l == nil || l.logger == nil {
		return msg
	}

	requestFields := l.getRequestFields()
	if err != nil {
		requestFields["error"] = err.Error()
	}
	l.logger.withFields(requestFields).Warn(message)
	return msg
}

//LogError prints the log at error level with given message and error
//	Returns combined error message as string
func (l *Log) LogError(message string, err error) string {
	msg := fmt.Sprintf("%s: %s", message, errors.Root(err))
	if l == nil || l.logger == nil {
		return msg
	}

	requestFields := l.getRequestFields()
	if err != nil {
		requestFields["error"] = err.Error()
	}
	l.logger.withFields(requestFields).Error(message)
	return msg
}

//Error prints the log at error level with given message
// Note: If possible, use LogError() instead
func (l *Log) Error(message string) {
	if l == nil || l.logger == nil {
		return
	}

	requestFields := l.getRequestFields()
	l.logger.withFields(requestFields).Error(message)
}

//ErrorWithDetails prints the log at error level with given details and message
func (l *Log) ErrorWithDetails(message string, details logutils.Fields) {
	if l == nil || l.logger == nil {
		return
	}

	requestFields := l.getRequestFields()
	requestFields["details"] = details
	l.logger.withFields(requestFields).Error(message)
}

//Errorf prints the log at error level with given formatted string
// Note: If possible, use LogError() instead
func (l *Log) Errorf(format string, args ...interface{}) {
	if l == nil || l.logger == nil {
		return
	}

	requestFields := l.getRequestFields()
	l.logger.withFields(requestFields).Errorf(format, args...)
}

//RequestSuccess sets "Success" as the HTTP response, sets standard headers, and stores the message
// 	to the log context
//	Params:
//		w: The http response writer for the active request
//		msg: The success message
func (l *Log) RequestSuccess(w http.ResponseWriter) {
	l.SetContext("status_code", http.StatusOK)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}

//RequestSuccessMessage sets the provided success message as the HTTP response, sets standard headers, and stores the message
// 	to the log context
//	Params:
//		w: The http response writer for the active request
//		msg: The success message
func (l *Log) RequestSuccessMessage(w http.ResponseWriter, message string) {
	l.SetContext("status_code", http.StatusOK)
	l.SetContext("success", message)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}

//RequestSuccessJSON sets the provided JSON as the HTTP response body, sets standard headers, and stores the request status
// 	to the log context
//	Params:
//		w: The http response writer for the active request
//		responseJSON: JSON encoded response data
func (l *Log) RequestSuccessJSON(w http.ResponseWriter, responseJSON []byte) {
	l.SetContext("status_code", http.StatusOK)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func (l *Log) requestErrorHelper(message string, err error, code int, showDetails bool) string {
	l.addLayer(1)
	defer l.resetLayer()

	l.SetContext("status_code", code)

	status := errors.Status(err)
	if len(status) == 0 {
		status = strings.ReplaceAll(strings.ToLower(http.StatusText(code)), " ", "-")
	}
	l.SetContext("status", status)

	detailMsg := l.LogError(message, err)
	if showDetails {
		message = detailMsg
	}

	response := map[string]string{"status": status, "message": message}
	jsonMessage, _ := json.Marshal(response)
	message = string(jsonMessage)
	return message
}

// HttpJsonError replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
// The error message should be JSON encoded.
func HttpJsonError(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, err)
}

//RequestError logs the provided message and error and sets it as the HTTP response
//	Params:
//		w: The http response writer for the active request
//		message: The error message
//		err: The error received from the application
//		code: The HTTP response code to be set
//		showDetails: Only provide 'message' not 'err' in HTTP response when false
func (l *Log) RequestError(w http.ResponseWriter, message string, err error, code int, showDetails bool) {
	l.addLayer(1)
	defer l.resetLayer()

	message = l.requestErrorHelper(message, err, code, showDetails)
	HttpJsonError(w, message, code)
}

//HttpResponseSuccess generates an HttpResponse with the message "Success", sets standard headers, and stores the status
// 	to the log context
func (l *Log) HttpResponseSuccess() HttpResponse {
	l.SetContext("status_code", http.StatusOK)

	headers := map[string][]string{}
	headers["Content-Type"] = []string{"text/plain; charset=utf-8"}
	return HttpResponse{ResponseCode: http.StatusOK, Headers: headers, Body: []byte("Success")}
}

//HttpResponseSuccess generates an HttpResponse with the provided success message, sets standard headers, and stores the message
// 	and status to the log context
//	Params:
//		msg: The success message
func (l *Log) HttpResponseSuccessMessage(message string) HttpResponse {
	l.SetContext("status_code", http.StatusOK)
	l.SetContext("success", message)

	headers := map[string][]string{}
	headers["Content-Type"] = []string{"text/plain; charset=utf-8"}
	return HttpResponse{ResponseCode: http.StatusOK, Headers: headers, Body: []byte(message)}
}

//HttpResponseSuccessJSON generates an HttpResponse with the provided JSON as the HTTP response body, sets standard headers,
// 	and stores the status to the log context
//	Params:
//		json: JSON encoded response data
func (l *Log) HttpResponseSuccessJSON(json []byte) HttpResponse {
	l.SetContext("status_code", http.StatusOK)

	headers := map[string][]string{}
	headers["Content-Type"] = []string{"application/json; charset=utf-8"}
	return HttpResponse{ResponseCode: http.StatusOK, Headers: headers, Body: json}
}

//HttpResponseError logs the provided message and error and generates an HttpResponse
//	Params:
//		message: The error message
//		err: The error received from the application
//		code: The HTTP response code to be set
//		showDetails: Only provide 'message' not 'err' in HTTP response when false
func (l *Log) HttpResponseError(message string, err error, code int, showDetails bool) HttpResponse {
	l.addLayer(1)
	defer l.resetLayer()

	message = l.requestErrorHelper(message, err, code, showDetails)
	return NewErrorJsonHttpResponse(message, code)
}

//AddContext adds any relevant unstructured data to context map
// If the provided key already exists in the context, an error is returned
func (l *Log) AddContext(fieldName string, value interface{}) error {
	if l == nil {
		return fmt.Errorf("error adding context: nil log")
	}

	if _, ok := l.context[fieldName]; ok {
		return fmt.Errorf("error adding context: %s already exists", fieldName)
	}

	l.context[fieldName] = value
	return nil
}

//SetContext sets the provided context key to the provided value
func (l *Log) SetContext(fieldName string, value interface{}) {
	l.context[fieldName] = value
}

//RequestReceived prints the request context of a log object
func (l *Log) RequestReceived() {
	if l == nil || l.logger == nil || l.suppress {
		return
	}

	fields := l.getRequestFields()
	fields["request"] = l.request
	l.logger.InfoWithFields("Request Received", fields)
}

//RequestComplete prints the context of a log object
func (l *Log) RequestComplete() {
	if l == nil || l.logger == nil {
		return
	}

	hasLogged := l.hasLogged
	fields := l.getRequestFields()

	if l.suppress {
		if hasLogged {
			fields["request"] = l.request
		} else {
			return
		}
	}

	fields["context"] = l.context
	l.logger.InfoWithFields("Request Complete", fields)
}

//getLogPrevFuncName - fetches the calling function name when logging
//	layer: Number of internal library function calls above caller
func getLogPrevFuncName(layer int) string {
	return logutils.GetFuncName(5 + layer)
}
