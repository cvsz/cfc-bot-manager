package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/cvsz/cfc-bot-manager/backend/internal/domain"
	"github.com/cvsz/cfc-bot-manager/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Provider implements RepositoryProvider
type Provider struct {
	pool *pgxpool.Pool

	// Repositories (concrete types for internal access)
	organizations   *organizationRepo
	users           *userRepo
	memberships     *membershipRepo
	roles           *roleRepo
	permissions     *permissionRepo
	sessions        *sessionRepo
	rooms           *roomRepo
	bots            *botRepo
	plugins         *pluginRepo
	pluginConfigs   *pluginConfigRepo
	auditLogs       *auditLogRepo
	apiKeys         *apiKeyRepo
	botConnections  *botConnectionRepo
}

func NewProvider(pool *pgxpool.Pool) *Provider {
	return &Provider{pool: pool}
}

// Repository accessors (non-transactional)
func (p *Provider) Organizations() repository.OrganizationRepository {
	if p.organizations == nil {
		p.organizations = NewOrganizationRepository(p.pool)
	}
	return p.organizations
}

func (p *Provider) Users() repository.UserRepository {
	if p.users == nil {
		p.users = NewUserRepository(p.pool)
	}
	return p.users
}

func (p *Provider) Memberships() repository.MembershipRepository {
	if p.memberships == nil {
		p.memberships = NewMembershipRepository(p.pool)
	}
	return p.memberships
}

func (p *Provider) Roles() repository.RoleRepository {
	if p.roles == nil {
		p.roles = NewRoleRepository(p.pool)
	}
	return p.roles
}

func (p *Provider) Permissions() repository.PermissionRepository {
	if p.permissions == nil {
		p.permissions = NewPermissionRepository(p.pool)
	}
	return p.permissions
}

func (p *Provider) Sessions() repository.SessionRepository {
	if p.sessions == nil {
		p.sessions = NewSessionRepository(p.pool)
	}
	return p.sessions
}

func (p *Provider) Rooms() repository.RoomRepository {
	if p.rooms == nil {
		p.rooms = NewRoomRepository(p.pool)
	}
	return p.rooms
}

func (p *Provider) Bots() repository.BotRepository {
	if p.bots == nil {
		p.bots = NewBotRepository(p.pool)
	}
	return p.bots
}

func (p *Provider) Plugins() repository.PluginRepository {
	if p.plugins == nil {
		p.plugins = NewPluginRepository(p.pool)
	}
	return p.plugins
}

func (p *Provider) PluginConfigs() repository.PluginConfigRepository {
	if p.pluginConfigs == nil {
		p.pluginConfigs = NewPluginConfigRepository(p.pool)
	}
	return p.pluginConfigs
}

func (p *Provider) AuditLogs() repository.AuditLogRepository {
	if p.auditLogs == nil {
		p.auditLogs = NewAuditLogRepository(p.pool)
	}
	return p.auditLogs
}

func (p *Provider) APIKeys() repository.APIKeyRepository {
	if p.apiKeys == nil {
		p.apiKeys = NewAPIKeyRepository(p.pool)
	}
	return p.apiKeys
}

func (p *Provider) BotConnections() repository.BotConnectionRepository {
	if p.botConnections == nil {
		p.botConnections = NewBotConnectionRepository(p.pool)
	}
	return p.botConnections
}

// WithTransaction executes a function within a database transaction
func (p *Provider) WithTransaction(ctx context.Context, fn func(repository.UnitOfWork) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}

	txProvider := &txProvider{
		tx:     tx,
		parent: p,
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(txProvider); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// txProvider implements UnitOfWork for transactional operations
type txProvider struct {
	tx     pgx.Tx
	parent *Provider
}

func (t *txProvider) Begin(ctx context.Context) (repository.UnitOfWork, error) {
	// Create savepoint for nested transaction
	savepointName := fmt.Sprintf("sp_%d", time.Now().UnixNano())
	_, err := t.tx.Exec(ctx, "SAVEPOINT "+savepointName)
	if err != nil {
		return nil, err
	}
	return &nestedTxProvider{parent: t, savepointName: savepointName}, nil
}

// nestedTxProvider implements nested transaction support via savepoints
type nestedTxProvider struct {
	parent        *txProvider
	savepointName string
}

func (n *nestedTxProvider) Begin(ctx context.Context) (repository.UnitOfWork, error) {
	// Create another savepoint for deeper nesting
	savepointName := fmt.Sprintf("sp_%d", time.Now().UnixNano())
	_, err := n.parent.tx.Exec(ctx, "SAVEPOINT "+savepointName)
	if err != nil {
		return nil, err
	}
	return &nestedTxProvider{parent: n.parent, savepointName: savepointName}, nil
}

func (n *nestedTxProvider) Organizations() repository.OrganizationRepository {
	return &txOrgRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) Users() repository.UserRepository {
	return &txUserRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) Memberships() repository.MembershipRepository {
	return &txMembershipRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) Roles() repository.RoleRepository {
	return &txRoleRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) Permissions() repository.PermissionRepository {
	return &txPermissionRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) Sessions() repository.SessionRepository {
	return &txSessionRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) Rooms() repository.RoomRepository {
	return &txRoomRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) Bots() repository.BotRepository {
	return &txBotRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) Plugins() repository.PluginRepository {
	return &txPluginRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) PluginConfigs() repository.PluginConfigRepository {
	return &txPluginConfigRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) AuditLogs() repository.AuditLogRepository {
	return &txAuditLogRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) APIKeys() repository.APIKeyRepository {
	return &txAPIKeyRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) BotConnections() repository.BotConnectionRepository {
	return &txBotConnectionRepo{tx: n.parent.tx}
}

func (n *nestedTxProvider) Commit(ctx context.Context) error {
	// Release savepoint
	_, err := n.parent.tx.Exec(ctx, "RELEASE SAVEPOINT "+n.savepointName)
	return err
}

func (n *nestedTxProvider) Rollback(ctx context.Context) error {
	// Rollback to savepoint
	_, err := n.parent.tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+n.savepointName)
	return err
}

func (t *txProvider) Organizations() repository.OrganizationRepository {
	return &txOrgRepo{tx: t.tx}
}

func (t *txProvider) Users() repository.UserRepository {
	return &txUserRepo{tx: t.tx}
}

func (t *txProvider) Memberships() repository.MembershipRepository {
	return &txMembershipRepo{tx: t.tx}
}

func (t *txProvider) Roles() repository.RoleRepository {
	return &txRoleRepo{tx: t.tx}
}

func (t *txProvider) Permissions() repository.PermissionRepository {
	return &txPermissionRepo{tx: t.tx}
}

func (t *txProvider) Sessions() repository.SessionRepository {
	return &txSessionRepo{tx: t.tx}
}

func (t *txProvider) Rooms() repository.RoomRepository {
	return &txRoomRepo{tx: t.tx}
}

func (t *txProvider) Bots() repository.BotRepository {
	return &txBotRepo{tx: t.tx}
}

func (t *txProvider) Plugins() repository.PluginRepository {
	return &txPluginRepo{tx: t.tx}
}

func (t *txProvider) PluginConfigs() repository.PluginConfigRepository {
	return &txPluginConfigRepo{tx: t.tx}
}

func (t *txProvider) AuditLogs() repository.AuditLogRepository {
	return &txAuditLogRepo{tx: t.tx}
}

func (t *txProvider) APIKeys() repository.APIKeyRepository {
	return &txAPIKeyRepo{tx: t.tx}
}

func (t *txProvider) BotConnections() repository.BotConnectionRepository {
	return &txBotConnectionRepo{tx: t.tx}
}

func (t *txProvider) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *txProvider) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// ============================================================
// Transaction-aware repository implementations
// These use the transaction instead of the pool
// ============================================================

type txOrgRepo struct{ tx pgx.Tx }

func (r *txOrgRepo) Create(ctx context.Context, org *domain.Organization) error {
	query := `INSERT INTO organizations (id, name, slug, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.tx.Exec(ctx, query, org.ID, org.Name, org.Slug, org.Description, org.CreatedAt, org.UpdatedAt)
	return err
}

func (r *txOrgRepo) GetByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error) {
	query := `SELECT id, name, slug, description, created_at, updated_at FROM organizations WHERE id = $1`
	var org domain.Organization
	err := r.tx.QueryRow(ctx, query, id).Scan(&org.ID, &org.Name, &org.Slug, &org.Description, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, err
	}
	return &org, nil
}

func (r *txOrgRepo) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	query := `SELECT id, name, slug, description, created_at, updated_at FROM organizations WHERE slug = $1`
	var org domain.Organization
	err := r.tx.QueryRow(ctx, query, slug).Scan(&org.ID, &org.Name, &org.Slug, &org.Description, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, err
	}
	return &org, nil
}

func (r *txOrgRepo) List(ctx context.Context, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.Organization], error) {
	// For transactions, we typically don't list, but implement for completeness
	whereClause, args := buildFilterClause(filter, 1)
	argIndex := len(args) + 1

	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"name": true, "slug": true, "created_at": true, "updated_at": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	countQuery := "SELECT COUNT(*) FROM organizations" + whereClause
	var totalCount int64
	err := r.tx.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.Organization]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, name, slug, description, created_at, updated_at
		FROM organizations
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.tx.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.Organization]{}, err
	}
	defer rows.Close()

	var orgs []domain.Organization
	for rows.Next() {
		var org domain.Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.Description, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return repository.ListResult[domain.Organization]{}, err
		}
		orgs = append(orgs, org)
	}

	return repository.ListResult[domain.Organization]{
		Items:      orgs,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txOrgRepo) Update(ctx context.Context, org *domain.Organization) error {
	query := `UPDATE organizations SET name = $2, slug = $3, description = $4, updated_at = $5 WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, org.ID, org.Name, org.Slug, org.Description, org.UpdatedAt)
	return err
}

func (r *txOrgRepo) Delete(ctx context.Context, id domain.OrganizationID) error {
	query := `DELETE FROM organizations WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, id)
	return err
}

func (r *txOrgRepo) ExistsBySlug(ctx context.Context, slug string, excludeID *domain.OrganizationID) (bool, error) {
	query := `SELECT 1 FROM organizations WHERE slug = $1`
	args := []any{slug}
	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}
	var exists int
	err := r.tx.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type txUserRepo struct{ tx pgx.Tx }

func (r *txUserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, organization_id, email, display_name, password_hash, mfa_secret, mfa_enabled, is_active, last_login_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.tx.Exec(ctx, query, user.ID, user.Organization, user.Email, user.DisplayName, user.PasswordHash, user.MFASecret, user.MFAEnabled, user.IsActive, user.LastLoginAt, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *txUserRepo) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	query := `SELECT id, organization_id, email, display_name, password_hash, mfa_secret, mfa_enabled, is_active, last_login_at, created_at, updated_at FROM users WHERE id = $1`
	var user domain.User
	err := r.tx.QueryRow(ctx, query, id).Scan(&user.ID, &user.Organization, &user.Email, &user.DisplayName, &user.PasswordHash, &user.MFASecret, &user.MFAEnabled, &user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *txUserRepo) GetByEmail(ctx context.Context, orgID domain.OrganizationID, email string) (*domain.User, error) {
	query := `SELECT id, organization_id, email, display_name, password_hash, mfa_secret, mfa_enabled, is_active, last_login_at, created_at, updated_at FROM users WHERE organization_id = $1 AND email = $2`
	var user domain.User
	err := r.tx.QueryRow(ctx, query, orgID, email).Scan(&user.ID, &user.Organization, &user.Email, &user.DisplayName, &user.PasswordHash, &user.MFASecret, &user.MFAEnabled, &user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *txUserRepo) List(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.User], error) {
	whereClause, args := buildFilterClauseWithOrg(filter, orgID, 1)
	argIndex := len(args) + 1

	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"email": true, "display_name": true, "created_at": true, "updated_at": true, "is_active": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	countQuery := "SELECT COUNT(*) FROM users" + whereClause
	var totalCount int64
	err := r.tx.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.User]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, organization_id, email, display_name, password_hash, mfa_secret, mfa_enabled, is_active, last_login_at, created_at, updated_at
		FROM users
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.tx.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.User]{}, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID, &user.Organization, &user.Email, &user.DisplayName, &user.PasswordHash,
			&user.MFASecret, &user.MFAEnabled, &user.IsActive, &user.LastLoginAt,
			&user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return repository.ListResult[domain.User]{}, err
		}
		users = append(users, user)
	}

	return repository.ListResult[domain.User]{
		Items:      users,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txUserRepo) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET email = $2, display_name = $3, password_hash = $4, mfa_secret = $5, mfa_enabled = $6, is_active = $7, updated_at = $8 WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, user.ID, user.Email, user.DisplayName, user.PasswordHash, user.MFASecret, user.MFAEnabled, user.IsActive, user.UpdatedAt)
	return err
}

func (r *txUserRepo) UpdateLastLogin(ctx context.Context, id domain.UserID, t time.Time) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3`
	_, err := r.tx.Exec(ctx, query, t, time.Now(), id)
	return err
}

func (r *txUserRepo) Delete(ctx context.Context, id domain.UserID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, id)
	return err
}

func (r *txUserRepo) ExistsByEmail(ctx context.Context, orgID domain.OrganizationID, email string, excludeID *domain.UserID) (bool, error) {
	query := `SELECT 1 FROM users WHERE organization_id = $1 AND email = $2`
	args := []any{orgID, email}
	if excludeID != nil {
		query += ` AND id != $3`
		args = append(args, *excludeID)
	}
	var exists int
	err := r.tx.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type txMembershipRepo struct{ tx pgx.Tx }

func (r *txMembershipRepo) Create(ctx context.Context, m *domain.Membership) error {
	query := `INSERT INTO memberships (id, user_id, organization_id, role_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.tx.Exec(ctx, query, m.ID, m.UserID, m.OrganizationID, m.RoleID, m.CreatedAt, m.UpdatedAt)
	return err
}

func (r *txMembershipRepo) GetByUserAndOrg(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID) (*domain.Membership, error) {
	query := `SELECT id, user_id, organization_id, role_id, created_at, updated_at FROM memberships WHERE user_id = $1 AND organization_id = $2`
	var m domain.Membership
	err := r.tx.QueryRow(ctx, query, userID, orgID).Scan(&m.ID, &m.UserID, &m.OrganizationID, &m.RoleID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrMembershipNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *txMembershipRepo) ListByUser(ctx context.Context, userID domain.UserID) ([]domain.Membership, error) {
	query := `SELECT id, user_id, organization_id, role_id, created_at, updated_at FROM memberships WHERE user_id = $1`
	rows, err := r.tx.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []domain.Membership
	for rows.Next() {
		var m domain.Membership
		if err := rows.Scan(&m.ID, &m.UserID, &m.OrganizationID, &m.RoleID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}

func (r *txMembershipRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort) (repository.ListResult[domain.Membership], error) {
	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"created_at": true, "updated_at": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	countQuery := `SELECT COUNT(*) FROM memberships WHERE organization_id = $1`
	var totalCount int64
	err := r.tx.QueryRow(ctx, countQuery, orgID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.Membership]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, organization_id, role_id, created_at, updated_at
		FROM memberships WHERE organization_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, sortField, sortOrder)

	rows, err := r.tx.Query(ctx, dataQuery, orgID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.Membership]{}, err
	}
	defer rows.Close()

	var memberships []domain.Membership
	for rows.Next() {
		var m domain.Membership
		if err := rows.Scan(&m.ID, &m.UserID, &m.OrganizationID, &m.RoleID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return repository.ListResult[domain.Membership]{}, err
		}
		memberships = append(memberships, m)
	}

	return repository.ListResult[domain.Membership]{
		Items:      memberships,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txMembershipRepo) UpdateRole(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID, roleID domain.RoleID) error {
	query := `UPDATE memberships SET role_id = $1, updated_at = $2 WHERE user_id = $3 AND organization_id = $4`
	_, err := r.tx.Exec(ctx, query, roleID, time.Now(), userID, orgID)
	return err
}

func (r *txMembershipRepo) Delete(ctx context.Context, userID domain.UserID, orgID domain.OrganizationID) error {
	query := `DELETE FROM memberships WHERE user_id = $1 AND organization_id = $2`
	_, err := r.tx.Exec(ctx, query, userID, orgID)
	return err
}

func (r *txMembershipRepo) CountByOrgAndRole(ctx context.Context, orgID domain.OrganizationID, roleID domain.RoleID) (int64, error) {
	query := `SELECT COUNT(*) FROM memberships WHERE organization_id = $1 AND role_id = $2`
	var count int64
	err := r.tx.QueryRow(ctx, query, orgID, roleID).Scan(&count)
	return count, err
}

type txRoleRepo struct{ tx pgx.Tx }

func (r *txRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	query := `INSERT INTO roles (id, organization_id, name, description, is_system, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.tx.Exec(ctx, query, role.ID, role.OrganizationID, role.Name, role.Description, role.IsSystem, role.CreatedAt, role.UpdatedAt)
	return err
}

func (r *txRoleRepo) GetByID(ctx context.Context, id domain.RoleID) (*domain.Role, error) {
	query := `SELECT id, organization_id, name, description, is_system, created_at, updated_at FROM roles WHERE id = $1`
	var role domain.Role
	err := r.tx.QueryRow(ctx, query, id).Scan(&role.ID, &role.OrganizationID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *txRoleRepo) GetByName(ctx context.Context, orgID domain.OrganizationID, name string) (*domain.Role, error) {
	query := `SELECT id, organization_id, name, description, is_system, created_at, updated_at FROM roles WHERE organization_id = $1 AND name = $2`
	var role domain.Role
	err := r.tx.QueryRow(ctx, query, orgID, name).Scan(&role.ID, &role.OrganizationID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *txRoleRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, includeSystem bool) ([]domain.Role, error) {
	query := `SELECT id, organization_id, name, description, is_system, created_at, updated_at FROM roles WHERE organization_id = $1`
	if !includeSystem {
		query += ` AND is_system = FALSE`
	}
	query += ` ORDER BY name`
	rows, err := r.tx.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.OrganizationID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *txRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	query := `UPDATE roles SET name = $2, description = $3, updated_at = $4 WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, role.ID, role.Name, role.Description, role.UpdatedAt)
	return err
}

func (r *txRoleRepo) Delete(ctx context.Context, id domain.RoleID) error {
	var isSystem bool
	err := r.tx.QueryRow(ctx, `SELECT is_system FROM roles WHERE id = $1`, id).Scan(&isSystem)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrRoleNotFound
		}
		return err
	}
	if isSystem {
		return domain.ErrSystemRoleImmutable
	}

	query := `DELETE FROM roles WHERE id = $1`
	_, err = r.tx.Exec(ctx, query, id)
	return err
}

func (r *txRoleRepo) AddPermission(ctx context.Context, roleID domain.RoleID, permID domain.PermissionID) error {
	query := `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.tx.Exec(ctx, query, roleID, permID)
	return err
}

func (r *txRoleRepo) RemovePermission(ctx context.Context, roleID domain.RoleID, permID domain.PermissionID) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err := r.tx.Exec(ctx, query, roleID, permID)
	return err
}

func (r *txRoleRepo) ListPermissions(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error) {
	query := `
		SELECT p.id, p.name, p.description, p.resource, p.action, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := r.tx.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Resource, &p.Action, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *txRoleRepo) IsSystemRole(ctx context.Context, id domain.RoleID) (bool, error) {
	var isSystem bool
	err := r.tx.QueryRow(ctx, `SELECT is_system FROM roles WHERE id = $1`, id).Scan(&isSystem)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, domain.ErrRoleNotFound
		}
		return false, err
	}
	return isSystem, err
}

type txPermissionRepo struct{ tx pgx.Tx }

func (r *txPermissionRepo) Create(ctx context.Context, perm *domain.Permission) error {
	query := `INSERT INTO permissions (id, name, description, resource, action, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.tx.Exec(ctx, query, perm.ID, perm.Name, perm.Description, perm.Resource, perm.Action, perm.CreatedAt)
	return err
}

func (r *txPermissionRepo) GetByID(ctx context.Context, id domain.PermissionID) (*domain.Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at FROM permissions WHERE id = $1`
	var p domain.Permission
	err := r.tx.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &p.Description, &p.Resource, &p.Action, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPermissionNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *txPermissionRepo) GetByName(ctx context.Context, name string) (*domain.Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at FROM permissions WHERE name = $1`
	var p domain.Permission
	err := r.tx.QueryRow(ctx, query, name).Scan(&p.ID, &p.Name, &p.Description, &p.Resource, &p.Action, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPermissionNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *txPermissionRepo) List(ctx context.Context) ([]domain.Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at FROM permissions ORDER BY resource, action`
	rows, err := r.tx.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Resource, &p.Action, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *txPermissionRepo) ListByResource(ctx context.Context, resource string) ([]domain.Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at FROM permissions WHERE resource = $1 ORDER BY action`
	rows, err := r.tx.Query(ctx, query, resource)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Resource, &p.Action, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *txPermissionRepo) Update(ctx context.Context, perm *domain.Permission) error {
	query := `UPDATE permissions SET name = $2, description = $3, resource = $4, action = $5 WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, perm.ID, perm.Name, perm.Description, perm.Resource, perm.Action)
	return err
}

func (r *txPermissionRepo) Delete(ctx context.Context, id domain.PermissionID) error {
	query := `DELETE FROM permissions WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, id)
	return err
}

type txSessionRepo struct{ tx pgx.Tx }

func (r *txSessionRepo) Create(ctx context.Context, session *domain.Session) error {
	query := `INSERT INTO sessions (id, user_id, organization_id, token_hash, csrf_token, expires_at, revoked_at, user_agent, ip_address, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.tx.Exec(ctx, query, session.ID, session.UserID, session.Organization, session.TokenHash, session.CSRFToken, session.ExpiresAt, session.RevokedAt, session.UserAgent, session.IPAddress, session.CreatedAt)
	return err
}

func (r *txSessionRepo) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	query := `SELECT id, user_id, organization_id, token_hash, csrf_token, expires_at, revoked_at, user_agent, ip_address, created_at FROM sessions WHERE id = $1`
	var s domain.Session
	err := r.tx.QueryRow(ctx, query, id).Scan(&s.ID, &s.UserID, &s.Organization, &s.TokenHash, &s.CSRFToken, &s.ExpiresAt, &s.RevokedAt, &s.UserAgent, &s.IPAddress, &s.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *txSessionRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	query := `SELECT id, user_id, organization_id, token_hash, csrf_token, expires_at, revoked_at, user_agent, ip_address, created_at FROM sessions WHERE token_hash = $1`
	var s domain.Session
	err := r.tx.QueryRow(ctx, query, tokenHash).Scan(&s.ID, &s.UserID, &s.Organization, &s.TokenHash, &s.CSRFToken, &s.ExpiresAt, &s.RevokedAt, &s.UserAgent, &s.IPAddress, &s.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *txSessionRepo) Update(ctx context.Context, session *domain.Session) error {
	query := `UPDATE sessions SET token_hash = $2, csrf_token = $3, expires_at = $4, revoked_at = $5, user_agent = $6, ip_address = $7 WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, session.ID, session.TokenHash, session.CSRFToken, session.ExpiresAt, session.RevokedAt, session.UserAgent, session.IPAddress)
	return err
}

func (r *txSessionRepo) Revoke(ctx context.Context, id domain.SessionID) error {
	query := `UPDATE sessions SET revoked_at = $1 WHERE id = $2`
	_, err := r.tx.Exec(ctx, query, time.Now(), id)
	return err
}

func (r *txSessionRepo) RevokeAllForUser(ctx context.Context, userID domain.UserID) error {
	query := `UPDATE sessions SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, err := r.tx.Exec(ctx, query, time.Now(), userID)
	return err
}

func (r *txSessionRepo) RevokeAllForOrg(ctx context.Context, orgID domain.OrganizationID) error {
	query := `UPDATE sessions SET revoked_at = $1 WHERE organization_id = $2 AND revoked_at IS NULL`
	_, err := r.tx.Exec(ctx, query, time.Now(), orgID)
	return err
}

func (r *txSessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < $1 OR revoked_at IS NOT NULL`
	result, err := r.tx.Exec(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *txSessionRepo) CleanupRevoked(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM sessions WHERE revoked_at IS NOT NULL AND revoked_at < $1`
	result, err := r.tx.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

type txRoomRepo struct{ tx pgx.Tx }

func (r *txRoomRepo) Create(ctx context.Context, room *domain.Room) error {
	query := `INSERT INTO rooms (id, organization_id, name, access_key_hash, bot_id, language, is_paused, status, last_seen_at, config, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.tx.Exec(ctx, query, room.ID, room.OrganizationID, room.Name, room.AccessKeyHash, room.BotID, room.Language, room.IsPaused, room.Status, room.LastSeenAt, room.Config, room.CreatedAt, room.UpdatedAt)
	return err
}

func (r *txRoomRepo) GetByID(ctx context.Context, id domain.RoomID) (*domain.Room, error) {
	query := `SELECT id, organization_id, name, access_key_hash, bot_id, language, is_paused, status, last_seen_at, config, created_at, updated_at FROM rooms WHERE id = $1`
	var room domain.Room
	var botID pgtype.UUID
	err := r.tx.QueryRow(ctx, query, id).Scan(
		&room.ID, &room.OrganizationID, &room.Name, &room.AccessKeyHash,
		&botID, &room.Language, &room.IsPaused, &room.Status,
		&room.LastSeenAt, &room.Config, &room.CreatedAt, &room.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRoomNotFound
		}
		return nil, err
	}
	if botID.Valid {
		bid := domain.BotID(botID.Bytes)
		room.BotID = &bid
	}
	return &room, nil
}

func (r *txRoomRepo) GetByOrgAndName(ctx context.Context, orgID domain.OrganizationID, name string) (*domain.Room, error) {
	query := `SELECT id, organization_id, name, access_key_hash, bot_id, language, is_paused, status, last_seen_at, config, created_at, updated_at FROM rooms WHERE organization_id = $1 AND name = $2`
	var room domain.Room
	var botID pgtype.UUID
	err := r.tx.QueryRow(ctx, query, orgID, name).Scan(
		&room.ID, &room.OrganizationID, &room.Name, &room.AccessKeyHash,
		&botID, &room.Language, &room.IsPaused, &room.Status,
		&room.LastSeenAt, &room.Config, &room.CreatedAt, &room.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRoomNotFound
		}
		return nil, err
	}
	if botID.Valid {
		bid := domain.BotID(botID.Bytes)
		room.BotID = &bid
	}
	return &room, nil
}

func (r *txRoomRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.Room], error) {
	whereClause, args := buildFilterClauseWithOrg(filter, orgID, 1)
	argIndex := len(args) + 1

	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"name": true, "created_at": true, "updated_at": true, "status": true, "is_paused": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	countQuery := "SELECT COUNT(*) FROM rooms" + whereClause
	var totalCount int64
	err := r.tx.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.Room]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, organization_id, name, access_key_hash, bot_id, language, is_paused, status, last_seen_at, config, created_at, updated_at
		FROM rooms
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.tx.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.Room]{}, err
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var room domain.Room
		var botID pgtype.UUID
		if err := rows.Scan(
			&room.ID, &room.OrganizationID, &room.Name, &room.AccessKeyHash,
			&botID, &room.Language, &room.IsPaused, &room.Status,
			&room.LastSeenAt, &room.Config, &room.CreatedAt, &room.UpdatedAt,
		); err != nil {
			return repository.ListResult[domain.Room]{}, err
		}
		if botID.Valid {
			bid := domain.BotID(botID.Bytes)
			room.BotID = &bid
		}
		rooms = append(rooms, room)
	}

	return repository.ListResult[domain.Room]{
		Items:      rooms,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txRoomRepo) ListByBot(ctx context.Context, botID domain.BotID) ([]domain.Room, error) {
	query := `SELECT id, organization_id, name, access_key_hash, bot_id, language, is_paused, status, last_seen_at, config, created_at, updated_at FROM rooms WHERE bot_id = $1`
	rows, err := r.tx.Query(ctx, query, botID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var room domain.Room
		var bid pgtype.UUID
		if err := rows.Scan(
			&room.ID, &room.OrganizationID, &room.Name, &room.AccessKeyHash,
			&bid, &room.Language, &room.IsPaused, &room.Status,
			&room.LastSeenAt, &room.Config, &room.CreatedAt, &room.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if bid.Valid {
			room.BotID = &botID
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *txRoomRepo) Update(ctx context.Context, room *domain.Room) error {
	query := `UPDATE rooms SET name = $2, access_key_hash = $3, bot_id = $4, language = $5, is_paused = $6, status = $7, last_seen_at = $8, config = $9, updated_at = $10 WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, room.ID, room.Name, room.AccessKeyHash, room.BotID, room.Language, room.IsPaused, room.Status, room.LastSeenAt, room.Config, room.UpdatedAt)
	return err
}

func (r *txRoomRepo) UpdateStatus(ctx context.Context, id domain.RoomID, status string, lastSeen *time.Time) error {
	query := `UPDATE rooms SET status = $1, last_seen_at = $2, updated_at = $3 WHERE id = $4`
	_, err := r.tx.Exec(ctx, query, status, lastSeen, time.Now(), id)
	return err
}

func (r *txRoomRepo) UpdateConfig(ctx context.Context, id domain.RoomID, config domain.RoomConfig) error {
	query := `UPDATE rooms SET config = $1, updated_at = $2 WHERE id = $3`
	_, err := r.tx.Exec(ctx, query, config, time.Now(), id)
	return err
}

func (r *txRoomRepo) AssignBot(ctx context.Context, roomID domain.RoomID, botID *domain.BotID) error {
	query := `UPDATE rooms SET bot_id = $1, updated_at = $2 WHERE id = $3`
	_, err := r.tx.Exec(ctx, query, botID, time.Now(), roomID)
	return err
}

func (r *txRoomRepo) Delete(ctx context.Context, id domain.RoomID) error {
	query := `DELETE FROM rooms WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, id)
	return err
}

func (r *txRoomRepo) ExistsByName(ctx context.Context, orgID domain.OrganizationID, name string, excludeID *domain.RoomID) (bool, error) {
	query := `SELECT 1 FROM rooms WHERE organization_id = $1 AND name = $2`
	args := []any{orgID, name}
	if excludeID != nil {
		query += ` AND id != $3`
		args = append(args, *excludeID)
	}
	var exists int
	err := r.tx.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type txBotRepo struct{ tx pgx.Tx }

func (r *txBotRepo) Create(ctx context.Context, bot *domain.Bot) error {
	query := `INSERT INTO bots (id, organization_id, name, username, password_hash, status, room_id, last_connected_at, last_error, config, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.tx.Exec(ctx, query, bot.ID, bot.OrganizationID, bot.Name, bot.Username, bot.PasswordHash, bot.Status, bot.RoomID, bot.LastConnectedAt, bot.LastError, bot.Config, bot.CreatedAt, bot.UpdatedAt)
	return err
}

func (r *txBotRepo) GetByID(ctx context.Context, id domain.BotID) (*domain.Bot, error) {
	query := `SELECT id, organization_id, name, username, password_hash, status, room_id, last_connected_at, last_error, config, created_at, updated_at FROM bots WHERE id = $1`
	var bot domain.Bot
	var roomID pgtype.UUID
	err := r.tx.QueryRow(ctx, query, id).Scan(
		&bot.ID, &bot.OrganizationID, &bot.Name, &bot.Username, &bot.PasswordHash,
		&bot.Status, &roomID, &bot.LastConnectedAt, &bot.LastError,
		&bot.Config, &bot.CreatedAt, &bot.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrBotNotFound
		}
		return nil, err
	}
	if roomID.Valid {
		rid := domain.RoomID(roomID.Bytes)
		bot.RoomID = &rid
	}
	return &bot, nil
}

func (r *txBotRepo) GetByUsername(ctx context.Context, orgID domain.OrganizationID, username string) (*domain.Bot, error) {
	query := `SELECT id, organization_id, name, username, password_hash, status, room_id, last_connected_at, last_error, config, created_at, updated_at FROM bots WHERE organization_id = $1 AND username = $2`
	var bot domain.Bot
	var roomID pgtype.UUID
	err := r.tx.QueryRow(ctx, query, orgID, username).Scan(
		&bot.ID, &bot.OrganizationID, &bot.Name, &bot.Username, &bot.PasswordHash,
		&bot.Status, &roomID, &bot.LastConnectedAt, &bot.LastError,
		&bot.Config, &bot.CreatedAt, &bot.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrBotNotFound
		}
		return nil, err
	}
	if roomID.Valid {
		rid := domain.RoomID(roomID.Bytes)
		bot.RoomID = &rid
	}
	return &bot, nil
}

func (r *txBotRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.Bot], error) {
	whereClause, args := buildFilterClauseWithOrg(filter, orgID, 1)
	argIndex := len(args) + 1

	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"name": true, "username": true, "status": true, "created_at": true, "updated_at": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	countQuery := "SELECT COUNT(*) FROM bots" + whereClause
	var totalCount int64
	err := r.tx.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.Bot]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, organization_id, name, username, password_hash, status, room_id, last_connected_at, last_error, config, created_at, updated_at
		FROM bots
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.tx.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.Bot]{}, err
	}
	defer rows.Close()

	var bots []domain.Bot
	for rows.Next() {
		var bot domain.Bot
		var roomID pgtype.UUID
		if err := rows.Scan(
			&bot.ID, &bot.OrganizationID, &bot.Name, &bot.Username, &bot.PasswordHash,
			&bot.Status, &roomID, &bot.LastConnectedAt, &bot.LastError,
			&bot.Config, &bot.CreatedAt, &bot.UpdatedAt,
		); err != nil {
			return repository.ListResult[domain.Bot]{}, err
		}
		if roomID.Valid {
			rid := domain.RoomID(roomID.Bytes)
			bot.RoomID = &rid
		}
		bots = append(bots, bot)
	}

	return repository.ListResult[domain.Bot]{
		Items:      bots,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txBotRepo) ListByRoom(ctx context.Context, roomID domain.RoomID) ([]domain.Bot, error) {
	query := `SELECT id, organization_id, name, username, password_hash, status, room_id, last_connected_at, last_error, config, created_at, updated_at FROM bots WHERE room_id = $1`
	rows, err := r.tx.Query(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bots []domain.Bot
	for rows.Next() {
		var bot domain.Bot
		var rid pgtype.UUID
		if err := rows.Scan(
			&bot.ID, &bot.OrganizationID, &bot.Name, &bot.Username, &bot.PasswordHash,
			&bot.Status, &rid, &bot.LastConnectedAt, &bot.LastError,
			&bot.Config, &bot.CreatedAt, &bot.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if rid.Valid {
			bot.RoomID = &roomID
		}
		bots = append(bots, bot)
	}
	return bots, nil
}

func (r *txBotRepo) Update(ctx context.Context, bot *domain.Bot) error {
	query := `UPDATE bots SET name = $2, username = $3, password_hash = $4, status = $5, room_id = $6, last_connected_at = $7, last_error = $8, config = $9, updated_at = $10 WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, bot.ID, bot.Name, bot.Username, bot.PasswordHash, bot.Status, bot.RoomID, bot.LastConnectedAt, bot.LastError, bot.Config, bot.UpdatedAt)
	return err
}

func (r *txBotRepo) UpdateStatus(ctx context.Context, id domain.BotID, status string, lastError string, lastConnected *time.Time) error {
	query := `UPDATE bots SET status = $1, last_error = $2, last_connected_at = $3, updated_at = $4 WHERE id = $5`
	_, err := r.tx.Exec(ctx, query, status, lastError, lastConnected, time.Now(), id)
	return err
}

func (r *txBotRepo) UpdateConfig(ctx context.Context, id domain.BotID, config domain.BotConfig) error {
	query := `UPDATE bots SET config = $1, updated_at = $2 WHERE id = $3`
	_, err := r.tx.Exec(ctx, query, config, time.Now(), id)
	return err
}

func (r *txBotRepo) AssignRoom(ctx context.Context, botID domain.BotID, roomID *domain.RoomID) error {
	query := `UPDATE bots SET room_id = $1, updated_at = $2 WHERE id = $3`
	_, err := r.tx.Exec(ctx, query, roomID, time.Now(), botID)
	return err
}

func (r *txBotRepo) Delete(ctx context.Context, id domain.BotID) error {
	query := `DELETE FROM bots WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, id)
	return err
}

func (r *txBotRepo) ExistsByUsername(ctx context.Context, orgID domain.OrganizationID, username string, excludeID *domain.BotID) (bool, error) {
	query := `SELECT 1 FROM bots WHERE organization_id = $1 AND username = $2`
	args := []any{orgID, username}
	if excludeID != nil {
		query += ` AND id != $3`
		args = append(args, *excludeID)
	}
	var exists int
	err := r.tx.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type txPluginRepo struct{ tx pgx.Tx }

func (r *txPluginRepo) Create(ctx context.Context, plugin *domain.Plugin) error {
	query := `INSERT INTO plugins (id, plugin_number, name, description, version, compatibility, capability_flags, config_schema, is_enabled, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.tx.Exec(ctx, query, plugin.ID, plugin.PluginNumber, plugin.Name, plugin.Description, plugin.Version, plugin.Compatibility, plugin.CapabilityFlags, plugin.ConfigSchema, plugin.IsEnabled, plugin.CreatedAt, plugin.UpdatedAt)
	return err
}

func (r *txPluginRepo) GetByID(ctx context.Context, id domain.PluginID) (*domain.Plugin, error) {
	query := `SELECT id, plugin_number, name, description, version, compatibility, capability_flags, config_schema, is_enabled, created_at, updated_at FROM plugins WHERE id = $1`
	var plugin domain.Plugin
	err := r.tx.QueryRow(ctx, query, id).Scan(
		&plugin.ID, &plugin.PluginNumber, &plugin.Name, &plugin.Description,
		&plugin.Version, &plugin.Compatibility, &plugin.CapabilityFlags,
		&plugin.ConfigSchema, &plugin.IsEnabled, &plugin.CreatedAt, &plugin.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPluginNotFound
		}
		return nil, err
	}
	return &plugin, nil
}

func (r *txPluginRepo) GetByNumber(ctx context.Context, number int) (*domain.Plugin, error) {
	query := `SELECT id, plugin_number, name, description, version, compatibility, capability_flags, config_schema, is_enabled, created_at, updated_at FROM plugins WHERE plugin_number = $1`
	var plugin domain.Plugin
	err := r.tx.QueryRow(ctx, query, number).Scan(
		&plugin.ID, &plugin.PluginNumber, &plugin.Name, &plugin.Description,
		&plugin.Version, &plugin.Compatibility, &plugin.CapabilityFlags,
		&plugin.ConfigSchema, &plugin.IsEnabled, &plugin.CreatedAt, &plugin.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPluginNotFound
		}
		return nil, err
	}
	return &plugin, nil
}

func (r *txPluginRepo) List(ctx context.Context, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.Plugin], error) {
	whereClause, args := buildFilterClause(filter, 1)
	argIndex := len(args) + 1

	sortField := "plugin_number"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"plugin_number": true, "name": true, "compatibility": true, "is_enabled": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "ASC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	} else {
		sortOrder = "DESC"
	}

	countQuery := "SELECT COUNT(*) FROM plugins" + whereClause
	var totalCount int64
	err := r.tx.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.Plugin]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, plugin_number, name, description, version, compatibility, capability_flags, config_schema, is_enabled, created_at, updated_at
		FROM plugins
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.tx.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.Plugin]{}, err
	}
	defer rows.Close()

	var plugins []domain.Plugin
	for rows.Next() {
		var plugin domain.Plugin
		if err := rows.Scan(
			&plugin.ID, &plugin.PluginNumber, &plugin.Name, &plugin.Description,
			&plugin.Version, &plugin.Compatibility, &plugin.CapabilityFlags,
			&plugin.ConfigSchema, &plugin.IsEnabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return repository.ListResult[domain.Plugin]{}, err
		}
		plugins = append(plugins, plugin)
	}

	return repository.ListResult[domain.Plugin]{
		Items:      plugins,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txPluginRepo) ListEnabled(ctx context.Context) ([]domain.Plugin, error) {
	query := `SELECT id, plugin_number, name, description, version, compatibility, capability_flags, config_schema, is_enabled, created_at, updated_at FROM plugins WHERE is_enabled = TRUE ORDER BY plugin_number`
	rows, err := r.tx.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plugins []domain.Plugin
	for rows.Next() {
		var plugin domain.Plugin
		if err := rows.Scan(
			&plugin.ID, &plugin.PluginNumber, &plugin.Name, &plugin.Description,
			&plugin.Version, &plugin.Compatibility, &plugin.CapabilityFlags,
			&plugin.ConfigSchema, &plugin.IsEnabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, err
		}
		plugins = append(plugins, plugin)
	}
	return plugins, nil
}

func (r *txPluginRepo) Update(ctx context.Context, plugin *domain.Plugin) error {
	query := `UPDATE plugins SET name = $2, description = $3, version = $4, compatibility = $5, capability_flags = $6, config_schema = $7, is_enabled = $8, updated_at = $9 WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, plugin.ID, plugin.Name, plugin.Description, plugin.Version, plugin.Compatibility, plugin.CapabilityFlags, plugin.ConfigSchema, plugin.IsEnabled, plugin.UpdatedAt)
	return err
}

func (r *txPluginRepo) UpdateCompatibility(ctx context.Context, id domain.PluginID, compat string) error {
	query := `UPDATE plugins SET compatibility = $1, updated_at = $2 WHERE id = $3`
	_, err := r.tx.Exec(ctx, query, compat, time.Now(), id)
	return err
}

func (r *txPluginRepo) Delete(ctx context.Context, id domain.PluginID) error {
	query := `DELETE FROM plugins WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, id)
	return err
}

func (r *txPluginRepo) ExistsByNumber(ctx context.Context, number int, excludeID *domain.PluginID) (bool, error) {
	query := `SELECT 1 FROM plugins WHERE plugin_number = $1`
	args := []any{number}
	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}
	var exists int
	err := r.tx.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, err
}

type txPluginConfigRepo struct{ tx pgx.Tx }

func (r *txPluginConfigRepo) Create(ctx context.Context, config *domain.PluginConfiguration) error {
	query := `INSERT INTO plugin_configurations (id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.tx.Exec(ctx, query, config.ID, config.PluginID, config.RoomID, config.BotID, config.Settings, config.IsEnabled, config.Version, config.CreatedAt, config.UpdatedAt)
	return err
}

func (r *txPluginConfigRepo) GetByID(ctx context.Context, id domain.PluginConfigID) (*domain.PluginConfiguration, error) {
	query := `SELECT id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at FROM plugin_configurations WHERE id = $1`
	var config domain.PluginConfiguration
	var roomID, botID pgtype.UUID
	err := r.tx.QueryRow(ctx, query, id).Scan(
		&config.ID, &config.PluginID, &roomID, &botID,
		&config.Settings, &config.IsEnabled, &config.Version,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPluginConfigNotFound
		}
		return nil, err
	}
	if roomID.Valid {
		rid := domain.RoomID(uuid.MustParse(roomID.String()))
		config.RoomID = &rid
	}
	if botID.Valid {
		bid := domain.BotID(uuid.MustParse(botID.String()))
		config.BotID = &bid
	}
	return &config, nil
}

func (r *txPluginConfigRepo) GetByPluginAndRoom(ctx context.Context, pluginID domain.PluginID, roomID domain.RoomID) (*domain.PluginConfiguration, error) {
	query := `SELECT id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at FROM plugin_configurations WHERE plugin_id = $1 AND room_id = $2`
	var config domain.PluginConfiguration
	var rID, bID pgtype.UUID
	err := r.tx.QueryRow(ctx, query, pluginID, roomID).Scan(
		&config.ID, &config.PluginID, &rID, &bID,
		&config.Settings, &config.IsEnabled, &config.Version,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPluginConfigNotFound
		}
		return nil, err
	}
	if rID.Valid {
		rid := domain.RoomID(uuid.MustParse(rID.String()))
		config.RoomID = &rid
	}
	return &config, err
}

func (r *txPluginConfigRepo) GetByPluginAndBot(ctx context.Context, pluginID domain.PluginID, botID domain.BotID) (*domain.PluginConfiguration, error) {
	query := `SELECT id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at FROM plugin_configurations WHERE plugin_id = $1 AND bot_id = $2`
	var config domain.PluginConfiguration
	var rID, bID pgtype.UUID
	err := r.tx.QueryRow(ctx, query, pluginID, botID).Scan(
		&config.ID, &config.PluginID, &rID, &bID,
		&config.Settings, &config.IsEnabled, &config.Version,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPluginConfigNotFound
		}
		return nil, err
	}
	if bID.Valid {
		bid := domain.BotID(uuid.MustParse(bID.String()))
		config.BotID = &bid
	}
	return &config, err
}

func (r *txPluginConfigRepo) ListByRoom(ctx context.Context, roomID domain.RoomID) ([]domain.PluginConfiguration, error) {
	query := `SELECT id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at FROM plugin_configurations WHERE room_id = $1`
	rows, err := r.tx.Query(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []domain.PluginConfiguration
	for rows.Next() {
		var config domain.PluginConfiguration
		var rID, bID pgtype.UUID
		if err := rows.Scan(
			&config.ID, &config.PluginID, &rID, &bID,
			&config.Settings, &config.IsEnabled, &config.Version,
			&config.CreatedAt, &config.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if rID.Valid {
			rid := domain.RoomID(uuid.MustParse(rID.String()))
			config.RoomID = &rid
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (r *txPluginConfigRepo) ListByBot(ctx context.Context, botID domain.BotID) ([]domain.PluginConfiguration, error) {
	query := `SELECT id, plugin_id, room_id, bot_id, settings, is_enabled, version, created_at, updated_at FROM plugin_configurations WHERE bot_id = $1`
	rows, err := r.tx.Query(ctx, query, botID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []domain.PluginConfiguration
	for rows.Next() {
		var config domain.PluginConfiguration
		var rID, bID pgtype.UUID
		if err := rows.Scan(
			&config.ID, &config.PluginID, &rID, &bID,
			&config.Settings, &config.IsEnabled, &config.Version,
			&config.CreatedAt, &config.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if bID.Valid {
			bid := domain.BotID(uuid.MustParse(bID.String()))
			config.BotID = &bid
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (r *txPluginConfigRepo) Update(ctx context.Context, config *domain.PluginConfiguration) error {
	query := `UPDATE plugin_configurations SET plugin_id = $2, room_id = $3, bot_id = $4, settings = $5, is_enabled = $6, version = $7, updated_at = $8 WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, config.ID, config.PluginID, config.RoomID, config.BotID, config.Settings, config.IsEnabled, config.Version, config.UpdatedAt)
	return err
}

func (r *txPluginConfigRepo) UpdateSettings(ctx context.Context, id domain.PluginConfigID, settings string, version int) error {
	query := `UPDATE plugin_configurations SET settings = $1, version = $2, updated_at = $3 WHERE id = $4 AND version = $5`
	result, err := r.tx.Exec(ctx, query, settings, version, time.Now(), id, version-1)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrOptimisticLock
	}
	return nil
}

func (r *txPluginConfigRepo) ToggleEnabled(ctx context.Context, id domain.PluginConfigID, enabled bool) error {
	query := `UPDATE plugin_configurations SET is_enabled = $1, updated_at = $2 WHERE id = $3`
	_, err := r.tx.Exec(ctx, query, enabled, time.Now(), id)
	return err
}

func (r *txPluginConfigRepo) Delete(ctx context.Context, id domain.PluginConfigID) error {
	query := `DELETE FROM plugin_configurations WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, id)
	return err
}

func (r *txPluginConfigRepo) DeleteByRoom(ctx context.Context, roomID domain.RoomID) error {
	query := `DELETE FROM plugin_configurations WHERE room_id = $1`
	_, err := r.tx.Exec(ctx, query, roomID)
	return err
}

func (r *txPluginConfigRepo) DeleteByBot(ctx context.Context, botID domain.BotID) error {
	query := `DELETE FROM plugin_configurations WHERE bot_id = $1`
	_, err := r.tx.Exec(ctx, query, botID)
	return err
}

type txAuditLogRepo struct{ tx pgx.Tx }

func (r *txAuditLogRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `INSERT INTO audit_logs (id, organization_id, actor_user_id, actor_type, action, resource_type, resource_id, old_value, new_value, ip_address, user_agent, correlation_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.tx.Exec(ctx, query, log.ID, log.OrganizationID, log.ActorUserID, log.ActorType, log.Action, log.ResourceType, log.ResourceID, log.OldValue, log.NewValue, log.IPAddress, log.UserAgent, log.CorrelationID, log.CreatedAt)
	return err
}

func (r *txAuditLogRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination, sort repository.Sort, filter repository.Filter) (repository.ListResult[domain.AuditLog], error) {
	whereClause, args := buildFilterClauseWithOrg(filter, orgID, 1)
	argIndex := len(args) + 1

	sortField := "created_at"
	if sort.Field != "" {
		allowedSortFields := map[string]bool{"created_at": true, "action": true, "resource_type": true}
		if allowedSortFields[sort.Field] {
			sortField = sort.Field
		}
	}
	sortOrder := "DESC"
	if sort.Order == repository.SortAsc {
		sortOrder = "ASC"
	}

	countQuery := "SELECT COUNT(*) FROM audit_logs" + whereClause
	var totalCount int64
	err := r.tx.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, organization_id, actor_user_id, actor_type, action, resource_type, resource_id, old_value, new_value, ip_address, user_agent, correlation_id, created_at
		FROM audit_logs
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.tx.Query(ctx, dataQuery, args...)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var actorUserID pgtype.UUID
		if err := rows.Scan(
			&log.ID, &log.OrganizationID, &actorUserID, &log.ActorType,
			&log.Action, &log.ResourceType, &log.ResourceID,
			&log.OldValue, &log.NewValue, &log.IPAddress, &log.UserAgent,
			&log.CorrelationID, &log.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.AuditLog]{}, err
		}
		if actorUserID.Valid {
			uid := domain.UserID(uuid.MustParse(actorUserID.String()))
			log.ActorUserID = &uid
		}
		logs = append(logs, log)
	}

	return repository.ListResult[domain.AuditLog]{
		Items:      logs,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txAuditLogRepo) ListByResource(ctx context.Context, orgID domain.OrganizationID, resourceType, resourceID string, pagination repository.Pagination) (repository.ListResult[domain.AuditLog], error) {
	query := `
		SELECT id, organization_id, actor_user_id, actor_type, action, resource_type, resource_id, old_value, new_value, ip_address, user_agent, correlation_id, created_at
		FROM audit_logs
		WHERE organization_id = $1 AND resource_type = $2 AND resource_id = $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.tx.Query(ctx, query, orgID, resourceType, resourceID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var actorUserID pgtype.UUID
		if err := rows.Scan(
			&log.ID, &log.OrganizationID, &actorUserID, &log.ActorType,
			&log.Action, &log.ResourceType, &log.ResourceID,
			&log.OldValue, &log.NewValue, &log.IPAddress, &log.UserAgent,
			&log.CorrelationID, &log.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.AuditLog]{}, err
		}
		if actorUserID.Valid {
			uid := domain.UserID(uuid.MustParse(actorUserID.String()))
			log.ActorUserID = &uid
		}
		logs = append(logs, log)
	}

	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1 AND resource_type = $2 AND resource_id = $3`
	var totalCount int64
	err = r.tx.QueryRow(ctx, countQuery, orgID, resourceType, resourceID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}

	return repository.ListResult[domain.AuditLog]{
		Items:      logs,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txAuditLogRepo) ListByActor(ctx context.Context, orgID domain.OrganizationID, actorUserID domain.UserID, pagination repository.Pagination) (repository.ListResult[domain.AuditLog], error) {
	query := `
		SELECT id, organization_id, actor_user_id, actor_type, action, resource_type, resource_id, old_value, new_value, ip_address, user_agent, correlation_id, created_at
		FROM audit_logs
		WHERE organization_id = $1 AND actor_user_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.tx.Query(ctx, query, orgID, actorUserID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var aUID pgtype.UUID
		if err := rows.Scan(
			&log.ID, &log.OrganizationID, &aUID, &log.ActorType,
			&log.Action, &log.ResourceType, &log.ResourceID,
			&log.OldValue, &log.NewValue, &log.IPAddress, &log.UserAgent,
			&log.CorrelationID, &log.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.AuditLog]{}, err
		}
		if aUID.Valid {
			uid := domain.UserID(uuid.MustParse(aUID.String()))
			log.ActorUserID = &uid
		}
		logs = append(logs, log)
	}

	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1 AND actor_user_id = $2`
	var totalCount int64
	err = r.tx.QueryRow(ctx, countQuery, orgID, actorUserID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.AuditLog]{}, err
	}

	return repository.ListResult[domain.AuditLog]{
		Items:      logs,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txAuditLogRepo) DeleteOlderThan(ctx context.Context, orgID domain.OrganizationID, olderThan time.Time) (int64, error) {
	query := `DELETE FROM audit_logs WHERE organization_id = $1 AND created_at < $2`
	result, err := r.tx.Exec(ctx, query, orgID, olderThan)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

type txAPIKeyRepo struct{ tx pgx.Tx }

func (r *txAPIKeyRepo) Create(ctx context.Context, key *domain.APIKey) error {
	query := `INSERT INTO api_keys (id, organization_id, name, key_hash, prefix, permissions, expires_at, last_used_at, revoked_at, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.tx.Exec(ctx, query, key.ID, key.OrganizationID, key.Name, key.KeyHash, key.Prefix, key.Permissions, key.ExpiresAt, key.LastUsedAt, key.RevokedAt, key.CreatedAt)
	return err
}

func (r *txAPIKeyRepo) GetByID(ctx context.Context, id domain.APIKeyID) (*domain.APIKey, error) {
	query := `SELECT id, organization_id, name, key_hash, prefix, permissions, expires_at, last_used_at, revoked_at, created_at FROM api_keys WHERE id = $1`
	var key domain.APIKey
	err := r.tx.QueryRow(ctx, query, id).Scan(&key.ID, &key.OrganizationID, &key.Name, &key.KeyHash, &key.Prefix, &key.Permissions, &key.ExpiresAt, &key.LastUsedAt, &key.RevokedAt, &key.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrAPIKeyNotFound
		}
		return nil, err
	}
	return &key, nil
}

func (r *txAPIKeyRepo) GetByPrefix(ctx context.Context, orgID domain.OrganizationID, prefix string) (*domain.APIKey, error) {
	query := `SELECT id, organization_id, name, key_hash, prefix, permissions, expires_at, last_used_at, revoked_at, created_at FROM api_keys WHERE organization_id = $1 AND prefix = $2`
	var key domain.APIKey
	err := r.tx.QueryRow(ctx, query, orgID, prefix).Scan(&key.ID, &key.OrganizationID, &key.Name, &key.KeyHash, &key.Prefix, &key.Permissions, &key.ExpiresAt, &key.LastUsedAt, &key.RevokedAt, &key.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrAPIKeyNotFound
		}
		return nil, err
	}
	return &key, nil
}

func (r *txAPIKeyRepo) ListByOrg(ctx context.Context, orgID domain.OrganizationID, pagination repository.Pagination) (repository.ListResult[domain.APIKey], error) {
	countQuery := `SELECT COUNT(*) FROM api_keys WHERE organization_id = $1`
	var totalCount int64
	err := r.tx.QueryRow(ctx, countQuery, orgID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.APIKey]{}, err
	}

	query := `
		SELECT id, organization_id, name, key_hash, prefix, permissions, expires_at, last_used_at, revoked_at, created_at
		FROM api_keys WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.tx.Query(ctx, query, orgID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.APIKey]{}, err
	}
	defer rows.Close()

	var keys []domain.APIKey
	for rows.Next() {
		var key domain.APIKey
		if err := rows.Scan(
			&key.ID, &key.OrganizationID, &key.Name, &key.KeyHash,
			&key.Prefix, &key.Permissions, &key.ExpiresAt,
			&key.LastUsedAt, &key.RevokedAt, &key.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.APIKey]{}, err
		}
		keys = append(keys, key)
	}

	return repository.ListResult[domain.APIKey]{
		Items:      keys,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txAPIKeyRepo) Update(ctx context.Context, key *domain.APIKey) error {
	query := `UPDATE api_keys SET name = $2, permissions = $3, expires_at = $4, revoked_at = $5 WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, key.ID, key.Name, key.Permissions, key.ExpiresAt, key.RevokedAt)
	return err
}

func (r *txAPIKeyRepo) UpdateLastUsed(ctx context.Context, id domain.APIKeyID, t time.Time) error {
	query := `UPDATE api_keys SET last_used_at = $1 WHERE id = $2`
	_, err := r.tx.Exec(ctx, query, t, id)
	return err
}

func (r *txAPIKeyRepo) Revoke(ctx context.Context, id domain.APIKeyID) error {
	query := `UPDATE api_keys SET revoked_at = $1 WHERE id = $2`
	_, err := r.tx.Exec(ctx, query, time.Now(), id)
	return err
}

func (r *txAPIKeyRepo) Delete(ctx context.Context, id domain.APIKeyID) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	_, err := r.tx.Exec(ctx, query, id)
	return err
}

type txBotConnectionRepo struct{ tx pgx.Tx }

func (r *txBotConnectionRepo) Create(ctx context.Context, conn *domain.BotConnection) error {
	query := `INSERT INTO bot_connections (id, bot_id, room_id, event_type, message, metadata, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.tx.Exec(ctx, query, conn.ID, conn.BotID, conn.RoomID, conn.EventType, conn.Message, conn.Metadata, conn.CreatedAt)
	return err
}

func (r *txBotConnectionRepo) ListByBot(ctx context.Context, botID domain.BotID, pagination repository.Pagination) (repository.ListResult[domain.BotConnection], error) {
	countQuery := `SELECT COUNT(*) FROM bot_connections WHERE bot_id = $1`
	var totalCount int64
	err := r.tx.QueryRow(ctx, countQuery, botID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.BotConnection]{}, err
	}

	query := `
		SELECT id, bot_id, room_id, event_type, message, metadata, created_at
		FROM bot_connections WHERE bot_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.tx.Query(ctx, query, botID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.BotConnection]{}, err
	}
	defer rows.Close()

	var conns []domain.BotConnection
	for rows.Next() {
		var conn domain.BotConnection
		var roomID pgtype.UUID
		if err := rows.Scan(
			&conn.ID, &conn.BotID, &roomID, &conn.EventType,
			&conn.Message, &conn.Metadata, &conn.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.BotConnection]{}, err
		}
		if roomID.Valid {
			rid := domain.RoomID(uuid.MustParse(roomID.String()))
			conn.RoomID = &rid
		}
		conns = append(conns, conn)
	}

	return repository.ListResult[domain.BotConnection]{
		Items:      conns,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txBotConnectionRepo) ListByRoom(ctx context.Context, roomID domain.RoomID, pagination repository.Pagination) (repository.ListResult[domain.BotConnection], error) {
	countQuery := `SELECT COUNT(*) FROM bot_connections WHERE room_id = $1`
	var totalCount int64
	err := r.tx.QueryRow(ctx, countQuery, roomID).Scan(&totalCount)
	if err != nil {
		return repository.ListResult[domain.BotConnection]{}, err
	}

	query := `
		SELECT id, bot_id, room_id, event_type, message, metadata, created_at
		FROM bot_connections WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.tx.Query(ctx, query, roomID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repository.ListResult[domain.BotConnection]{}, err
	}
	defer rows.Close()

	var conns []domain.BotConnection
	for rows.Next() {
		var conn domain.BotConnection
		var rID pgtype.UUID
		if err := rows.Scan(
			&conn.ID, &conn.BotID, &rID, &conn.EventType,
			&conn.Message, &conn.Metadata, &conn.CreatedAt,
		); err != nil {
			return repository.ListResult[domain.BotConnection]{}, err
		}
		if rID.Valid {
			conn.RoomID = &roomID
		}
		conns = append(conns, conn)
	}

	return repository.ListResult[domain.BotConnection]{
		Items:      conns,
		TotalCount: totalCount,
		Pagination: pagination,
	}, nil
}

func (r *txBotConnectionRepo) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM bot_connections WHERE created_at < $1`
	result, err := r.tx.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

var _ repository.UnitOfWork = (*txProvider)(nil)
var _ repository.RepositoryProvider = (*Provider)(nil)