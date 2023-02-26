package logic

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePDFFromZip(t *testing.T) {
	t.Run("IsValidExtension", func(t *testing.T) {
		var tests = []struct {
			name     string
			expected bool
			givenExt string
		}{
			{
				name:     "JPG is valid",
				expected: true,
				givenExt: ".jpg",
			},
			{
				name:     "JPEG is valid",
				expected: true,
				givenExt: ".jpeg",
			},
			{
				name:     "PNG is valid",
				expected: true,
				givenExt: ".png",
			},
			{
				name:     "Tiff is invalid",
				expected: false,
				givenExt: ".tiff",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				actual := IsValidExtension(tt.givenExt)
				assert.Equal(t, tt.expected, actual)
			})
		}
	})

	createValidZipfile := func(t *testing.T) string {
		file, err := ioutil.TempFile("/tmp/", "valid-*.zip")
		assert.NoError(t, err)
		defer file.Close()

		zipWriter := zip.NewWriter(file)
		defer zipWriter.Close()

		for i := 0; i < 10; i++ {
			img := image.NewNRGBA(image.Rect(0, 0, 10, 10))

			w, err := zipWriter.Create(fmt.Sprintf("%d.jpg", i))
			assert.NoError(t, err)

			jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
		}

		return file.Name()
	}

	createZipWithInvalidFileExt := func(t *testing.T) string {
		file, err := ioutil.TempFile("/tmp/", "invalid-ext-*.zip")
		assert.NoError(t, err)
		defer file.Close()

		zipWriter := zip.NewWriter(file)
		defer zipWriter.Close()

		img := image.NewNRGBA(image.Rect(0, 0, 10, 10))

		w, err := zipWriter.Create(fmt.Sprintf("%d.random", 0))
		assert.NoError(t, err)

		jpeg.Encode(w, img, &jpeg.Options{Quality: 90})

		return file.Name()
	}

	createZipWithoutIndexOnFilename := func(t *testing.T) string {
		file, err := ioutil.TempFile("/tmp/", "invalid-filename-*.zip")
		assert.NoError(t, err)
		defer file.Close()

		zipWriter := zip.NewWriter(file)
		defer zipWriter.Close()

		img := image.NewNRGBA(image.Rect(0, 0, 10, 10))

		w, err := zipWriter.Create("random.jpg")
		assert.NoError(t, err)

		jpeg.Encode(w, img, &jpeg.Options{Quality: 90})

		return file.Name()
	}

	createEmptyZip := func(t *testing.T) string {
		file, err := ioutil.TempFile("/tmp/", "invalid-filename-*.zip")
		assert.NoError(t, err)
		defer file.Close()

		zipWriter := zip.NewWriter(file)
		defer zipWriter.Close()

		return file.Name()
	}
	t.Run("NewChapter", func(t *testing.T) {
		var tests = []struct {
			name   string
			setup  func(t *testing.T) (r *zip.ReadCloser, zipname string)
			assert func(t *testing.T, gotChapter Chapter, gotErr error)
		}{
			{
				name: "should create a chapter with all pages",
				setup: func(t *testing.T) (*zip.ReadCloser, string) {
					zipname := createValidZipfile(t)
					r, err := zip.OpenReader(zipname)
					assert.NoError(t, err)
					return r, zipname
				},
				assert: func(t *testing.T, gotChapter Chapter, gotErr error) {
					assert.NoError(t, gotErr)
					gotChapter.SortPages()

					for i := 0; i < 10; i++ {
						assert.Equal(t, i, gotChapter.Pages[i].PageNumber)
						assert.Equal(t, fmt.Sprintf("%d.jpg", i), gotChapter.Pages[i].FileInfo.Name())
						img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
						gotImg, err := jpeg.Decode(gotChapter.Pages[i].Content)
						assert.NoError(t, err)
						assert.Equal(t, img.Bounds(), gotImg.Bounds())
					}
				},
			},
			{
				name: "should return an error if the zip contains unexpected content extension",
				setup: func(t *testing.T) (*zip.ReadCloser, string) {
					zipname := createZipWithInvalidFileExt(t)
					r, err := zip.OpenReader(zipname)
					assert.NoError(t, err)
					return r, zipname
				},
				assert: func(t *testing.T, _ Chapter, gotErr error) {
					assert.Error(t, gotErr)
				},
			},
			{
				name: "should return an error if the zip contains a file without the page number",
				setup: func(t *testing.T) (*zip.ReadCloser, string) {
					zipname := createZipWithoutIndexOnFilename(t)
					r, err := zip.OpenReader(zipname)
					assert.NoError(t, err)
					return r, zipname
				},
				assert: func(t *testing.T, _ Chapter, gotErr error) {
					assert.Error(t, gotErr)
				},
			},
			{
				name: "should return an error if the zip is empty",
				setup: func(t *testing.T) (*zip.ReadCloser, string) {
					zipname := createEmptyZip(t)
					r, err := zip.OpenReader(zipname)
					assert.NoError(t, err)
					return r, zipname
				},
				assert: func(t *testing.T, _ Chapter, gotErr error) {
					assert.Error(t, gotErr)
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				input, zipname := tt.setup(t)
				defer func() {
					input.Close()
					if assert.NotEmpty(t, zipname) {
						os.Remove(zipname)
					}
				}()
				chapter, err := NewChapter(input)
				tt.assert(t, chapter, err)
			})
		}
	})
	t.Run("GeneratePDFFromZip", func(t *testing.T) {
		l := NewLogic()
		var tests = []struct {
			name   string
			setup  func(t *testing.T) (*zip.ReadCloser, io.WriteCloser, string)
			assert func(t *testing.T, gotErr error)
		}{
			{
				name: "should return error when reader is nil",
				setup: func(t *testing.T) (*zip.ReadCloser, io.WriteCloser, string) {
					return nil, nil, ""
				},
				assert: func(t *testing.T, gotErr error) {
					assert.Error(t, gotErr)
				},
			},
			{
				name: "should return error when output is nil",
				setup: func(t *testing.T) (*zip.ReadCloser, io.WriteCloser, string) {
					return new(zip.ReadCloser), nil, ""
				},
				assert: func(t *testing.T, gotErr error) {
					assert.Error(t, gotErr)
				},
			},
			{
				name: "should return error if there's anything wrong with the zip content",
				setup: func(t *testing.T) (*zip.ReadCloser, io.WriteCloser, string) {
					zipname := createZipWithInvalidFileExt(t)
					r, err := zip.OpenReader(zipname)
					assert.NoError(t, err)
					buffer := new(bytes.Buffer)
					bw := bufio.NewWriter(buffer)
					wc := &WC{bw}
					return r, wc, zipname
				},
				assert: func(t *testing.T, gotErr error) {
					assert.Error(t, gotErr)
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				input, output, zipname := tt.setup(t)
				defer func() {
					if input != nil {
						input.Close()
					}
					if output != nil {
						output.Close()
					}
					if zipname != "" {
						os.Remove(zipname)
					}
				}()
				err := l.GeneratePDFFromZip(input, output)
				tt.assert(t, err)
			})
		}
	})
}

type WC struct {
	*bufio.Writer
}

func (wc *WC) Close() error {
	return nil
}
