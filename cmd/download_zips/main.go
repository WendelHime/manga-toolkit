package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/WendelHime/manga-toolkit/internal/logic"
	"github.com/WendelHime/manga-toolkit/internal/services"
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

	service := services.NewMangaFreakService(endpoint)
	l := logic.NewLogic(service)

	err := l.DownloadChapters(context.Background(), mangaTerm, outputDir, fromChapter, toChapter)
	log.Fatal(err)
}
