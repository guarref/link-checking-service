package web

import (
	"encoding/json"
	"net/http"

	"github.com/guarref/link-checking-service/dto"
	"github.com/guarref/link-checking-service/internal/links"
)

type Handler struct {
	service *links.Service
}

func NewHandler(service *links.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetStatusToJSON(rw http.ResponseWriter, r *http.Request) {

	var req dto.LinksToJSONRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	id, res := h.service.ValidLinks(req.Links)

	resp := dto.LinksToJSONResponseDTO{
		LinksNum: id,
		Links:    linksInformationSliceToDTO(res),
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(resp)
}

func (h *Handler) GetStatusToPDF(rw http.ResponseWriter, r *http.Request) {

	var req dto.LinksToPDFRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	pdfBytes, err := h.service.GeneratePDF(req.LinksList)
	if err != nil {
		http.Error(rw, "error generating report", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/pdf")
	rw.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
	rw.Write(pdfBytes)
}
