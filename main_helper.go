package main

import (
	"fmt"
	"sync"

	"github.com/manifoldco/promptui"
)

func GetAsOptions(pahe_response []SearchResponse) []string {
	var options []string
	for _, anime := range pahe_response {
		var details string = anime.Title + " - " + anime.Type_ + " (" + fmt.Sprint(anime.Year) + ")"
		options = append(options, details)
	}
	return options
}

func PromptAndSelectAvailableLanguage(episode_links map[string][]Episode) string {

	languages := make([]string, 0, len(episode_links))
	for k := range episode_links {
		languages = append(languages, k)
	}

	fmt.Println("Languages: ", languages)

	// display options to select language
	prompt := promptui.Select{
		Label: "Select a Language",
		Items: languages,
	}
	_, language, err := prompt.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	return language
}

func GetAvailableQualityOptions(episode_link []Episode) []string {
	quality := make(map[string]int)

	for _, episode := range episode_link {
		quality[episode.Quality] = 1
	}

	// fmt.Println("Quality : ", quality)

	keys := []string{}
	for key := range quality {
		keys = append(keys, key)
	}

	// fmt.Println("Quality Keys :", keys)
	return keys
}

func PromptAndSelectAvailableQuality(qualities []string) string {

	// display options to select quality
	prompt := promptui.Select{
		Label: "Select a Quality",
		Items: qualities,
	}

	_, quality, err := prompt.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// fmt.Println("Selected Quality: ", quality)

	return quality

}

func FilterEpisodes(episode_list []Episode, quality string) []Episode {

	filteredEpisodes := []Episode{}

	for _, episode := range episode_list {
		if episode.Quality == quality {
			filteredEpisodes = append(filteredEpisodes, episode)
		}
	}
	return filteredEpisodes
}

func IterateOverEpisodes(kwik_episodes []Episode) map[int]string {

	download_url := make(map[int]string)

	var wg sync.WaitGroup
	resp_chan := make(chan map[int]string)

	for _, kwik_ep := range kwik_episodes {

		// fmt.Println("inside For Loop : ", kwik_ep.Number)
		wg.Add(1)
		go func() {
			defer wg.Done()
			body, httpClient := makeGetRequestWithSession(kwik_ep.Url)

			// fmt.Println("Response :", kwik_ep.Url, string(body)[0:10])

			location := Parse_html_token_getDownloadLink(string(body), kwik_ep.Url, httpClient)
			// fmt.Println("Download Location: ", location)
			ep_map := make(map[int]string)
			ep_map[kwik_ep.Number] = location
			resp_chan <- ep_map
		}()

	}

	go func() {
		wg.Wait()
		close(resp_chan)
	}()

	for resp := range resp_chan {
		// fmt.Println(resp)
		for n, u := range resp {
			download_url[n] = u
			break
		}
	}

	// fmt.Println(download_url)

	return download_url

}
