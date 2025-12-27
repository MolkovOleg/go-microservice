package handlers

import (
	"context"
	"encoding/json"
	"go-microservice/services"
	"go-microservice/utils"
	"net/http"
	"time"
)

type IntegrationHandler struct {
	service *services.IntegrationService
}

func NewIntegrationHandler(service *services.IntegrationService) *IntegrationHandler {
	return &IntegrationHandler{service: service}
}

type uploadRequest struct {
	Bucket      string `json:"bucket,omitempty"`
	ObjectName  string `json:"object_name"`
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
}
type presignRequest struct {
	Bucket        string `json:"bucket,omitempty"`
	ObjectName    string `json:"object_name"`
	ExpirySeconds int    `json:"expiry_seconds,omitempty"`
}

func (h *IntegrationHandler) UploadObject(w http.ResponseWriter, r *http.Request) {
	var req uploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		go utils.LogError("UploadObject", err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	info, err := h.service.UploadObject(ctx, req.Bucket, req.ObjectName, []byte(req.Content), req.ContentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		go utils.LogError("UploadObject", err)
		return
	}
	go utils.LogUserAction("UPLOAD_OBJECT", 0)
	response := map[string]interface{}{
		"bucket":      info.Bucket,
		"object_name": req.ObjectName,
		"etag":        info.ETag,
		"size":        info.Size,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}
func (h *IntegrationHandler) GetPresignedURL(w http.ResponseWriter, r *http.Request) {
	var req presignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		go utils.LogError("GetPresignedURL", err)
		return
	}
	expiry := time.Duration(req.ExpirySeconds) * time.Second
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	url, err := h.service.PresignedURL(ctx, req.Bucket, req.ObjectName, expiry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		go utils.LogError("GetPresignedURL", err)
		return
	}
	go utils.LogUserAction("PRESIGN_OBJECT", 0)
	response := map[string]interface{}{
		"url":            url,
		"expiry_seconds": req.ExpirySeconds,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}