package ruller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"github.com/sirupsen/logrus"
	"net/http/httptest"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestRullerFunction(t *testing.T) {
	t.Parallel()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infof("====Starting Ruller Sample====")

	AddRequiredInput("test", "age", Float64)
	AddRequiredInput("test", "children", Bool)

	err := Add("test", "rule1", func(ctx Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["opt1"] = "Some tests rule 1"
		output["rule1-opt2"] = 129.99
		rnd := fmt.Sprintf("v%d", rand.Int())
		if ctx.Input["menu"] == true {
			child := make(map[string]interface{})
			child["rule1-c1"] = "123"
			child["rule1-c2"] = rnd
			output["menu"] = child
		}
		output["rule1"] = true
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	SetRequestFilter(func(r *http.Request, input map[string]interface{}) error {
		logrus.Debugf("filtering request. input=%s", input)
		input["_something"] = "test"
		return nil
	})

	SetResponseFilter(func(w http.ResponseWriter, input map[string]interface{}, output map[string]interface{}, outBytes []byte) (bool, error) {
		logrus.Debugf("filtering response. input=%s", input)
		output["filter-attribute"] = "added by sample filter"
		if input["_something"] == "test" {
			w.Write([]byte("{\"a\":"))
			w.Write(outBytes)
			w.Write([]byte("}"))
		}
		return true, nil
	})

	data := make(map[string]interface{})
	data["age"] = 22
	data["children"] = false
	requestBody, err := json.Marshal(data)
	if err != nil {
		logrus.Fatalln(err)
	}
	r, _ := http.NewRequest("POST", "/rules/test", bytes.NewBuffer(requestBody))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	//Hack to try to fake gorilla/mux vars
	vars := map[string]string{
			"groupName": "test",
	}

	// CHANGE THIS LINE!!!
	r = mux.SetURLVars(r, vars)

	HandleRuleGroup(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []byte(`{"a":{"_items":[{"opt1":"Some tests rule 1","rule1":true,"rule1-opt2":129.99}]}}`), w.Body.Bytes())
}