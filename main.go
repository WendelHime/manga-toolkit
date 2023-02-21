package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/go-pdf/fpdf"
)

func main() {
	var inputDir string
	var outputDir string
	flag.StringVar(&inputDir, "input_dir", "", "directory with manga ZIP images")
	flag.StringVar(&outputDir, "output_dir", "", "output directory that will contain generated PDFs")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, `This script expects a directory containing ZIP files with JPG images.
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

	err := filepath.Walk(inputDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Printf("failure when accessing input dir: %v", err)
			return err
		}

		if info.IsDir() {
			log.Println("skipping dir", path)
			return nil
		}

		filename := info.Name()
		if !strings.Contains(filename, ".zip") {
			return fmt.Errorf("[%s] is not a zip file", filename)
		}

		pdf := fpdf.New(fpdf.OrientationPortrait, fpdf.UnitMillimeter, fpdf.PageSizeA5, "")
		imgs, err := GetImagesFromZip(path)
		if err != nil {
			return err
		}

		sort.Slice(imgs, func(i, j int) bool {
			return imgs[i].Index < imgs[j].Index
		})

		for i, jpg := range imgs {
			img, err := imaging.Decode(jpg.Content, imaging.AutoOrientation(true))
			if err != nil {
				return err
			}

			// image in landscape we need to rotate
			if img.Bounds().Max.X > img.Bounds().Max.Y {
				img = imaging.Rotate270(img)
			}

			var byteBuffer bytes.Buffer
			err = imaging.Encode(&byteBuffer, img, imaging.JPEG, imaging.JPEGQuality(100))
			if err != nil {
				return err
			}

			opt := fpdf.ImageOptions{
				ImageType:             "jpg",
				AllowNegativePosition: false,
			}
			name := strconv.Itoa(i)
			pdf.RegisterImageOptionsReader(name, opt, &byteBuffer)
			pdf.AddPage()
			wd, hd, _ := pdf.PageSize(i)
			jpg.Content.Close()
			pdf.ImageOptions(name, 0, 0, wd, hd, false, opt, 0, "")
		}

		ext := filepath.Ext(filename)
		outputFile := fmt.Sprintf("%s.pdf", strings.TrimSuffix(filename, ext))
		log.Printf("output path: %s", outputFile)

		return pdf.OutputFileAndClose(filepath.Join(outputDir, outputFile))
	})
	if err != nil {
		log.Fatal(err)
	}
}

type Image struct {
	Content  io.ReadCloser
	FileInfo fs.FileInfo
	Index    int
}

func GetImagesFromZip(source string) ([]Image, error) {
	imgs := make([]Image, 0)
	reader, err := zip.OpenReader(source)
	if err != nil {
		return imgs, err
	}

	for _, f := range reader.File {
		if !strings.Contains(strings.ToLower(f.FileInfo().Name()), ".jpg") {
			continue
		}
		content, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}

		imgName := f.FileInfo().Name()
		ext := filepath.Ext(imgName)
		imgName = strings.TrimSuffix(imgName, ext)
		imgNameSplited := strings.Split(imgName, "_")
		img := Image{
			Content:  content,
			FileInfo: f.FileInfo(),
		}
		img.Index, err = strconv.Atoi(imgNameSplited[len(imgNameSplited)-1])
		if err != nil {
			log.Fatal(err)
		}

		imgs = append(imgs, img)
	}

	return imgs, nil
}
