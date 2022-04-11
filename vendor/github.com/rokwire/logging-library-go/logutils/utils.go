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

package logutils

import (
	"fmt"
	"runtime"
)

//MessageData generates a message string for a data element
//	status: The status of the data
//	dataType: The data type
//	args: Any args that should be included in the message (nil if none)
func MessageData(status MessageDataStatus, dataType MessageDataType, args MessageArgs) string {
	argStr := ""
	if args != nil {
		argStr = args.String()
		if argStr != "" {
			argStr = ": " + argStr
		}
	}

	return fmt.Sprintf("%s %s%s", status, dataType, argStr)
}

//MessageAction generates a message string for an action
//	status: The status of the action
//	action: The action that is occurring
//	dataType: The data type that the action is occurring on
//	args: Any args that should be included in the message (nil if none)
func MessageAction(status MessageActionStatus, action MessageActionType, dataType MessageDataType, args MessageArgs) string {
	argStr := ""
	if args != nil {
		argStr = args.String()
		if argStr != "" {
			argStr = " for " + argStr
		}
	}

	return fmt.Sprintf("%s %s %s%s", status, action, dataType, argStr)
}

func ContainsString(slice []string, val string) bool {
	for _, v := range slice {
		if val == v {
			return true
		}
	}
	return false
}

//GetFuncName fetches the name of a function caller based on the numFrames
func GetFuncName(numFrames int) string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(numFrames, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}
