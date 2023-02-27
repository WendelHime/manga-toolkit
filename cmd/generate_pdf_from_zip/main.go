package main

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/WendelHime/manga-toolkit/internal/logic"
)

func main() {
	var inputDir string
	var outputDir string
	flag.StringVar(&inputDir, "input_dir", "", "directory with manga ZIP images")
	flag.StringVar(&outputDir, "output_dir", "", "output directory that will contain generated PDFs")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, `This script generates PDFs from a provided directory containing ZIP files with JPG images.
/home/someuser/mangadir
| chapter1.zip
  - img1.jpg
  - img2.jpg
| chapter2.zip
| chapter3.zip
`)
		flag.PrintDefaults()
	}
	flag.Parse()

	l := logic.NewLogic(nil)

	err := filepath.Walk(inputDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Printf("failure when accessing input dir: %v", err)
			return err
		}

		if info.IsDir() {
			log.Println("skipping dir", path)
			return nil
		}

		zipFilename := info.Name()
		if !strings.Contains(zipFilename, ".zip") {
			return fmt.Errorf("[%s] is not a zip file", zipFilename)
		}

		zipReader, err := zip.OpenReader(path)
		if err != nil {
			return err
		}
		defer zipReader.Close()

		ext := filepath.Ext(zipFilename)
		outputFile := fmt.Sprintf("%s.pdf", strings.TrimSuffix(zipFilename, ext))
		log.Printf("output path: %s", outputFile)

		output, err := os.Create(filepath.Join(outputDir, outputFile))
		if err != nil {
			return err
		}
		defer output.Close()

		return l.GeneratePDFFromZip(context.Background(), zipReader, output)
	})
	if err != nil {
		log.Fatal(err)
	}
}
