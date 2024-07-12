package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"tecdsa/pkg/database/repository"
	"tecdsa/pkg/utils"
)

type RegisterClientSecurityRequest struct {
	PublicKey string `json:"public_key"`
}

type RegisterClientSecurityResponse struct {
	ID uint32 `json:"id"`
}

type RegisterClientSecurityHandler struct {
	clientSecurityRepo repository.ClientSecurityRepository
}

func NewRegisterClientSecurityHandler(repo repository.ClientSecurityRepository) *RegisterClientSecurityHandler {
	return &RegisterClientSecurityHandler{
		clientSecurityRepo: repo,
	}
}

func (h *RegisterClientSecurityHandler) Serve(w http.ResponseWriter, r *http.Request) {
	var req RegisterClientSecurityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// // 공개키 유효성 검사
	// if err := utils.ValidatePublicKey(req.PublicKey); err != nil {
	// 	http.Error(w, "Invalid public key: "+err.Error(), http.StatusBadRequest)
	// 	return
	// }

	// IP 주소 추출
	ip := utils.GetClientIP(r)

	// IP로 기존 레코드 조회
	existingRecord, err := h.clientSecurityRepo.FindByIP(ip)
	if err != nil {
		// 데이터베이스 조회 중 오류 발생
		if err != sql.ErrNoRows { // 레코드가 없는 경우가 아닌 다른 오류
			http.Error(w, "Failed to check existing record", http.StatusInternalServerError)
			return
		}
		// sql.ErrNoRows 오류는 무시하고 계속 진행 (레코드가 없음을 의미)
	}

	if existingRecord != nil {
		// 이미 존재하는 IP인 경우 에러 반환
		http.Error(w, "IP already registered", http.StatusConflict)
		return
	}

	// 새 ClientSecurity 레코드 생성
	newClientSecurity, err := h.clientSecurityRepo.Create(req.PublicKey, ip)
	if err != nil {
		http.Error(w, "Failed to register client security", http.StatusInternalServerError)
		return
	}

	// 응답 생성
	resp := RegisterClientSecurityResponse{ID: newClientSecurity.ID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
