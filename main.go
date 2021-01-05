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
	Type   string `json:"asset"`
	Amount int32  `json:"amount"`
}

func main() {

	var secretKeeper SecretKeeper
	err := readJsonFile("secret.json", &secretKeeper)
	checkError("read secret file fail %s", err)

	var targetConfig TargetConfig
	err = readJsonFile("target.json", &targetConfig)
	checkError("read config file fail %s", err)

	c := cron.New()

	c.AddFunc("CRON_TZ=Asia/Tokyo 00 08 * * *", func() {
		for _, target := range targetConfig.Targets {
			rst, err := buyAssetFromJYP(secretKeeper, target.Type, float64(target.Amount))
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(rst)
			}
		}
	})
	c.Start()

	for {
	}
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func readJsonFile(filename string, rst interface{}) error {
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
