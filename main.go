package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

func getAsOptions(pahe_response []SearchResponse) []string {
	var options []string
	for _, anime := range pahe_response {
		var details string = anime.Title + " - " + anime.Type_ + " (" + fmt.Sprint(anime.Year) + ")"
		options = append(options, details)
	}
	return options
}

func promptAndSelectAvailableLanguage(episode_links map[string][]Episode) string {

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

func getAvailableQualityOptions(episode_link []Episode) []string {
	quality := make(map[string]int)

	for _, episode := range episode_link {
		quality[episode.Quality] = 1
	}

	// fmt.Println("Quality : ", quality)

	keys := []string{}
	for key, _ := range quality {
		keys = append(keys, key)
	}

	// fmt.Println("Quality Keys :", keys)
	return keys
}

func promptAndSelectAvailableQuality(qualities []string) string {

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

func filterEpisodes(episode_list []Episode, quality string) []Episode {

	filteredEpisodes := []Episode{}

	for _, episode := range episode_list {
		if episode.Quality == quality {
			filteredEpisodes = append(filteredEpisodes, episode)
		}
	}
	return filteredEpisodes
}

func main() {

	// current_directory, _ := os.Getwd()
	// fmt.Println("Current directory:", current_directory)

	fmt.Print("Enter input: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var input_anime string = scanner.Text()

	// search for anime
	pahe_response, size := SearchAnime(input_anime)
	if size == 0 {
		fmt.Println("No anime found")
		return
	}

	// display options to select anime
	options := getAsOptions(pahe_response)
	prompt := promptui.Select{
		Label: "Select an Anime",
		Items: options,
	}
	index, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	selected_anime := pahe_response[index]

	var anime_id string = selected_anime.Id
	// episode_count := selected_anime.Episodes

	total_episodes_aired, first_episode_number := GetAnimeEpisodesList(anime_id)
	latest_episode := first_episode_number + total_episodes_aired - 1

	// display selected anime
	fmt.Println("\033[35m" + selected_anime.Title + " - " + fmt.Sprint(selected_anime.Year))
	fmt.Println("\033[36m"+"Type: ", selected_anime.Type_)
	fmt.Println("\033[33m"+"Episodes: ", selected_anime.Episodes, "("+fmt.Sprint(first_episode_number)+"-"+fmt.Sprint(latest_episode)+")")
	fmt.Println("\033[32m"+"Status: ", selected_anime.Status)
	fmt.Println("\033[0m"+"Score: ", selected_anime.Score)
	fmt.Printf("\n\n")

	if selected_anime.Episodes == 0 {
		fmt.Println("\033[31m" + "No episodes found")
		return
	}
	

	// ask for episode numbers
	fmt.Print("Enter episode range (e.g. 5, " + fmt.Sprint(first_episode_number) + "-" + fmt.Sprint(latest_episode) + ", all) : ")
	scanner.Scan()
	var input_range string = scanner.Text()


	episode_range := [2]int{}
	if input_range == "" || input_range == "all" {
		episode_range[0] = first_episode_number
		episode_range[1] = latest_episode
	} else if strings.Contains(input_range, "-") {
		var split_range []string = strings.Split(input_range, "-")
		episode_range[0], _ = strconv.Atoi(split_range[0])
		episode_range[1], _ = strconv.Atoi(split_range[1])
	} else {
		episode_range[0], _ = strconv.Atoi(input_range)
		episode_range[1], _ = strconv.Atoi(input_range)
	}

	fmt.Println("Selected Episode Range: ", episode_range)

	// fetch episode ids
	// var episode_ids map[int]string = FetchEpisodeIds(anime_id, episode_range)
	var episode_ids map[int]string = FetchEpisodeIdsWithGoRoutine(anime_id, episode_range)
	fmt.Println("Episode IDs: ", episode_ids)



	// fetch episode links
	// episode_links := FetchEpisodeLinks(anime_id, episode_ids)
	episode_links := FetchEpisodeLinksWithGoRoutine(anime_id, episode_ids)

	// select language from available list
	selected_language := promptAndSelectAvailableLanguage(episode_links)
	fmt.Println("Selected Language: ", selected_language)

	// display and select quality from available list
	var quality_options []string = getAvailableQualityOptions(episode_links[selected_language])
	selected_quality := promptAndSelectAvailableQuality(quality_options)
	fmt.Println("Selected Quality:", selected_quality)

	// filter Episode based on language and Quality
	episode_list := filterEpisodes(episode_links[selected_language], selected_quality)
	fmt.Println("Final Episodes: ", episode_list)

}
