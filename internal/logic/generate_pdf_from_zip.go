package logic

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
)

func (logic) GeneratePDFFromZip(reader *zip.ReadCloser, output io.WriteCloser) error {
	imgs, err := GetImagesFromZip(reader)
	if err != nil {
		return err
	}

	sort.Slice(imgs, func(i, j int) bool {
		return imgs[i].PageNumber < imgs[j].PageNumber
	})

	pdf := fpdf.New(fpdf.OrientationPortrait, fpdf.UnitMillimeter, fpdf.PageSizeA5, "")

	for _, img := range imgs {
		buffer := new(bytes.Buffer)
		EncodeImage(img, buffer)
		img.Content.Close()

		imgType, err := GetImageFormat(img)
		if err != nil {
			return err
		}

		AddPage(pdf, imgType, buffer)
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
func GetImageFormat(image MangaImagePage) (string, error) {
	format, err := imaging.FormatFromFilename(image.FileInfo.Name())
	if err != nil {
		return "", err
	}
	return format.String(), nil
}

// EncodeImage encode a manga image page into the provided buffer
// If the page is on landscape mode it'll be rotated -90 degrees and will became
// portrait mode
func EncodeImage(image MangaImagePage, buffer *bytes.Buffer) error {
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

// MangaImagePage represents a manga image page
type MangaImagePage struct {
	Content    io.ReadCloser
	FileInfo   fs.FileInfo
	PageNumber int
}

// GetImagesFromZip iterate through the ZIP file, open the content and return
// a list of images. Remember to close the images after using it!
func GetImagesFromZip(reader *zip.ReadCloser) ([]MangaImagePage, error) {
	imgs := make([]MangaImagePage, 0)
	var allErrors error

	for _, imageFile := range reader.File {
		imgName := imageFile.FileInfo().Name()
		imgExtension := strings.ToLower(filepath.Ext(imgName))
		if !IsValidExtension(imgExtension) {
			errors.Join(allErrors, fmt.Errorf("image [%s] have a unexpected image format", imgExtension))
			continue
		}
		content, err := imageFile.Open()
		if err != nil {
			errors.Join(allErrors, fmt.Errorf("failed opening image [%s] with error: %w", imgName, err))
			continue
		}

		imgName = strings.TrimSuffix(imgName, imgExtension)
		imgNameSplited := strings.Split(imgName, "_")
		img := MangaImagePage{
			Content:  content,
			FileInfo: imageFile.FileInfo(),
		}
		img.PageNumber, err = strconv.Atoi(imgNameSplited[len(imgNameSplited)-1])
		if err != nil {
			errors.Join(allErrors, fmt.Errorf("couldn't extract page number from [%s]: %w", imgName, err))
			continue
		}

		imgs = append(imgs, img)
	}

	return imgs, allErrors
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
