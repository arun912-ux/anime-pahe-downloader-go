package main

import (
	"fmt"
	"sync"

	"github.com/manifoldco/promptui"
	"github.com/schollz/progressbar/v3"
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

// Download Methods

func IterateOverEpisodesWithDownloadLinks(kwik_episodes []Episode) []Episode {

	// download_url := make(map[int]string)
	download_ep := make([]Episode, 0, len(kwik_episodes))

	var wg sync.WaitGroup
	resp_chan := make(chan Episode, 8)
	bar := progressbar.New(len(kwik_episodes))
	for _, kwik_ep := range kwik_episodes {

		// fmt.Println("inside For Loop : ", kwik_ep.Number)
		wg.Add(1)
		go func() {
			defer wg.Done()
			body, httpClient := makeGetRequestWithSession(kwik_ep.Url)

			// fmt.Println("Response :", kwik_ep.Url, string(body)[0:10])

			location := Parse_html_token_getDownloadLink(string(body), kwik_ep.Url, httpClient)
			// fmt.Println("Download Location: ", location)
			ep := Episode{
				Number:  kwik_ep.Number,
				Url:     location,
				Quality: kwik_ep.Quality,
			}
			resp_chan <- ep
		}()

	}

	go func() {
		wg.Wait()
		close(resp_chan)
	}()

	// bar := progressbar.New(len(kwik_episodes))
	for resp := range resp_chan {
		// fmt.Println(resp)
		// download_url[resp.Number] = resp.Url
		ep := Episode{
			Number:  resp.Number,
			Quality: resp.Quality,
			Url:     resp.Url,
		}
		download_ep = append(download_ep, ep)
		bar.Add(1)
	}

	// fmt.Println(download_url)

	// return download_url
	return download_ep

}




// func DownloadEpisodes(download_url []Episode, language string, download_dir string) {

// 	os.Chdir(download_dir)

// 	pwd, _ := os.Getwd()
// 	fmt.Println("Current Directory: ", pwd)

// 	var wg sync.WaitGroup
// 	resp_chan := make(chan string, 8)

// 	for _, ep := range download_url {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			filePath := pwd + "/" + fmt.Sprint(ep.Number) + "_" + language + "_" + ep.Quality + ".mp4"
// 			// TODO: create bars outside go routine and pass them to the go routine

// 			DownloadFromUrl(ep.Url, filePath)
// 			resp_chan <- "Downloading: " + filePath + " " + "from: " + ep.Url
// 		}()
// 	}

// 	go func() {
// 		wg.Wait()
// 		close(resp_chan)
// 	}()

// 	for resp := range resp_chan {
// 		fmt.Println(resp)
// 	}

// }

// func CreateDirectory(name string) string {
// 	// fmt.Println("Creating Directory: ", name)
// 	directory := "./Download/" + name
// 	err := os.MkdirAll(directory, os.ModePerm)
// 	if !(err == nil || os.IsExist(err)) {
// 		fmt.Println(err)
// 	}

// 	return directory
// }

// func DownloadFromUrl(url string, filePath string) error {

// 	// fmt.Println("Downloading: ", url)
// 	// Check if the file already exists
// 	fileInfo, err := os.Stat(filePath)
// 	var fileSize int64
// 	if err == nil {
// 		// If the file exists, get the size
// 		fileSize = fileInfo.Size()
// 		fmt.Printf("File already exists. Resuming download from byte %d...\n", fileSize)
// 	} else if os.IsNotExist(err) {
// 		// If the file doesn't exist, start from the beginning
// 		fmt.Println("File does not exist. Starting download...")
// 	} else {
// 		return fmt.Errorf("failed to check file: %v", err)
// 	}

// 	// Create HTTP client
// 	client := &http.Client{}

// 	// Add the Range header if file exists and is incomplete
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to create request: %v", err)
// 	}
// 	if fileSize > 0 {
// 		req.Header.Add("Range", "bytes="+strconv.FormatInt(fileSize, 10)+"-")
// 	}

// 	// Send the HTTP request
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return fmt.Errorf("failed to send request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	// Check if the server supports resuming the download
// 	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
// 		return fmt.Errorf("server responded with status %d: %s", resp.StatusCode, resp.Status)
// 	}

// 	// Get the total size of the file (either from the server or the local file)
// 	var totalSize int64
// 	if resp.StatusCode == http.StatusOK {
// 		totalSize = resp.ContentLength
// 	} else {
// 		totalSize = resp.ContentLength + fileSize
// 	}

// 	// Create progress bar
// 	bar := progressbar.NewOptions64(
// 		totalSize,
// 		progressbar.OptionShowTotalBytes(true),
// 		progressbar.OptionSetDescription("Downloading "+strings.Split(filePath, "/")[len(strings.Split(filePath, "/"))-1]),
// 		progressbar.OptionSetWidth(60),
// 		progressbar.OptionSetRenderBlankState(true),
// 		progressbar.OptionSetTheme(progressbar.ThemeASCII),
// 		progressbar.OptionShowBytes(true),
// 		progressbar.OptionShowTotalBytes(true),
// 	)

// 	bar.Set64(fileSize)
// 	// bar.Reset()

// 	// Open the file for appending (create if doesn't exist)
// 	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
// 	if err != nil {
// 		return fmt.Errorf("failed to open file: %v", err)
// 	}
// 	defer file.Close()

// 	// Create a buffer to read data
// 	buffer := make([]byte, 1024*8) // 8 KB buffer
// 	var totalDownloaded int64

// 	// Download the file in chunks
// 	for {
// 		n, err := resp.Body.Read(buffer)
// 		if err != nil && err != io.EOF {
// 			return fmt.Errorf("failed to read response body: %v", err)
// 		}
// 		if n == 0 {
// 			break
// 		}

// 		// Write to file and update progress bar
// 		if _, err := file.Write(buffer[:n]); err != nil {
// 			return fmt.Errorf("failed to write to file: %v", err)
// 		}

// 		// Update the total downloaded bytes
// 		totalDownloaded += int64(n)
// 		bar.Add(n) // Update the progress bar
// 		fmt.Print()
// 	}

// 	// Print completion message
// 	fmt.Println("\nDownload completed successfully.")
// 	return nil

// }

// DownloadFromUrl handles downloading the file from the provided URL
// and displays the progress bar.

// func DownloadFromUrl(url, filePath string) error {
// 	// Send a GET request to the URL
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return fmt.Errorf("error getting file: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	// Check for valid response
// 	if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("failed to download file: %v", resp.Status)
// 	}

// 	// Get the size of the file to set up the progress bar
// 	contentLength := resp.ContentLength
// 	if contentLength == -1 {
// 		fmt.Println("Warning: Content-Length not provided, proceeding without progress bar.")
// 	}

// 	// Create the destination file
// 	file, err := os.Create(filePath)
// 	if err != nil {
// 		return fmt.Errorf("error creating file: %v", err)
// 	}
// 	defer file.Close()

// 	// Initialize the multibar
// 	bar, _ := multibar.New()
// 	go bar.Listen()

// 	// Add a progress bar for downloading
// 	progressBar := bar.MakeBar(int(contentLength), fmt.Sprintf("Downloading %s", filePath))

// 	// Set up the buffer size (2MB)
// 	bufferSize := 2 * 1024 * 1024 // 2MB
// 	buffer := make([]byte, bufferSize)

// 	// Total number of bytes read
// 	var totalBytesRead int

// 	// Read the file in chunks of 2MB and write to file + progress bar
// 	for {
// 		n, err := resp.Body.Read(buffer)
// 		if err != nil && err != io.EOF {
// 			return fmt.Errorf("error reading response body: %v", err)
// 		}

// 		// Write to both the file and progress bar
// 		_, writeErr := file.Write(buffer[:n])
// 		if writeErr != nil {
// 			return fmt.Errorf("error writing to file: %v", writeErr)
// 		}

// 		// Update the progress bar
// 		totalBytesRead += int(n)
// 		progressBar(totalBytesRead)

// 		// Check if we have reached the end of the file
// 		if err == io.EOF {
// 			break
// 		}
// 	}

// 	// Download complete
// 	return nil
// }
