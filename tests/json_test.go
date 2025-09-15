/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

type Tested struct {
	Updated int64 `json:"updated,omitempty"`
}
type Encloser struct {
	Enclosed *Tested `json:"enclosed"`
}

func checkConcrete(js string, t Encloser) {
	fmt.Printf("	Given     : %s\n", js)

	if err := json.Unmarshal([]byte(js), &t); err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("	Struct     : %s\n", t)

	newJS, err := json.Marshal(t)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("	Marshalled: %s\n\n", string(newJS))
}

func checkAsInterface(js string, t interface{}) {
	fmt.Printf("	Given     : %s\n", js)

	if err := json.Unmarshal([]byte(js), &t); err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("	Struct     : %s\n", t)

	newJS, err := json.Marshal(t)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("	Marshalled: %s\n\n", string(newJS))
}

func TestOne(t *testing.T) {
	var encloser Encloser
	fmt.Printf("\nAsInterface\n\n")
	checkAsInterface(`{"enclosed":{"updated":4}}`, encloser)
	checkAsInterface(`{"enclosed":{"updated":0}}`, encloser)
	checkAsInterface(`{"enclosed":{}}`, encloser)

	fmt.Printf("\nAsConcrete\n\n")

	checkConcrete(`{"enclosed":{"updated":4}}`, encloser)
	checkConcrete(`{"enclosed":{"updated":0}}`, encloser)
	checkConcrete(`{"enclosed":{}}`, encloser)

}
