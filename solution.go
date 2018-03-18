package main

import (
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
)

type Solution struct {
	que             *Queue
	timeWindows     []data
	lastPrepareTime map[int64]time.Time
	writeCh         chan data
}

type Queue struct {
	buf []data
}

func (q *Queue) push(dat data, lastTime time.Time) {
	q.buf = append(q.buf, dat)
	i := len(q.buf) - 2
	for ; i >= 0 && lastTime.Before(q.buf[i].sendTime); i-- {
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

	o.que.push(dat, lastTime)

	early := o.getEarliestPrepareTime()
	//log.Println(dat.commit, "send time:", dat.sendTime.Format(time.RFC3339Nano), "prepare time:", lastTime.Format(time.RFC3339Nano), early.Format(time.RFC3339Nano))
	for !o.que.empty() {
		frontData := o.que.front()
		if frontData.sendTime.Before(early) {
			if len(finals) != 0 {
				if finals[len(finals)-1].commit > frontData.commit {
					log.Println("invalid:", finals[len(finals)-1].commit, frontData.commit)
					os.Exit(0)
				}
			}
			finals = append(finals, frontData)
			o.writeCh <- frontData
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
	result := time.Now().Add(time.Hour)
	for _, t := range o.lastPrepareTime {
		if result.After(t) {
			result = t
		}
	}

	return result
}

func (o *Solution) simpleSort() {
	go o.writerDemon()

	streamLen := len(dataStreaming)
	selectCases := make([]reflect.SelectCase, streamLen)
	for i := 0; i < streamLen; i++ {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(dataStreaming[i]),
		}
	}

	const windowMaxLength = 5
	for {
		_, recvValue, _ := reflect.Select(selectCases)
		dat, ok := recvValue.Interface().(data)
		if !ok {
			log.Printf("[Error] can't convert type %v", recvValue.Type)
		} else {
			o.timeWindows = append(o.timeWindows, dat)
			winLen := len(o.timeWindows)
			for i := winLen - 1; i > 0; i-- {
				if o.timeWindows[i].sendTime.Before(o.timeWindows[i-1].sendTime) {
					o.timeWindows[i], o.timeWindows[i-1] = o.timeWindows[i-1], o.timeWindows[i]
				} else {
					break
				}
			}
			if winLen == windowMaxLength {
				o.putIt(o.timeWindows[0])
				o.timeWindows = o.timeWindows[1:]
			}
		}
	}
}
