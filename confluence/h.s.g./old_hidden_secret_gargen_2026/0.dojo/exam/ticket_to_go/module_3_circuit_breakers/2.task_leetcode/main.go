package main

import "fmt"

type ListNode struct {
	Val  int
	Next *ListNode
}

// mergeTwoLists — dummy head, два указателя. O(n+m) time, O(1) extra space.
func mergeTwoLists(l1, l2 *ListNode) *ListNode {
	dummy := &ListNode{}
	cur := dummy
	for l1 != nil && l2 != nil {
		if l1.Val < l2.Val {
			cur.Next = l1
			l1 = l1.Next
		} else {
			cur.Next = l2
			l2 = l2.Next
		}
		cur = cur.Next
	}
	if l1 != nil {
		cur.Next = l1
	} else {
		cur.Next = l2
	}
	return dummy.Next
}

func fromSlice(xs []int) *ListNode {
	dummy := &ListNode{}
	cur := dummy
	for _, v := range xs {
		cur.Next = &ListNode{Val: v}
		cur = cur.Next
	}
	return dummy.Next
}

func toSlice(h *ListNode) []int {
	var out []int
	for h != nil {
		out = append(out, h.Val)
		h = h.Next
	}
	return out
}

func main() {
	a := fromSlice([]int{1, 2, 4})
	b := fromSlice([]int{1, 3, 4})
	fmt.Println(toSlice(mergeTwoLists(a, b))) // [1 1 2 3 4 4]
}
