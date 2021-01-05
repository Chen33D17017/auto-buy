package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CryptMsg struct {
	Msg  int8      `json:"success"`
	Data CryptData `json:"data"`
}

type CryptData struct {
	Sell      string `json:"sell"`
	Buy       string `json:"buy"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Last      string `json:"last"`
	Vol       string `json:"vol"`
	Timestamp int64  `json:"timestamp"`
}

func getPrice(cryp string) (CryptMsg, error) {
	var cryptMsg CryptMsg
	url := fmt.Sprintf("https://public.bitbank.cc/%s_jpy/ticker", cryp)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return cryptMsg, fmt.Errorf("getPrice err: %s", err.Error())
	}
	res, err := client.Do(req)
	if err != nil {
		return cryptMsg, fmt.Errorf("getPrice err: %s", err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return cryptMsg, fmt.Errorf("getPrice err: %s", err.Error())
	}

	err = json.Unmarshal(body, &cryptMsg)
	if err != nil {
		return cryptMsg, fmt.Errorf("getPrice err: %s", err.Error())
	}
	return cryptMsg, nil
}
