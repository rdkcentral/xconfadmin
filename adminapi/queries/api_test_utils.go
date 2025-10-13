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
package queries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"

	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/rdkcentral/xconfadmin/common"

	"github.com/rdkcentral/xconfadmin/util"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"gotest.tools/assert"
)

const (
	NO_INPUT         = ""
	NO_POSTERMS      = ""
	NO_PRETERMS      = ""
	JSON_SUFFIX      = ".json"
	DATA_LOCN_SUFFIX = "_DATA_LOCATION"
)

type apiUnitTestCase struct {
	api       string
	inputs    string
	preTerms  string
	preP      func(tcase apiUnitTestCase, reqBytes *[]byte)
	method    string
	endpoint  string
	expRetVal int
	postTerms string
	postP     func(tcase apiUnitTestCase, rsp *http.Response, reqBody *bytes.Buffer)
}

func buildBytes(t *testing.T, tcase apiUnitTestCase, locn string, baseFileNames string) *bytes.Buffer {
	if strings.Contains(baseFileNames, "[") {
		newStr := strings.ReplaceAll(baseFileNames, "[", "")
		newStr = strings.ReplaceAll(newStr, "]", "")
		subStrs := strings.Split(newStr, " ")
		jsonBytes := buildBytesFromManyJsonFiles(t, tcase, locn, subStrs)
		return bytes.NewBuffer(bytes.Join(jsonBytes, []byte{}))
	}
	if strings.Contains(baseFileNames, "=") {
		kvMap, err := url.ParseQuery(baseFileNames)
		assert.NilError(t, err)
		return bytes.NewBuffer([]byte(kvMap.Encode()))
	}
	jsonBytes := buildBytesFromOneJsonFile(t, tcase, locn, baseFileNames)
	return bytes.NewBuffer(jsonBytes)

}

func buildBytesFromOneJsonFile(t *testing.T, tcase apiUnitTestCase, locn string, baseName string) (jsonBytes []byte) {
	if util.IsBlank(baseName) {
		return jsonBytes
	}
	var err error
	jsonBytes, err = os.ReadFile(locn + baseName + JSON_SUFFIX)
	assert.NilError(t, err)
	if tcase.preP != nil {
		tcase.preP(tcase, &jsonBytes)
	}

	return jsonBytes
}

func buildBytesFromManyJsonFiles(t *testing.T, tcase apiUnitTestCase, locn string, baseNames []string) (jsonBytes [][]byte) {
	jsonBytes = append(jsonBytes, []byte{'['})
	for i, v := range baseNames {
		jsonBytes = append(jsonBytes, buildBytesFromOneJsonFile(t, tcase, locn, v))
		if i != 0 {
			jsonBytes = append(jsonBytes, []byte{','})
		}
	}
	jsonBytes = append(jsonBytes, []byte{']'})
	return jsonBytes
}

func ExecRequest(r *http.Request, handler http.Handler) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

type apiUnitTest struct {
	t        *testing.T
	router   *mux.Router
	savedMap map[string]string
}

func (aut *apiUnitTest) replaceKeysByValues(tcase apiUnitTestCase, reqBytes *[]byte) {
	kvMap, err := url.ParseQuery(tcase.preTerms)
	assert.NilError(aut.t, err)

	for k, v := range kvMap {
		*reqBytes = []byte(strings.Replace(string(*reqBytes), k, v[0], -1))
	}
}

func (aut *apiUnitTest) end() {
}

func (aut *apiUnitTest) run(testCases []apiUnitTestCase) {
	oldLevel := log.GetLevel()
	log.SetLevel(log.WarnLevel)
	for _, tcase := range testCases {
		ipval := ""
		if tcase.inputs != NO_INPUT {
			ipval = fmt.Sprintf("--data-binary \"@%s.json\"", tcase.inputs)
		}
		fmt.Printf("\ncurl -i -H \"Accept: application/json\" -H  \"Content-Type: application/json\" --request %s \"http://localhost:9000%s%s\" %s\n", tcase.method, tcase.api, tcase.endpoint, ipval)
		_, present := os.LookupEnv("RUN_IN_LOCAL")
		if !present {
			aut.t.Skip("Running this test only on local till we figure out why getting deleted object succeeds on Jenkins")
		}
		if tcase.postTerms != "" {
			assert.Equal(aut.t, tcase.postP != nil, true)
		}
		if tcase.postP != nil {
			assert.Equal(aut.t, tcase.postTerms != NO_POSTERMS, true)
		}
		assert.Equal(aut.t, tcase.api != "", true)
		jsonBytes := buildBytes(aut.t, tcase, aut.getValOf(tcase.api+DATA_LOCN_SUFFIX), tcase.inputs)
		jsonBytesCopy := *jsonBytes
		jsonBytesCopy2 := *jsonBytes // make a copy because each set can be unmarshalled only once.

		req, err := http.NewRequest(tcase.method, tcase.api+tcase.endpoint, jsonBytes)
		assert.NilError(aut.t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		res := ExecRequest(req, aut.router).Result()
		defer res.Body.Close()
		fmt.Printf("%s\n", res.Status)

		var resBytes []byte
		if res.Body != nil {
			resBytes, _ = ioutil.ReadAll(res.Body)
		}

		res.Body = ioutil.NopCloser(bytes.NewBuffer(resBytes))
		aut.apiErrorMessageReporter(tcase, res, &jsonBytesCopy2)

		assert.Equal(aut.t, res.StatusCode, tcase.expRetVal)
		if tcase.postP == nil {
			continue
		}
		res.Body = ioutil.NopCloser(bytes.NewBuffer(resBytes))
		tcase.postP(tcase, res, &jsonBytesCopy)
	}
	log.SetLevel(oldLevel)
}

func (aut *apiUnitTest) apiImportValidator(tcase apiUnitTestCase, rsp *http.Response, inputBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	assert.Equal(aut.t, strings.Contains(tcase.endpoint, "importAll"), true)
	bodyMap := map[string][]string{}
	err := json.Unmarshal(rspBody, &bodyMap)
	assert.NilError(aut.t, err)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	imported, err := strconv.Atoi(kvMap["imported"][0])
	assert.NilError(aut.t, err)

	not_imported, err := strconv.Atoi(kvMap["not_imported"][0])
	assert.NilError(aut.t, err)
	assert.Equal(aut.t, len(bodyMap["IMPORTED"]), imported)
	assert.Equal(aut.t, len(bodyMap["NOT_IMPORTED"]), not_imported)
}

func (aut *apiUnitTest) apiNameMapValidator(tcase apiUnitTestCase, rsp *http.Response, inputBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	bodyMap := map[string]string{}
	err := json.Unmarshal(rspBody, &bodyMap)
	assert.NilError(aut.t, err)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	aut.assertFetched(kvMap, len(bodyMap))
	aut.saveFetchedCntIn(kvMap, len(bodyMap))
}

func (aut *apiUnitTest) ErrorValidator(tcase apiUnitTestCase, genRsp *http.Response, reqBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(genRsp.Body)
	var xconfError *common.XconfError
	err := json.Unmarshal(rspBody, &xconfError)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling xconf error"))
	}

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)
	entry, ok := kvMap["error_message"]
	if ok {
		assert.Equal(aut.t, xconfError.Message, entry[0])
	}
}

func (aut *apiUnitTest) apiNameListValidator(tcase apiUnitTestCase, rsp *http.Response, inputBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	bodyMap := []string{}
	err := json.Unmarshal(rspBody, &bodyMap)
	assert.NilError(aut.t, err)

	kvMap, err := url.ParseQuery(tcase.postTerms)
	assert.NilError(aut.t, err)

	aut.assertFetched(kvMap, len(bodyMap))
	aut.saveFetchedCntIn(kvMap, len(bodyMap))
}

func (aut *apiUnitTest) apiErrorMessageReporter(tcase apiUnitTestCase, rsp *http.Response, inputBody *bytes.Buffer) {
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	errRsp := ""
	err := json.Unmarshal(rspBody, &errRsp)
	if err == nil {
		log.Printf("-------- Api Returned error = %s --------- ", errRsp)
	} else {
		log.Printf("-------- Error in unmarshalling response = %s --------- ", err.Error())

	}
}

func (aut *apiUnitTest) getValOf(id string) string {
	val, ok := aut.savedMap[id]
	if ok {
		return val
	}
	return ""
}

func (aut *apiUnitTest) setValOf(id string, val string) {
	aut.savedMap[id] = val
}

func (aut *apiUnitTest) eval(val string) string {
	for k, v := range aut.savedMap {
		val = strings.Replace(val, k, v, -1)
	}

	evaled, err := ParseNEval(val)
	assert.NilError(aut.t, err)
	return strconv.Itoa(evaled)
}

const (
	MODEL_QUERYAPI         = "/xconfAdminService/queries/models"
	MODEL_UPAPI            = "/xconfAdminService/updates/models"
	MODEL_DELAPI           = "/xconfAdminService/delete/models"
	jsonModelTestDataLocan = "jsondata/model/"
)

func (aut *apiUnitTest) setupModelApi() {
	if aut.getValOf(MODEL_QUERYAPI) == "Done" {
		return
	}
	aut.setValOf(MODEL_QUERYAPI+DATA_LOCN_SUFFIX, jsonModelTestDataLocan)
	aut.setValOf(MODEL_UPAPI+DATA_LOCN_SUFFIX, jsonModelTestDataLocan)
	aut.setValOf(MODEL_DELAPI+DATA_LOCN_SUFFIX, jsonModelTestDataLocan)

	aut.setValOf(MODEL_QUERYAPI, "Done")
}
func ParseNEval(line string) (int, error) {
	exp, err := parser.ParseExpr(line)
	if err != nil {
		return 0, err
	}
	return Eval(exp), nil
}

func Eval(exp ast.Expr) int {
	switch exp := exp.(type) {
	case *ast.BinaryExpr:
		return EvalBinaryExpr(exp)
	case *ast.BasicLit:
		switch exp.Kind {
		case token.INT:
			i, _ := strconv.Atoi(exp.Value)
			return i
		}
	}

	return 0
}

func EvalBinaryExpr(exp *ast.BinaryExpr) int {
	left := Eval(exp.X)
	right := Eval(exp.Y)

	switch exp.Op {
	case token.ADD:
		return left + right
	case token.SUB:
		return left - right
	case token.MUL:
		return left * right
	case token.QUO:
		return left / right
	}

	return 0
}
func (aut *apiUnitTest) saveIdIn(kvMap map[string][]string, idVal string) {
	idName, ok := kvMap["saveIdIn"]
	if ok {
		aut.savedMap[idName[0]] = idVal
	}
}

func (aut *apiUnitTest) saveDescIn(kvMap map[string][]string, descVal string) {
	idName, ok := kvMap["saveDescIn"]
	if ok {
		aut.savedMap[idName[0]] = descVal
	}
}

func (aut *apiUnitTest) saveFetchedCntIn(kvMap map[string][]string, fetchedCnt int) {
	entry, ok := kvMap["saveFetchedCntIn"]
	if ok {
		aut.savedMap[entry[0]] = strconv.Itoa(fetchedCnt)
	}
}

func (aut *apiUnitTest) assertFetched(kvMap map[string][]string, fetchedCnt int) {
	entry, ok := kvMap["fetched"]
	if ok {
		expEntries, _ := strconv.Atoi(entry[0])
		assert.Equal(aut.t, fetchedCnt, expEntries)
	}
}

func (aut *apiUnitTest) assertPriority(kvMap map[string][]string, actPriority int) {
	entry, ok := kvMap["priority"]
	if ok {
		expPriority, _ := strconv.Atoi(entry[0])
		assert.Equal(aut.t, actPriority, expPriority)
	}
}
