package logic

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

func (l logic) DownloadChapters(ctx context.Context, mangaTerm string, outputPath string, fromChapter, toChapter int) error {
	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(4)
	for i := fromChapter; i <= toChapter; i++ {
		chapter := i
		group.Go(func() error {
			chapterName := fmt.Sprintf("%s_%d", mangaTerm, chapter)

			outputFilename := fmt.Sprintf("%s.zip", chapterName)
			outputFilepath := filepath.Join(outputPath, outputFilename)

			file, err := os.Create(outputFilepath)
			if err != nil {
				return err
			}
			defer file.Close()

			l.mangaFreakService.DownloadZip(ctx, chapterName, file)
			log.Printf("Downloaded [%s] on [%s]", chapterName, outputFilepath)
			return nil
		})
	}
	return group.Wait()
}
