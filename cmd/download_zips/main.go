package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	var mangaTerm string
	var outputDir string
	var fromChapter int
	var toChapter int
	flag.StringVar(&mangaTerm, "manga_term", "Chainsaw_Man", "Manga term used on URL for downloading ZIP files")
	flag.StringVar(&outputDir, "output_dir", "/Users/wotan/Documents/ChainsawMan", "output directory that will contain the ZIP files")
	flag.IntVar(&fromChapter, "from_chapter", 1, "From which chapter should be downloaded")
	flag.IntVar(&toChapter, "to_chapter", 120, "To which chapter to be downloaded")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "This script download the manga ZIPs from mangafreak")
		flag.PrintDefaults()
	}
	flag.Parse()

	const endpoint = "https://images.mangafreak.net/downloads"

	for i := fromChapter; i <= toChapter; i++ {
		chapterName := fmt.Sprintf("%s_%d", mangaTerm, i)
		chapterURL := fmt.Sprintf("%s/%s", endpoint, chapterName)

		outputFilename := fmt.Sprintf("%s.zip", chapterName)
		outputFilepath := filepath.Join(outputDir, outputFilename)

		file, err := os.Create(outputFilepath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		client := http.Client{
			CheckRedirect: func(r *http.Request, _ []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}

		resp, err := client.Get(chapterURL)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		size, err := io.Copy(file, resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Downloaded [%s] with size %d", outputFilepath, size)
	}
}
