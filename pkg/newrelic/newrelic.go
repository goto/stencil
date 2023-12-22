package newrelic

//go:generate mockery --name=Service -r --case underscore --with-expecter --structname NewRelic  --filename=newrelic.go --output=./mocks

import (
	"context"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type Service interface {
	StartGenericSegment(context.Context, string) func()
	StartDBSegment(ctx context.Context, op, table string) *newrelic.DatastoreSegment
}
type NewRelic struct {
}

func (nr *NewRelic) StartGenericSegment(ctx context.Context, name string) func() {
	txn := newrelic.FromContext(ctx)
	if txn == nil {
		return func() {}
	}
	gs := newrelic.Segment{
		Name: name,
	}
	gs.StartTime = txn.StartSegmentNow()
	return gs.End
}

func (nr *NewRelic) StartDBSegment(ctx context.Context, op, table string) *newrelic.DatastoreSegment {
	txn := newrelic.FromContext(ctx)
	s := newrelic.DatastoreSegment{
		Product:    newrelic.DatastorePostgres,
		Collection: table,
		Operation:  op,
		StartTime:  txn.StartSegmentNow(),
	}
	return &s
}
