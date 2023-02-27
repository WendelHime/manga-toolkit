// Package logic contains the logic used on manga-toolkit clients
package logic

import (
	"archive/zip"
	"context"
	"io"

	"github.com/WendelHime/manga-toolkit/internal/services"
)

// Logic abstracts the logic used on manga-toolkit
type Logic interface {
	GeneratePDFFromZip(ctx context.Context, reader *zip.ReadCloser, output io.WriteCloser) error
	DownloadChapters(ctx context.Context, mangaTerm string, outputDir string, fromChapter int, toChapter int) error
}

type logic struct {
	mangaFreakService services.MangaFreakService
}

// NewLogic builds the logic used on manga-toolkit
func NewLogic(mangaFreakService services.MangaFreakService) Logic {
	return logic{mangaFreakService: mangaFreakService}
}
