// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package groups

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
	"github.com/ory/x/contextx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/popx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
)

type ListGroupsParameters struct {
	// OrganizationID is the organization id to filter by, if not [uuid.Nil].
	OrganizationID uuid.UUID

	// Filters is a list of SCIM filters to apply.
	Filters []identity.SCIMFilter

	// Count specifies the maximum number of items to return.
	Count int

	// StartIndex is the 1-based index of the first element in the list.
	StartIndex int
}

type Persister interface {
	CreateGroup(ctx context.Context, g *Group) error
	ListGroups(ctx context.Context, params ListGroupsParameters) ([]Group, int, error)
	GetGroup(ctx context.Context, id uuid.UUID) (*Group, error)
	UpdateGroup(ctx context.Context, g *Group) ([]IdentityID, error)
	DeleteGroup(ctx context.Context, id uuid.UUID) ([]IdentityID, error)
	ListIdentityGroups(ctx context.Context, id uuid.UUID) ([]Group, error)
}

type dependencies interface {
	contextx.Provider
	x.TracingProvider
}

type GroupPersister struct {
	dependencies
	conn *pop.Connection
	nid  uuid.UUID
}

type IdentityID struct {
	ID uuid.UUID `db:"identity_id"`
}

func NewPersister(deps dependencies, conn *pop.Connection) *GroupPersister {
	return &GroupPersister{
		dependencies: deps,
		conn:         conn,
	}
}

func (p *GroupPersister) NetworkID(ctx context.Context) uuid.UUID {
	return p.Contextualizer().Network(ctx, p.nid)
}

func (p *GroupPersister) WithNetworkID(nid uuid.UUID) *GroupPersister {
	p.nid = nid
	return p
}

func (p *GroupPersister) CreateGroup(ctx context.Context, g *Group) error {
	g.NID = p.NetworkID(ctx)
	return sqlcon.HandleError(popx.GetConnection(ctx, p.conn.WithContext(ctx)).Create(g))
}

var IndexedSCIMAttributes = map[string]string{
	"displayName": "display_name",
}

func (p *GroupPersister) ListGroups(ctx context.Context, params ListGroupsParameters) (groups []Group, total int, err error) {
	ctx, span := p.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListGroups",
		trace.WithAttributes(attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	var (
		query           strings.Builder
		args            []any
		nid             = p.NetworkID(ctx)
		groupsWithTotal = make([]struct {
			Group
			Total int `db:"total"`
		}, 0)
	)

	query.WriteString(`
SELECT scim_groups.*, COUNT(*) OVER() AS total
FROM scim_groups
WHERE nid = ?
`)
	args = append(args, nid)

	for _, scimFilter := range params.Filters {
		// Do not allow filtering by SCIM attributes that are not indexed.
		col, ok := IndexedSCIMAttributes[scimFilter.SCIMAttribute]
		if !ok {
			return nil, 0, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The SCIM attribute %s is not indexed and cannot be used for filtering.", scimFilter.SCIMAttribute))
		}

		query.WriteString(fmt.Sprintf(`
				AND %s = ?
			`, col)) // This is not an SQL injection because we checked the SCIMAttribute above.
		args = append(args, scimFilter.Value)
	}

	if !params.OrganizationID.IsNil() {
		query.WriteString(`
				AND organization_id = ?
			`)
		args = append(args, params.OrganizationID.String())
	}
	if params.Count > 0 {
		query.WriteString(`
				LIMIT ?
			`)
		args = append(args, params.Count)
	}
	if params.StartIndex > 1 {
		query.WriteString(`
				OFFSET ?
			`)
		args = append(args, params.StartIndex-1)
	}

	err = sqlcon.HandleError(p.conn.WithContext(ctx).RawQuery(query.String(), args...).All(&groupsWithTotal))
	if err != nil {
		return nil, 0, err
	}

	for _, g := range groupsWithTotal {
		groups = append(groups, g.Group)
		total = g.Total
	}

	return
}

func (p *GroupPersister) GetGroup(ctx context.Context, id uuid.UUID) (g *Group, err error) {
	ctx, span := p.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetGroup",
		trace.WithAttributes(attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	results := make([]struct {
		Group
		MemberID          uuid.NullUUID    `db:"member_id"`
		MemberDisplayName sqlxx.NullString `db:"member_display_name"`
		SubgroupID        uuid.NullUUID    `db:"subgroup_id"`
		SubgroupName      sqlxx.NullString `db:"subgroup_display_name"`
	}, 0)

	query := `
SELECT scim_groups.*,
	   identities.id AS member_id,
	   identities.scim ->> 'userName' AS member_display_name,
	   subgroups.id AS subgroup_id,
	   subgroups.display_name AS subgroup_display_name

FROM scim_groups

LEFT JOIN scim_groups_members ON scim_groups_members.group_id = scim_groups.id
LEFT JOIN identities ON (
	scim_groups_members.identity_id = identities.id AND
	identities.nid = scim_groups.nid AND
    ((identities.organization_id = scim_groups.organization_id) OR (identities.organization_id IS NULL AND scim_groups.organization_id IS NULL))
)
    
LEFT JOIN scim_groups subgroups ON (
	scim_groups.id = subgroups.parent_id AND
	subgroups.nid = scim_groups.nid AND
    ((subgroups.organization_id = scim_groups.organization_id) OR (subgroups.organization_id IS NULL AND scim_groups.organization_id IS NULL))
)

WHERE scim_groups.id = ? AND scim_groups.nid = ?
`
	args := []any{id, p.NetworkID(ctx)}

	err = sqlcon.HandleError(p.conn.WithContext(ctx).RawQuery(query, args...).All(&results))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, herodot.ErrNotFound.WithReasonf("Group %q not found", id)
	}

	g = &results[0].Group
	for _, res := range results {
		if res.MemberID.Valid {
			g.Members = append(g.Members, GroupMember{
				ID:          res.MemberID.UUID,
				DisplayName: res.MemberDisplayName.String(),
			})
		}
		if res.SubgroupID.Valid {
			g.Subgroups = append(g.Subgroups, GroupMember{
				ID:          res.SubgroupID.UUID,
				DisplayName: res.SubgroupName.String(),
			})
		}
	}

	return g, nil
}

func (p *GroupPersister) UpdateGroup(ctx context.Context, g *Group) (affectedIdentityIDs []IdentityID, err error) {
	ctx, span := p.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateGroup",
		trace.WithAttributes(attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	for _, member := range g.Members {
		affectedIdentityIDs = append(affectedIdentityIDs, IdentityID{ID: member.ID})
	}

	return affectedIdentityIDs, errors.WithStack(sqlcon.HandleError(p.conn.WithContext(ctx).Transaction(func(tx *pop.Connection) (err error) {
		if err = p.removeOldSubgroups(ctx, tx, g); err != nil {
			return err
		}
		if err = p.addNewSubgroups(ctx, tx, g); err != nil {
			return err
		}
		if deletedIDs, err := p.removeOldMembers(ctx, tx, g); err != nil {
			return err
		} else {
			affectedIdentityIDs = append(affectedIdentityIDs, deletedIDs...)
		}
		if err = p.addNewMembers(ctx, tx, g); err != nil {
			return err
		}

		return nil
	})))
}

func (p *GroupPersister) removeOldSubgroups(ctx context.Context, conn *pop.Connection, g *Group) error {
	query := fmt.Sprintf(
		"UPDATE scim_groups SET parent_id = NULL WHERE parent_id = ? AND nid = ? AND organization_id = ? AND id NOT IN (%s)",
		strings.Join(slices.Repeat([]string{"?"}, len(g.Subgroups)), ", "),
	)
	args := []any{g.ID, p.NetworkID(ctx), g.OrganizationID}
	for _, subgroup := range g.Subgroups {
		args = append(args, subgroup.ID)
	}

	return conn.RawQuery(query, args...).Exec()
}

func (p *GroupPersister) addNewSubgroups(ctx context.Context, conn *pop.Connection, g *Group) error {
	query := fmt.Sprintf(
		"UPDATE scim_groups SET parent_id = ? WHERE nid = ? AND organization_id = ? AND id IN (%s)",
		strings.Join(slices.Repeat([]string{"?"}, len(g.Subgroups)), ", "),
	)
	args := []any{g.ID, p.NetworkID(ctx), g.OrganizationID}
	for _, subgroup := range g.Subgroups {
		args = append(args, subgroup.ID)
	}

	return conn.RawQuery(query, args...).Exec()
}

func (p *GroupPersister) removeOldMembers(ctx context.Context, conn *pop.Connection, g *Group) (ids []IdentityID, err error) {
	query := fmt.Sprintf(`
DELETE FROM scim_groups_members
WHERE
	group_id = ? AND
	identity_id NOT IN (SELECT id FROM identities WHERE nid = ? AND id IN (%s))
RETURNING identity_id
`,
		strings.Join(slices.Repeat([]string{"?"}, len(g.Members)), ", "),
	)
	args := []any{g.ID, p.NetworkID(ctx)}
	for _, member := range g.Members {
		args = append(args, member.ID)
	}

	return ids, conn.RawQuery(query, args...).All(&ids)
}

func (p *GroupPersister) addNewMembers(ctx context.Context, conn *pop.Connection, g *Group) error {
	args := []any{p.NetworkID(ctx)}

	var whereOrg string
	if g.OrganizationID.Valid {
		whereOrg = "organization_id = ?"
		args = append(args, g.OrganizationID.UUID)
	} else {
		whereOrg = "organization_id IS NULL"
	}

	query := fmt.Sprintf(`
INSERT INTO scim_groups_members (group_id, identity_id)
SELECT	'%s' as group_id, id AS identity_id
FROM	identities
WHERE	nid = ? AND %s AND id IN (%s)
ON CONFLICT DO NOTHING
`,
		g.ID,
		whereOrg,
		strings.Join(slices.Repeat([]string{"?"}, len(g.Members)), ", "),
	)
	for _, member := range g.Members {
		args = append(args, member.ID)
	}

	return conn.RawQuery(query, args...).Exec()
}

func (p *GroupPersister) updateGroupFields(ctx context.Context, conn *pop.Connection, g *Group) error {
	query := "UPDATE scim_groups SET display_name = ? WHERE id = ? AND nid = ?"
	args := []any{g.DisplayName, g.ID, p.NetworkID(ctx)}

	return conn.RawQuery(query, args...).Exec()
}

func (p *GroupPersister) DeleteGroup(ctx context.Context, id uuid.UUID) (affectedIdentityIDs []IdentityID, err error) {
	ctx, span := p.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteGroup",
		trace.WithAttributes(attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	return affectedIdentityIDs, errors.WithStack(sqlcon.HandleError(p.conn.WithContext(ctx).Transaction(func(tx *pop.Connection) (err error) {
		if err = tx.RawQuery(
			"SELECT identity_id FROM scim_groups_members WHERE group_id = ?", id,
		).All(&affectedIdentityIDs); err != nil {
			return err
		}
		count, err := tx.RawQuery(
			"DELETE FROM scim_groups WHERE id = ? AND nid = ?",
			id, p.NetworkID(ctx),
		).ExecWithCount()
		if err != nil {
			return err
		}
		if count == 0 {
			return sqlcon.ErrNoRows
		}
		return nil
	})))
}

func (p *GroupPersister) ListIdentityGroups(ctx context.Context, id uuid.UUID) (groups []Group, err error) {
	ctx, span := p.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ListIdentityGroups",
		trace.WithAttributes(attribute.Stringer("network.id", p.NetworkID(ctx))))
	defer otelx.End(span, &err)

	return groups, errors.WithStack(sqlcon.HandleError(p.conn.WithContext(ctx).RawQuery(`
SELECT	scim_groups.*
FROM	identities
JOIN	scim_groups_members ON scim_groups_members.identity_id = identities.id
JOIN	scim_groups ON scim_groups_members.group_id = scim_groups.id
WHERE	identities.id = ? AND
		identities.nid = ? AND
		identities.nid = scim_groups.nid
`,
		id, p.NetworkID(ctx)).All(&groups)))
}

type GroupMember struct {
	ID          uuid.UUID
	DisplayName string
}

type Group struct {
	ID       uuid.UUID     `db:"id"`
	NID      uuid.UUID     `db:"nid"`
	ParentID uuid.NullUUID `db:"parent_id"`

	OrganizationID uuid.NullUUID `db:"organization_id"`
	DisplayName    string        `db:"display_name"`

	Members   []GroupMember `db:"-"`
	Subgroups []GroupMember `db:"-"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (g Group) TableName() string {
	return "scim_groups"
}
