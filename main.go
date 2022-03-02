package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/spf13/pflag"
)

type ListChannels struct {
	AllChannels []struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"channels"`
}

func getChannel(authtoken string, channel string) (string, error) {
	var chanId string

	// Connect to slack using auth token
	// Get URL first
	url := "https://slack.com/api/conversations.list?limit=1000"

	bearer := "Bearer " + authtoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error on building Request: \n[ERROR] - ", err)
	}
	req.Header.Add("Authorization", bearer)

	// Create a client and Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error on Response: %s\v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error while reading the response body: %v\n", err)
	}

	fmt.Printf("Body: %v\n", string(body))

	var Channels ListChannels
	err = json.Unmarshal(body, &Channels)
	if err != nil {
		fmt.Printf("Failed to unmarshal: %v\n", err)
	}

	fmt.Printf("Results: %v\n", Channels)

	// If the channel requested equals the one in the list, get the ID
	for _, v := range Channels.AllChannels {
		if v.Name == channel {
			chanId = v.Id
			log.Printf("FOUND\n")
		}
	}

	return chanId, nil
}

func main() {
	fmt.Printf("Preparing to export channel to file......\n")
	var (
		channel   string
		authtoken string
		fpath     string
	)

	pflag.StringVarP(&authtoken, "authtoken", "a", "", "Authorization Bearer Token for Slack API")
	pflag.StringVarP(&channel, "channel", "c", "", "Channel that you want to export - String")
	pflag.StringVarP(&fpath, "filepath", "f", "", "File path for output")
	pflag.Parse()

	//Get a list of channels
	id, err := getChannel(authtoken, channel)
	if err != nil {
		log.Printf("Error in Getting Channel: %v\n", err)
	}

	fmt.Printf("ID: %s\n", id)

	// Error out if the channel isn't found
	// Find ID
	//Use the ID to get dump of all the posts
	//Will probably require a loop to keep asking for pages and append it to existing list
	//iterate through list
	//download images and add date and caption information
}
