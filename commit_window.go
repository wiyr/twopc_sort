package main

import "time"

type DataWithCommitTime struct {
	data
	possibleCommitTime time.Time // time that be close to commit time but >= commit time
}

type orderByCommit struct {
	buf []DataWithCommitTime
}

func (o *orderByCommit) push(dat DataWithCommitTime) {
	o.buf = append(o.buf, dat)
	for i := len(o.buf) - 1; i > 0; i-- {
		if o.buf[i].commit < o.buf[i-1].commit {
			o.buf[i].possibleCommitTime = o.buf[i-1].possibleCommitTime
			o.swap(i, i-1)
		} else {
			break
		}
	}
}

func (o *orderByCommit) swap(i, j int) {
	o.buf[i], o.buf[j] = o.buf[j], o.buf[i]
}

func (o *orderByCommit) popFront() {
	o.buf = o.buf[1:]
}

func (o *orderByCommit) front() DataWithCommitTime {
	return o.buf[0]
}

func (o *orderByCommit) size() int {
	return len(o.buf)
}
