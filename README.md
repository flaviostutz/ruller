# ruller
A simple REST based rule engine in which rules are written in Go

1. You create and register some rules as Go functions in a "main" program and run it
2. You POST to /[group-name] along with some json input.
3. All rules for that group are processed using the input
4. Depending on your implementation, some rules returns data and some rules not
4. Finally, all rule's results are merged and returned to the REST caller as a json

Ruller works by invocating a bunch of rules with the same input from a single REST call, merging rules outputs and returning the result to the caller. 

You can use this for feature enablement, dynamic systems configuration and other applications where a static key value store wouldn't help you because you need some logic on an input value in order to determine the output.

Checkout some [benchmarks](BENCHMARK.md) we made this far too.

## Example

* main.go file

```
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"ruller"

	"github.com/Sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infof("====Starting Ruller Sample====")
	err := ruller.Add("test", "rule1", func(ctx ruller.Context) (map[string]interface{}, error) {
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

	ruller.AddRequiredInput("test", "samplestring", ruller.String)
	ruller.AddRequiredInput("test", "samplefloat", ruller.Numeric)

	err = ruller.AddChild("test", "rule1.1", "rule1", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["rule1.1-mydata"] = "myvalue"
		output["rule1.1"] = true
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	err = ruller.Add("test", "rule2", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["opt1"] = "Lots of tests rule 2"
		output["rule2"] = true
		// logrus.Debugf("children output from rule 2.1 is %s", ctx.ChildrenOutput)
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	err = ruller.AddChild("test", "rule2.1", "rule2", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		age, ok := ctx.Input["age"].(float64)
		if !ok {
			return nil, fmt.Errorf("Invalid 'age' detected. age=%s", ctx.Input["age"])
		}
		if age > 60 {
			output["category"] = "elder rule2.1"
		} else {
			output["category"] = "young rule2.1"
		}
		output["city"] = ctx.Input["_ip_city"]
		output["rule2.1"] = true
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	ruller.SetRequestFilter(func(r *http.Request, input map[string]interface{}) error {
		logrus.Debugf("filtering request. input=%s", input)
		input["_something"] = "test"
		return nil
	})

	ruller.SetResponseFilter(func(w http.ResponseWriter, input map[string]interface{}, output map[string]interface{}, outBytes []byte) (bool, error) {
		logrus.Debugf("filtering response. input=%s", input)
		output["filter-attribute"] = "added by sample filter"
		if input["_something"] == "test" {
			w.Write([]byte("{\"a\":"))
			w.Write(outBytes)
			w.Write([]byte("}"))
		}
		return true, nil
	})

	err = ruller.AddChild("test", "rule2.2", "rule2", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["rule2.2-type"] = "any"
		output["rule2.2"] = true
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	for a := 0; a < 10; a++ {
		err = ruller.AddChild("test", fmt.Sprintf("rule2.2-%d", a), "rule2.2", func(ctx ruller.Context) (map[string]interface{}, error) {
			output := make(map[string]interface{})
			output["opt1"] = "any1"
			return output, nil
		})
		if err != nil {
			panic(err)
		}
		for b := 0; b < 1; b++ {
			err = ruller.AddChild("test", fmt.Sprintf("rule2.2-%d-%d", a, b), fmt.Sprintf("rule2.2-%d", a), func(ctx ruller.Context) (map[string]interface{}, error) {
				output := make(map[string]interface{})
				output["opt2"] = "any2"
				return output, nil
			})
			if err != nil {
				panic(err)
			}
			for c := 0; c < 1; c++ {
				err = ruller.AddChild("test", fmt.Sprintf("rule2.2-%d-%d-%d", a, b, c), fmt.Sprintf("rule2.2-%d-%d", a, b), func(ctx ruller.Context) (map[string]interface{}, error) {
					output := make(map[string]interface{})
					output["opt3"] = "any3"
					return output, nil
				})
				if err != nil {
					panic(err)
				}
			}
		}
	}

	ruller.StartServer()
}

```

* run the sample program
  * if you cloned this repo, just type ```docker-compose up --build```

* execute some tests

```
curl -X POST \
  http://localhost:3000/test \
  -H 'Content-Type: application/json' \
  -d '{
	"age": 22,
	"children": false
}'
```
```
{"category":"young","opt1":"Lots of tests","opt2":129.99}
```

-----
```
curl -X POST \
  http://localhost:3000/test \
  -H 'Content-Type: application/json' \
  -d '{
	"age": 77,
	"children": true
}'
```
```
{"category":"elder","children":{"c1":"v1","c2":"v2"},"opt1":"Lots of tests","opt2":129.99}
```

## Special parameters on POST body

* "_flatten" - true|false. If true, a flat map with all keys returned by all rules, with results merged, will be returned. If false, will return the results with the same tree shape as the rules itself. Defaults to true

* "_keepFirst" - true|false. When using flat map as result, this determines whetever to keep the value from the first or the last rule processed during merge. Default is true

* "_info" - true|false. If true, will add attribute "_rule" with the name of the rule that generated the node on the result tree (if not using flat map as result). Default to true

## Input parameters used as rules input

* The POST body JSON elements will be converted to a map and used as input parameters
* Additionally, "\_remote\_ip" is set with client remote address

* If you define a geolite2 database using "--geolite2-db", Ruller will use GeoLite to determine City and Country names corresponding to client IP. It will determine the source IP by first looking at the "X-Forwarded-For" header. If not present, it will use the IP of the direct requestor.
* When Geolite is activated, the following attributes will be placed on input:
   * "\_ip\_country": Country name
   * "\_ip\_city": City name
   * "\_ip\_longitude": Longitude
   * "\_ip\_latitude: Latitude
   * "\_ip\_accuracy_radius: Accuracy radius

* When you pass a csv file in format "[country iso code],[City],[State]" using "--city-state-db", you will have an additional input:
   * "\_ip\_state: State based on city info

* You can define required inputs along with their associated types so that before processing rules Ruller will perform a basic check if they are present (ruller.AddRequiredInput(..)). This is usedful so that you don't have to perform those verifications inside each rule, as it was already verified before executing the rules.

## Request/Response filtering

* ```ruller.setRequestFilter(func(r *http.Request, input map[string]interface{}) error { return nil })```
   * You can verify http request attributes and change input map as you need

* ```func(w http.ResponseWriter, input map[string]interface{}, output map[string]interface{}, outBytes []byte) (bool, error) {return false, nil}```
   * You can verify the input and output map and write something to the response. If you return true, the default JSON marshal that ruller performs will be skipped.

* See an example at ```ruller-sample/main.go```

## More resources

* http://github.com/flaviostutz/ruller-sample-feature-flag - an example on how to build a DSL tool to generate Go Ruller code from a JSON and to build a Docker container with the REST api for your compiled rules

## Thanks, Maxmind!

This product includes GeoLite2 data created by MaxMind, available from
<a href="https://www.maxmind.com">https://www.maxmind.com</a>.