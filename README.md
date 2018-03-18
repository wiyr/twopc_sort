# twopc_sort

## 观察现象
    1. 每次获取commit值之前会发送一个prepare数据
    2. prepare数据跟commit数据的`prepare`值都是一样的
    3. commit value乱序的原因
        1. 获取commit值后,发送落后于其他goroutine的发送
    4. 对于当前goroutine而言, 其他goroutine 产生commit值大且先发送的数据的sendTime <
       当前goroutine的上一次发送preapre数据的sendTime
    5. prepare的值大小并没有任何意义
    6. 数据的sendTime有几率大的比sendTime小先接收到


## 做法
记
- `SPT`: send prepare data time
- `GCT`: get commit value time
- `SCT`: send commit data time
- `CV`: commit value
- `SPTi`: goroutine i send prepare data time

    1. 定义possibleCommitTime: 最接近GCT的time，且time >= GCT and time <= SCT. 维护一个commit value已经是递增的window,  并且window中数据最小的possibleCommitTime > min{PSTi}, `i`为发送了prepare data 但是还未发送commit data的goroutine
    2. 每次插入一个数据i，做一次插入排序的算法, 用满足 data[j].commit > data[i].commit的data[j]的SCT来更新data[i]的possibleCommitTime
    2. 维护一个SPT递增且其对应的commit data还没收到的队列
    3. 由于现象6, 维护一个`clientNums*2+1`的buffer, 保证收到的数据的send time 是递增的. 因为只有多个goroutine同时将数据push到channel时，接收端select 的结果会有一定的随机性.
    4. 每来一个commit data, 把那些possibleCommitTime < min{PSTi} 的数据都输出


----

## 证明

```
-----------------------------------------> groutine a
    ^           ^                ^
    |           |                |
    |           |                |
    |           |                |
    |SPTa       |GCTa            |SCTa
-----------------------------------------> groutine b
                  ^         ^
                  |         |
                  |         |
                  |         |
                  |GCTb     |SCTb
-------------------------------------------> groutine c
                    ^         ^
                    |         |
                    |         |
                    |         |
                    |GCTc     |SCTc

```


当出现收到的commit value 乱序时
```
有 CVa < CVb && SCTa > SCTb
.·. GCTa < GCTb
·.· GCTa > SPTa, SCTb > GCTb
.·. SPTa < SCTb
```
由于无法准确知道GCT, 我们放宽条件, 只要
```
condition 1: `SCT` < `min{SPTi}`, `i`: the goroutine that sent prepare data but not sent commit data
```

那么这些数据都可以输出到文件
