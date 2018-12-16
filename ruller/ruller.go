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
	ruleGroups = make(map[string]map[string]ruleInfo)
	rules      = make(map[string]*ruleInfo)
)

//Rule Function that defines a rule. The rule accepts a map as input and returns a map as output. The output map maybe nil
type Rule func(Context) (map[string]interface{}, error)

//Context used as input for rule processing
type Context struct {
	Input          map[string]interface{}
	ChildrenOutput map[string]interface{}
}

//ProcessOptions options for rule process
type ProcessOptions struct {
	MergeKeepFirst bool
}

type ruleInfo struct {
	rule     Rule
	children map[string]ruleInfo
}

//Add adds a rule implementation to a group
func Add(groupName string, ruleName string, rule Rule) error {
	return AddChild(groupName, ruleName, "", rule)
}

//AddChild adds a rule implementation to a group
func AddChild(groupName string, ruleName string, parentRuleName string, rule Rule) error {
	logrus.Debugf("Adding rule '%s' '%v' to group '%s'. parent=%s", ruleName, rule, groupName, parentRuleName)
	if _, exists := ruleGroups[groupName]; !exists {
		ruleGroups[groupName] = make(map[string]ruleInfo)
	}
	if _, exists := ruleGroups[groupName][ruleName]; exists {
		logrus.Warnf("Rule '%s' already exists in group '%s'. Skipping Add", ruleName, groupName)
		return fmt.Errorf("Rule '%s' already exists in group '%s'", ruleName, groupName)
	}

	rulei := ruleInfo{
		rule:     rule,
		children: make(map[string]ruleInfo, 0),
	}
	rules[ruleName] = &rulei

	if parentRuleName == "" {
		logrus.Debugf("Found root rule %s", ruleName)
		ruleGroups[groupName][ruleName] = rulei

	} else {
		logrus.Debugf("Adding child rule '%s' to parent", ruleName)
		parentRule, exists := rules[parentRuleName]
		if !exists {
			return fmt.Errorf("parent rule '%s' not found", parentRuleName)
		}
		logrus.Debugf("Parent of %v is %v", rule, parentRule.rule)
		pr := *parentRule
		pr.children[ruleName] = rulei
	}
	return nil
}

//Process process all rules in a group and return a resulting map with all values returned by the rules
func Process(groupName string, input map[string]interface{}, options ProcessOptions) (map[string]interface{}, error) {
	logrus.Debugf(">>>Processing rules from group '%s' with input map %s", groupName, input)
	ruleGroup, exists := ruleGroups[groupName]
	if !exists {
		return nil, fmt.Errorf("Group %s doesn't exist", groupName)
	}
	logrus.Debugf("Invoking all rules from group %s", groupName)
	return processRules(ruleGroup, input, options)
}

func processRules(rules map[string]ruleInfo, input map[string]interface{}, options ProcessOptions) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	for k, ruleInfo := range rules {
		coutput := make(map[string]interface{})
		if len(ruleInfo.children) > 0 {
			logrus.Debugf("Rule '%s': processing %d children rules before itself", k, len(ruleInfo.children))
			coutput2, err := processRules(ruleInfo.children, input, options)
			if err != nil {
				return nil, err
			}
			coutput = coutput2
		}

		rule := ruleInfo.rule
		logrus.Debugf("Invoking rule '%s' '%v'", k, rule)
		ctx := Context{Input: input, ChildrenOutput: coutput}
		routput, err := rule(ctx)
		if err != nil {
			return nil, fmt.Errorf("Error processing rule %s. err=%s", k, err)
		}
		if routput != nil {
			logrus.Debugf("Output is %v", routput)
			for k, v := range routput {
				_, exists := output[k]
				if exists {
					if options.MergeKeepFirst {
						logrus.Debugf("Skipping key '%s' because it already exists in output", k)
					} else {
						output[k] = v
						logrus.Debugf("Replacing existing key '%s' in output", k)
					}
				} else {
					output[k] = v
				}
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

	mergeKeepFirst, exists := pinput["_mergeKeepFirst"]
	keepFirst := true
	if exists {
		switch mergeKeepFirst.(type) {
		case bool:
			keepFirst = mergeKeepFirst.(bool)
		default:
			logrus.Warnf("Input attribute '_mergeKeepFirst' must be boolean")
			http.Error(w, "Error processing rules", 500)
			return
		}
	}
	poutput, err := Process(groupName, pinput, ProcessOptions{MergeKeepFirst: keepFirst})
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
