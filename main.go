package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
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
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
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
		fmt.Println(path, info.IsDir(), info.Name())
		if err != nil {
			log.Printf("failure when accessing input dir: %v", err)
			return err
		}

		if info.IsDir() {
			log.Println("skipping dir")
			return nil
		}

		filename := info.Name()
		if !strings.Contains(filename, ".zip") {
			return fmt.Errorf("[%s] is not a zip file", filename)
		}

		m := pdf.NewMaroto(consts.Portrait, consts.A5)
		m.SetPageMargins(0, 0, 0)
		imgs, err := GetImagesFromZip(path)
		if err != nil {
			return err
		}

		sort.Slice(imgs, func(i, j int) bool {
			log.Println(imgs[i].Index, imgs[j].Index)
			return imgs[i].Index < imgs[j].Index
		})

		for _, jpg := range imgs {
			img, err := imaging.Decode(jpg.Content, imaging.AutoOrientation(true))
			if err != nil {
				return err
			}

			log.Printf("img %s bounds: %+v", jpg.FileInfo.Name(), img.Bounds())

			// image in landscape we need to rotate
			if img.Bounds().Max.X > img.Bounds().Max.Y {
				img = imaging.Rotate270(img)
				log.Printf("rotated img %s bounds: %+v", jpg.FileInfo.Name(), img.Bounds())
			}

			var byteBuffer bytes.Buffer
			err = imaging.Encode(&byteBuffer, img, imaging.JPEG, imaging.JPEGQuality(100))
			if err != nil {
				return err
			}

			//imgB64 := "data:image/jpeg;base64,"
			imgB64 := base64.StdEncoding.EncodeToString(byteBuffer.Bytes())
			m.Row(210, func() {
				m.Col(148, func() {
					err = m.Base64Image(imgB64, consts.Jpg, props.Rect{
						Left:    0.0,
						Top:     0.0,
						Percent: 100,
						Center:  false,
					})
					if err != nil {
						log.Fatal(err)
					}

				})
			})

			left, top, right, bottom := m.GetPageMargins()
			width, height := m.GetPageSize()
			log.Printf("page margins - left: %.2f top: %.2f right: %.2f bottom: %.2f ", left, top, right, bottom)
			log.Printf("page size - width: %.2f, height: %.2f", width, height)
			m.AddPage()
			log.Printf("current page %d", m.GetCurrentPage())

			jpg.Content.Close()
		}

		ext := filepath.Ext(filename)
		outputFile := fmt.Sprintf("%s.pdf", strings.TrimSuffix(filename, ext))
		log.Printf("output path: %s", outputFile)

		return m.OutputFileAndClose(filepath.Join(outputDir, outputFile))
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
		log.Println(imgNameSplited)
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
