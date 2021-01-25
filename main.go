package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

type DiscordContent struct {
	Content string `json:"content"`
}

func main() {
	var keyReader KeyReader
	err := readJSONFile("secret.json", &keyReader)
	checkError("read secret file fail %s", err)

	var targetConfig TargetConfig
	err = readJSONFile("target.json", &targetConfig)
	checkError("read config file fail %s", err)

	c := cron.New()

	for _, secretKeeper := range keyReader.SecretKeepers {
		secretKeeper := secretKeeper
		for _, target := range targetConfig.Targets {
			target := target
			c.AddFunc(target.TimeRule, func() {
				_, err := buyAssetFromJYP(secretKeeper, target.Type, float64(target.Amount))
				if err != nil {
					infoDiscord(fmt.Sprintf("err %s on buying %s for %s", err, target.Type, secretKeeper.Name))
				} else {
					infoDiscord(fmt.Sprintf("I have brought %v JYP of %s for %s", target.Amount, target.Type, secretKeeper.Name))
				}
			})
		}
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

func infoDiscord(msg string) error {
	url := "https://discordapp.com/api/webhooks/803232063594823740/2ieDrkArEEwIAMSfT-YhNn9IMdlmuhCvy3o656aGlL8wrVFQpmA0DjcYvqBYxIBqUVJl"
	method := "POST"

	msgJSON, _ := json.Marshal(DiscordContent{msg})
	payload := bytes.NewReader(msgJSON)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")
	client.Do(req)
	return nil
}
