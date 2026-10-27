package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"news-portal/apps/api/internal/pkg/response"
	"news-portal/apps/api/internal/usecase"
)

type AdminMediaHandler struct {
	mediaUsecase *usecase.MediaAdminUsecase
}

func NewAdminMediaHandler(mediaUsecase *usecase.MediaAdminUsecase) *AdminMediaHandler {
	return &AdminMediaHandler{mediaUsecase: mediaUsecase}
}

func (h *AdminMediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form up to 25MB
	if err := r.ParseMultipartForm(25 * 1024 * 1024); err != nil {
		response.Error(w, http.StatusBadRequest, "FORM_PARSE_ERROR", "Failed to parse multipart form", nil)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "FILE_REQUIRED", "No file found under 'file' key", nil)
		return
	}
	defer file.Close()

	altText := r.FormValue("alt_text")
	if altText == "" {
		response.Error(w, http.StatusBadRequest, "ALT_TEXT_REQUIRED", "Alt text is strictly required for media accessibility", nil)
		return
	}

	caption := r.FormValue("caption")
	photographer := r.FormValue("photographer")

	resp, err := h.mediaUsecase.Upload(r.Context(), file, header.Filename, altText, caption, photographer)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "UPLOAD_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusCreated, resp, nil)
}

func (h *AdminMediaHandler) List(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := int32(20)
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = int32(l)
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := int32(0)
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = int32(o)
	}

	items, err := h.mediaUsecase.List(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "FETCH_MEDIA_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, items, nil)
}

func (h *AdminMediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Media ID must be a number", nil)
		return
	}

	if err := h.mediaUsecase.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "DELETE_MEDIA_FAILED", err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "media deleted successfully"}, nil)
}
