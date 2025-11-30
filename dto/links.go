package dto

type LinkStatus string

const (
	LinkStatusAvailable    LinkStatus = "available"
	LinkStatusNotAvailable LinkStatus = "not available"
)

type LinkInformationDTO struct {
	URL    string     `json:"url"`
	Status LinkStatus `json:"status"`
}

type LinksToJSONRequestDTO struct {
	Links []string `json:"links"`
}

type LinksToJSONResponseDTO struct {
	Links    []LinkInformationDTO `json:"links"`
	LinksNum int                  `json:"links_num"`
}

type LinksToPDFRequestDTO struct {
	LinksList []int `json:"links_list"`
}
