package links

type Repository interface {
	Get(id int) (links []LinkInformation, isExists bool, expired bool)
	Set(links []LinkInformation) int
	Update(id int, links []LinkInformation)
}
