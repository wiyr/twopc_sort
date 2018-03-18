package main

type orderBySendTime struct {
	buf []DataWithCommitTime
}

func (o *orderBySendTime) push(dat DataWithCommitTime) {
	o.buf = append(o.buf, dat)
	for i := o.size() - 1; i > 0; i-- {
		if o.buf[i].sendTime.Before(o.buf[i-1].sendTime) {
			o.swap(i, i-1)
		} else {
			break
		}
	}
}

func (o *orderBySendTime) popFront() {
	o.buf = o.buf[1:]
}

func (o *orderBySendTime) front() DataWithCommitTime {
	return o.buf[0]
}

func (o *orderBySendTime) swap(i, j int) {
	o.buf[i], o.buf[j] = o.buf[j], o.buf[i]
}

func (o *orderBySendTime) size() int {
	return len(o.buf)
}
