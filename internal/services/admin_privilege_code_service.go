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

	"go.mongodb.org/mongo-driver/bson"

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
	accessMode := normalizePrivilegeAccessMode(req.AccessMode)

	grants, grantActions, err := normalizePrivilegeGrants(req.Grants, accessMode)
	if err != nil {
		return nil, err
	}

	combinedPermissions := permissions.RemoveDuplicatePermissions(
		append(req.Permissions, grantActions...),
	)

	if len(combinedPermissions) == 0 {
		return nil, errors.New("at least one permission or one grant action is required")
	}

	invalid := permissions.InvalidPermissions(combinedPermissions)
	if len(invalid) > 0 {
		return nil, fmt.Errorf("invalid permissions: %s", strings.Join(invalid, ", "))
	}

	if accessMode != models.PrivilegeAccessModeGlobal &&
		strings.TrimSpace(req.OrganisationID) == "" {
		return nil, errors.New("organisationId is required unless accessMode is global")
	}

	rawUUID := uuid.NewString()
	codeHash := hashPrivilegeCode(rawUUID)
	codePrefix := privilegeCodePrefix(rawUUID)

	var expiresAt *time.Time

	if strings.TrimSpace(req.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, errors.New("expiresAt must be a valid RFC3339 datetime, for example 2026-09-30T23:59:00Z")
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
		OrganisationType: strings.TrimSpace(req.OrganisationType),
		LevelID:          strings.TrimSpace(req.LevelID),
		LevelName:        strings.TrimSpace(req.LevelName),
		Permissions:      combinedPermissions,
		Grants:           grants,
		AccessMode:       accessMode,
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
			"label":            code.Label,
			"purpose":          code.Purpose,
			"permissions":      code.Permissions,
			"grants":           code.Grants,
			"accessMode":       code.AccessMode,
			"organisationType": code.OrganisationType,
			"createdBy":        createdBy,
		},
		CreatedAt: now,
	})

	return toPrivilegeCodeResponse(code, rawUUID), nil
}

func (s *AdminPrivilegeCodeService) UpdatePrivilegeCode(
	ctx context.Context,
	id string,
	req dto.UpdateAdminPrivilegeCodeRequest,
	updatedBy string,
	ipAddress string,
	userAgent string,
) (*dto.AdminPrivilegeCodeResponse, error) {
	existing, err := s.codeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, errors.New("privilege code not found")
	}

	if existing.Status != models.PrivilegeCodeStatusActive {
		return nil, errors.New("only active privilege codes can be updated")
	}

	nextOrganisationID := existing.OrganisationID
	nextAccessMode := normalizePrivilegeAccessMode(existing.AccessMode)

	if req.OrganisationID != nil {
		nextOrganisationID = strings.TrimSpace(*req.OrganisationID)
	}

	if req.AccessMode != nil {
		nextAccessMode = normalizePrivilegeAccessMode(*req.AccessMode)
	}

	nextPermissions := existing.Permissions
	nextGrants := existing.Grants

	if req.Permissions != nil {
		nextPermissions = permissions.RemoveDuplicatePermissions(req.Permissions)

		invalid := permissions.InvalidPermissions(nextPermissions)
		if len(invalid) > 0 {
			return nil, fmt.Errorf("invalid permissions: %s", strings.Join(invalid, ", "))
		}
	}

	if req.Grants != nil {
		grants, grantActions, err := normalizePrivilegeGrants(req.Grants, nextAccessMode)
		if err != nil {
			return nil, err
		}

		nextGrants = grants
		nextPermissions = permissions.RemoveDuplicatePermissions(
			append(nextPermissions, grantActions...),
		)
	}

	if len(nextPermissions) == 0 {
		return nil, errors.New("at least one permission or one grant action is required")
	}

	if nextAccessMode != models.PrivilegeAccessModeGlobal &&
		strings.TrimSpace(nextOrganisationID) == "" {
		return nil, errors.New("organisationId is required unless accessMode is global")
	}

	setFields := bson.M{
		"organisationId": nextOrganisationID,
		"permissions":    nextPermissions,
		"grants":         nextGrants,
		"accessMode":     nextAccessMode,
	}

	unsetFields := bson.M{}

	if req.Label != nil {
		setFields["label"] = strings.TrimSpace(*req.Label)
	}

	if req.Purpose != nil {
		setFields["purpose"] = strings.TrimSpace(*req.Purpose)
	}

	if req.OrganisationName != nil {
		setFields["organisationName"] = strings.TrimSpace(*req.OrganisationName)
	}

	if req.OrganisationType != nil {
		setFields["organisationType"] = strings.TrimSpace(*req.OrganisationType)
	}

	if req.LevelID != nil {
		setFields["levelId"] = strings.TrimSpace(*req.LevelID)
	}

	if req.LevelName != nil {
		setFields["levelName"] = strings.TrimSpace(*req.LevelName)
	}

	if req.ExpiresAt != nil {
		value := strings.TrimSpace(*req.ExpiresAt)

		if value == "" {
			unsetFields["expiresAt"] = ""
		} else {
			parsed, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return nil, errors.New("expiresAt must be a valid RFC3339 datetime, for example 2026-09-30T23:59:00Z")
			}

			setFields["expiresAt"] = &parsed
		}
	}

	updated, err := s.codeRepo.Update(ctx, id, setFields, unsetFields)
	if err != nil {
		return nil, err
	}

	if updated == nil {
		return nil, errors.New("privilege code not found")
	}

	_ = s.createLog(ctx, &models.AdminPrivilegeLog{
		PrivilegeCodeID:  updated.ID.Hex(),
		CodePrefix:       updated.CodePrefix,
		OrganisationID:   updated.OrganisationID,
		OrganisationName: updated.OrganisationName,
		LevelID:          updated.LevelID,
		LevelName:        updated.LevelName,
		Action:           "code_updated",
		Allowed:          true,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Message:          "Privilege code updated",
		Metadata: map[string]interface{}{
			"updatedBy":        updatedBy,
			"permissions":      updated.Permissions,
			"grants":           updated.Grants,
			"accessMode":       updated.AccessMode,
			"organisationType": updated.OrganisationType,
		},
		CreatedAt: time.Now().UTC(),
	})

	return toPrivilegeCodeResponse(updated, ""), nil
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
	grants := make([]dto.PrivilegeGrantResponse, 0, len(code.Grants))

	for _, grant := range code.Grants {
		grants = append(grants, dto.PrivilegeGrantResponse{
			CategoryID:   grant.CategoryID,
			CategorySlug: grant.CategorySlug,
			CategoryName: grant.CategoryName,
			Actions:      grant.Actions,
			AccessMode:   grant.AccessMode,
			Countries:    grant.Countries,
			Regions:      grant.Regions,
			Districts:    grant.Districts,
		})
	}

	return &dto.AdminPrivilegeCodeResponse{
		ID:               code.ID.Hex(),
		UUID:             rawUUID,
		CodePrefix:       code.CodePrefix,
		Label:            code.Label,
		Purpose:          code.Purpose,
		OrganisationID:   code.OrganisationID,
		OrganisationName: code.OrganisationName,
		OrganisationType: code.OrganisationType,
		LevelID:          code.LevelID,
		LevelName:        code.LevelName,
		Permissions:      code.Permissions,
		Grants:           grants,
		AccessMode:       code.AccessMode,
		Status:           code.Status,
		UsageCount:       code.UsageCount,
		LastUsedAt:       code.LastUsedAt,
		ExpiresAt:        code.ExpiresAt,
		CreatedBy:        code.CreatedBy,
		CreatedAt:        code.CreatedAt,
		UpdatedAt:        code.UpdatedAt,
	}
}

func normalizePrivilegeAccessMode(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case models.PrivilegeAccessModeGlobal:
		return models.PrivilegeAccessModeGlobal
	case models.PrivilegeAccessModeOwnedOnly:
		return models.PrivilegeAccessModeOwnedOnly
	case models.PrivilegeAccessModeScoped:
		return models.PrivilegeAccessModeScoped
	case models.PrivilegeAccessModeAssignedOnly:
		return models.PrivilegeAccessModeAssignedOnly
	default:
		return models.PrivilegeAccessModeAssignedOnly
	}
}

func normalizePrivilegeGrants(
	input []dto.PrivilegeGrantRequest,
	defaultAccessMode string,
) ([]models.PrivilegeGrant, []string, error) {
	grants := make([]models.PrivilegeGrant, 0, len(input))
	allActions := []string{}

	for _, item := range input {
		actions := permissions.RemoveDuplicatePermissions(item.Actions)

		if len(actions) == 0 {
			return nil, nil, errors.New("each privilege grant must contain at least one action")
		}

		invalid := permissions.InvalidPermissions(actions)
		if len(invalid) > 0 {
			return nil, nil, fmt.Errorf("invalid grant actions: %s", strings.Join(invalid, ", "))
		}

		grantAccessMode := normalizePrivilegeAccessMode(item.AccessMode)
		if strings.TrimSpace(item.AccessMode) == "" {
			grantAccessMode = defaultAccessMode
		}

		grants = append(grants, models.PrivilegeGrant{
			CategoryID:   strings.TrimSpace(item.CategoryID),
			CategorySlug: normalizeCategorySlug(item.CategorySlug),
			CategoryName: strings.TrimSpace(item.CategoryName),
			Actions:      actions,
			AccessMode:   grantAccessMode,
			Countries:    normalizeStringList(item.Countries),
			Regions:      normalizeStringList(item.Regions),
			Districts:    normalizeStringList(item.Districts),
		})

		allActions = append(allActions, actions...)
	}

	return grants, permissions.RemoveDuplicatePermissions(allActions), nil
}

func normalizeStringList(input []string) []string {
	result := []string{}
	seen := map[string]bool{}

	for _, value := range input {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		key := strings.ToLower(value)
		if seen[key] {
			continue
		}

		seen[key] = true
		result = append(result, value)
	}

	return result
}

func normalizeCategorySlug(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}
