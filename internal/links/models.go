package links

type LinkStatus string

const (
	LinkStatusAvailable    LinkStatus = "available"
	LinkStatusNotAvailable LinkStatus = "not available"
)

type LinkInformation struct {
	URL     string     `json:"url"`
	Status  LinkStatus `json:"status"`
	LinkNum int        `json:"link_num"`
}
