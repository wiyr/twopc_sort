package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Solution struct {
	que             *Queue
	lastPrepareTime map[int64]time.Time
	writeCh         chan data
}

type Queue struct {
	buf []data
}

func (q *Queue) push(dat data, lastTime time.Time) {
	q.buf = append(q.buf, dat)
	i := len(q.buf) - 2
	for ; i >= 0; i-- {
		if q.buf[i].commit > q.buf[i+1].commit {
			q.buf[i], q.buf[i+1] = q.buf[i+1], q.buf[i]
		} else {
			break
		}
	}
}

func (q *Queue) pop() {
	q.buf = q.buf[1:]
}

func (q *Queue) front() data {
	return q.buf[0]
}

func (q *Queue) empty() bool {
	return len(q.buf) == 0
}

func NewOrderBuffer() *Solution {
	return &Solution{
		que:             &Queue{},
		lastPrepareTime: map[int64]time.Time{},
		writeCh:         make(chan data, 1000),
	}
}

var finals []data
var lastPut int

func (o *Solution) putIt(dat data) {
	var lastTime time.Time
	if dat.kind == "commit" {
		lastTime = o.lastPrepareTime[dat.prepare]
		delete(o.lastPrepareTime, dat.prepare)
	} else {
		o.lastPrepareTime[dat.prepare] = dat.sendTime
		return
	}

	//	log.Println(dat.commit, dat.sendTime.UnixNano(), lastTime.UnixNano(), dat.which)
	o.que.push(dat, lastTime)

	early := o.getEarliestPrepareTime()
	for !o.que.empty() {
		if o.que.front().sendTime.Before(early) {
			/*
			 *if len(finals) != 0 {
			 *    if finals[len(finals)-1].commit > o.que.front().commit {
			 *        log.Println("invalid")
			 *    }
			 *}
			 *finals = append(finals, o.que.front())
			 */
			o.writeCh <- o.que.front()
			o.que.pop()
		} else {
			break
		}
	}
}

func (o *Solution) writerDemon() {
	filename := `./result.txt`
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Panicf("open file %s failed: %s", filename, err)
	}
	defer f.Close()
	for {
		select {
		case d := <-o.writeCh:
			_, err := f.WriteString(strconv.FormatInt(d.commit, 10))
			if err != nil {
				log.Printf("[Error] write error %s", err)
			}
			f.WriteString("\n")
		}
	}
}

func (o *Solution) getEarliestPrepareTime() time.Time {
	result := time.Now()
	for _, t := range o.lastPrepareTime {
		if result.After(t) {
			result = t
		}
	}

	return result
}

func (o *Solution) simpleSort() {
	go o.writerDemon()
	for {
		select {
		case dat := <-dataStreaming[0]:
			o.putIt(dat)
		case dat := <-dataStreaming[1]:
			o.putIt(dat)
		}
	}
}
