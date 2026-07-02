package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/permissions"
	"disaster_alert_backend/internal/repositories"
	
	"github.com/google/uuid"
)

type AdminPrivilegeCodeService struct {
	codeRepo *repositories.AdminPrivilegeCodeRepository
	logRepo  *repositories.AdminPrivilegeLogRepository
}

type PermissionCheckResult struct {
	Allowed bool
	Code    *models.AdminPrivilegeCode
	Message string
}

func NewAdminPrivilegeCodeService(
	codeRepo *repositories.AdminPrivilegeCodeRepository,
	logRepo *repositories.AdminPrivilegeLogRepository,
) *AdminPrivilegeCodeService {
	return &AdminPrivilegeCodeService{
		codeRepo: codeRepo,
		logRepo:  logRepo,
	}
}

func (s *AdminPrivilegeCodeService) CreatePrivilegeCode(
	ctx context.Context,
	req dto.CreateAdminPrivilegeCodeRequest,
	createdBy string,
	ipAddress string,
	userAgent string,
) (*dto.AdminPrivilegeCodeResponse, error) {
	req.Permissions = permissions.RemoveDuplicatePermissions(req.Permissions)

	invalid := permissions.InvalidPermissions(req.Permissions)
	if len(invalid) > 0 {
		return nil, fmt.Errorf("invalid permissions: %s", strings.Join(invalid, ", "))
	}

	rawUUID := uuid.NewString()
	codeHash := hashPrivilegeCode(rawUUID)
	codePrefix := privilegeCodePrefix(rawUUID)

	var expiresAt *time.Time

	if strings.TrimSpace(req.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, errors.New("expiresAt must be a valid RFC3339 datetime, for example 2026-07-28T12:00:00Z")
		}

		expiresAt = &parsed
	}

	now := time.Now().UTC()

	code := &models.AdminPrivilegeCode{
		CodeHash:         codeHash,
		CodePrefix:       codePrefix,
		Label:            strings.TrimSpace(req.Label),
		Purpose:          strings.TrimSpace(req.Purpose),
		OrganisationID:   strings.TrimSpace(req.OrganisationID),
		OrganisationName: strings.TrimSpace(req.OrganisationName),
		LevelID:          strings.TrimSpace(req.LevelID),
		LevelName:        strings.TrimSpace(req.LevelName),
		Permissions:      req.Permissions,
		Status:           models.PrivilegeCodeStatusActive,
		UsageCount:       0,
		ExpiresAt:        expiresAt,
		CreatedBy:        createdBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.codeRepo.Create(ctx, code); err != nil {
		return nil, err
	}

	_ = s.createLog(ctx, &models.AdminPrivilegeLog{
		PrivilegeCodeID:  code.ID.Hex(),
		CodePrefix:       code.CodePrefix,
		OrganisationID:   code.OrganisationID,
		OrganisationName: code.OrganisationName,
		LevelID:          code.LevelID,
		LevelName:        code.LevelName,
		Action:           models.PrivilegeLogCodeCreated,
		Allowed:          true,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Message:          "Privilege code created",
		Metadata: map[string]interface{}{
			"label":       code.Label,
			"purpose":     code.Purpose,
			"permissions": code.Permissions,
			"createdBy":   createdBy,
		},
		CreatedAt: now,
	})

	return toPrivilegeCodeResponse(code, rawUUID), nil
}

func (s *AdminPrivilegeCodeService) GetPrivilegeCodes(
	ctx context.Context,
	limit int64,
) ([]dto.AdminPrivilegeCodeResponse, error) {
	codes, err := s.codeRepo.GetAll(ctx, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.AdminPrivilegeCodeResponse, 0, len(codes))

	for _, code := range codes {
		copyCode := code
		responses = append(responses, *toPrivilegeCodeResponse(&copyCode, ""))
	}

	return responses, nil
}

func (s *AdminPrivilegeCodeService) GetPrivilegeCodeByID(
	ctx context.Context,
	id string,
) (*dto.AdminPrivilegeCodeResponse, error) {
	code, err := s.codeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if code == nil {
		return nil, errors.New("privilege code not found")
	}

	return toPrivilegeCodeResponse(code, ""), nil
}

func (s *AdminPrivilegeCodeService) ValidatePrivilegeCode(
	ctx context.Context,
	rawCode string,
	ipAddress string,
	userAgent string,
) (*dto.AdminPrivilegeCodeResponse, error) {
	code, err := s.getUsableCode(ctx, rawCode)
	if err != nil {
		_ = s.createLog(ctx, &models.AdminPrivilegeLog{
			CodePrefix: privilegeCodePrefix(rawCode),
			Action:     models.PrivilegeLogCodeValidated,
			Allowed:    false,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			Message:    err.Error(),
			CreatedAt:  time.Now().UTC(),
		})

		return nil, err
	}

	if err := s.codeRepo.MarkUsed(ctx, code.ID); err != nil {
		return nil, err
	}

	_ = s.createLog(ctx, &models.AdminPrivilegeLog{
		PrivilegeCodeID:  code.ID.Hex(),
		CodePrefix:       code.CodePrefix,
		OrganisationID:   code.OrganisationID,
		OrganisationName: code.OrganisationName,
		LevelID:          code.LevelID,
		LevelName:        code.LevelName,
		Action:           models.PrivilegeLogCodeValidated,
		Allowed:          true,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Message:          "Privilege code validated",
		CreatedAt:        time.Now().UTC(),
	})


	return toPrivilegeCodeResponse(code, ""), nil
}

func (s *AdminPrivilegeCodeService) RevokePrivilegeCode(
	ctx context.Context,
	id string,
	reason string,
	revokedBy string,
	ipAddress string,
	userAgent string,
) (*dto.AdminPrivilegeCodeResponse, error) {
	code, err := s.codeRepo.Revoke(ctx, id, revokedBy, reason)
	if err != nil {
		return nil, err
	}

	_ = s.createLog(ctx, &models.AdminPrivilegeLog{
		PrivilegeCodeID:  code.ID.Hex(),
		CodePrefix:       code.CodePrefix,
		OrganisationID:   code.OrganisationID,
		OrganisationName: code.OrganisationName,
		LevelID:          code.LevelID,
		LevelName:        code.LevelName,
		Action:           models.PrivilegeLogCodeRevoked,
		Allowed:          true,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Message:          "Privilege code revoked",
		Metadata: map[string]interface{}{
			"reason":    reason,
			"revokedBy": revokedBy,
		},
		CreatedAt: time.Now().UTC(),
	})


	return toPrivilegeCodeResponse(code, ""), nil
}

func (s *AdminPrivilegeCodeService) CheckPermission(
	ctx context.Context,
	rawCode string,
	requiredPermission string,
	endpoint string,
	method string,
	ipAddress string,
	userAgent string,
) (*PermissionCheckResult, error) {
	code, err := s.getUsableCode(ctx, rawCode)
	if err != nil {
		_ = s.createLog(ctx, &models.AdminPrivilegeLog{
			CodePrefix:         privilegeCodePrefix(rawCode),
			Action:             models.PrivilegeLogPermissionCheck,
			Endpoint:           endpoint,
			Method:             method,
			RequiredPermission: requiredPermission,
			Allowed:            false,
			IPAddress:          ipAddress,
			UserAgent:          userAgent,
			Message:            err.Error(),
			CreatedAt:          time.Now().UTC(),
		})

		return &PermissionCheckResult{
			Allowed: false,
			Message: err.Error(),
		}, nil
	}

	allowed := permissions.ContainsPermission(code.Permissions, requiredPermission)

	message := "Permission allowed"
	if !allowed {
		message = "You do not have permission to perform this action"
	}

	_ = s.createLog(ctx, &models.AdminPrivilegeLog{
		PrivilegeCodeID:    code.ID.Hex(),
		CodePrefix:         code.CodePrefix,
		OrganisationID:     code.OrganisationID,
		OrganisationName:   code.OrganisationName,
		LevelID:            code.LevelID,
		LevelName:          code.LevelName,
		Action:             models.PrivilegeLogPermissionCheck,
		Endpoint:           endpoint,
		Method:             method,
		RequiredPermission: requiredPermission,
		Allowed:            allowed,
		IPAddress:          ipAddress,
		UserAgent:          userAgent,
		Message:            message,
		CreatedAt:          time.Now().UTC(),
	})

	if allowed {
		_ = s.codeRepo.MarkUsed(ctx, code.ID)
	}

	return &PermissionCheckResult{
		Allowed: allowed,
		Code:    code,
		Message: message,
	}, nil
}

func (s *AdminPrivilegeCodeService) GetPrivilegeLogs(
	ctx context.Context,
	limit int64,
	privilegeCodeID string,
) ([]models.AdminPrivilegeLog, error) {
	return s.logRepo.GetAll(ctx, limit, privilegeCodeID)
}

func (s *AdminPrivilegeCodeService) getUsableCode(
	ctx context.Context,
	rawCode string,
) (*models.AdminPrivilegeCode, error) {
	rawCode = strings.TrimSpace(rawCode)
	if rawCode == "" {
		return nil, errors.New("privilege code is required")
	}

	codeHash := hashPrivilegeCode(rawCode)

	code, err := s.codeRepo.FindByCodeHash(ctx, codeHash)
	if err != nil {
		return nil, err
	}

	if code == nil {
		return nil, errors.New("privilege code not found")
	}

	if code.Status != models.PrivilegeCodeStatusActive {
		return nil, errors.New("privilege code is not active")
	}

	if code.ExpiresAt != nil && time.Now().UTC().After(*code.ExpiresAt) {
		return nil, errors.New("privilege code has expired")
	}

	return code, nil
}

func (s *AdminPrivilegeCodeService) createLog(
	ctx context.Context,
	log *models.AdminPrivilegeLog,
) error {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}

	return s.logRepo.Create(ctx, log)
}

func hashPrivilegeCode(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}

func privilegeCodePrefix(raw string) string {
	raw = strings.TrimSpace(raw)

	if len(raw) <= 12 {
		return raw
	}

	return raw[:8] + "..." + raw[len(raw)-4:]
}

func toPrivilegeCodeResponse(
	code *models.AdminPrivilegeCode,
	rawUUID string,
) *dto.AdminPrivilegeCodeResponse {
	return &dto.AdminPrivilegeCodeResponse{
		ID:               code.ID.Hex(),
		UUID:             rawUUID,
		CodePrefix:       code.CodePrefix,
		Label:            code.Label,
		Purpose:          code.Purpose,
		OrganisationID:   code.OrganisationID,
		OrganisationName: code.OrganisationName,
		LevelID:          code.LevelID,
		LevelName:        code.LevelName,
		Permissions:      code.Permissions,
		Status:           code.Status,
		UsageCount:       code.UsageCount,
		LastUsedAt:       code.LastUsedAt,
		ExpiresAt:        code.ExpiresAt,
		CreatedBy:        code.CreatedBy,
		CreatedAt:        code.CreatedAt,
		UpdatedAt:        code.UpdatedAt,
	}
}

