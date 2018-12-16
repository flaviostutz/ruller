package main

import (
	"fmt"
	"math/rand"

	"ruller"

	"github.com/Sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infof("====Starting Ruller Sample====")
	err := ruller.Add("test", "rule1", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["opt1"] = "Some tests"
		output["opt2"] = 129.99
		rnd := fmt.Sprintf("v%d", rand.Int())
		if ctx.Input["menu"] == true {
			child := make(map[string]interface{})
			child["c1"] = "123"
			child["c2"] = rnd
			output["children"] = child
		}
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	err = ruller.Add("test", "rule2", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["opt1"] = "Lots of tests"
		logrus.Debugf("children output from rule 2.1 is %s", ctx.ChildrenOutput)
		output["from-child-category"] = ctx.ChildrenOutput["category"]
		output["from-child-type"] = ctx.ChildrenOutput["type"]
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
			output["category"] = "elder"
		} else {
			output["category"] = "young"
		}
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	err = ruller.AddChild("test", "rule2.2", "rule2", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["type"] = "any"
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	ruller.StartServer()
}
