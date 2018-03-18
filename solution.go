package main

import (
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
)

type Window interface {
	push(DataWithCommitTime)
	popFront()
	front() DataWithCommitTime
	size() int
}

type Solution struct {
	commitTimeWindow Window // data in this window is ordered by commit value
	receiveWindow    Window // data in this window is ordered by send time
	maxRecvWindow    int
	prepareData      *prepareQueue
	lastPrepareTime  map[int64]time.Time
	writeCh          chan data
}

func NewOrderBuffer(maxRecvWindow int) *Solution {
	return &Solution{
		commitTimeWindow: &orderByCommit{},
		receiveWindow:    &orderBySendTime{},
		prepareData:      newPrepareQueue(),
		maxRecvWindow:    maxRecvWindow,
		lastPrepareTime:  map[int64]time.Time{},
		writeCh:          make(chan data, 1000),
	}
}

var finals []data

func (o *Solution) putIt(dat DataWithCommitTime) {
	if dat.kind == "commit" {
		o.prepareData.removePrepareID(dat.prepare)
		delete(o.lastPrepareTime, dat.prepare)
	} else {
		o.prepareData.addPrepareData(dat.data)
		return
	}

	o.commitTimeWindow.push(dat)
	early := o.prepareData.earliestPrepareTime()
	//log.Println(dat.commit, "send time:", dat.sendTime.Format(time.RFC3339Nano), "prepare time:", lastTime.Format(time.RFC3339Nano), early.Format(time.RFC3339Nano))
	for o.commitTimeWindow.size() != 0 {
		frontData := o.commitTimeWindow.front()
		if frontData.possibleCommitTime.Before(early) {
			if len(finals) != 0 {
				if finals[len(finals)-1].commit > frontData.commit {
					log.Println("invalid:", finals[len(finals)-1].commit, frontData.commit)
					os.Exit(0)
				}
			}
			finals = append(finals, frontData.data)
			o.writeCh <- frontData.data
			o.commitTimeWindow.popFront()
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

func (o *Solution) simpleSort() {
	go o.writerDemon()

	streamLen := len(dataStreaming)
	selectCases := make([]reflect.SelectCase, streamLen+1)
	for i := 0; i < streamLen; i++ {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(dataStreaming[i]),
		}
	}

	selectCases[streamLen] = reflect.SelectCase{
		Dir: reflect.SelectDefault,
	}

	var ticker <-chan time.Time
LOOP:
	for {
		which, recvValue, _ := reflect.Select(selectCases)
		if which == streamLen {
			select {
			case <-ticker:
				log.Printf("No stream data in one second, exit")
				break LOOP
			default:
			}
		} else {
			dat, ok := recvValue.Interface().(data)
			if !ok {
				log.Printf("[Error] can't convert type %v", recvValue.Type)
			} else {
				o.putInRecvWindow(dat)
			}
			ticker = time.After(time.Second)
			// TODO: save all remain data into file
		}
	}
}

func (o *Solution) putInRecvWindow(dat data) {
	dataWithCommitTime := DataWithCommitTime{
		data:               dat,
		possibleCommitTime: dat.sendTime,
	}
	o.receiveWindow.push(dataWithCommitTime)
	if o.receiveWindow.size() == o.maxRecvWindow {
		o.putIt(o.receiveWindow.front())
		o.receiveWindow.popFront()
	}
}
