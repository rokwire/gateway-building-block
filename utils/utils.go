// Copyright 2022 Board of Trustees of the University of Illinois.
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

package utils

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Filter represents find filter for finding entities by the their fields
type Filter struct {
	Items []FilterItem
}

// FilterItem represents find filter pair - field/value
type FilterItem struct {
	Field string
	Value []string
}

// ConstructFilter constructs Filter from the http request params
func ConstructFilter(r *http.Request) *Filter {
	values := r.URL.Query()
	if len(values) == 0 {
		return nil
	}

	var filter Filter
	var items []FilterItem
	for k, v := range values {
		if len(v) > 0 {
			items = append(items, FilterItem{Field: k, Value: v})
		}
	}
	filter.Items = items
	return &filter
}

// ModifyHTMLContent removes all not web href links. It also remove web links which points to pdf document
// For example:
// <a href="mailto:email@abc.abc">email@abc.abc</a> -> email@abc.abc
// <a href="ftp://server/file">Some text</a> -> Some text
// <a href="tel:1234">1234</a> -> 1234
//
// <a href="https://humanresources.illinois.edu/assets/docs/COVID-19-Pay-Continuation-Protocol-Final-3-22-2020.pdf">the university's pay continuation protocol</a> ->
// the university's pay continuation protocol(https://humanresources.illinois.edu/assets/docs/COVID-19-Pay-Continuation-Protocol-Final-3-22-2020.pdf)
func ModifyHTMLContent(input string) string {
	reader := strings.NewReader(input)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Printf("error creating reader from the html string - %s\n", err)
		//there is no what to do so return the input
		return input
	}

	//process
	doc.Find("a").Each(func(_ int, link *goquery.Selection) {
		text := strings.TrimSpace(link.Text())
		href, ok := link.Attr("href")
		if ok && len(href) > 0 {

			splitHref := strings.Split(href, ":")
			if len(splitHref) > 0 {
				protocol := splitHref[0]

				if protocol == "http" || protocol == "https" {
					//it is a web protocol, so we just need to look for .pdf resources
					if strings.HasSuffix(href, ".pdf") {
						log.Printf("modifying.. href - %s\ttext - %s\n", href, text)
						link.ReplaceWithHtml(text + "(" + href + ")")
					}
				} else {
					//it is not а web protocol, so here we need to apply modifications

					log.Printf("modifying.. href - %s\ttext - %s\n", href, text)
					link.ReplaceWithHtml(text)
				}
			}

		}
	})

	body := doc.Find("body")
	if body == nil {
		log.Printf("body is nil for some reasons - %s\n", input)
		//there is no what to do so return the input
		return input
	}
	final, err := body.Html()
	if err != nil {
		log.Printf("error getting html from body - %s\n", err)
		//there is no what to do so return the input
		return input
	}
	return final
}

// LogRequest logs the request as hide some header fields because of security reasons
func LogRequest(req *http.Request) {
	if req == nil {
		return
	}

	method := req.Method
	path := req.URL.Path

	val, ok := req.Header["User-Agent"]
	if ok && len(val) != 0 && val[0] == "ELB-HealthChecker/2.0" {
		return
	}

	header := make(map[string][]string)
	for key, value := range req.Header {
		var logValue []string
		//do not log api keys, cookies and Authorization
		if key == "Rokwire-Api-Key" || key == "User-Id" || key == "Cookie" ||
			key == "Authorization" || key == "Rokwire-Hs-Api-Key" || key == "Group" ||
			key == "Rokwire-Acc-Id" || key == "Csrf" {
			logValue = append(logValue, "---")
		} else {
			logValue = value
		}
		header[key] = logValue
	}
	log.Printf("%s %s %s", method, path, header)
}

// GetLogUUIDValue prepares UUID to be logged.
func GetLogUUIDValue(identifier string) string {
	if len(identifier) < 26 {
		return fmt.Sprintf("bad identifier - %s", identifier)
	}

	sub := identifier[:26]
	return fmt.Sprintf("%s***", sub)
}

// GetLogValue prepares a sensitive data to be logged.
func GetLogValue(value string) string {
	if len(value) <= 3 {
		return "***"
	}
	last3 := value[len(value)-3:]
	return fmt.Sprintf("***%s", last3)
}

// Equal compares two slices
func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// EqualPointers compares two pointers slices
func EqualPointers(a, b *[]string) bool {
	if a == nil && b == nil {
		return true //equals
	}
	if a != nil && b == nil {
		return false // not equals
	}
	if a == nil && b != nil {
		return false // not equals
	}

	//both are not nil
	return Equal(*a, *b)
}

// GetInt gives the value which this pointer points. Gives 0 if the pointer is nil
func GetInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

// GetBool gives the value which this pointer points. Gives false if the pointer is nil
func GetBool(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

// GetString gives the value which this pointer points. Gives empty string if the pointer is nil
func GetString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// GetTime gives the value which this pointer points. Gives empty string if the pointer is nil
func GetTime(time *time.Time) string {
	if time == nil {
		return ""
	}
	return fmt.Sprintf("%s", time)
}

// SortVersions sorts the versions list. The format is x.x.x or x.x which is the short for x.x.0
func SortVersions(versions []string) {
	//sort
	sort.Slice(versions, func(i, j int) bool {
		v1 := versions[i]
		v2 := versions[j]
		return !IsVersionLess(v1, v2)
	})
}

// IsVersionLess checks if v1 is less than v2. The format is x.x.x or x.x which is the short for x.x.0
func IsVersionLess(v1 string, v2 string) bool {
	var v1Major, v1Minor, v1Patch int
	var v2Major, v2Minor, v2Patch int

	v1Elements := strings.Split(v1, ".")
	v2Elements := strings.Split(v2, ".")

	v1Major, _ = strconv.Atoi(v1Elements[0])
	v1Minor, _ = strconv.Atoi(v1Elements[1])
	if len(v1Elements) == 2 {
		v1Patch = 0
	} else {
		v1Patch, _ = strconv.Atoi(v1Elements[2])
	}

	v2Major, _ = strconv.Atoi(v2Elements[0])
	v2Minor, _ = strconv.Atoi(v2Elements[1])
	if len(v2Elements) == 2 {
		v2Patch = 0
	} else {
		v2Patch, _ = strconv.Atoi(v2Elements[2])
	}

	//1 first check major
	if v1Major < v2Major {
		return true
	}
	if v1Major > v2Major {
		return false
	}

	//2. majors are equals so check minors
	if v1Minor < v2Minor {
		return true
	}
	if v1Minor > v2Minor {
		return false
	}

	//3. minors are equals so check patch
	if v1Patch < v2Patch {
		return true
	}
	if v1Patch > v2Patch {
		return false
	}

	// they are equals
	return false
}
