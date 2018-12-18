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
		logrus.Debugf("children output from rule 2.1 is %s", ctx.ChildrenOutput)
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
		output["rule2.1"] = true
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	err = ruller.AddChild("test", "rule2.2", "rule2", func(ctx ruller.Context) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["rule2.2-type"] = "any"
		output["rule2.2"] = true
		return output, nil
	})
	if err != nil {
		panic(err)
	}

	for a := 0; a < 100; a++ {
		err = ruller.AddChild("test", fmt.Sprintf("rule2.2-%d", a), "rule2.2", func(ctx ruller.Context) (map[string]interface{}, error) {
			output := make(map[string]interface{})
			output["opt"] = "any"
			return output, nil
		})
		if err != nil {
			panic(err)
		}
		for b := 0; b < 20; b++ {
			err = ruller.AddChild("test", fmt.Sprintf("rule2.2-%d-%d", a, b), fmt.Sprintf("rule2.2-%d", a), func(ctx ruller.Context) (map[string]interface{}, error) {
				output := make(map[string]interface{})
				output["opt"] = "any"
				return output, nil
			})
			if err != nil {
				panic(err)
			}
		}
	}

	ruller.StartServer()
}
