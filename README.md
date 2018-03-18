# twopc_sort

## 观察现象
    1. 每次获取commit值之前会发送一个prepare数据
    2. prepare数据跟commit数据的`prepare`值都是一样的
    3. 乱序的原因
        1. 获取commit值后,发送落后于其他goroutine的发送
    4. 对于当前goroutine而言, 其他goroutine 产生commit值大且先发送的数据的sendTime <
       当前goroutine的上一次发送preapre数据的sendTime
    5. prepare的值大小并没有任何意义
    6. 数据的sendTime有几率大的比sendTime小先接收到

## 做法
    1. 维护一个已经是递增的序列, 每次插入一个数据，做一次插入排序的算法
    2. 对每个收到的prepare数据记录prepare --> sendTime的映射关系,
       如果该prepare对应的commit数据已经收到了,则移除该prepare的映射关系
    3. 维护prepare数据的sendTime的有序性, 因为这个时间戳是拿来判断数据的完整性的.
    4. 每次把那些sendTime < min( `2`中维护的sendTime ) 的数据都写到文件


## 证明
    1. 完整性. 由于数据没有被丢弃，除了无法确定其commit order的最新数据, 这部分我觉得没办法
    2. 顺序性. 因为每次采用插入排序，所以必定是有序的. 且写入到文件的数据不会再被访问了由现象`4`所得
    3. sleep时间在获取commit值之后最大是5 millisecond, 所以最多保存5 millisecond的数据量, 内存可以支撑.

