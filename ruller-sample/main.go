package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/flaviostutz/ruller/ruller"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infof("====Starting Ruller Sample====")
	ruller.Add("test", "rule1", func(input map[string]interface{}) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["opt1"] = "Lots of tests"
		output["opt2"] = 129.99
		if input["param1"] == "value1" {
			child := make(map[string]interface{})
			child["c1"] = "v1"
			child["c2"] = "v2"
			output["w/child"] = child
		}
		return output, nil
	})
	ruller.Add("test", "rule2", func(input map[string]interface{}) (map[string]interface{}, error) {
		output := make(map[string]interface{})
		output["opt1"] = "More tests to come!"
		output["opt3"] = "Maybe"
		return output, nil
	})

	ruller.StartServer()

	// input := make(map[string]interface{})
	// output, err := Process("test", input)
	// if err != nil {
	// 	logrus.Errorf("Error evaluating rules. err=%s", err)
	// 	os.Exit(1)
	// }
	// logrus.Infof("Rules evaluated. output=%s", output)
}
