package links

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-pdf/fpdf"
)

type Service struct {
	repo   Repository
	client *http.Client
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:   repo,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *Service) ValidLinks(urls []string) (int, []LinkInformation) {

	if len(urls) == 0 {
		return 0, nil
	}

	res := s.checkURLs(urls)
	id := s.repo.Set(res)

	return id, res
}

func (s *Service) GetStatuses(id int) ([]LinkInformation, bool) {

	links, isExists, isExpired := s.repo.Get(id)
	if !isExists {
		return nil, false
	}

	if isExpired {
		var urls []string
		for _, l := range links {
			urls = append(urls, l.URL)
		}

		newRes := s.checkURLs(urls)
		s.repo.Update(id, newRes)

		return newRes, true
	}

	return links, true
}

func (s *Service) checkURLs(urls []string) []LinkInformation {

	if len(urls) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	res := make([]LinkInformation, len(urls))

	for i, rawUrl := range urls {
		wg.Add(1)
		go func(i int, rawURL string) {
			defer wg.Done()

			scheme := "https"
			host := rawURL

			if strings.Contains(rawURL, "://") {
				parsed, err := url.Parse(rawURL)
				if err == nil {
					scheme = parsed.Scheme
					host = parsed.Host
				}
			}

			status := checkURL(s.client, host, scheme)
			res[i] = LinkInformation{URL: rawURL, Status: status}
		}(i, rawUrl)
	}

	wg.Wait()

	return res
}

func checkURL(client *http.Client, page, scheme string) LinkStatus {

	rawURL := scheme + "://" + page

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return LinkStatusNotAvailable
	}

	if parsedURL.Scheme != scheme {
		return LinkStatusNotAvailable
	}

	resp, err := client.Get(rawURL)
	if err != nil {
		return LinkStatusNotAvailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return LinkStatusAvailable
	} else {
		return LinkStatusNotAvailable
	}
}

func (s *Service) GeneratePDF(ids []int) ([]byte, error) {

	pdf := fpdf.New(fpdf.OrientationPortrait, fpdf.UnitPoint, fpdf.PageSizeA4, "")
	pdf.AddPage()
	pdf.SetFont("Courier", "", 12)

	for _, id := range ids {
		links, found := s.GetStatuses(id)
		if !found || len(links) == 0 {
			continue
		}

		for _, link := range links {
			pdf.Cell(0, 12, fmt.Sprintf("%s - %s", link.URL, string(link.Status)))
			pdf.Ln(10)
		}
		pdf.Ln(6)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("error generating pdf: %w", err)
	}

	return buf.Bytes(), nil
}
