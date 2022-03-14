package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

type ListChannels struct {
	AllChannels []struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"channels"`
}

type ListMessages struct {
	BulkMessages []struct {
		Type      string `json:"type"`
		Text      string `json:"text"`
		MessageTS string `json:"ts"`
		Files     []struct {
			Name                 string `json:"name"`
			Timestamp            int    `json:"timestamp"`
			Url_Private          string `json:"url_private"`
			Url_Private_Download string `json:"url_private_download"`
		}
	} `json:"messages"`
	More    bool `json:"has_more"`
	ResMeta struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

type FinalMessages struct {
	Type      string    `json:"type"`
	Text      string    `json:"text"`
	MessageTS time.Time `json:"ts"`
	Files     []struct {
		Name                 string `json:"name"`
		Timestamp            int    `json:"timestamp"`
		Url_Private          string `json:"url_private"`
		Url_Private_Download string `json:"url_private_download"`
	}
}

type Messages struct {
	MessageList []FinalMessages
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

func getPosts(auth string, channelID string) Messages {
	base_url := "https://slack.com/api/conversations.history?limit=1000&channel=" + channelID

	bearer := "Bearer " + auth

	Url := base_url

	// Create a new container with all the messages that will receive everything
	var FinalMessageList Messages
	times := 1

	for {

		fmt.Printf("Running through Page %v\n", times)
		req, err := http.NewRequest("GET", Url, nil)
		if err != nil {
			log.Println("Error on building Request: \n[ERROR] - ", err)
		}
		req.Header.Add("Authorization", bearer)

		client := &http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error on Response: %s\v", err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error while reading the response body: %v\n", err)
		}

		//fmt.Printf("Body: %v\n", string(body))

		var TempList ListMessages
		err = json.Unmarshal(body, &TempList)
		if err != nil {
			fmt.Printf("Failed to unmarshal: %v\n", err)
		}

		for _, i := range TempList.BulkMessages {
			Float, err := strconv.ParseFloat(i.MessageTS, 64)
			if err != nil {
				log.Printf("Error Parsing TimeStamp: %v\n", err)
			}

			timeInt := int64(Float)
			ts := time.Unix(timeInt, 0)

			f := FinalMessages{
				Type:      i.Type,
				Text:      i.Text,
				MessageTS: ts,
				Files:     i.Files,
			}
			FinalMessageList.MessageList = append(FinalMessageList.MessageList, f)
		}

		if TempList.More == true {
			Url = base_url + "&cursor=" + TempList.ResMeta.NextCursor
			times++
		}

		if TempList.More == false {
			return FinalMessageList
		}
	}

}

func PrettyJSON(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func FileDownload(url string, filepath string, auth string) error {
	/*
		fileURL, err := url.Parse(url)
		if err != nil {
			log.Printf("URL Parse Error: %v\n")
		}
	*/
	bearer := "Bearer " + auth

	segments := strings.Split(url, "/")
	fileName := segments[len(segments)-1]

	fullFile := filepath + "/" + fileName

	file, err := os.Create(fullFile)
	if err != nil {
		log.Printf("Error Creating File: %v\n", err)
	}
	/*
		client := http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}

		resp, err := client.Get(url)
		if err != nil {
			log.Printf("Error Connecting to URL: %v\n", err)
		}
	*/
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error on building Request: \n[ERROR] - ", err)
	}
	req.Header.Add("Authorization", bearer)

	client := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error on Response: %s\v", err)
	}

	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("Error Creating File: ", err.Error())
	}

	defer file.Close()

	fmt.Printf("Downloading file %s with size %d\n", fileName, size)
	return nil
}

func main() {
	fmt.Printf("Preparing to export channel to file......\n")
	var (
		channel   string
		authtoken string
		dirpath   string
	)

	pflag.StringVarP(&authtoken, "authtoken", "a", "", "Authorization Bearer Token for Slack API")
	pflag.StringVarP(&channel, "channel", "c", "", "Channel that you want to export - String")
	pflag.StringVarP(&dirpath, "dirpath", "p", "", "Directory path for output")
	pflag.Parse()

	//Get a list of channels
	id, err := getChannel(authtoken, channel)
	if err != nil {
		log.Printf("Error in Getting Channel: %v\n", err)
	}

	fmt.Printf("ID: %s\n", id)

	//Use the ID to get dump of all the posts
	FullList := getPosts(authtoken, id)
	fmt.Printf("Full Message List: %v\n", FullList)

	rawfile, err := json.Marshal(FullList)
	if err != nil {
		log.Printf("Error Marshalling: %v\n", err)
	}

	// Let's make the JSON Pretty - Shall We?
	PrettyData, err := PrettyJSON(FullList)
	if err != nil {
		log.Printf("Error Making JSON Pretty: %v\n", err)
	}

	// create full directory path and output to file
	// TODO ErrorHandling
	err = os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Printf("Error Making Directory: %v\n", err)
	}

	rawdir := dirpath + "/" + channel + "_raw_outfile.txt"
	prettydir := dirpath + "/" + channel + "_pretty_outfile.txt"

	// Output the to file
	os.WriteFile(rawdir, rawfile, 0644)
	os.WriteFile(prettydir, []byte(PrettyData), 0644)

	// Download images and add date and caption information
	for _, i := range FullList.MessageList {
		for _, f := range i.Files {
			FileDownload(f.Url_Private, dirpath, authtoken)
		}
	}

}
