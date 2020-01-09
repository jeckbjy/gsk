package zset

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// Callback 遍历回调函数
type Callback func(rank uint64, el *Element)

// Element 存储信息
type Element struct {
	Key   string
	Score int64
	Data  interface{}
}

// SortedSet 实现类似Redis的zset功能,用于支持排行榜
// 与redis区别:
//   SkipList支持递减排序,zset支持设置最大个数
//   去除了通过Score查找，因为并不常用，而且需要额外定义一个结构体标识开闭区间
// 期望实现的功能：
// 1：插入，删除，查询
// 2：支持增序，减序
// 3：支持设置最大个数
// 4：支持安区间遍历
// 5：支持简单的序列化和反序列化
type SortedSet struct {
	dict      map[string]*Element
	sl        *zSkipList
	countMax  uint64
	ascending bool
}

// New create SortedSet
func New() *SortedSet {
	s := new(SortedSet)
	s.dict = make(map[string]*Element)
	s.sl = zNewSkipList(true)
	s.countMax = math.MaxUint64
	s.ascending = true
	return s
}

// SetMax set the zset max element count
func (ss *SortedSet) SetMax(v uint64) {
	ss.countMax = v
}

// SetDescending 设置为降序排列
func (ss *SortedSet) SetDescending() {
	if ss.ascending && ss.Len() == 0 {
		ss.ascending = false
		ss.sl.SetDescending()
	}
}

// Len return SortedSet element count
func (ss *SortedSet) Len() int {
	return len(ss.dict)
}

// Clear 清空数据
func (ss *SortedSet) Clear() {
	ss.dict = make(map[string]*Element)
	ss.sl.Clear()
}

// Insert 插入数据,存在则更新数据
func (ss *SortedSet) Insert(key string, score int64) *Element {
	v, ok := ss.dict[key]
	if ok {
		if score != v.Score {
			// 插入删除
			ss.sl.Delete(v)
			v.Score = score
			ss.sl.Insert(v)
		}
	} else {
		// 新建
		v = new(Element)
		v.Key = key
		v.Score = score
		ss.dict[key] = v
		ss.sl.Insert(v)

		if ss.sl.Len() > ss.countMax {
			// delete tail
			ss.sl.DeleteRangeByRank(ss.countMax+1, math.MaxUint64, func(rank uint64, el *Element) {
				delete(ss.dict, el.Key)
			})
		}
	}

	return v
}

// Delete 删除Key
func (ss *SortedSet) Delete(key string) bool {
	v, ok := ss.dict[key]
	if ok {
		ss.sl.Delete(v)
		delete(ss.dict, key)
		return true
	}

	return false
}

// GetScore 查询得分
func (ss *SortedSet) GetScore(key string) (int64, interface{}, bool) {
	v, ok := ss.dict[key]
	if !ok {
		return 0, nil, false
	}

	return v.Score, v.Data, true
}

// GetRank 查询rank
func (ss *SortedSet) GetRank(key string) uint64 {
	v, ok := ss.dict[key]
	if !ok {
		return 0
	}

	return ss.sl.GetRank(v)
}

// GetRankWithScore 查询rank和score等信息
func (ss *SortedSet) GetRankWithScore(key string) (uint64, int64, interface{}) {
	v, ok := ss.dict[key]
	if !ok {
		return 0, 0, nil
	}

	return ss.sl.GetRank(v), v.Score, v.Data
}

func (ss *SortedSet) Scan(cb Callback) error {
	return ss.ScanRange(0, 0, cb)
}

// Scan 遍历rank区间为[start,end]的数据,start 和end都可以为0，表示遍历全部,1base
func (ss *SortedSet) ScanRange(start, end uint64, cb Callback) error {
	if start == 0 {
		start = 1
	}
	if end == 0 {
		end = math.MaxUint64
	}

	if end < start {
		return fmt.Errorf("zset scan[%v-%v] end must great than start", start, end)
	}

	ss.sl.Scan(start, end, cb)
	return nil
}

// Load 加载数据
func (ss *SortedSet) Load(rd io.Reader) error {
	ss.Clear()

	scanner := bufio.NewScanner(rd)
	if !scanner.Scan() {
		return fmt.Errorf("zset read version fail")
	}

	if !scanner.Scan() {
		return fmt.Errorf("zset read head fail")
	}

	head := strings.SplitN(scanner.Text(), ",", 3)
	if len(head) != 3 {
		return fmt.Errorf("zset parse head fail,%+v", scanner.Text())
	}

	ascending, err := strconv.ParseBool(head[0])
	if err != nil {
		return fmt.Errorf("zset parse ascending fail")
	}

	countmax, err := strconv.ParseInt(head[1], 10, 64)
	if err != nil {
		return fmt.Errorf("zset parse countmax fail")
	}

	count, err := strconv.ParseUint(head[2], 10, 64)
	if err != nil {
		return fmt.Errorf("zset parse count fail")
	}

	if !ascending {
		ss.SetDescending()
	}

	ss.countMax = uint64(countmax)

	for i := uint64(0); i < count; i++ {
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()

		tokens := strings.SplitN(line, ",", 2)
		if len(tokens) != 2 {
			return fmt.Errorf("zset parse data fail:%+v", line)
		}

		score, err := strconv.ParseInt(tokens[0], 10, 64)
		if err != nil {
			return fmt.Errorf("zset parse score fail:%+v", tokens[0])
		}

		tokens[1] = strings.TrimSpace(tokens[1])
		ss.Insert(tokens[1], score)
	}

	return nil
}

// Save 保存数据
func (ss *SortedSet) Save(w io.Writer) {
	writer := bufio.NewWriter(w)
	// version
	_, _ = fmt.Fprintln(writer, "v1")
	// head
	head := fmt.Sprintf("%+v,%+v,%+v", ss.ascending, int64(ss.countMax), ss.Len())
	_, _ = fmt.Fprintln(writer, head)
	// count
	_ = ss.Scan(func(rank uint64, e *Element) {
		line := fmt.Sprintf("%v,%v", e.Score, e.Key)
		_, _ = fmt.Fprintln(writer, line)
	})

	_ = writer.Flush()
}

// LoadFile 从文件中加载
func (ss *SortedSet) LoadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file fail:%v", err)
	}

	return ss.Load(file)
}

func (ss *SortedSet) SaveFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file fail:%v", err)
	}

	ss.Save(file)
	file.Close()
	return nil
}
