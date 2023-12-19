package postgres

import (
	"context"
	newrelic2 "github.com/goto/stencil/pkg/newrelic"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/goto/stencil/core/namespace"
)

const namespaceListQuery = `
SELECT id, format, compatibility from namespaces
`

const namespaceGetQuery = `
SELECT * from namespaces where id=$1
`

const namespaceDeleteQuery = `
DELETE from namespaces where id=$1
`

const namespaceUpdateQuery = `
UPDATE namespaces SET format=$2,compatibility=$3,description=$4,updated_at=now()
WHERE id = $1
RETURNING *
`

const namespaceInsertQuery = `
INSERT INTO namespaces (id, format, compatibility, description, created_at, updated_at)
    VALUES ($1, $2, $3, $4, now(), now())
RETURNING *
`

type NamespaceRepository struct {
	db       *DB
	newrelic newrelic2.Service
}

func NewNamespaceRepository(dbc *DB, nr newrelic2.Service) *NamespaceRepository {
	return &NamespaceRepository{
		db:       dbc,
		newrelic: nr,
	}
}

func (r *NamespaceRepository) Create(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error) {
	endFunc := r.newrelic.StartGenericSegment(ctx, "DatabaseCall")
	defer endFunc()
	newNamespace := namespace.Namespace{}
	err := pgxscan.Get(ctx, r.db, &newNamespace, namespaceInsertQuery, ns.ID, ns.Format, ns.Compatibility, ns.Description)
	return newNamespace, wrapError(err, ns.ID)
}

func (r *NamespaceRepository) Update(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error) {
	endFunc := r.newrelic.StartGenericSegment(ctx, "DatabaseCall")
	defer endFunc()
	newNamespace := namespace.Namespace{}
	err := pgxscan.Get(ctx, r.db, &newNamespace, namespaceUpdateQuery, ns.ID, ns.Format, ns.Compatibility, ns.Description)
	return newNamespace, wrapError(err, ns.ID)
}

func (r *NamespaceRepository) Get(ctx context.Context, id string) (namespace.Namespace, error) {
	endFunc := r.newrelic.StartGenericSegment(ctx, "DatabaseCall")
	defer endFunc()
	newNamespace := namespace.Namespace{}
	err := pgxscan.Get(ctx, r.db, &newNamespace, namespaceGetQuery, id)
	return newNamespace, wrapError(err, id)
}

func (r *NamespaceRepository) Delete(ctx context.Context, id string) error {
	endFunc := r.newrelic.StartGenericSegment(ctx, "DatabaseCall")
	defer endFunc()
	_, err := r.db.Exec(ctx, namespaceDeleteQuery, id)
	r.db.Exec(ctx, deleteOrphanedData)
	return wrapError(err, id)
}

func (r *NamespaceRepository) List(ctx context.Context) ([]namespace.Namespace, error) {
	endFunc := r.newrelic.StartGenericSegment(ctx, "DatabaseCall")
	defer endFunc()
	var namespaces []namespace.Namespace
	err := pgxscan.Select(ctx, r.db, &namespaces, namespaceListQuery)
	return namespaces, wrapError(err, "")
}
