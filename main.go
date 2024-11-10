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

	// display options to select
	options := getAsOptions(pahe_response)
	prompt :=promptui.Select{
		Label: "Select an option",
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
	fmt.Println("\033[36m" + "Type: ", selected_anime.Type_)
	fmt.Println("\033[33m" + "Episodes: ", selected_anime.Episodes, "(" + fmt.Sprint(first_episode_number) + "-" + fmt.Sprint(latest_episode) + ")")
	fmt.Println("\033[32m" + "Status: ", selected_anime.Status)
	fmt.Println("\033[0m" + "Score: ", selected_anime.Score)
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
	jpn_episodes := episode_links["jpn"]
	fmt.Println("Episode Links: size: ", len(jpn_episodes))

	// fmt.Println("Episode Links: ", episode_links)

	// for lang, episode := range episode_links {
	// 	fmt.Println(lang)
	// 	for _, ep := range episode {
	// 		fmt.Println(ep)
	// 	}
	// }

	ep_count := 0
	for _, ep := range jpn_episodes {
		if (ep.Quality == "720p") {
			ep_count += 1
		}
	}

	fmt.Println("Episode count: ", ep_count)
}


