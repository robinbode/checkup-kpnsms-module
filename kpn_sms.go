package checkup

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// KPNsms Slack consist of all the sub components required to use KPN SMS Api
type KPNsms struct {
	AppConsumerKey       string `json:"APP_CONSUMER_KEY"`
	AppConsumerSecret    string `json:"APP_CONSUMER_SECRET"`
	Sender               string `json:"SENDER"`
	Phonenumberslocation string `json:"PHONENUMBERSLOCATION"`
}

// Notify implements notifier interface
func (s KPNsms) Notify(results []Result) error {
	for _, result := range results {
		if !result.Healthy {
			s.Send(result)
		}
	}
	return nil
}

func getSMStoken(AppConsumerKey string, AppConsumerSecret string) string {
	client := &http.Client{}

	data := url.Values{}
	data.Set("client_id", AppConsumerKey)
	data.Add("client_secret", AppConsumerSecret)

	req, err := http.NewRequest("POST", "https://api-prd.kpn.com/oauth/client_credential/accesstoken?grant_type=client_credentials", bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	if err != nil {
		log.Println(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	f, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	jsondata := make(map[string]interface{})
	json.Unmarshal(f, &jsondata)
	if err != nil {
	}
	bearertoken := jsondata["access_token"].(string)
	return string(bearertoken)
}
func sendSMS(AppToken string, Content string, MobileNumber string, Sender string) string {

	jsonData := string(fmt.Sprintf("{\"messages\":[{\"content\":\"%s\",\"mobile_number\": \"%s\"}],\"sender\":\"%s\"}", Content, MobileNumber, Sender))

	var jsonStr = []byte(jsonData)

	req, err := http.NewRequest("POST", "https://api-prd.kpn.com/messaging/sms-kpn/v1/send", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "BearerToken "+AppToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func readPhonenumberslist(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Send request via KPN SMS API to send SMS notification to mobile phone(s)
func (s KPNsms) Send(result Result) error {
	phonenumbers, err := readPhonenumberslist(s.Phonenumberslocation)
	if err != nil {
		fmt.Println("Could not load phonenumber list")
	}
	for _, phonenumber := range phonenumbers {
		sendSMS(getSMStoken(s.AppConsumerKey, s.AppConsumerSecret), (result.Title + " = " + strings.ToUpper(fmt.Sprint(result.Status()))), phonenumber, s.Sender)
	}
	return nil
}
