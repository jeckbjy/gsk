package cache

// LFU Least-Frequently Used
// 简单实现可以使用hash_map+小顶堆,但有排序消耗
// O(1)复杂实现,双链表?
// http://dhruvbird.com/lfu.pdf
// https://my.oschina.net/manmao/blog/603253
// https://leetcode.com/problems/lfu-cache/?tab=Solutions
type LFU struct {
}
