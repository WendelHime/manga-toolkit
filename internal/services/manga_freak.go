package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type MangaFreakService interface {
	DownloadZip(ctx context.Context, chapter string, output io.Writer) error
}

type service struct {
	client   *http.Client
	endpoint string
}

func NewMangaFreakService(endpoint string) MangaFreakService {
	return &service{
		client: &http.Client{
			CheckRedirect: func(r *http.Request, _ []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		},
		endpoint: endpoint,
	}
}

func (s *service) DownloadZip(ctx context.Context, chapter string, output io.Writer) error {
	chapterURL := fmt.Sprintf("%s/%s", s.endpoint, chapter)

	request, err := http.NewRequestWithContext(ctx, "GET", chapterURL, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(output, resp.Body)
	return err
}
