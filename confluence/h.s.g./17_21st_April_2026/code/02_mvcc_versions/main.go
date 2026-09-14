// MVCC toy: UPDATE создаёт новую версию строки (xmin/xmax), не затирает in-place.
package main

import "fmt"

type version struct {
	xmin, xmax int // xmax=0 → жива
	value      string
}

type row struct {
	versions []version
}

func (r *row) insert(txid int, value string) {
	r.versions = append(r.versions, version{xmin: txid, value: value})
}

func (r *row) update(txid int, value string) {
	for i := range r.versions {
		if r.versions[i].xmax == 0 {
			r.versions[i].xmax = txid
		}
	}
	r.versions = append(r.versions, version{xmin: txid, value: value})
}

func (r *row) visible(snapshotTX int) string {
	for i := len(r.versions) - 1; i >= 0; i-- {
		v := r.versions[i]
		if v.xmin <= snapshotTX && (v.xmax == 0 || v.xmax > snapshotTX) {
			return v.value
		}
	}
	return ""
}

func main() {
	var r row
	r.insert(1, "Alice")
	r.update(2, "Alice Smith")
	fmt.Println("tx1 sees:", r.visible(1)) // Alice
	fmt.Println("tx2 sees:", r.visible(2)) // Alice Smith
	fmt.Println("versions:", len(r.versions), "(bloat until VACUUM)")
}
