package groups

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Fista6k/disClone/internal/auth"
	"github.com/Fista6k/disClone/internal/httpapi"
)

type GroupHandler struct {
	groupService *GroupService
	groupHub     IClientGroupManager
}

func NewGroupHandler(groupService *GroupService, groupHub IClientGroupManager) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		groupHub:     groupHub,
	}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	owner_id, err := auth.UserIdFromContext(r.Context())
	if err != nil {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req CreateGroupRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	err = h.groupService.CreateGroup(r.Context(), req.Name, owner_id)
	if err != nil {
		httpapi.WriteDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *GroupHandler) GetMyGroups(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.UserIdFromContext(r.Context())
	if err != nil {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	groups, err := h.groupService.GetMyGroups(r.Context(), userID)
	if err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(groups)
	if err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}
}

func (h *GroupHandler) AddNewMembers(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "group_id must be a valid integer")
		return
	}

	var req AddMembersRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	err = h.groupService.AddNewMembers(r.Context(), groupID, req.MembersIDs)
	if err != nil {
		httpapi.WriteDomainError(w, err)
		return
	}

	for _, memberID := range req.MembersIDs {
		h.groupHub.AddClientToGroup(memberID, groupID)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) GetGroupInfo(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "group_id must be a valid integer")
		return
	}

	groupResponse, err := h.groupService.GetGroupInfo(r.Context(), groupID)
	if err != nil {
		httpapi.WriteDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(groupResponse); err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}
}

func (h *GroupHandler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("group_id")
	memberIDStr := r.PathValue("member_id")

	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "group_id must be a valid integer")
		return
	}

	memberID, err := strconv.ParseInt(memberIDStr, 10, 64)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "member_id must be a valid integer")
		return
	}

	userID, err := auth.UserIdFromContext(r.Context())
	if err != nil {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	err = h.groupService.DeleteMember(r.Context(), groupID, memberID, userID)
	if err != nil {
		httpapi.WriteDomainError(w, err)
		return
	}

	h.groupHub.RemoveClientFromGroup(memberID, groupID)

	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) GetGroupHistory(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "group_id must be a valid integer")
		return
	}

	messages, err := h.groupService.GetGroupHistory(r.Context(), groupID)
	if err != nil {
		httpapi.WriteDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(messages); err != nil {
		httpapi.WriteInternalError(w, err)
		return
	}
}
