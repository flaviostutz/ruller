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

	"github.com/Sirupsen/logrus"
	"github.com/flaviostutz/ruller/ruller"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infof("====Starting Ruller Sample====")

	ruller.Add("test", "rule1", func(input map[string]interface{}) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["opt1"] = "Some tests"
		output["opt2"] = 129.99
		if input["children"] == true {
			child := make(map[string]interface{})
			child["c1"] = "v1"
			child["c2"] = "v2"
			output["children"] = child
		}
		return output, nil
	})

	ruller.Add("test", "rule2", func(input map[string]interface{}) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["opt1"] = "Lots of tests"
		age, ok := input["age"].(float64)
		if !ok {
			return nil, fmt.Errorf("Invalid 'age' detected. age=%s", input["age"])
		}
		if age > 60 {
			output["category"] = "elder"
		} else {
			output["category"] = "young"
		}
		return output, nil
	})
     
    //this will start the rest server
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

## Estimated Country/City based on request IPs

* If you define a geolite2 database using "--geolite2-db", Ruller will use GeoLite to determine City and Country names corresponding to client IP. It will determine the source IP by first looking at the "X-Forwarded-For" header. If not found present, it will use the IP of the direct requestor.
* When Geolite is activated, the following attributes will be placed on input:
   * "\_ip\_country": Country name
   * "\_ip\_city": City name
   * "\_ip\_longitude": Longitude
   * "\_ip\_latitude: Latitude
   * "\_ip\_accuracy_radius: Accuracy radius

## More resources

* http://github.com/flaviostutz/ruller-sample-dsl - an example on how to build a DSL tool to generate Go Ruller code from an specific rule domain and build a Docker container with the REST api for your compiled rules

## Thanks, Maxmind!

This product includes GeoLite2 data created by MaxMind, available from
<a href="https://www.maxmind.com">https://www.maxmind.com</a>.