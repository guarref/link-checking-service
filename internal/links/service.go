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

// CheckLinks checks a list of URLs and saves the result.
func (s *Service) CheckLinks(urls []string) (int, []LinkInformation) {
	results := s.checkURLs(urls)
	id := s.repo.Set(results)
	return id, results
}

// GetLinks retrieves links by ID. If expired, it re-checks and updates them.
func (s *Service) GetLinks(id int) ([]LinkInformation, bool) {
	links, isExists, expired := s.repo.Get(id)
	if !isExists {
		return nil, false
	}

	if expired {
		// Extract URLs from the expired data
		var urls []string
		for _, l := range links {
			urls = append(urls, l.URL)
		}

		// Re-check
		newResults := s.checkURLs(urls)

		// Update storage
		s.repo.Update(id, newResults)

		return newResults, true
	}

	return links, true
}

func (s *Service) checkURLs(urls []string) []LinkInformation {
	var wg sync.WaitGroup
	results := make([]LinkInformation, len(urls))

	for i, u := range urls {
		wg.Add(1)
		go func(index int, rawURL string) {
			defer wg.Done()

			// Determine scheme and host for the user's checkURL function
			// Assuming input might be "google.com" or "https://google.com"
			scheme := "http"
			host := rawURL

			if strings.Contains(rawURL, "://") {
				parsed, err := url.Parse(rawURL)
				if err == nil {
					scheme = parsed.Scheme
					host = parsed.Host
				}
			}

			status, _ := checkURL(host, scheme)

			results[index] = LinkInformation{
				URL:    rawURL,
				Status: LinkStatus(status), // Casting string to LinkStatus
			}
		}(i, u)
	}

	wg.Wait()
	return results
}

// User's provided checkURL function (slightly adapted for package usage if needed)
func checkURL(page, scheme string) (string, time.Duration) {
	rawURL := scheme + "://" + page

	// Fix: url.Parse might fail if page contains path, but assuming simple host for now
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "not available", 0
	}

	// Basic validation
	if parsedURL.Scheme != scheme {
		return "not available", 0
	}

	start := time.Now()
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(rawURL)
	duration := time.Since(start)
	if err != nil {
		return "not available", duration
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "available", duration
	} else {
		return "not available", duration
	}
}

// GenerateReport generates a PDF report for the given IDs.
func (s *Service) GenerateReport(ids []int) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Link Check Report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)

	for _, id := range ids {
		links, found := s.GetLinks(id)
		if !found {
			continue
		}

		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 10, fmt.Sprintf("Request ID: %d", id))
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		for _, link := range links {
			pdf.Cell(100, 8, link.URL)
			pdf.Cell(50, 8, string(link.Status))
			pdf.Ln(6)
		}
		pdf.Ln(4)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
