package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// var jar, _ = cookiejar.New(nil)
// var client = &http.Client{
// 	Jar:     jar,
// 	Timeout: time.Second * 10,
// 	// Returning an error to prevent redirects
// 	CheckRedirect: func(req *http.Request, via []*http.Request) error {
// 		return http.ErrUseLastResponse
// 	},
// }

func makeGetRequestWithSession(url string) ([]byte, *http.Client) {

	var jar, _ = cookiejar.New(nil)
	var client = &http.Client{
		Jar:     jar,
		Timeout: time.Second * 10,
		// Returning an error to prevent redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	// req.Header.Set("Cookie", "__ddg8_=9KCSctc4iOruShN0; __ddgid_=7lWyc52yRS7YgOpW; __ddgmark_=wfZpnxacF2nXdVTE; __ddg2_=qtEE5nKN3PCJ7c2Z; __ddg1_=5Qh2v4L5z7LpVnQx; __ddg3_=fHwWwZbYpI3fHcQx")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, client
}

func makePostRequestWithSession(url string, headers map[string]string, payload io.Reader, client *http.Client) string {

	req, _ := http.NewRequest("POST", url, payload)
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	download_url := resp.Header.Get("Location")

	defer resp.Body.Close()
	// fmt.Println("Response Status: ", resp.StatusCode)
	// fmt.Println("Response Headers Location: ", download_url)

	return download_url
}

func Parse_html_token_getDownloadLink(html_body string, orig_link string, httpClient *http.Client) string {

	pattern := `\("(\S+)",\d+,"(\S+)",(\d+),(\d+)`
	regex := regexp.MustCompile(pattern)

	matches := regex.FindAllStringSubmatch(html_body, -1)
	// fmt.Println("Matches : ", matches)
	match := matches[0]
	// for _, m := range match {
	// 	fmt.Println("Match", m)
	// }
	data, key := match[1], match[2]
	load, seperator := match[3], match[4]

	target_url, token := step_1(data, key, load, seperator)
	fmt.Println("Token : ", token, " Url : ", target_url)

	// payload := "_token=zYCTmeU3hNE84ly3aroK9Yjv4qRepX6o7CsJy0Is"
	payload := strings.NewReader("_token=" + token)

	headers := make(map[string]string)
	headers["Referer"] = orig_link
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	// fmt.Println("Headers:", headers, "body :", payload, "url:", target_url)
	// loc, _ := makeRequestWithHeaders(target_url, headers, payload)
	loc := makePostRequestWithSession(target_url, headers, payload, httpClient)
	return loc
}

// Helper function to reverse a string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func step_2(s string, seperator int, base int) string {
	mappedRange := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/"
	numbers := mappedRange[:base]
	maxIter := 0
	for index, value := range reverseString(s) {
		// Only consider digits for calculation
		if value >= '0' && value <= '9' {
			// Convert character to int (assuming it's a digit)
			digitValue, _ := strconv.Atoi(string(value))
			maxIter += digitValue * int(math.Pow(float64(seperator), float64(index)))
		}
	}

	// fmt.Println("max_iter:", maxIter)
	mid := ""
	for maxIter > 0 {
		mid = string(numbers[maxIter%base]) + mid
		maxIter = maxIter / base
	}

	// fmt.Println("mid:", mid)
	if mid == "" {
		return "0"
	}
	return mid
}

func step_1(data, key, load, seperator string) (string, string) {

	payload := ""
	i := 0
	seperator_int, err := strconv.Atoi(seperator)
	if err != nil {
		fmt.Println("Error converting seperator to int", err)
	}
	load_int, err := strconv.Atoi(load)
	if err != nil {
		fmt.Println("Error converting load to int", err)
	}

	for i < len(data) {

		s := ""
		for i < len(data) && data[i] != key[seperator_int] {
			s += string(data[i])
			i++
		}
		for index, value := range key {
			s = strings.ReplaceAll(s, string(value), strconv.Itoa(index))
		}

		val, _ := strconv.Atoi(step_2(s, seperator_int, 10))
		payload += string(val - load_int)
		i++

	}

	// fmt.Println("91 Payload : ", payload)
	pattern := `action="([^\"]+)" method="POST"><input type="hidden" name="_token"\s+value="([^\"]+)`
	regex := regexp.MustCompile(pattern)
	payload_matches := regex.FindAllStringSubmatch(payload, -1)

	// fmt.Println("Payload_matches : ", payload_matches)
	return payload_matches[0][1], payload_matches[0][2]
}

// GetDownloadLinksFromRedirectedLinks takes a slice of Episode structs, makes a GET request to each
// of the URLs, and extracts the download link from each of the HTML responses.
// The function returns a slice of Episode structs, each of which contains the
// download link for the corresponding episode.
// func GetDownloadLinksFromRedirectedLinks(kwik_episodes []Episode) []Episode {

// 	returned_episodes := []Episode{}

// 	var wg sync.WaitGroup
// 	response_channel := make(chan string, 8)

// 	for _, kwik_ep := range kwik_episodes {

// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			// fmt.Println("link :" , kwik_ep.Url)
// 			body, err := makeGetRequestWithSession(kwik_ep.Url)
// 			if err != nil {
// 				fmt.Println("Error getting redirected link")
// 			}

// 			html_body := string(body)
// 			response_channel <- html_body
// 		}()

// 	}

// 	go func() {
// 		wg.Wait()
// 		close(response_channel)
// 	}()

// 	for _, kwik_ep := range kwik_episodes {

// 		html_body := <-response_channel
// 		// fmt.Println("Redirected link html : ", html_body)
// 		download_link := Parse_html_token_getDownloadLink(html_body, kwik_ep.Url)
// 		fmt.Println("Download link: ep:", kwik_ep.Number, " : ", download_link)
// 		// create Episode struct
// 		episode := Episode{
// 			Url:      download_link,
// 			Number:   kwik_ep.Number,
// 			Quality:  kwik_ep.Quality,
// 			FileSize: kwik_ep.FileSize,
// 		}
// 		returned_episodes = append(returned_episodes, episode)
// 	}

// 	return returned_episodes

// }
