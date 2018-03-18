package main

import "time"

type prepareQueue struct {
	exist map[int64]time.Time
	buf   []int64
}

func newPrepareQueue() *prepareQueue {
	return &prepareQueue{
		exist: map[int64]time.Time{},
	}
}

func (p *prepareQueue) addPrepareData(dat data) {
	p.buf = append(p.buf, dat.prepare)
	p.exist[dat.prepare] = dat.sendTime
}

func (p *prepareQueue) removePrepareID(id int64) {
	delete(p.exist, id)
}

func (p *prepareQueue) earliestPrepareTime() time.Time {
	for i, prepareID := range p.buf {
		if t, e := p.exist[prepareID]; e {
			p.buf = p.buf[i:]
			return t
		}
	}
	p.buf = p.buf[:0]
	return time.Now().Add(time.Hour)
}
