package logic

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func (logic) GeneratePDFFromZip(ctx context.Context, reader *zip.ReadCloser, output io.WriteCloser) error {
	if reader == nil {
		return errors.New("missing reader")
	}
	if output == nil {
		return errors.New("missing output")
	}

	chapter, err := NewChapter(ctx, reader)
	if err != nil {
		return err
	}
	chapter.SortPages()

	return GenerateOutput(chapter, output)
}

func GenerateOutput(chapter Chapter, output io.WriteCloser) error {
	pdf := fpdf.New(fpdf.OrientationPortrait, fpdf.UnitMillimeter, fpdf.PageSizeA5, "")

	err := chapter.ForeachPage(func(page Page) error {
		buffer := new(bytes.Buffer)
		EncodePage(page, buffer)
		page.Content.Close()

		imgType, err := GetImageFormat(page)
		if err != nil {
			return err
		}

		AddPage(pdf, imgType, buffer)
		return nil
	})

	if err != nil {
		return err
	}

	return pdf.OutputAndClose(output)
}

// AddPage adds the manga image page buffer as a page to the PDF
func AddPage(pdf fpdf.Pdf, imgType string, buffer *bytes.Buffer) {
	opt := fpdf.ImageOptions{
		ImageType:             imgType,
		AllowNegativePosition: false,
		ReadDpi:               true,
	}

	name := uuid.NewString()
	pdf.RegisterImageOptionsReader(name, opt, buffer)
	pdf.AddPage()

	wd, ht, _ := pdf.PageSize(pdf.PageNo())
	pdf.ImageOptions(name, 0, 0, wd, ht, false, opt, 0, "")
}

// GetImageFormat extracts the image format from the manga image page
func GetImageFormat(image Page) (string, error) {
	format, err := imaging.FormatFromFilename(image.FileInfo.Name())
	if err != nil {
		return "", err
	}
	return format.String(), nil
}

// EncodePage encode a manga image page into the provided buffer
// If the page is on landscape mode it'll be rotated -90 degrees and will became
// portrait mode
func EncodePage(image Page, buffer *bytes.Buffer) error {
	format, err := imaging.FormatFromFilename(image.FileInfo.Name())
	if err != nil {
		return err
	}

	img, err := imaging.Decode(image.Content, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	// image in landscape we need to rotate
	if img.Bounds().Max.X > img.Bounds().Max.Y {
		img = imaging.Rotate270(img)
		img = imaging.FlipH(img)
		img = imaging.FlipV(img)
	}

	opts := make([]imaging.EncodeOption, 0)
	if format == imaging.JPEG {
		opts = append(opts, imaging.JPEGQuality(100))
	}

	err = imaging.Encode(buffer, img, format, opts...)
	if err != nil {
		return err
	}

	return nil
}

type Chapter struct {
	Pages []Page
}

// Page represents a manga image page
type Page struct {
	Content    io.ReadCloser
	FileInfo   fs.FileInfo
	PageNumber int
}

func (c Chapter) SortPages() {
	sort.Slice(c.Pages, func(i, j int) bool {
		return c.Pages[i].PageNumber < c.Pages[j].PageNumber
	})
}

func (c Chapter) ForeachPage(do func(page Page) error) error {
	var err error
	for _, page := range c.Pages {
		errors.Join(err, do(page))
	}
	return err
}

// NewChapter iterate through the ZIP file, open the content and return
// a list of images. Remember to close the images after using it!
func NewChapter(ctx context.Context, reader *zip.ReadCloser) (Chapter, error) {
	pages := make([]Page, 0)
	group, _ := errgroup.WithContext(ctx)
	mutex := new(sync.Mutex)

	if len(reader.File) == 0 {
		return Chapter{}, errors.New("empty zip")
	}

	for i := range reader.File {
		index := i
		group.Go(func() error {
			mutex.Lock()
			imageFile := reader.File[index]
			mutex.Unlock()
			imgName := imageFile.FileInfo().Name()
			imgExtension := strings.ToLower(filepath.Ext(imgName))
			if !IsValidExtension(imgExtension) {
				return fmt.Errorf("image [%s] have a unexpected image format", imgExtension)
			}
			content, err := imageFile.Open()
			if err != nil {
				return fmt.Errorf("failed opening image [%s] with error: %w", imgName, err)
			}

			imgName = strings.TrimSuffix(imgName, imgExtension)
			imgNameSplited := strings.Split(imgName, "_")
			page := Page{
				Content:  content,
				FileInfo: imageFile.FileInfo(),
			}
			page.PageNumber, err = strconv.Atoi(imgNameSplited[len(imgNameSplited)-1])
			if err != nil {
				return fmt.Errorf("couldn't extract page number from [%s]: %w", imgName, err)
			}
			mutex.Lock()
			pages = append(pages, page)
			mutex.Unlock()
			return nil
		})
	}

	return Chapter{Pages: pages}, group.Wait()
}

// IsValidExtension check if the image has a valid extension
func IsValidExtension(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}
