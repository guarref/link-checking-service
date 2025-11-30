package links

type Repository interface {
	Get(id int) ([]LinkInformation, bool, bool)
	Set(links []LinkInformation) int
	Update(id int, links []LinkInformation)
}
