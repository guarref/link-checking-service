package web

import (
	"github.com/guarref/link-checking-service/dto"
	"github.com/guarref/link-checking-service/internal/links"
)

func linkInformationToDTO(link links.LinkInformation) dto.LinkInformationDTO {
	return dto.LinkInformationDTO{
		URL:    link.URL,
		Status: dto.LinkStatus(link.Status),
	}
}

func linksInformationSliceToDTO(links []links.LinkInformation) []dto.LinkInformationDTO {
	
	res := make([]dto.LinkInformationDTO, 0, len(links))

	for _, it := range links {
		res = append(res, linkInformationToDTO(it))
	}

	return res
}
