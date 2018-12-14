package ruller

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var (
	ruleGroups = make(map[string]map[string]Rule)
)

//Rule Function that defines a rule. The rule accepts a map as input and returns a map as output. The output map maybe nil
type Rule func(map[string]interface{}) (map[string]interface{}, error)

//Add adds a rule implementation to a group
func Add(groupName string, ruleName string, rule Rule) error {
	logrus.Debugf("Adding rule '%s' '%v' to group '%s'", ruleName, rule, groupName)
	if _, exists := ruleGroups[groupName]; !exists {
		ruleGroups[groupName] = make(map[string]Rule)
	}
	if _, exists := ruleGroups[groupName][ruleName]; exists {
		logrus.Warnf("Rule '%s' already exists in group '%s'. Skipping Add", ruleName, groupName)
		return fmt.Errorf("Rule '%s' already exists in group '%s'", ruleName, groupName)
	}
	ruleGroups[groupName][ruleName] = rule
	return nil
}

//Process process all rules in a group and return a resulting map with all values returned by the rules
func Process(groupName string, input map[string]interface{}) (map[string]interface{}, error) {
	logrus.Debugf("Processing rules from group '%s' with input map %s", groupName, input)
	ruleGroup, exists := ruleGroups[groupName]
	if !exists {
		return nil, fmt.Errorf("Group %s doesn't exist", groupName)
	}
	logrus.Debugf("Invoking all rules from group %s", groupName)
	output := make(map[string]interface{})
	for k, rule := range ruleGroup {
		logrus.Debugf("Processing rule '%s' '%v'", k, rule)
		routput, err := rule(input)
		if err != nil {
			return nil, fmt.Errorf("Error processing rule %s.%s. err=%s", groupName, k, err)
		}
		if routput != nil {
			logrus.Debugf("Output is %v", routput)
			for k, v := range routput {
				if _, exists := output[k]; exists {
					logrus.Debugf("A rule has replaced an existing key (%s) in output", k)
				}
				output[k] = v
			}
		} else {
			logrus.Debugf("Output is nil")
		}
	}
	return output, nil
}

//StartServer Initialize and start REST server
func StartServer() error {
	listenPort := flag.Int("listen-port", 3000, "REST API server listen port")
	listenIP := flag.String("listen-address", "0.0.0.0", "REST API server listen ip address")
	logLevel := flag.String("log-level", "info", "debug, info, warning or error")
	flag.Parse()

	switch *logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	router := mux.NewRouter()
	router.HandleFunc("/{groupName}", processRuleGroup).Methods("POST")
	listen := fmt.Sprintf("%s:%d", *listenIP, *listenPort)
	logrus.Infof("Listening at %s", listen)
	err := http.ListenAndServe(listen, router)
	if err != nil {
		return err
	}
	return nil
}

func processRuleGroup(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("processRuleGroup r=%v", r)
	params := mux.Vars(r)

	groupName := params["groupName"]

	logrus.Debugf("Parsing input json to map")
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Warnf("Error reading request body. err=%s", err)
		http.Error(w, "Error reading request body", 500)
		return
	}

	pinput := make(map[string]interface{})
	if len(bodyBytes) > 0 {
		err = json.Unmarshal(bodyBytes, &pinput)
		if err != nil {
			logrus.Warnf("Error parsing json body to map. err=%s", err)
			http.Error(w, "Invalid input JSON. err="+err.Error(), 500)
			return
		}
	}

	logrus.Debugf("input=%s", pinput)

	poutput, err := Process(groupName, pinput)
	if err != nil {
		logrus.Warnf("Error processing rules. err=%s", err)
		http.Error(w, "Error processing rules", 500)
		return
	}

	logrus.Debugf("Parsing output map to json. output=%s", poutput)
	w.Header().Set("Content-Type", "application/json")
	outBytes, err := json.Marshal(poutput)
	_, err1 := w.Write(outBytes)
	if err1 != nil {
		logrus.Warnf("Error writing response. err=%s", err1)
		http.Error(w, "Error writing response", 500)
		return
	}
}
