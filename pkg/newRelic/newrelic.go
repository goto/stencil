package newRelic

import (
	"github.com/newrelic/go-agent/v3/newrelic"
)

func NewTransaction(nr *newrelic.Transaction) *NewRelicTransaction {
	return &NewRelicTransaction{
		nr: nr,
	}
}

type NewRelicTransaction struct {
	nr *newrelic.Transaction
}

func (txn *NewRelicTransaction) StartGenericSegment(name string) EndSegment {
	if txn == nil {
		return endSegmentFunc(func() {})
	}

	gs := newrelic.Segment{
		Name: name,
	}
	//fmt.Println("starting segment", name)
	gs.StartTime = txn.nr.StartSegmentNow()

	return endSegmentFunc(func() {
		//fmt.Println("ending segment", name)
		gs.End()
	})
}

type EndSegment interface {
	End()
}

type endSegmentFunc func()

func (esf endSegmentFunc) End() {
	esf()
}
