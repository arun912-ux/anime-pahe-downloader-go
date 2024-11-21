package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func DownloadEpisodes(download_url []Episode, language string, download_dir string) {

	os.Chdir(download_dir)

	pwd, _ := os.Getwd()
	fmt.Println("Current Directory: ", pwd)

	var wg sync.WaitGroup
	resp_chan := make(chan string, 2)
	progress := mpb.New(mpb.WithWaitGroup(&wg))

	// barTotal := progressbar.New(len(download_url))
	barTotal := progress.AddBar(int64(len(download_url)),
		mpb.BarStyle("[=> ]<+"),
		mpb.PrependDecorators(decor.Name("Total : ")),
		mpb.AppendDecorators(decor.Percentage(decor.WCSyncWidth)),
	)

	for _, ep := range download_url {
		wg.Add(1)
		resp_chan <- "Downloaded: " + fmt.Sprint(ep.Number) + "_" + language + "_" + ep.Quality + ".mp4 " + " from: " + ep.Url
		go func() {
			defer wg.Done()

			fileName := fmt.Sprint(ep.Number) + "_" + language + "_" + ep.Quality + ".mp4"
			filePath := pwd + "/" + fileName
			// TODO: create bars outside go routine and pass them to the go routine

			resp, _ := DownloadFromUrl(ep.Url, filePath)

			// bar := progressbar.New64(resp.ContentLength)
			bar := progress.AddBar(int64(resp.ContentLength),
				mpb.BarStyle("[=> ]<+"),

				mpb.PrependDecorators(
					decor.Name(fileName),
					decor.Percentage(decor.WCSyncSpace),
				),
				mpb.AppendDecorators(
					decor.CountersKibiByte("% .2f / % .2f    "),
					decor.Percentage(decor.WCSyncWidth),
				),
			)
			fmt.Println("Downloading: " + fileName + " " + "from: " + ep.Url)

			SaveToFile(resp, filePath, bar)
			// resp_chan <- "Downloading: " + fileName + " " + "from: " + ep.Url
			<-resp_chan
			fmt.Println("Downloading: " + fileName + " " + "from: " + ep.Url)
			barTotal.Increment()
		}()
	}

	go func() {
		wg.Wait()
		close(resp_chan)
	}()

	// for resp := range resp_chan {
	// 	fmt.Println(resp)
	// 	barTotal.Add(1)
	// }

}

func CreateDirectory(name string) string {
	// fmt.Println("Creating Directory: ", name)
	directory := "./Download/" + name
	err := os.MkdirAll(directory, os.ModePerm)
	if !(err == nil || os.IsExist(err)) {
		fmt.Println(err)
	}

	return directory
}

func DownloadFromUrl(url, filePath string) (*http.Response, error) {
	// Send a GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting file: %v", err)
	}

	// Check for valid response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file: %v", resp.Status)
	}
	return resp, nil
}

func SaveToFile(resp *http.Response, filePath string, progressBar *mpb.Bar) error {

	// Get the size of the file to set up the progress bar
	contentLength := resp.ContentLength
	if contentLength == -1 {
		fmt.Println("Warning: Content-Length not provided, proceeding without progress bar.")
	}

	// Create the destination file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Set up the buffer size (2MB)
	bufferSize := 2 * 1024 * 1024 // 2MB
	buffer := make([]byte, bufferSize)

	// Total number of bytes read
	var totalBytesRead int

	// Read the file in chunks of 2MB and write to file + progress bar
	for {
		n, err := resp.Body.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading response body: %v", err)
		}

		// Write to both the file and progress bar
		_, writeErr := file.Write(buffer[:n])
		if writeErr != nil {
			return fmt.Errorf("error writing to file: %v", writeErr)
		}

		// Update the progress bar
		totalBytesRead += int(n)
		progressBar.IncrBy(n)
		// Check if we have reached the end of the file
		if err == io.EOF {
			break
		}
	}

	// io.MultiWriter is used to write to both the file and the progress bar
	// io.Copy(io.MultiWriter(progressBar, file), resp.Body)

	// Download complete
	return nil
}
