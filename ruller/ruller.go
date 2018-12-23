package ruller

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	groupRules                        = make(map[string][]*ruleInfo)
	requiredInputNames                = make(map[string]map[string]bool)
	rulesMap                          = make(map[string]map[string]*ruleInfo)
	requestFilter      RequestFilter  = func(r *http.Request, input map[string]interface{}) error { return nil }
	responseFilter     ResponseFilter = func(w http.ResponseWriter, input map[string]interface{}, output map[string]interface{}, outBytes []byte) (bool, error) {
		return false, nil
	}
	geodb = (*geoip2.Reader)(nil)
)

var rulesProcessingHist = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "ruller_rules_calculation_seconds",
	Help:    "Ruller rules group calculation duration buckets",
	Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
}, []string{
	"group",
	"status",
})

var groupRuleCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "ruller_rules_active_count",
	Help: "Number of active rules in each rule group",
}, []string{
	"group",
})

//Rule Function that defines a rule. The rule accepts a map as input and returns a map as output. The output map maybe nil
type Rule func(Context) (map[string]interface{}, error)

//RequestFilter Function called on every HTTP call before rules processing.
//params: request, input attributes
//returns: error
type RequestFilter func(r *http.Request, input map[string]interface{}) error

//ResponseFilter Function called on every HTTP call after rules processing.
//params: http response writer, input attribute, output attributes.
//returns: bool true if ruller should interrupt renderization and rely on what the filter did, error
type ResponseFilter func(w http.ResponseWriter, input map[string]interface{}, output map[string]interface{}, outBytes []byte) (bool, error)

//Context used as input for rule processing
type Context struct {
	Input          map[string]interface{}
	ChildrenOutput map[string]interface{}
}

//ProcessOptions options for rule process
type ProcessOptions struct {
	//MergeKeepFirst When merging output results from rules, if there is a duplicate key, keep the first or the last result found. applies when using flatten output. defaults to true
	MergeKeepFirst bool
	//AddRuleInfo Add rule info attributes (name etc) to the output tree when not flatten. defaults to false
	AddRuleInfo bool
	//Get all rules's results and merge all outputs into a single flat map. If false, the output will come the same way as the hierarchy of rules. Defaults to true
	FlattenOutput bool
}

type ruleInfo struct {
	name       string
	parentName string
	rule       Rule
	children   []*ruleInfo
}

//SetRequestFilter set the function that will be called at every call
func SetRequestFilter(rf RequestFilter) {
	requestFilter = rf
}

//SetResponseFilter set the function that will be called at every call with output. If returns true, won't perform the default JSON renderization
func SetResponseFilter(rf ResponseFilter) {
	responseFilter = rf
}

//AddRequiredInput adds a input attribute name that is required before processing the rules
func AddRequiredInput(groupName string, inputName string) {
	logrus.Debugf("Adding required input. group=%s. attribute=%s", groupName, inputName)
	rgi, exists := requiredInputNames[groupName]
	if !exists {
		rgi = make(map[string]bool)
		requiredInputNames[groupName] = rgi
	}
	requiredInputNames[groupName][inputName] = true
}

//Add adds a rule implementation to a group
func Add(groupName string, ruleName string, rule Rule) error {
	return AddChild(groupName, ruleName, "", rule)
}

//AddChild adds a rule implementation to a group
func AddChild(groupName string, ruleName string, parentRuleName string, rule Rule) error {
	logrus.Debugf("Adding rule '%s' '%v' to group '%s'. parent=%s", ruleName, rule, groupName, parentRuleName)
	if _, exists := rulesMap[groupName]; !exists {
		rulesMap[groupName] = make(map[string]*ruleInfo)
	}
	if _, exists := rulesMap[groupName][ruleName]; exists {
		logrus.Warnf("Rule '%s' already exists in group '%s'", ruleName, groupName)
		return fmt.Errorf("Rule '%s' already exists in group '%s'", ruleName, groupName)
	}

	rulei := ruleInfo{
		name:       ruleName,
		parentName: parentRuleName,
		rule:       rule,
		children:   make([]*ruleInfo, 0),
	}
	rulesMap[groupName][ruleName] = &rulei

	if parentRuleName == "" {
		logrus.Debugf("Rule %s is a root rule", ruleName)
		groupRules[groupName] = append(groupRules[groupName], &rulei)

	} else {
		logrus.Debugf("Adding child rule '%s' to parent", ruleName)
		parentRule, exists := rulesMap[groupName][parentRuleName]
		if !exists {
			return fmt.Errorf("Parent rule '%s' not found", parentRuleName)
		}
		logrus.Debugf("Parent of %v is %v", rule, parentRule.rule)
		parentRule.children = append(parentRule.children, &rulei)
	}
	groupRuleCount.WithLabelValues(groupName).Inc()
	return nil
}

//Process process all rules in a group and return a resulting map with all values returned by the rules
func Process(groupName string, input map[string]interface{}, options ProcessOptions) (map[string]interface{}, error) {
	logrus.Debugf(">>>Processing rules from group '%s' with input map %s", groupName, input)

	logrus.Debugf("Validating required input attributes")
	missingInput := ""
	for k := range requiredInputNames[groupName] {
		_, exists := input[k]
		if !exists {
			missingInput = missingInput + " " + k
		}
	}
	if missingInput != "" {
		return nil, fmt.Errorf("Missing required input attributes: %s", missingInput)
	}

	rules, exists := groupRules[groupName]
	if !exists {
		return nil, fmt.Errorf("Group %s doesn't exist", groupName)
	}
	logrus.Debugf("Invoking all rules from group %s", groupName)
	start := time.Now()
	result, err := processRules(rules, input, options)
	status := "2xx"
	if err != nil {
		status = "5xx"
	}
	rulesProcessingHist.WithLabelValues(groupName, status).Observe(time.Since(start).Seconds())
	return result, err
}

func processRules(rules []*ruleInfo, input map[string]interface{}, options ProcessOptions) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	for _, rinfo := range rules {
		childrenOutput := make(map[string]interface{})
		if len(rinfo.children) > 0 {
			logrus.Debugf("Rule '%s': processing %d children rules before itself", rinfo.name, len(rinfo.children))
			co, err := processRules(rinfo.children, input, options)
			if err != nil {
				return nil, err
			}
			childrenOutput = co
		} else {
			logrus.Debugf("No children found for %v", rinfo)
		}

		rule := rinfo.rule
		logrus.Debugf("Invoking rule '%s' '%v'", rinfo.name, rule)
		ctx := Context{Input: input, ChildrenOutput: childrenOutput}
		routput, err := rule(ctx)
		if err != nil {
			return nil, fmt.Errorf("Error processing rule %s. err=%s", rinfo.name, err)
		}
		if routput == nil {
			routput = make(map[string]interface{})
		}

		if len(routput) == 0 {
			logrus.Debugf("Rule '%s' has no output", rinfo.name)
			continue
		}

		if options.AddRuleInfo && options.FlattenOutput {
			routput["_rule"] = rinfo.name
		}

		for k, v := range childrenOutput {
			routput[k] = v
		}

		mergeMaps(rinfo, routput, &output, options)
	}
	return output, nil
}

func mergeMaps(rinfo *ruleInfo, sourceMap map[string]interface{}, destMapP *map[string]interface{}, options ProcessOptions) {
	destMap := *destMapP
	logrus.Debugf("Merging map %v to %v", sourceMap, destMap)
	if len(sourceMap) > 0 {
		if options.FlattenOutput {
			logrus.Debugf("Merge results (flatten)")
			for k, v := range sourceMap {
				_, exists := destMap[k]
				if exists {
					if options.MergeKeepFirst {
						logrus.Debugf("Skipping key '%s' because it already exists in output", k)
					} else {
						destMap[k] = v
						logrus.Debugf("Replacing existing key '%s' in output", k)
					}
				} else {
					destMap[k] = v
				}
			}
		} else {
			logrus.Debugf("Appending rule %s output to parent %s", rinfo.name, rinfo.parentName)
			rmap, exists := destMap["_items"].([]map[string]interface{})
			if !exists {
				rmap = make([]map[string]interface{}, 0)
			}
			rmap = append(rmap, sourceMap)
			destMap["_items"] = rmap
		}
	}
}

//StartServer Initialize and start REST server
func StartServer() error {
	listenPort := flag.Int("listen-port", 3000, "REST API server listen port")
	listenIP := flag.String("listen-address", "0.0.0.0", "REST API server listen ip address")
	golitedb := flag.String("geolite2-db", "", "Geolite mmdb database file. If not defined, localization info based on IP will be disabled")
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

	prometheus.MustRegister(rulesProcessingHist)
	prometheus.MustRegister(groupRuleCount)

	gf := *golitedb
	if gf == "" {
		logrus.Debugf("Geolite database file not found. Localization capabilities based on IP will be disabled")
	} else {
		logrus.Debugf("Loading GeoIP2 database")
		gdb, err := geoip2.Open(gf)
		if err != nil {
			return err
		}
		geodb = gdb
		defer geodb.Close()
		logrus.Infof("GeoIP2 database loaded")
	}

	router := mux.NewRouter()
	router.HandleFunc("/rules/{groupName}", handleRuleGroup).Methods("POST")
	router.Handle("/metrics", promhttp.Handler())
	listen := fmt.Sprintf("%s:%d", *listenIP, *listenPort)
	logrus.Infof("Listening at %s", listen)
	err := http.ListenAndServe(listen, router)
	if err != nil {
		return err
	}
	return nil
}

func handleRuleGroup(w http.ResponseWriter, r *http.Request) {
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

	ipStr := r.Header.Get("X-Forwarded-For")
	if ipStr == "" {
		ra := strings.Split(r.RemoteAddr, ":")
		if len(ra) > 0 {
			ipStr = ra[0]
		}
	}
	if ipStr == "" {
		ipStr = "0.0.0.0"
	}
	pinput["_remote_ip"] = ipStr
	if geodb != nil {
		pinput["_remote_ip"] = ipStr
		ip := net.ParseIP(ipStr)
		start := time.Now()
		ipRecord, err := geodb.City(ip)
		logrus.Debugf("Time to find getIp data: %s", time.Since(start))
		if err != nil {
			logrus.Warnf("Couldn't find geo info for ip %s. err=%s", ipStr, err)
		} else {
			pinput["_ip_country"] = ipRecord.Country.Names["en"]
			pinput["_ip_city"] = ipRecord.City.Names["en"]
			pinput["_ip_latitude"] = ipRecord.Location.Latitude
			pinput["_ip_longitude"] = ipRecord.Location.Longitude
			pinput["_ip_accuracy_radius"] = ipRecord.Location.AccuracyRadius
		}
	}

	logrus.Debugf("input=%s", pinput)

	keepFirst, err := getBool(pinput, "_keepFirst", true)
	if err != nil {
		logrus.Warnf(err.Error())
		http.Error(w, "Error processing rules", 500)
		return
	}

	flatten, err := getBool(pinput, "_flatten", true)
	if err != nil {
		logrus.Warnf(err.Error())
		http.Error(w, "Error processing rules", 500)
		return
	}

	info, err := getBool(pinput, "_info", true)
	if err != nil {
		logrus.Warnf(err.Error())
		http.Error(w, "Error processing rules", 500)
		return
	}

	logrus.Debugf("Calling request filter")
	err = requestFilter(r, pinput)
	if err != nil {
		logrus.Warnf("Error processing rules. err=%s", err)
		http.Error(w, "Error processing rules", 500)
	}

	poutput, err := Process(groupName, pinput, ProcessOptions{MergeKeepFirst: keepFirst, FlattenOutput: flatten, AddRuleInfo: info})
	if err != nil {
		logrus.Warnf("Error processing rules. err=%s", err)
		http.Error(w, fmt.Sprintf("Error processing rules: %s", err), 500)
		return
	}

	logrus.Debugf("Parsing output map to json. output=%s", poutput)
	w.Header().Set("Content-Type", "application/json")
	outBytes, err := json.Marshal(poutput)

	logrus.Debugf("Calling response filter")
	interrupt, err1 := responseFilter(w, pinput, poutput, outBytes)
	if err1 != nil {
		logrus.Warnf("Error processing rules. err=%s", err1)
		http.Error(w, "Error processing rules", 500)
	}
	if interrupt {
		return
	}

	_, err1 = w.Write(outBytes)
	if err1 != nil {
		logrus.Warnf("Error writing response. err=%s", err1)
		http.Error(w, "Error writing response", 500)
		return
	}
}

func getBool(vmap map[string]interface{}, vkey string, defaultValue bool) (bool, error) {
	valueOpt, exists1 := vmap[vkey]
	value := defaultValue
	if exists1 {
		switch valueOpt.(type) {
		case bool:
			value = valueOpt.(bool)
		default:
			return false, fmt.Errorf("'%s' must be a boolean value", vkey)
		}
	}
	return value, nil
}
