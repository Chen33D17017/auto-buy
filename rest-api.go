package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type SecretKeeper struct {
	ApiKey    string `json:"key"`
	ApiSecret string `json:"secret"`
}

type AssitRespnonse struct {
	Msg  int8 `json:"success"`
	Data struct {
		Assets []Asset
	} `json:"data"`
}

type Asset struct {
	Asset           string `json:"asset"`
	AmountPrecision int    `json:"amount_precision"`
	OnhandAmount    string `json:"onhand_amount"`
	FreeAmount      string `json:"free_amount"`
}

type TradesResponse struct {
	Msg  int8 `json:"success"`
	Data struct {
		Trades []Trade `json:"trades"`
	} `json:"data"`
}

type Trade struct {
	TradeId        int64  `json:"trade_id"`
	Pair           string `json:"pair"`
	OrderId        int64  `json:"order_id"`
	Side           string `json:"side"`
	Type           string `json:"type"`
	Amount         string `json:"amount"`
	Price          string `json:"price"`
	MakerTaker     string `json:"maker_taker"`
	FeeAmountBase  string `json:"fee_amount_base"`
	FeeAmountQuote string `json:"fee_amount_quote"`
	ExecutedAt     int64  `json:"executed_at"`
}

type TransactionReq struct {
	Pair   string `json:"pair"`
	Amount string `json:"amount"`
	Side   string `json:"side"`
	Type   string `json:"type"`
}

type TransactionRes struct {
	Msg int8           `json:"success"`
	Rst TransactionRst `json:"data"`
}

type TransactionRst struct {
	OrderId         int64  `json:"order_id"`
	Pair            string `json:"pair"`
	Type            string `json:"type"`
	StartAmount     string `json:"start_amount"`
	RemainingAmount string `json:"remaining_amount"`
	ExecutedAmount  string `json:"executed_amount"`
	Price           string `json:"Price"`
	AveragePrice    string `json:"average_price"`
	OrderedAt       int64  `json:"ordered_at"`
	Status          string `json:"status"`
}

func (tr TransactionRst) String() string {
	return fmt.Sprintf("%v - Buy Pair %s with amount %s @%v", tr.OrderId, tr.Pair, tr.ExecutedAmount, tr.OrderedAt)
}

func (asset Asset) String() string {
	return fmt.Sprintf("Free amount on %s: %s", asset.Asset, asset.FreeAmount)
}

func (tr Trade) String() string {
	return fmt.Sprintf("%v: %s %s with price %s amount %s", readUTC(tr.ExecutedAt), tr.Side, tr.Pair, tr.Price, tr.Amount)
}

func (tr Trade) CSV() []string {
	var weights string
	if tr.Side == "buy" {
		weights = "1"
	} else {
		weights = "-1"
	}
	return []string{readUTC(tr.ExecutedAt), weights, tr.Price, tr.Amount}
}

func (sk SecretKeeper) encode(content string) string {
	h := hmac.New(sha256.New, []byte(sk.ApiSecret))
	h.Write([]byte(content))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}

func readUTC(timestamp int64) string {
	return time.Unix(timestamp/1000, 0).Format("2006-01-02")
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func addHeader(req *http.Request, sk SecretKeeper, content string) {
	nonce := fmt.Sprint(makeTimestamp())
	req.Header.Add("ACCESS-KEY", sk.ApiKey)
	req.Header.Add("ACCESS-NONCE", nonce)
	req.Header.Add("ACCESS-SIGNATURE", sk.encode(nonce+content))
	req.Header.Add("Content-Type", "application/json")
}

func apiRequest(req *http.Request, response interface{}) error {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("apiRequest err %s", err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("apiRequest err: %s", err.Error())
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("apiRequest err %s", err.Error())
	}

	return nil
}

func (sk SecretKeeper) getRequest(query string, response interface{}) error {
	url := fmt.Sprintf("https://api.bitbank.cc%s", query)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("getRequest err: %s", err.Error())
	}

	addHeader(req, sk, query)
	err = apiRequest(req, response)
	if err != nil {
		return fmt.Errorf("getRequest err: %s", err.Error())
	}
	return nil
}

func (sk SecretKeeper) postRequest(endpoint string, payload []byte, response interface{}) error {
	url := fmt.Sprintf("https://api.bitbank.cc%s", endpoint)
	payloadReader := bytes.NewReader(payload)
	req, err := http.NewRequest("POST", url, payloadReader)
	if err != nil {
		return fmt.Errorf("postRequest err: %s", err)
	}

	addHeader(req, sk, string(payload))
	err = apiRequest(req, response)
	if err != nil {
		return fmt.Errorf("postRequest err: %s", err)
	}
	return nil
}

func checkAssets(sk SecretKeeper) ([]Asset, error) {
	var response AssitRespnonse
	err := sk.getRequest("/v1/user/assets", &response)
	if err != nil {
		return nil, fmt.Errorf("checkAssets err: %s", err.Error())
	}
	return response.Data.Assets, nil
}

func buyAssetFromJYP(sk SecretKeeper, assetType string, price float64) (TransactionRst, error) {
	var response TransactionRes
	url := fmt.Sprintf("/v1/user/spot/order")
	cryptmsg, err := getPrice(assetType)
	if err != nil {
		fmt.Println(err.Error())
	}
	cryptPrice, _ := strconv.Atoi(cryptmsg.Data.Buy)
	amount := float64(price / float64(cryptPrice))
	msgJson, _ := json.Marshal(TransactionReq{fmt.Sprintf("%s_jpy", assetType), fmt.Sprintf("%.8f", amount), "buy", "market"})
	err = sk.postRequest(url, msgJson, &response)
	if err != nil {
		return TransactionRst{}, fmt.Errorf("buyAssetFromJYP err : %s", err.Error())
	}

	return response.Rst, nil
}

func getTradeHistory(sk SecretKeeper, assetType string) ([]Trade, error) {
	var response TradesResponse
	url := fmt.Sprintf("/v1/user/spot/trade_history?pair=%s_jpy", assetType)
	err := sk.getRequest(url, &response)
	if err != nil {
		return nil, fmt.Errorf("getTradeHistory err: %s", err.Error())
	}
	return response.Data.Trades, nil
}
