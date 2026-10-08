package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/identity"
	"github.com/bangweiz/dubhu-designer/internal/mapper"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"github.com/bangweiz/dubhu-designer/internal/service/util"
	"github.com/bangweiz/dubhu-designer/internal/validator"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

const passwordCost = 12
const sessionLifetime = 24 * time.Hour

type AuthService struct {
	repo              *repository.AuthRepository
	dummyPasswordHash []byte
}

func NewAuthService(repo *repository.AuthRepository) (*AuthService, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("not-a-real-account-password"), passwordCost)
	if err != nil {
		return nil, err
	}
	return &AuthService{repo: repo, dummyPasswordHash: dummy}, nil
}

func (s *AuthService) CreateOrganisation(ctx context.Context, input dto.CreateOrganisationDTO) (*dto.CreateOrganisationResponseDTO, error) {
	input.Trim()
	if errs := validator.ValidateCreateOrganisation(&input); len(errs) > 0 {
		return nil, errs
	}
	password, err := bcrypt.GenerateFromPassword([]byte(input.RootAccount.Password), passwordCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Truncate(time.Millisecond)
	orgID, rootID := bson.NewObjectID(), bson.NewObjectID()
	org := &models.Organisation{ID: orgID, Name: input.Name, Description: input.Description, RootAccountID: rootID, AuditFields: models.AuditFields{CreatedBy: rootID, UpdatedBy: rootID, CreatedAt: now, UpdatedAt: now}}
	root := &models.Account{ID: rootID, OrganisationID: orgID, Name: input.RootAccount.Name, Email: input.RootAccount.Email, PasswordHash: string(password), Role: models.RoleRoot, AuditFields: models.AuditFields{CreatedBy: rootID, UpdatedBy: rootID, CreatedAt: now, UpdatedAt: now}}
	return util.RunInTransaction(ctx, s.repo.Client(), func(tx context.Context) (*dto.CreateOrganisationResponseDTO, error) {
		if err := s.repo.CreateOrganisation(tx, org); err != nil {
			if errors.Is(err, repository.ErrOrganisationNameExists) {
				return nil, ErrOrganisationNameExists
			}
			return nil, err
		}
		if err := s.repo.CreateAccount(tx, root); err != nil {
			return nil, err
		}
		return &dto.CreateOrganisationResponseDTO{Organisation: mapper.ToOrganisationResponseDTO(org), RootAccount: mapper.ToAccountResponseDTO(root)}, nil
	})
}
func (s *AuthService) CreateAccount(ctx context.Context, input dto.CreateAccountDTO) (*dto.AccountResponseDTO, error) {
	principal, ok := identity.FromContext(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}
	if principal.Role != string(models.RoleRoot) {
		return nil, ErrForbidden
	}
	input.Trim()
	if errs := validator.ValidateCreateAccount(&input); len(errs) > 0 {
		return nil, errs
	}
	password, err := bcrypt.GenerateFromPassword([]byte(input.Password), passwordCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Truncate(time.Millisecond)
	account := &models.Account{ID: bson.NewObjectID(), OrganisationID: principal.OrganisationID, Name: input.Name, Email: input.Email, PasswordHash: string(password), Role: input.Role, AuditFields: models.AuditFields{CreatedBy: principal.AccountID, UpdatedBy: principal.AccountID, CreatedAt: now, UpdatedAt: now}}
	if err := s.repo.CreateAccount(ctx, account); err != nil {
		if errors.Is(err, repository.ErrAccountConflict) {
			return nil, ErrAccountEmailExists
		}
		return nil, err
	}
	response := mapper.ToAccountResponseDTO(account)
	return &response, nil
}
func (s *AuthService) Login(ctx context.Context, orgIDStr string, input dto.LoginDTO) (*dto.LoginResponseDTO, error) {
	input.Trim()
	if errs := validator.ValidateLogin(&input); len(errs) > 0 {
		return nil, ErrUnauthenticated
	}
	orgID, parseErr := bson.ObjectIDFromHex(orgIDStr)
	var account *models.Account
	if parseErr == nil {
		var err error
		account, err = s.repo.GetAccountByEmail(ctx, orgID, input.Email)
		if err != nil {
			return nil, err
		}
	}
	hash := s.dummyPasswordHash
	if account != nil {
		hash = []byte(account.PasswordHash)
	}
	passwordErr := bcrypt.CompareHashAndPassword(hash, []byte(input.Password))
	if parseErr != nil || account == nil || passwordErr != nil {
		return nil, ErrUnauthenticated
	}
	if err := s.validateAccount(ctx, account); err != nil {
		return nil, err
	}
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(random[:])
	now := time.Now().UTC().Truncate(time.Millisecond)
	session := &models.AuthSession{TokenHash: hashToken(token), OrganisationID: orgID, AccountID: account.ID, CreatedAt: now, ExpiresAt: now.Add(sessionLifetime)}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	return &dto.LoginResponseDTO{AccessToken: token, TokenType: "Bearer", ExpiresAt: session.ExpiresAt, Account: mapper.ToAccountResponseDTO(account)}, nil
}
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
func (s *AuthService) Authenticate(ctx context.Context, token string) (*models.Account, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(decoded) != 32 {
		return nil, ErrUnauthenticated
	}
	session, err := s.repo.GetSession(ctx, hashToken(token))
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrUnauthenticated
	}
	account, err := s.repo.GetAccount(ctx, session.OrganisationID, session.AccountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrUnauthenticated
	}
	if err := s.validateAccount(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}
func (s *AuthService) validateAccount(ctx context.Context, account *models.Account) error {
	if account.Role != models.RoleRoot && account.Role != models.RoleAdmin && account.Role != models.RoleUser {
		return ErrUnauthenticated
	}
	org, err := s.repo.GetOrganisation(ctx, account.OrganisationID)
	if err != nil {
		return err
	}
	if org == nil || (account.Role == models.RoleRoot && org.RootAccountID != account.ID) {
		return ErrUnauthenticated
	}
	return nil
}
func (s *AuthService) Logout(ctx context.Context, token string) error {
	p, ok := identity.FromContext(ctx)
	if !ok {
		return ErrUnauthenticated
	}
	return s.repo.DeleteSession(ctx, p.OrganisationID, p.AccountID, hashToken(token))
}
