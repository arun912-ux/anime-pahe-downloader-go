package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"strings"
)

var host string = "https://animepahe.ru/"

var num_of_cores int = runtime.NumCPU() / 2

var headers map[string]string = map[string]string{
	"Cookie": "__ddg8_=9KCSctc4iOruShN0; __ddgid_=7lWyc52yRS7YgOpW; __ddgmark_=wfZpnxacF2nXdVTE; __ddg2_=qtEE5nKN3PCJ7c2Z; __ddg1_=5Qh2v4L5z7LpVnQx; __ddg3_=fHwWwZbYpI3fHcQx",
}

type PaheResponse struct {
	Total int              `json:"total"`
	Data  []SearchResponse `json:"data"`
}

type SearchResponse struct {
	Title    string  `json:"title"`
	Type_    string  `json:"type"`
	Episodes int     `json:"episodes"`
	Status   string  `json:"status"`
	Year     int     `json:"year"`
	Score    float64 `json:"score"`
	Poster   string  `json:"poster"`
	Id       string  `json:"session"`
}

type AnimeResponse struct {
	Total int               `json:"total"`
	Data  []EpisodeResponse `json:"data"`
}

type EpisodeResponse struct {
	Episode int    `json:"episode"`
	Id      string `json:"session"`
}

type Episode struct {
	Url      string
	Number   int
	Quality  string
	FileSize string
}

func makeRequest(url string) ([]byte, error) {

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Cookie", "__ddg8_=9KCSctc4iOruShN0; __ddgid_=7lWyc52yRS7YgOpW; __ddgmark_=wfZpnxacF2nXdVTE; __ddg2_=qtEE5nKN3PCJ7c2Z; __ddg1_=5Qh2v4L5z7LpVnQx; __ddg3_=fHwWwZbYpI3fHcQx")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return body, nil
}

func SearchAnime(name string) ([]SearchResponse, int) {

	name = strings.ReplaceAll(name, " ", "%20")

	search_url := host + "api?m=search&q=" + name

	fmt.Println("Search URL: ", search_url)

	body, err := makeRequest(search_url)
	if err != nil {
		fmt.Println(err)
	}

	// fmt.Println(string(body))

	var response PaheResponse
	json.Unmarshal(body, &response)

	// fmt.Printf("+%v", response)

	return response.Data, response.Total

}

func GetAnimeEpisodesList(anime_id string) (int, int) {

	url := host + "api?m=release&id=" + anime_id + "&sort=episode_asc&page=1"
	body, err := makeRequest(url)
	if err != nil {
		fmt.Println("Error retrieving episodes : ", err)
	}

	var animeResponse AnimeResponse
	json.Unmarshal(body, &animeResponse)

	return animeResponse.Total, animeResponse.Data[0].Episode
}

func FetchEpisodeIds(anime_id string, episode_range [2]int) map[int]string {

	pages := [2]int{1, 1}
	pages[0] += episode_range[0] / 30
	pages[1] += episode_range[1] / 30

	episode_numbers := []int{}

	for i := episode_range[0]; i <= episode_range[1]; i++ {
		episode_numbers = append(episode_numbers, i)
	}

	fmt.Println("Episode Numbers : ", episode_numbers)
	fmt.Println("Pages: ", pages)

	episode_id_number_map := make(map[int]string)

	for i := pages[0] - 1; i <= pages[1]+1; i++ {

		fmt.Println("Page: ", i)

		url := host + "api?m=release&id=" + anime_id + "&sort=episode_asc&page=" + fmt.Sprint(i)
		body, err := makeRequest(url)

		if err != nil {
			fmt.Println("Error retrieving episodes : ", err)
		}

		// fmt.Println("Body: ", string(body))

		var animeResponse AnimeResponse
		json.Unmarshal(body, &animeResponse)

		// episode_id_number_map[episode] = episode_id
		for _, ep := range episode_numbers {
			for _, episode_data := range animeResponse.Data {
				if episode_data.Episode == ep {
					episode_id_number_map[ep] = episode_data.Id
				}
			}
		}

	}

	// fmt.Println("Episode Map : ", episode_id_number_map)

	return episode_id_number_map

}

func FetchEpisodeLinks(animeId string, episode_ids map[int]string) map[string][]Episode {
	episode_links := make(map[int]string)

	lang_episode_map := make(map[string][]Episode)

	for i, episode_id := range episode_ids {
		url := host + "play/" + animeId + "/" + episode_id
		episode_links[i] = url

		// body, err := makeRequest(url)
		// if err != nil {
		// 	fmt.Println("Error retrieving episodes link : ", err)
		// }

		response_channel := make(chan string, 8)
		go func() {
			body, err := makeRequest(url)
			if err != nil {
				fmt.Println("Error retrieving episodes link : ", err)
			}
			response_channel <- string(body)
		}()

		// html_body := string(body)
		html_body := <-response_channel
		pattern := `href="(?:([^\"]+)" target="_blank" class="dropdown-item">(?:[^\&]+)&middot; ([^\<]+))(?:<span class="badge badge-primary">(?:[^\&]+)</span> <span class="badge badge-warning text-capitalize">([^\<]+))?`
		regex := regexp.MustCompile(pattern)

		matches := regex.FindAllStringSubmatch(html_body, -1)

		for _, match := range matches {
			// fmt.Println("Match: ", match)
			// for _, s := range match {
			// 	fmt.Println(s)
			// }
			splits := strings.Split(match[2], " ")
			quality := splits[0]
			file_size := strings.ReplaceAll(splits[1], "(", "")
			file_size = strings.ReplaceAll(file_size, ")", "")
			language := ""
			if len(match) > 2 {
				language = match[3]
			}
			if language == "" {
				language = "jpn"
			}
			fmt.Println("Episode:", i, "Link:", match[1], "Quality:", quality, "Language:", language, "Size:", file_size)

			lang_episode_map[language] = append(lang_episode_map[language], Episode{Url: match[1], Number: i, Quality: quality, FileSize: file_size})
		}
		fmt.Println()

		// map[lang][]Episode

	}

	// fmt.Printf("%+v", lang_episode_map)
	log.Default().Printf("%v", lang_episode_map)

	return lang_episode_map

}



func FetchEpisodeIdsWithGoRoutine(anime_id string, episode_range [2]int) map[int]string {

	pages := [2]int{1, 1}
	pages[0] += episode_range[0] / 30
	pages[1] += episode_range[1] / 30

	episode_numbers := []int{}

	for i := episode_range[0]; i <= episode_range[1]; i++ {
		episode_numbers = append(episode_numbers, i)
	}

	fmt.Println("Episode Numbers : ", episode_numbers)
	fmt.Println("Pages: ", pages)

	episode_id_number_map := make(map[int]string)

	response_channel := make(chan []byte, 8)
	

	for i := pages[0] - 1; i <= pages[1]+1; i++ {

		fmt.Println("Page: ", i)

		url := host + "api?m=release&id=" + anime_id + "&sort=episode_asc&page=" + fmt.Sprint(i)

		go func() {
			body, err := makeRequest(url)
			if err != nil {
				fmt.Println("Error retrieving episodes : ", err)
			}
			response_channel <- (body)
		}()

	}



	for i := pages[0] - 1; i <= pages[1]+1; i++ {

		body := <-response_channel

		var animeResponse AnimeResponse
		json.Unmarshal(body, &animeResponse)

		// episode_id_number_map[episode] = episode_id
		for _, ep := range episode_numbers {
			for _, episode_data := range animeResponse.Data {
				if episode_data.Episode == ep {
					episode_id_number_map[ep] = episode_data.Id
				}
			}
		}

	}

	// fmt.Println("Episode Map : ", episode_id_number_map)

	return episode_id_number_map

}



func FetchEpisodeLinksWithGoRoutine(animeId string, episode_ids map[int]string) map[string][]Episode {
	episode_links := make(map[int]string)

	lang_episode_map := make(map[string][]Episode)
	response_channel := make(chan string, 8)

	for i, episode_id := range episode_ids {
		url := host + "play/" + animeId + "/" + episode_id
		episode_links[i] = url

		go func() {
			body, err := makeRequest(url)
			if err != nil {
				fmt.Println("Error retrieving episodes link : ", err)
			}
			response_channel <- string(body)
		}()
	}

	for i := range episode_ids {
		// html_body := string(body)
		html_body := <-response_channel
		pattern := `href="(?:([^\"]+)" target="_blank" class="dropdown-item">(?:[^\&]+)&middot; ([^\<]+))(?:<span class="badge badge-primary">(?:[^\&]+)</span> <span class="badge badge-warning text-capitalize">([^\<]+))?`
		regex := regexp.MustCompile(pattern)

		matches := regex.FindAllStringSubmatch(html_body, -1)

		for _, match := range matches {
			// fmt.Println("Match: ", match)
			// for _, s := range match {
			// 	fmt.Println(s)
			// }
			splits := strings.Split(match[2], " ")
			quality := splits[0]
			file_size := strings.ReplaceAll(splits[1], "(", "")
			file_size = strings.ReplaceAll(file_size, ")", "")
			language := ""
			if len(match) > 2 {
				language = match[3]
			}
			if language == "" {
				language = "jpn"
			}
			fmt.Println("Episode:", i, "Link:", match[1], "Quality:", quality, "Language:", language, "Size:", file_size)

			lang_episode_map[language] = append(lang_episode_map[language], Episode{Url: match[1], Number: i, Quality: quality, FileSize: file_size})
		}
		fmt.Println()

		// map[lang][]Episode

	}

	// fmt.Printf("%+v", lang_episode_map)
	// log.Default().Printf("%v", lang_episode_map)

	return lang_episode_map

}