package groups

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Fista6k/disClone/internal/auth"
)

type GroupHandler struct {
	groupService *GroupService
}

func NewGroupHandler(groupService *GroupService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	owner_id, err := auth.UserIdFromContext(r.Context())
	if err != nil {
		http.Error(w, "error with auth", http.StatusUnauthorized)
		return
	}

	var req CreateGroupRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	err = h.groupService.CreateGroup(r.Context(), req.Name, owner_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *GroupHandler) GetGroups(w http.ResponseWriter, r *http.Request) {
	owner_id, err := auth.UserIdFromContext(r.Context())
	if err != nil {
		http.Error(w, "error with auth", http.StatusUnauthorized)
		return
	}

	groups, err := h.groupService.GetGroups(r.Context(), owner_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(groups)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *GroupHandler) AddNewMembers(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid group_id value", http.StatusBadRequest)
		return
	}

	var req AddMembersRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	err = h.groupService.AddNewMembers(r.Context(), groupID, req.MembersIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
