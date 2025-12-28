package handlers

import (
	"context"
	"encoding/json"
	"go-microservice/models"
	"go-microservice/services"
	"go-microservice/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)


type UserHandler struct {
	service      *services.UserService
	integration  *services.IntegrationService
	exportBucket string
}

func NewUserHandler(service *services.UserService, integration *services.IntegrationService, exportBucket string) *UserHandler {
	return &UserHandler{
		service:      service,
		integration:  integration,
		exportBucket: exportBucket,
	}
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users := h.service.GetAll()
	go utils.LogUserAction("GET_ALL_USERS", 0)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		go utils.LogError("GET_BY_ID", err)
		return
	}
	user, err := h.service.GetById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		go utils.LogError("GET_BY_ID", err)
		return
	}
	go utils.LogUserAction("GET_BY_ID", user.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		go utils.LogError("CREATE_USER", err)
		return
	}
	savedUser, err := h.service.Create(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		go utils.LogError("CREATE_USER", err)
		return
	}
	go utils.LogUserAction("CREATE_USER", savedUser.ID)
	go utils.SendNotification(savedUser.ID, "User created successfully")
	h.exportUserSnapshot(savedUser)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(savedUser)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		go utils.LogError("UPDATE_USER", err)
		return
	}
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		go utils.LogError("UPDATE_USER", err)
		return
	}
	updatedUser, err := h.service.Update(id, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		go utils.LogError("UPDATE_USER", err)
		return
	}
	go utils.LogUserAction("UPDATE_USER", updatedUser.ID)
	go utils.SendNotification(updatedUser.ID, "User updated successfully")
	h.exportUserSnapshot(updatedUser)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		go utils.LogError("DELETE_USER", err)
		return
	}
	err = h.service.Delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		go utils.LogError("DELETE_USER", err)
		return
	}
	go utils.LogUserAction("DELETE_USER", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) exportUserSnapshot(user *models.User) {
	if h.integration == nil || user == nil {
		return
	}
	userCopy := *user
	go func(u models.User) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := h.integration.ExportUserSnapshot(ctx, h.exportBucket, &u); err != nil {
			utils.LogError("ExportUserSnapshot", err)
		}
	}(userCopy)
}