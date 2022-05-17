package concurrency

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

var nilT *lockState

const (
	degree   = 16
	maxItems = 2*degree - 1
	minItems = degree - 1
)

func cmp(a, b *lockState) int {
	__antithesis_instrumentation__.Notify(100425)
	c := bytes.Compare(a.Key(), b.Key())
	if c != 0 {
		__antithesis_instrumentation__.Notify(100428)
		return c
	} else {
		__antithesis_instrumentation__.Notify(100429)
	}
	__antithesis_instrumentation__.Notify(100426)
	c = bytes.Compare(a.EndKey(), b.EndKey())
	if c != 0 {
		__antithesis_instrumentation__.Notify(100430)
		return c
	} else {
		__antithesis_instrumentation__.Notify(100431)
	}
	__antithesis_instrumentation__.Notify(100427)
	if a.ID() < b.ID() {
		__antithesis_instrumentation__.Notify(100432)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(100433)
		if a.ID() > b.ID() {
			__antithesis_instrumentation__.Notify(100434)
			return 1
		} else {
			__antithesis_instrumentation__.Notify(100435)
			return 0
		}
	}
}

type keyBound struct {
	key []byte
	inc bool
}

func (b keyBound) compare(o keyBound) int {
	__antithesis_instrumentation__.Notify(100436)
	c := bytes.Compare(b.key, o.key)
	if c != 0 {
		__antithesis_instrumentation__.Notify(100440)
		return c
	} else {
		__antithesis_instrumentation__.Notify(100441)
	}
	__antithesis_instrumentation__.Notify(100437)
	if b.inc == o.inc {
		__antithesis_instrumentation__.Notify(100442)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(100443)
	}
	__antithesis_instrumentation__.Notify(100438)
	if b.inc {
		__antithesis_instrumentation__.Notify(100444)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(100445)
	}
	__antithesis_instrumentation__.Notify(100439)
	return -1
}

func (b keyBound) contains(a *lockState) bool {
	__antithesis_instrumentation__.Notify(100446)
	c := bytes.Compare(a.Key(), b.key)
	if c == 0 {
		__antithesis_instrumentation__.Notify(100448)
		return b.inc
	} else {
		__antithesis_instrumentation__.Notify(100449)
	}
	__antithesis_instrumentation__.Notify(100447)
	return c < 0
}

func upperBound(c *lockState) keyBound {
	__antithesis_instrumentation__.Notify(100450)
	if len(c.EndKey()) != 0 {
		__antithesis_instrumentation__.Notify(100452)
		return keyBound{key: c.EndKey()}
	} else {
		__antithesis_instrumentation__.Notify(100453)
	}
	__antithesis_instrumentation__.Notify(100451)
	return keyBound{key: c.Key(), inc: true}
}

type leafNode struct {
	ref   int32
	count int16
	leaf  bool
	max   keyBound
	items [maxItems]*lockState
}

type node struct {
	leafNode
	children [maxItems + 1]*node
}

func leafToNode(ln *leafNode) *node {
	__antithesis_instrumentation__.Notify(100454)
	return (*node)(unsafe.Pointer(ln))
}

func nodeToLeaf(n *node) *leafNode {
	__antithesis_instrumentation__.Notify(100455)
	return (*leafNode)(unsafe.Pointer(n))
}

var leafPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(100456)
		return new(leafNode)
	},
}

var nodePool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(100457)
		return new(node)
	},
}

func newLeafNode() *node {
	__antithesis_instrumentation__.Notify(100458)
	n := leafToNode(leafPool.Get().(*leafNode))
	n.leaf = true
	n.ref = 1
	return n
}

func newNode() *node {
	__antithesis_instrumentation__.Notify(100459)
	n := nodePool.Get().(*node)
	n.ref = 1
	return n
}

func mut(n **node) *node {
	__antithesis_instrumentation__.Notify(100460)
	if atomic.LoadInt32(&(*n).ref) == 1 {
		__antithesis_instrumentation__.Notify(100462)

		return *n
	} else {
		__antithesis_instrumentation__.Notify(100463)
	}
	__antithesis_instrumentation__.Notify(100461)

	c := (*n).clone()
	(*n).decRef(true)
	*n = c
	return *n
}

func (n *node) incRef() {
	__antithesis_instrumentation__.Notify(100464)
	atomic.AddInt32(&n.ref, 1)
}

func (n *node) decRef(recursive bool) {
	__antithesis_instrumentation__.Notify(100465)
	if atomic.AddInt32(&n.ref, -1) > 0 {
		__antithesis_instrumentation__.Notify(100467)

		return
	} else {
		__antithesis_instrumentation__.Notify(100468)
	}
	__antithesis_instrumentation__.Notify(100466)

	if n.leaf {
		__antithesis_instrumentation__.Notify(100469)
		ln := nodeToLeaf(n)
		*ln = leafNode{}
		leafPool.Put(ln)
	} else {
		__antithesis_instrumentation__.Notify(100470)

		if recursive {
			__antithesis_instrumentation__.Notify(100472)
			for i := int16(0); i <= n.count; i++ {
				__antithesis_instrumentation__.Notify(100473)
				n.children[i].decRef(true)
			}
		} else {
			__antithesis_instrumentation__.Notify(100474)
		}
		__antithesis_instrumentation__.Notify(100471)
		*n = node{}
		nodePool.Put(n)
	}
}

func (n *node) clone() *node {
	__antithesis_instrumentation__.Notify(100475)
	var c *node
	if n.leaf {
		__antithesis_instrumentation__.Notify(100478)
		c = newLeafNode()
	} else {
		__antithesis_instrumentation__.Notify(100479)
		c = newNode()
	}
	__antithesis_instrumentation__.Notify(100476)

	c.count = n.count
	c.max = n.max
	c.items = n.items
	if !c.leaf {
		__antithesis_instrumentation__.Notify(100480)

		c.children = n.children
		for i := int16(0); i <= c.count; i++ {
			__antithesis_instrumentation__.Notify(100481)
			c.children[i].incRef()
		}
	} else {
		__antithesis_instrumentation__.Notify(100482)
	}
	__antithesis_instrumentation__.Notify(100477)
	return c
}

func (n *node) insertAt(index int, item *lockState, nd *node) {
	__antithesis_instrumentation__.Notify(100483)
	if index < int(n.count) {
		__antithesis_instrumentation__.Notify(100486)
		copy(n.items[index+1:n.count+1], n.items[index:n.count])
		if !n.leaf {
			__antithesis_instrumentation__.Notify(100487)
			copy(n.children[index+2:n.count+2], n.children[index+1:n.count+1])
		} else {
			__antithesis_instrumentation__.Notify(100488)
		}
	} else {
		__antithesis_instrumentation__.Notify(100489)
	}
	__antithesis_instrumentation__.Notify(100484)
	n.items[index] = item
	if !n.leaf {
		__antithesis_instrumentation__.Notify(100490)
		n.children[index+1] = nd
	} else {
		__antithesis_instrumentation__.Notify(100491)
	}
	__antithesis_instrumentation__.Notify(100485)
	n.count++
}

func (n *node) pushBack(item *lockState, nd *node) {
	__antithesis_instrumentation__.Notify(100492)
	n.items[n.count] = item
	if !n.leaf {
		__antithesis_instrumentation__.Notify(100494)
		n.children[n.count+1] = nd
	} else {
		__antithesis_instrumentation__.Notify(100495)
	}
	__antithesis_instrumentation__.Notify(100493)
	n.count++
}

func (n *node) pushFront(item *lockState, nd *node) {
	__antithesis_instrumentation__.Notify(100496)
	if !n.leaf {
		__antithesis_instrumentation__.Notify(100498)
		copy(n.children[1:n.count+2], n.children[:n.count+1])
		n.children[0] = nd
	} else {
		__antithesis_instrumentation__.Notify(100499)
	}
	__antithesis_instrumentation__.Notify(100497)
	copy(n.items[1:n.count+1], n.items[:n.count])
	n.items[0] = item
	n.count++
}

func (n *node) removeAt(index int) (*lockState, *node) {
	__antithesis_instrumentation__.Notify(100500)
	var child *node
	if !n.leaf {
		__antithesis_instrumentation__.Notify(100502)
		child = n.children[index+1]
		copy(n.children[index+1:n.count], n.children[index+2:n.count+1])
		n.children[n.count] = nil
	} else {
		__antithesis_instrumentation__.Notify(100503)
	}
	__antithesis_instrumentation__.Notify(100501)
	n.count--
	out := n.items[index]
	copy(n.items[index:n.count], n.items[index+1:n.count+1])
	n.items[n.count] = nilT
	return out, child
}

func (n *node) popBack() (*lockState, *node) {
	__antithesis_instrumentation__.Notify(100504)
	n.count--
	out := n.items[n.count]
	n.items[n.count] = nilT
	if n.leaf {
		__antithesis_instrumentation__.Notify(100506)
		return out, nil
	} else {
		__antithesis_instrumentation__.Notify(100507)
	}
	__antithesis_instrumentation__.Notify(100505)
	child := n.children[n.count+1]
	n.children[n.count+1] = nil
	return out, child
}

func (n *node) popFront() (*lockState, *node) {
	__antithesis_instrumentation__.Notify(100508)
	n.count--
	var child *node
	if !n.leaf {
		__antithesis_instrumentation__.Notify(100510)
		child = n.children[0]
		copy(n.children[:n.count+1], n.children[1:n.count+2])
		n.children[n.count+1] = nil
	} else {
		__antithesis_instrumentation__.Notify(100511)
	}
	__antithesis_instrumentation__.Notify(100509)
	out := n.items[0]
	copy(n.items[:n.count], n.items[1:n.count+1])
	n.items[n.count] = nilT
	return out, child
}

func (n *node) find(item *lockState) (index int, found bool) {
	__antithesis_instrumentation__.Notify(100512)

	i, j := 0, int(n.count)
	for i < j {
		__antithesis_instrumentation__.Notify(100514)
		h := int(uint(i+j) >> 1)

		v := cmp(item, n.items[h])
		if v == 0 {
			__antithesis_instrumentation__.Notify(100515)
			return h, true
		} else {
			__antithesis_instrumentation__.Notify(100516)
			if v > 0 {
				__antithesis_instrumentation__.Notify(100517)
				i = h + 1
			} else {
				__antithesis_instrumentation__.Notify(100518)
				j = h
			}
		}
	}
	__antithesis_instrumentation__.Notify(100513)
	return i, false
}

func (n *node) split(i int) (*lockState, *node) {
	__antithesis_instrumentation__.Notify(100519)
	out := n.items[i]
	var next *node
	if n.leaf {
		__antithesis_instrumentation__.Notify(100524)
		next = newLeafNode()
	} else {
		__antithesis_instrumentation__.Notify(100525)
		next = newNode()
	}
	__antithesis_instrumentation__.Notify(100520)
	next.count = n.count - int16(i+1)
	copy(next.items[:], n.items[i+1:n.count])
	for j := int16(i); j < n.count; j++ {
		__antithesis_instrumentation__.Notify(100526)
		n.items[j] = nilT
	}
	__antithesis_instrumentation__.Notify(100521)
	if !n.leaf {
		__antithesis_instrumentation__.Notify(100527)
		copy(next.children[:], n.children[i+1:n.count+1])
		for j := int16(i + 1); j <= n.count; j++ {
			__antithesis_instrumentation__.Notify(100528)
			n.children[j] = nil
		}
	} else {
		__antithesis_instrumentation__.Notify(100529)
	}
	__antithesis_instrumentation__.Notify(100522)
	n.count = int16(i)

	next.max = next.findUpperBound()
	if n.max.compare(next.max) != 0 && func() bool {
		__antithesis_instrumentation__.Notify(100530)
		return n.max.compare(upperBound(out)) != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(100531)

	} else {
		__antithesis_instrumentation__.Notify(100532)
		n.max = n.findUpperBound()
	}
	__antithesis_instrumentation__.Notify(100523)
	return out, next
}

func (n *node) insert(item *lockState) (replaced, newBound bool) {
	__antithesis_instrumentation__.Notify(100533)
	i, found := n.find(item)
	if found {
		__antithesis_instrumentation__.Notify(100538)
		n.items[i] = item
		return true, false
	} else {
		__antithesis_instrumentation__.Notify(100539)
	}
	__antithesis_instrumentation__.Notify(100534)
	if n.leaf {
		__antithesis_instrumentation__.Notify(100540)
		n.insertAt(i, item, nil)
		return false, n.adjustUpperBoundOnInsertion(item, nil)
	} else {
		__antithesis_instrumentation__.Notify(100541)
	}
	__antithesis_instrumentation__.Notify(100535)
	if n.children[i].count >= maxItems {
		__antithesis_instrumentation__.Notify(100542)
		splitLa, splitNode := mut(&n.children[i]).split(maxItems / 2)
		n.insertAt(i, splitLa, splitNode)

		switch cmp := cmp(item, n.items[i]); {
		case cmp < 0:
			__antithesis_instrumentation__.Notify(100543)

		case cmp > 0:
			__antithesis_instrumentation__.Notify(100544)
			i++
		default:
			__antithesis_instrumentation__.Notify(100545)
			n.items[i] = item
			return true, false
		}
	} else {
		__antithesis_instrumentation__.Notify(100546)
	}
	__antithesis_instrumentation__.Notify(100536)
	replaced, newBound = mut(&n.children[i]).insert(item)
	if newBound {
		__antithesis_instrumentation__.Notify(100547)
		newBound = n.adjustUpperBoundOnInsertion(item, nil)
	} else {
		__antithesis_instrumentation__.Notify(100548)
	}
	__antithesis_instrumentation__.Notify(100537)
	return replaced, newBound
}

func (n *node) removeMax() *lockState {
	__antithesis_instrumentation__.Notify(100549)
	if n.leaf {
		__antithesis_instrumentation__.Notify(100552)
		n.count--
		out := n.items[n.count]
		n.items[n.count] = nilT
		n.adjustUpperBoundOnRemoval(out, nil)
		return out
	} else {
		__antithesis_instrumentation__.Notify(100553)
	}
	__antithesis_instrumentation__.Notify(100550)

	i := int(n.count)
	if n.children[i].count <= minItems {
		__antithesis_instrumentation__.Notify(100554)

		n.rebalanceOrMerge(i)
		return n.removeMax()
	} else {
		__antithesis_instrumentation__.Notify(100555)
	}
	__antithesis_instrumentation__.Notify(100551)
	child := mut(&n.children[i])
	out := child.removeMax()
	n.adjustUpperBoundOnRemoval(out, nil)
	return out
}

func (n *node) remove(item *lockState) (out *lockState, newBound bool) {
	__antithesis_instrumentation__.Notify(100556)
	i, found := n.find(item)
	if n.leaf {
		__antithesis_instrumentation__.Notify(100561)
		if found {
			__antithesis_instrumentation__.Notify(100563)
			out, _ = n.removeAt(i)
			return out, n.adjustUpperBoundOnRemoval(out, nil)
		} else {
			__antithesis_instrumentation__.Notify(100564)
		}
		__antithesis_instrumentation__.Notify(100562)
		return nilT, false
	} else {
		__antithesis_instrumentation__.Notify(100565)
	}
	__antithesis_instrumentation__.Notify(100557)
	if n.children[i].count <= minItems {
		__antithesis_instrumentation__.Notify(100566)

		n.rebalanceOrMerge(i)
		return n.remove(item)
	} else {
		__antithesis_instrumentation__.Notify(100567)
	}
	__antithesis_instrumentation__.Notify(100558)
	child := mut(&n.children[i])
	if found {
		__antithesis_instrumentation__.Notify(100568)

		out = n.items[i]
		n.items[i] = child.removeMax()
		return out, n.adjustUpperBoundOnRemoval(out, nil)
	} else {
		__antithesis_instrumentation__.Notify(100569)
	}
	__antithesis_instrumentation__.Notify(100559)

	out, newBound = child.remove(item)
	if newBound {
		__antithesis_instrumentation__.Notify(100570)
		newBound = n.adjustUpperBoundOnRemoval(out, nil)
	} else {
		__antithesis_instrumentation__.Notify(100571)
	}
	__antithesis_instrumentation__.Notify(100560)
	return out, newBound
}

func (n *node) rebalanceOrMerge(i int) {
	__antithesis_instrumentation__.Notify(100572)
	switch {
	case i > 0 && func() bool {
		__antithesis_instrumentation__.Notify(100578)
		return n.children[i-1].count > minItems == true
	}() == true:
		__antithesis_instrumentation__.Notify(100573)

		left := mut(&n.children[i-1])
		child := mut(&n.children[i])
		xLa, grandChild := left.popBack()
		yLa := n.items[i-1]
		child.pushFront(yLa, grandChild)
		n.items[i-1] = xLa

		left.adjustUpperBoundOnRemoval(xLa, grandChild)
		child.adjustUpperBoundOnInsertion(yLa, grandChild)

	case i < int(n.count) && func() bool {
		__antithesis_instrumentation__.Notify(100579)
		return n.children[i+1].count > minItems == true
	}() == true:
		__antithesis_instrumentation__.Notify(100574)

		right := mut(&n.children[i+1])
		child := mut(&n.children[i])
		xLa, grandChild := right.popFront()
		yLa := n.items[i]
		child.pushBack(yLa, grandChild)
		n.items[i] = xLa

		right.adjustUpperBoundOnRemoval(xLa, grandChild)
		child.adjustUpperBoundOnInsertion(yLa, grandChild)

	default:
		__antithesis_instrumentation__.Notify(100575)

		if i >= int(n.count) {
			__antithesis_instrumentation__.Notify(100580)
			i = int(n.count - 1)
		} else {
			__antithesis_instrumentation__.Notify(100581)
		}
		__antithesis_instrumentation__.Notify(100576)
		child := mut(&n.children[i])

		_ = mut(&n.children[i+1])
		mergeLa, mergeChild := n.removeAt(i)
		child.items[child.count] = mergeLa
		copy(child.items[child.count+1:], mergeChild.items[:mergeChild.count])
		if !child.leaf {
			__antithesis_instrumentation__.Notify(100582)
			copy(child.children[child.count+1:], mergeChild.children[:mergeChild.count+1])
		} else {
			__antithesis_instrumentation__.Notify(100583)
		}
		__antithesis_instrumentation__.Notify(100577)
		child.count += mergeChild.count + 1

		child.adjustUpperBoundOnInsertion(mergeLa, mergeChild)
		mergeChild.decRef(false)
	}
}

func (n *node) findUpperBound() keyBound {
	__antithesis_instrumentation__.Notify(100584)
	var max keyBound
	for i := int16(0); i < n.count; i++ {
		__antithesis_instrumentation__.Notify(100587)
		up := upperBound(n.items[i])
		if max.compare(up) < 0 {
			__antithesis_instrumentation__.Notify(100588)
			max = up
		} else {
			__antithesis_instrumentation__.Notify(100589)
		}
	}
	__antithesis_instrumentation__.Notify(100585)
	if !n.leaf {
		__antithesis_instrumentation__.Notify(100590)
		for i := int16(0); i <= n.count; i++ {
			__antithesis_instrumentation__.Notify(100591)
			up := n.children[i].max
			if max.compare(up) < 0 {
				__antithesis_instrumentation__.Notify(100592)
				max = up
			} else {
				__antithesis_instrumentation__.Notify(100593)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(100594)
	}
	__antithesis_instrumentation__.Notify(100586)
	return max
}

func (n *node) adjustUpperBoundOnInsertion(item *lockState, child *node) bool {
	__antithesis_instrumentation__.Notify(100595)
	up := upperBound(item)
	if child != nil {
		__antithesis_instrumentation__.Notify(100598)
		if up.compare(child.max) < 0 {
			__antithesis_instrumentation__.Notify(100599)
			up = child.max
		} else {
			__antithesis_instrumentation__.Notify(100600)
		}
	} else {
		__antithesis_instrumentation__.Notify(100601)
	}
	__antithesis_instrumentation__.Notify(100596)
	if n.max.compare(up) < 0 {
		__antithesis_instrumentation__.Notify(100602)
		n.max = up
		return true
	} else {
		__antithesis_instrumentation__.Notify(100603)
	}
	__antithesis_instrumentation__.Notify(100597)
	return false
}

func (n *node) adjustUpperBoundOnRemoval(item *lockState, child *node) bool {
	__antithesis_instrumentation__.Notify(100604)
	up := upperBound(item)
	if child != nil {
		__antithesis_instrumentation__.Notify(100607)
		if up.compare(child.max) < 0 {
			__antithesis_instrumentation__.Notify(100608)
			up = child.max
		} else {
			__antithesis_instrumentation__.Notify(100609)
		}
	} else {
		__antithesis_instrumentation__.Notify(100610)
	}
	__antithesis_instrumentation__.Notify(100605)
	if n.max.compare(up) == 0 {
		__antithesis_instrumentation__.Notify(100611)

		n.max = n.findUpperBound()
		return n.max.compare(up) != 0
	} else {
		__antithesis_instrumentation__.Notify(100612)
	}
	__antithesis_instrumentation__.Notify(100606)
	return false
}

type btree struct {
	root   *node
	length int
}

func (t *btree) Reset() {
	__antithesis_instrumentation__.Notify(100613)
	if t.root != nil {
		__antithesis_instrumentation__.Notify(100615)
		t.root.decRef(true)
		t.root = nil
	} else {
		__antithesis_instrumentation__.Notify(100616)
	}
	__antithesis_instrumentation__.Notify(100614)
	t.length = 0
}

func (t *btree) Clone() btree {
	__antithesis_instrumentation__.Notify(100617)
	c := *t
	if c.root != nil {
		__antithesis_instrumentation__.Notify(100619)

		c.root.incRef()
	} else {
		__antithesis_instrumentation__.Notify(100620)
	}
	__antithesis_instrumentation__.Notify(100618)
	return c
}

func (t *btree) Delete(item *lockState) {
	__antithesis_instrumentation__.Notify(100621)
	if t.root == nil || func() bool {
		__antithesis_instrumentation__.Notify(100624)
		return t.root.count == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(100625)
		return
	} else {
		__antithesis_instrumentation__.Notify(100626)
	}
	__antithesis_instrumentation__.Notify(100622)
	if out, _ := mut(&t.root).remove(item); out != nilT {
		__antithesis_instrumentation__.Notify(100627)
		t.length--
	} else {
		__antithesis_instrumentation__.Notify(100628)
	}
	__antithesis_instrumentation__.Notify(100623)
	if t.root.count == 0 {
		__antithesis_instrumentation__.Notify(100629)
		old := t.root
		if t.root.leaf {
			__antithesis_instrumentation__.Notify(100631)
			t.root = nil
		} else {
			__antithesis_instrumentation__.Notify(100632)
			t.root = t.root.children[0]
		}
		__antithesis_instrumentation__.Notify(100630)
		old.decRef(false)
	} else {
		__antithesis_instrumentation__.Notify(100633)
	}
}

func (t *btree) Set(item *lockState) {
	__antithesis_instrumentation__.Notify(100634)
	if t.root == nil {
		__antithesis_instrumentation__.Notify(100636)
		t.root = newLeafNode()
	} else {
		__antithesis_instrumentation__.Notify(100637)
		if t.root.count >= maxItems {
			__antithesis_instrumentation__.Notify(100638)
			splitLa, splitNode := mut(&t.root).split(maxItems / 2)
			newRoot := newNode()
			newRoot.count = 1
			newRoot.items[0] = splitLa
			newRoot.children[0] = t.root
			newRoot.children[1] = splitNode
			newRoot.max = newRoot.findUpperBound()
			t.root = newRoot
		} else {
			__antithesis_instrumentation__.Notify(100639)
		}
	}
	__antithesis_instrumentation__.Notify(100635)
	if replaced, _ := mut(&t.root).insert(item); !replaced {
		__antithesis_instrumentation__.Notify(100640)
		t.length++
	} else {
		__antithesis_instrumentation__.Notify(100641)
	}
}

func (t *btree) MakeIter() iterator {
	__antithesis_instrumentation__.Notify(100642)
	return iterator{r: t.root, pos: -1}
}

func (t *btree) Height() int {
	__antithesis_instrumentation__.Notify(100643)
	if t.root == nil {
		__antithesis_instrumentation__.Notify(100646)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(100647)
	}
	__antithesis_instrumentation__.Notify(100644)
	h := 1
	n := t.root
	for !n.leaf {
		__antithesis_instrumentation__.Notify(100648)
		n = n.children[0]
		h++
	}
	__antithesis_instrumentation__.Notify(100645)
	return h
}

func (t *btree) Len() int {
	__antithesis_instrumentation__.Notify(100649)
	return t.length
}

func (t *btree) String() string {
	__antithesis_instrumentation__.Notify(100650)
	if t.length == 0 {
		__antithesis_instrumentation__.Notify(100652)
		return ";"
	} else {
		__antithesis_instrumentation__.Notify(100653)
	}
	__antithesis_instrumentation__.Notify(100651)
	var b strings.Builder
	t.root.writeString(&b)
	return b.String()
}

func (n *node) writeString(b *strings.Builder) {
	__antithesis_instrumentation__.Notify(100654)
	if n.leaf {
		__antithesis_instrumentation__.Notify(100656)
		for i := int16(0); i < n.count; i++ {
			__antithesis_instrumentation__.Notify(100658)
			if i != 0 {
				__antithesis_instrumentation__.Notify(100660)
				b.WriteString(",")
			} else {
				__antithesis_instrumentation__.Notify(100661)
			}
			__antithesis_instrumentation__.Notify(100659)
			b.WriteString(n.items[i].String())
		}
		__antithesis_instrumentation__.Notify(100657)
		return
	} else {
		__antithesis_instrumentation__.Notify(100662)
	}
	__antithesis_instrumentation__.Notify(100655)
	for i := int16(0); i <= n.count; i++ {
		__antithesis_instrumentation__.Notify(100663)
		b.WriteString("(")
		n.children[i].writeString(b)
		b.WriteString(")")
		if i < n.count {
			__antithesis_instrumentation__.Notify(100664)
			b.WriteString(n.items[i].String())
		} else {
			__antithesis_instrumentation__.Notify(100665)
		}
	}
}

type iterStack struct {
	a    iterStackArr
	aLen int16
	s    []iterFrame
}

type iterStackArr [3]iterFrame

type iterFrame struct {
	n   *node
	pos int16
}

func (is *iterStack) push(f iterFrame) {
	__antithesis_instrumentation__.Notify(100666)
	if is.aLen == -1 {
		__antithesis_instrumentation__.Notify(100667)
		is.s = append(is.s, f)
	} else {
		__antithesis_instrumentation__.Notify(100668)
		if int(is.aLen) == len(is.a) {
			__antithesis_instrumentation__.Notify(100669)
			is.s = make([]iterFrame, int(is.aLen)+1, 2*int(is.aLen))
			copy(is.s, is.a[:])
			is.s[int(is.aLen)] = f
			is.aLen = -1
		} else {
			__antithesis_instrumentation__.Notify(100670)
			is.a[is.aLen] = f
			is.aLen++
		}
	}
}

func (is *iterStack) pop() iterFrame {
	__antithesis_instrumentation__.Notify(100671)
	if is.aLen == -1 {
		__antithesis_instrumentation__.Notify(100673)
		f := is.s[len(is.s)-1]
		is.s = is.s[:len(is.s)-1]
		return f
	} else {
		__antithesis_instrumentation__.Notify(100674)
	}
	__antithesis_instrumentation__.Notify(100672)
	is.aLen--
	return is.a[is.aLen]
}

func (is *iterStack) len() int {
	__antithesis_instrumentation__.Notify(100675)
	if is.aLen == -1 {
		__antithesis_instrumentation__.Notify(100677)
		return len(is.s)
	} else {
		__antithesis_instrumentation__.Notify(100678)
	}
	__antithesis_instrumentation__.Notify(100676)
	return int(is.aLen)
}

func (is *iterStack) reset() {
	__antithesis_instrumentation__.Notify(100679)
	if is.aLen == -1 {
		__antithesis_instrumentation__.Notify(100680)
		is.s = is.s[:0]
	} else {
		__antithesis_instrumentation__.Notify(100681)
		is.aLen = 0
	}
}

type iterator struct {
	r   *node
	n   *node
	pos int16
	s   iterStack
	o   overlapScan
}

func (i *iterator) reset() {
	__antithesis_instrumentation__.Notify(100682)
	i.n = i.r
	i.pos = -1
	i.s.reset()
	i.o = overlapScan{}
}

func (i *iterator) descend(n *node, pos int16) {
	__antithesis_instrumentation__.Notify(100683)
	i.s.push(iterFrame{n: n, pos: pos})
	i.n = n.children[pos]
	i.pos = 0
}

func (i *iterator) ascend() {
	__antithesis_instrumentation__.Notify(100684)
	f := i.s.pop()
	i.n = f.n
	i.pos = f.pos
}

func (i *iterator) SeekGE(item *lockState) {
	__antithesis_instrumentation__.Notify(100685)
	i.reset()
	if i.n == nil {
		__antithesis_instrumentation__.Notify(100687)
		return
	} else {
		__antithesis_instrumentation__.Notify(100688)
	}
	__antithesis_instrumentation__.Notify(100686)
	for {
		__antithesis_instrumentation__.Notify(100689)
		pos, found := i.n.find(item)
		i.pos = int16(pos)
		if found {
			__antithesis_instrumentation__.Notify(100692)
			return
		} else {
			__antithesis_instrumentation__.Notify(100693)
		}
		__antithesis_instrumentation__.Notify(100690)
		if i.n.leaf {
			__antithesis_instrumentation__.Notify(100694)
			if i.pos == i.n.count {
				__antithesis_instrumentation__.Notify(100696)
				i.Next()
			} else {
				__antithesis_instrumentation__.Notify(100697)
			}
			__antithesis_instrumentation__.Notify(100695)
			return
		} else {
			__antithesis_instrumentation__.Notify(100698)
		}
		__antithesis_instrumentation__.Notify(100691)
		i.descend(i.n, i.pos)
	}
}

func (i *iterator) SeekLT(item *lockState) {
	__antithesis_instrumentation__.Notify(100699)
	i.reset()
	if i.n == nil {
		__antithesis_instrumentation__.Notify(100701)
		return
	} else {
		__antithesis_instrumentation__.Notify(100702)
	}
	__antithesis_instrumentation__.Notify(100700)
	for {
		__antithesis_instrumentation__.Notify(100703)
		pos, found := i.n.find(item)
		i.pos = int16(pos)
		if found || func() bool {
			__antithesis_instrumentation__.Notify(100705)
			return i.n.leaf == true
		}() == true {
			__antithesis_instrumentation__.Notify(100706)
			i.Prev()
			return
		} else {
			__antithesis_instrumentation__.Notify(100707)
		}
		__antithesis_instrumentation__.Notify(100704)
		i.descend(i.n, i.pos)
	}
}

func (i *iterator) First() {
	__antithesis_instrumentation__.Notify(100708)
	i.reset()
	if i.n == nil {
		__antithesis_instrumentation__.Notify(100711)
		return
	} else {
		__antithesis_instrumentation__.Notify(100712)
	}
	__antithesis_instrumentation__.Notify(100709)
	for !i.n.leaf {
		__antithesis_instrumentation__.Notify(100713)
		i.descend(i.n, 0)
	}
	__antithesis_instrumentation__.Notify(100710)
	i.pos = 0
}

func (i *iterator) Last() {
	__antithesis_instrumentation__.Notify(100714)
	i.reset()
	if i.n == nil {
		__antithesis_instrumentation__.Notify(100717)
		return
	} else {
		__antithesis_instrumentation__.Notify(100718)
	}
	__antithesis_instrumentation__.Notify(100715)
	for !i.n.leaf {
		__antithesis_instrumentation__.Notify(100719)
		i.descend(i.n, i.n.count)
	}
	__antithesis_instrumentation__.Notify(100716)
	i.pos = i.n.count - 1
}

func (i *iterator) Next() {
	__antithesis_instrumentation__.Notify(100720)
	if i.n == nil {
		__antithesis_instrumentation__.Notify(100724)
		return
	} else {
		__antithesis_instrumentation__.Notify(100725)
	}
	__antithesis_instrumentation__.Notify(100721)

	if i.n.leaf {
		__antithesis_instrumentation__.Notify(100726)
		i.pos++
		if i.pos < i.n.count {
			__antithesis_instrumentation__.Notify(100729)
			return
		} else {
			__antithesis_instrumentation__.Notify(100730)
		}
		__antithesis_instrumentation__.Notify(100727)
		for i.s.len() > 0 && func() bool {
			__antithesis_instrumentation__.Notify(100731)
			return i.pos >= i.n.count == true
		}() == true {
			__antithesis_instrumentation__.Notify(100732)
			i.ascend()
		}
		__antithesis_instrumentation__.Notify(100728)
		return
	} else {
		__antithesis_instrumentation__.Notify(100733)
	}
	__antithesis_instrumentation__.Notify(100722)

	i.descend(i.n, i.pos+1)
	for !i.n.leaf {
		__antithesis_instrumentation__.Notify(100734)
		i.descend(i.n, 0)
	}
	__antithesis_instrumentation__.Notify(100723)
	i.pos = 0
}

func (i *iterator) Prev() {
	__antithesis_instrumentation__.Notify(100735)
	if i.n == nil {
		__antithesis_instrumentation__.Notify(100739)
		return
	} else {
		__antithesis_instrumentation__.Notify(100740)
	}
	__antithesis_instrumentation__.Notify(100736)

	if i.n.leaf {
		__antithesis_instrumentation__.Notify(100741)
		i.pos--
		if i.pos >= 0 {
			__antithesis_instrumentation__.Notify(100744)
			return
		} else {
			__antithesis_instrumentation__.Notify(100745)
		}
		__antithesis_instrumentation__.Notify(100742)
		for i.s.len() > 0 && func() bool {
			__antithesis_instrumentation__.Notify(100746)
			return i.pos < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(100747)
			i.ascend()
			i.pos--
		}
		__antithesis_instrumentation__.Notify(100743)
		return
	} else {
		__antithesis_instrumentation__.Notify(100748)
	}
	__antithesis_instrumentation__.Notify(100737)

	i.descend(i.n, i.pos)
	for !i.n.leaf {
		__antithesis_instrumentation__.Notify(100749)
		i.descend(i.n, i.n.count)
	}
	__antithesis_instrumentation__.Notify(100738)
	i.pos = i.n.count - 1
}

func (i *iterator) Valid() bool {
	__antithesis_instrumentation__.Notify(100750)
	return i.pos >= 0 && func() bool {
		__antithesis_instrumentation__.Notify(100751)
		return i.pos < i.n.count == true
	}() == true
}

func (i *iterator) Cur() *lockState {
	__antithesis_instrumentation__.Notify(100752)
	return i.n.items[i.pos]
}

type overlapScan struct {
	constrMinN       *node
	constrMinPos     int16
	constrMinReached bool

	constrMaxN   *node
	constrMaxPos int16
}

func (i *iterator) FirstOverlap(item *lockState) {
	__antithesis_instrumentation__.Notify(100753)
	i.reset()
	if i.n == nil {
		__antithesis_instrumentation__.Notify(100755)
		return
	} else {
		__antithesis_instrumentation__.Notify(100756)
	}
	__antithesis_instrumentation__.Notify(100754)
	i.pos = 0
	i.o = overlapScan{}
	i.constrainMinSearchBounds(item)
	i.constrainMaxSearchBounds(item)
	i.findNextOverlap(item)
}

func (i *iterator) NextOverlap(item *lockState) {
	__antithesis_instrumentation__.Notify(100757)
	if i.n == nil {
		__antithesis_instrumentation__.Notify(100759)
		return
	} else {
		__antithesis_instrumentation__.Notify(100760)
	}
	__antithesis_instrumentation__.Notify(100758)
	i.pos++
	i.findNextOverlap(item)
}

func (i *iterator) constrainMinSearchBounds(item *lockState) {
	__antithesis_instrumentation__.Notify(100761)
	k := item.Key()
	j := sort.Search(int(i.n.count), func(j int) bool {
		__antithesis_instrumentation__.Notify(100763)
		return bytes.Compare(k, i.n.items[j].Key()) <= 0
	})
	__antithesis_instrumentation__.Notify(100762)
	i.o.constrMinN = i.n
	i.o.constrMinPos = int16(j)
}

func (i *iterator) constrainMaxSearchBounds(item *lockState) {
	__antithesis_instrumentation__.Notify(100764)
	up := upperBound(item)
	j := sort.Search(int(i.n.count), func(j int) bool {
		__antithesis_instrumentation__.Notify(100766)
		return !up.contains(i.n.items[j])
	})
	__antithesis_instrumentation__.Notify(100765)
	i.o.constrMaxN = i.n
	i.o.constrMaxPos = int16(j)
}

func (i *iterator) findNextOverlap(item *lockState) {
	__antithesis_instrumentation__.Notify(100767)
	for {
		__antithesis_instrumentation__.Notify(100768)
		if i.pos > i.n.count {
			__antithesis_instrumentation__.Notify(100773)

			i.ascend()
		} else {
			__antithesis_instrumentation__.Notify(100774)
			if !i.n.leaf {
				__antithesis_instrumentation__.Notify(100775)

				if i.o.constrMinReached || func() bool {
					__antithesis_instrumentation__.Notify(100776)
					return i.n.children[i.pos].max.contains(item) == true
				}() == true {
					__antithesis_instrumentation__.Notify(100777)
					par := i.n
					pos := i.pos
					i.descend(par, pos)

					if par == i.o.constrMinN && func() bool {
						__antithesis_instrumentation__.Notify(100780)
						return pos == i.o.constrMinPos == true
					}() == true {
						__antithesis_instrumentation__.Notify(100781)
						i.constrainMinSearchBounds(item)
					} else {
						__antithesis_instrumentation__.Notify(100782)
					}
					__antithesis_instrumentation__.Notify(100778)
					if par == i.o.constrMaxN && func() bool {
						__antithesis_instrumentation__.Notify(100783)
						return pos == i.o.constrMaxPos == true
					}() == true {
						__antithesis_instrumentation__.Notify(100784)
						i.constrainMaxSearchBounds(item)
					} else {
						__antithesis_instrumentation__.Notify(100785)
					}
					__antithesis_instrumentation__.Notify(100779)
					continue
				} else {
					__antithesis_instrumentation__.Notify(100786)
				}
			} else {
				__antithesis_instrumentation__.Notify(100787)
			}
		}
		__antithesis_instrumentation__.Notify(100769)

		if i.n == i.o.constrMaxN && func() bool {
			__antithesis_instrumentation__.Notify(100788)
			return i.pos == i.o.constrMaxPos == true
		}() == true {
			__antithesis_instrumentation__.Notify(100789)

			i.pos = i.n.count
			return
		} else {
			__antithesis_instrumentation__.Notify(100790)
		}
		__antithesis_instrumentation__.Notify(100770)
		if i.n == i.o.constrMinN && func() bool {
			__antithesis_instrumentation__.Notify(100791)
			return i.pos == i.o.constrMinPos == true
		}() == true {
			__antithesis_instrumentation__.Notify(100792)

			i.o.constrMinReached = true
		} else {
			__antithesis_instrumentation__.Notify(100793)
		}
		__antithesis_instrumentation__.Notify(100771)

		if i.pos < i.n.count {
			__antithesis_instrumentation__.Notify(100794)

			if i.o.constrMinReached {
				__antithesis_instrumentation__.Notify(100796)

				return
			} else {
				__antithesis_instrumentation__.Notify(100797)
			}
			__antithesis_instrumentation__.Notify(100795)
			if upperBound(i.n.items[i.pos]).contains(item) {
				__antithesis_instrumentation__.Notify(100798)
				return
			} else {
				__antithesis_instrumentation__.Notify(100799)
			}
		} else {
			__antithesis_instrumentation__.Notify(100800)
		}
		__antithesis_instrumentation__.Notify(100772)
		i.pos++
	}
}
