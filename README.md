# ruller
A simple REST based rule engine in which rules are written in Go

1. You create and register some rules as Go functions in a "main" program and run it
2. You POST to /[group-name] along with some json input.
3. All rules for that group are processed using the input
4. Depending on your implementation, some rules returns data and some rules not
4. Finally, all rule's results are merged and returned to the caller as a json

You can use this for feature enablement, dynamic systems configuration and other applications where a static key value store wouldn't help you because you need some logic on an input value in order to determine the output.

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
