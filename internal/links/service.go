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
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ValidLinks(urls []string) (int, []LinkInformation) {

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

			status := checkURL(host, scheme)
			res[i] = LinkInformation{URL: rawURL, Status: LinkStatus(status)}
		}(i, rawUrl)
	}

	wg.Wait()

	return res
}

func checkURL(page, scheme string) (string) {

	rawURL := scheme + "://" + page

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "not available"
	}

	if parsedURL.Scheme != scheme {
		return "not available"
	}

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return "not available"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "available"
	} else {
		return "not available"
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
		return nil, fmt.Errorf("error pdf parsing: %w", err)
	}

	return buf.Bytes(), nil
}