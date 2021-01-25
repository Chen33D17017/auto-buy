package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/robfig/cron/v3"
)

type TargetConfig struct {
	Targets []Target `json:"targets"`
}

type Target struct {
	Type     string `json:"asset"`
	Amount   int32  `json:"amount"`
	TimeRule string `json:"time"`
}

func main() {

	var secretKeeper SecretKeeper
	err := readJSONFile("secret.json", &secretKeeper)
	checkError("read secret file fail %s", err)

	var targetConfig TargetConfig
	err = readJSONFile("target.json", &targetConfig)
	checkError("read config file fail %s", err)

	c := cron.New()

	for _, target := range targetConfig.Targets {
		target := target
		c.AddFunc(target.TimeRule, func() {
			rst, err := buyAssetFromJYP(secretKeeper, target.Type, float64(target.Amount))
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(rst)
			}
		})
	}

	c.Start()

	for {
		c := make(chan int)
		<-c
	}
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func readJSONFile(filename string, rst interface{}) error {
	secretFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("readJsonFile err: %s", err.Error())
	}
	defer secretFile.Close()

	byteValue, err := ioutil.ReadAll(secretFile)
	if err != nil {
		return fmt.Errorf("readJsonFile err: %s", err.Error())
	}

	json.Unmarshal(byteValue, &rst)
	return nil
}
