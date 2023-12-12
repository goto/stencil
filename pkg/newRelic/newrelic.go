package newRelic

import (
	"github.com/newrelic/go-agent/v3/newrelic"
)

func NewTransaction(nr *newrelic.Transaction) *Transaction {
	return &Transaction{
		nr: nr,
	}
}

type Transaction struct {
	nr *newrelic.Transaction
}

func (txn *Transaction) StartGenericSegment(name string) func() {
	if txn == nil {
		return func() {}
	}
	gs := newrelic.Segment{
		Name: name,
	}
	gs.StartTime = txn.nr.StartSegmentNow()
	return gs.End
}
