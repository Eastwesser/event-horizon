package main

import "fmt"

type ListNode struct {
	Val  int
	Next *ListNode
}

// hasCycle — Floyd: slow/fast. O(n) time, O(1) space.
func hasCycle(head *ListNode) bool {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
		if slow == fast {
			return true
		}
	}
	return false
}

func main() {
	a, b, c := &ListNode{Val: 1}, &ListNode{Val: 2}, &ListNode{Val: 3}
	a.Next, b.Next, c.Next = b, c, a // cycle
	fmt.Println("cycle:", hasCycle(a))

	x := &ListNode{Val: 1, Next: &ListNode{Val: 2}}
	fmt.Println("no cycle:", hasCycle(x))
}
