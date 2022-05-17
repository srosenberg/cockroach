package spanlatch

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

var nilT *latch

const (
	degree   = 16
	maxItems = 2*degree - 1
	minItems = degree - 1
)

func cmp(a, b *latch) int {
	__antithesis_instrumentation__.Notify(122345)
	c := bytes.Compare(a.Key(), b.Key())
	if c != 0 {
		__antithesis_instrumentation__.Notify(122348)
		return c
	} else {
		__antithesis_instrumentation__.Notify(122349)
	}
	__antithesis_instrumentation__.Notify(122346)
	c = bytes.Compare(a.EndKey(), b.EndKey())
	if c != 0 {
		__antithesis_instrumentation__.Notify(122350)
		return c
	} else {
		__antithesis_instrumentation__.Notify(122351)
	}
	__antithesis_instrumentation__.Notify(122347)
	if a.ID() < b.ID() {
		__antithesis_instrumentation__.Notify(122352)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(122353)
		if a.ID() > b.ID() {
			__antithesis_instrumentation__.Notify(122354)
			return 1
		} else {
			__antithesis_instrumentation__.Notify(122355)
			return 0
		}
	}
}

type keyBound struct {
	key []byte
	inc bool
}

func (b keyBound) compare(o keyBound) int {
	__antithesis_instrumentation__.Notify(122356)
	c := bytes.Compare(b.key, o.key)
	if c != 0 {
		__antithesis_instrumentation__.Notify(122360)
		return c
	} else {
		__antithesis_instrumentation__.Notify(122361)
	}
	__antithesis_instrumentation__.Notify(122357)
	if b.inc == o.inc {
		__antithesis_instrumentation__.Notify(122362)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(122363)
	}
	__antithesis_instrumentation__.Notify(122358)
	if b.inc {
		__antithesis_instrumentation__.Notify(122364)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(122365)
	}
	__antithesis_instrumentation__.Notify(122359)
	return -1
}

func (b keyBound) contains(a *latch) bool {
	__antithesis_instrumentation__.Notify(122366)
	c := bytes.Compare(a.Key(), b.key)
	if c == 0 {
		__antithesis_instrumentation__.Notify(122368)
		return b.inc
	} else {
		__antithesis_instrumentation__.Notify(122369)
	}
	__antithesis_instrumentation__.Notify(122367)
	return c < 0
}

func upperBound(c *latch) keyBound {
	__antithesis_instrumentation__.Notify(122370)
	if len(c.EndKey()) != 0 {
		__antithesis_instrumentation__.Notify(122372)
		return keyBound{key: c.EndKey()}
	} else {
		__antithesis_instrumentation__.Notify(122373)
	}
	__antithesis_instrumentation__.Notify(122371)
	return keyBound{key: c.Key(), inc: true}
}

type leafNode struct {
	ref   int32
	count int16
	leaf  bool
	max   keyBound
	items [maxItems]*latch
}

type node struct {
	leafNode
	children [maxItems + 1]*node
}

func leafToNode(ln *leafNode) *node {
	__antithesis_instrumentation__.Notify(122374)
	return (*node)(unsafe.Pointer(ln))
}

func nodeToLeaf(n *node) *leafNode {
	__antithesis_instrumentation__.Notify(122375)
	return (*leafNode)(unsafe.Pointer(n))
}

var leafPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(122376)
		return new(leafNode)
	},
}

var nodePool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(122377)
		return new(node)
	},
}

func newLeafNode() *node {
	__antithesis_instrumentation__.Notify(122378)
	n := leafToNode(leafPool.Get().(*leafNode))
	n.leaf = true
	n.ref = 1
	return n
}

func newNode() *node {
	__antithesis_instrumentation__.Notify(122379)
	n := nodePool.Get().(*node)
	n.ref = 1
	return n
}

func mut(n **node) *node {
	__antithesis_instrumentation__.Notify(122380)
	if atomic.LoadInt32(&(*n).ref) == 1 {
		__antithesis_instrumentation__.Notify(122382)

		return *n
	} else {
		__antithesis_instrumentation__.Notify(122383)
	}
	__antithesis_instrumentation__.Notify(122381)

	c := (*n).clone()
	(*n).decRef(true)
	*n = c
	return *n
}

func (n *node) incRef() {
	__antithesis_instrumentation__.Notify(122384)
	atomic.AddInt32(&n.ref, 1)
}

func (n *node) decRef(recursive bool) {
	__antithesis_instrumentation__.Notify(122385)
	if atomic.AddInt32(&n.ref, -1) > 0 {
		__antithesis_instrumentation__.Notify(122387)

		return
	} else {
		__antithesis_instrumentation__.Notify(122388)
	}
	__antithesis_instrumentation__.Notify(122386)

	if n.leaf {
		__antithesis_instrumentation__.Notify(122389)
		ln := nodeToLeaf(n)
		*ln = leafNode{}
		leafPool.Put(ln)
	} else {
		__antithesis_instrumentation__.Notify(122390)

		if recursive {
			__antithesis_instrumentation__.Notify(122392)
			for i := int16(0); i <= n.count; i++ {
				__antithesis_instrumentation__.Notify(122393)
				n.children[i].decRef(true)
			}
		} else {
			__antithesis_instrumentation__.Notify(122394)
		}
		__antithesis_instrumentation__.Notify(122391)
		*n = node{}
		nodePool.Put(n)
	}
}

func (n *node) clone() *node {
	__antithesis_instrumentation__.Notify(122395)
	var c *node
	if n.leaf {
		__antithesis_instrumentation__.Notify(122398)
		c = newLeafNode()
	} else {
		__antithesis_instrumentation__.Notify(122399)
		c = newNode()
	}
	__antithesis_instrumentation__.Notify(122396)

	c.count = n.count
	c.max = n.max
	c.items = n.items
	if !c.leaf {
		__antithesis_instrumentation__.Notify(122400)

		c.children = n.children
		for i := int16(0); i <= c.count; i++ {
			__antithesis_instrumentation__.Notify(122401)
			c.children[i].incRef()
		}
	} else {
		__antithesis_instrumentation__.Notify(122402)
	}
	__antithesis_instrumentation__.Notify(122397)
	return c
}

func (n *node) insertAt(index int, item *latch, nd *node) {
	__antithesis_instrumentation__.Notify(122403)
	if index < int(n.count) {
		__antithesis_instrumentation__.Notify(122406)
		copy(n.items[index+1:n.count+1], n.items[index:n.count])
		if !n.leaf {
			__antithesis_instrumentation__.Notify(122407)
			copy(n.children[index+2:n.count+2], n.children[index+1:n.count+1])
		} else {
			__antithesis_instrumentation__.Notify(122408)
		}
	} else {
		__antithesis_instrumentation__.Notify(122409)
	}
	__antithesis_instrumentation__.Notify(122404)
	n.items[index] = item
	if !n.leaf {
		__antithesis_instrumentation__.Notify(122410)
		n.children[index+1] = nd
	} else {
		__antithesis_instrumentation__.Notify(122411)
	}
	__antithesis_instrumentation__.Notify(122405)
	n.count++
}

func (n *node) pushBack(item *latch, nd *node) {
	__antithesis_instrumentation__.Notify(122412)
	n.items[n.count] = item
	if !n.leaf {
		__antithesis_instrumentation__.Notify(122414)
		n.children[n.count+1] = nd
	} else {
		__antithesis_instrumentation__.Notify(122415)
	}
	__antithesis_instrumentation__.Notify(122413)
	n.count++
}

func (n *node) pushFront(item *latch, nd *node) {
	__antithesis_instrumentation__.Notify(122416)
	if !n.leaf {
		__antithesis_instrumentation__.Notify(122418)
		copy(n.children[1:n.count+2], n.children[:n.count+1])
		n.children[0] = nd
	} else {
		__antithesis_instrumentation__.Notify(122419)
	}
	__antithesis_instrumentation__.Notify(122417)
	copy(n.items[1:n.count+1], n.items[:n.count])
	n.items[0] = item
	n.count++
}

func (n *node) removeAt(index int) (*latch, *node) {
	__antithesis_instrumentation__.Notify(122420)
	var child *node
	if !n.leaf {
		__antithesis_instrumentation__.Notify(122422)
		child = n.children[index+1]
		copy(n.children[index+1:n.count], n.children[index+2:n.count+1])
		n.children[n.count] = nil
	} else {
		__antithesis_instrumentation__.Notify(122423)
	}
	__antithesis_instrumentation__.Notify(122421)
	n.count--
	out := n.items[index]
	copy(n.items[index:n.count], n.items[index+1:n.count+1])
	n.items[n.count] = nilT
	return out, child
}

func (n *node) popBack() (*latch, *node) {
	__antithesis_instrumentation__.Notify(122424)
	n.count--
	out := n.items[n.count]
	n.items[n.count] = nilT
	if n.leaf {
		__antithesis_instrumentation__.Notify(122426)
		return out, nil
	} else {
		__antithesis_instrumentation__.Notify(122427)
	}
	__antithesis_instrumentation__.Notify(122425)
	child := n.children[n.count+1]
	n.children[n.count+1] = nil
	return out, child
}

func (n *node) popFront() (*latch, *node) {
	__antithesis_instrumentation__.Notify(122428)
	n.count--
	var child *node
	if !n.leaf {
		__antithesis_instrumentation__.Notify(122430)
		child = n.children[0]
		copy(n.children[:n.count+1], n.children[1:n.count+2])
		n.children[n.count+1] = nil
	} else {
		__antithesis_instrumentation__.Notify(122431)
	}
	__antithesis_instrumentation__.Notify(122429)
	out := n.items[0]
	copy(n.items[:n.count], n.items[1:n.count+1])
	n.items[n.count] = nilT
	return out, child
}

func (n *node) find(item *latch) (index int, found bool) {
	__antithesis_instrumentation__.Notify(122432)

	i, j := 0, int(n.count)
	for i < j {
		__antithesis_instrumentation__.Notify(122434)
		h := int(uint(i+j) >> 1)

		v := cmp(item, n.items[h])
		if v == 0 {
			__antithesis_instrumentation__.Notify(122435)
			return h, true
		} else {
			__antithesis_instrumentation__.Notify(122436)
			if v > 0 {
				__antithesis_instrumentation__.Notify(122437)
				i = h + 1
			} else {
				__antithesis_instrumentation__.Notify(122438)
				j = h
			}
		}
	}
	__antithesis_instrumentation__.Notify(122433)
	return i, false
}

func (n *node) split(i int) (*latch, *node) {
	__antithesis_instrumentation__.Notify(122439)
	out := n.items[i]
	var next *node
	if n.leaf {
		__antithesis_instrumentation__.Notify(122444)
		next = newLeafNode()
	} else {
		__antithesis_instrumentation__.Notify(122445)
		next = newNode()
	}
	__antithesis_instrumentation__.Notify(122440)
	next.count = n.count - int16(i+1)
	copy(next.items[:], n.items[i+1:n.count])
	for j := int16(i); j < n.count; j++ {
		__antithesis_instrumentation__.Notify(122446)
		n.items[j] = nilT
	}
	__antithesis_instrumentation__.Notify(122441)
	if !n.leaf {
		__antithesis_instrumentation__.Notify(122447)
		copy(next.children[:], n.children[i+1:n.count+1])
		for j := int16(i + 1); j <= n.count; j++ {
			__antithesis_instrumentation__.Notify(122448)
			n.children[j] = nil
		}
	} else {
		__antithesis_instrumentation__.Notify(122449)
	}
	__antithesis_instrumentation__.Notify(122442)
	n.count = int16(i)

	next.max = next.findUpperBound()
	if n.max.compare(next.max) != 0 && func() bool {
		__antithesis_instrumentation__.Notify(122450)
		return n.max.compare(upperBound(out)) != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(122451)

	} else {
		__antithesis_instrumentation__.Notify(122452)
		n.max = n.findUpperBound()
	}
	__antithesis_instrumentation__.Notify(122443)
	return out, next
}

func (n *node) insert(item *latch) (replaced, newBound bool) {
	__antithesis_instrumentation__.Notify(122453)
	i, found := n.find(item)
	if found {
		__antithesis_instrumentation__.Notify(122458)
		n.items[i] = item
		return true, false
	} else {
		__antithesis_instrumentation__.Notify(122459)
	}
	__antithesis_instrumentation__.Notify(122454)
	if n.leaf {
		__antithesis_instrumentation__.Notify(122460)
		n.insertAt(i, item, nil)
		return false, n.adjustUpperBoundOnInsertion(item, nil)
	} else {
		__antithesis_instrumentation__.Notify(122461)
	}
	__antithesis_instrumentation__.Notify(122455)
	if n.children[i].count >= maxItems {
		__antithesis_instrumentation__.Notify(122462)
		splitLa, splitNode := mut(&n.children[i]).split(maxItems / 2)
		n.insertAt(i, splitLa, splitNode)

		switch cmp := cmp(item, n.items[i]); {
		case cmp < 0:
			__antithesis_instrumentation__.Notify(122463)

		case cmp > 0:
			__antithesis_instrumentation__.Notify(122464)
			i++
		default:
			__antithesis_instrumentation__.Notify(122465)
			n.items[i] = item
			return true, false
		}
	} else {
		__antithesis_instrumentation__.Notify(122466)
	}
	__antithesis_instrumentation__.Notify(122456)
	replaced, newBound = mut(&n.children[i]).insert(item)
	if newBound {
		__antithesis_instrumentation__.Notify(122467)
		newBound = n.adjustUpperBoundOnInsertion(item, nil)
	} else {
		__antithesis_instrumentation__.Notify(122468)
	}
	__antithesis_instrumentation__.Notify(122457)
	return replaced, newBound
}

func (n *node) removeMax() *latch {
	__antithesis_instrumentation__.Notify(122469)
	if n.leaf {
		__antithesis_instrumentation__.Notify(122472)
		n.count--
		out := n.items[n.count]
		n.items[n.count] = nilT
		n.adjustUpperBoundOnRemoval(out, nil)
		return out
	} else {
		__antithesis_instrumentation__.Notify(122473)
	}
	__antithesis_instrumentation__.Notify(122470)

	i := int(n.count)
	if n.children[i].count <= minItems {
		__antithesis_instrumentation__.Notify(122474)

		n.rebalanceOrMerge(i)
		return n.removeMax()
	} else {
		__antithesis_instrumentation__.Notify(122475)
	}
	__antithesis_instrumentation__.Notify(122471)
	child := mut(&n.children[i])
	out := child.removeMax()
	n.adjustUpperBoundOnRemoval(out, nil)
	return out
}

func (n *node) remove(item *latch) (out *latch, newBound bool) {
	__antithesis_instrumentation__.Notify(122476)
	i, found := n.find(item)
	if n.leaf {
		__antithesis_instrumentation__.Notify(122481)
		if found {
			__antithesis_instrumentation__.Notify(122483)
			out, _ = n.removeAt(i)
			return out, n.adjustUpperBoundOnRemoval(out, nil)
		} else {
			__antithesis_instrumentation__.Notify(122484)
		}
		__antithesis_instrumentation__.Notify(122482)
		return nilT, false
	} else {
		__antithesis_instrumentation__.Notify(122485)
	}
	__antithesis_instrumentation__.Notify(122477)
	if n.children[i].count <= minItems {
		__antithesis_instrumentation__.Notify(122486)

		n.rebalanceOrMerge(i)
		return n.remove(item)
	} else {
		__antithesis_instrumentation__.Notify(122487)
	}
	__antithesis_instrumentation__.Notify(122478)
	child := mut(&n.children[i])
	if found {
		__antithesis_instrumentation__.Notify(122488)

		out = n.items[i]
		n.items[i] = child.removeMax()
		return out, n.adjustUpperBoundOnRemoval(out, nil)
	} else {
		__antithesis_instrumentation__.Notify(122489)
	}
	__antithesis_instrumentation__.Notify(122479)

	out, newBound = child.remove(item)
	if newBound {
		__antithesis_instrumentation__.Notify(122490)
		newBound = n.adjustUpperBoundOnRemoval(out, nil)
	} else {
		__antithesis_instrumentation__.Notify(122491)
	}
	__antithesis_instrumentation__.Notify(122480)
	return out, newBound
}

func (n *node) rebalanceOrMerge(i int) {
	__antithesis_instrumentation__.Notify(122492)
	switch {
	case i > 0 && func() bool {
		__antithesis_instrumentation__.Notify(122498)
		return n.children[i-1].count > minItems == true
	}() == true:
		__antithesis_instrumentation__.Notify(122493)

		left := mut(&n.children[i-1])
		child := mut(&n.children[i])
		xLa, grandChild := left.popBack()
		yLa := n.items[i-1]
		child.pushFront(yLa, grandChild)
		n.items[i-1] = xLa

		left.adjustUpperBoundOnRemoval(xLa, grandChild)
		child.adjustUpperBoundOnInsertion(yLa, grandChild)

	case i < int(n.count) && func() bool {
		__antithesis_instrumentation__.Notify(122499)
		return n.children[i+1].count > minItems == true
	}() == true:
		__antithesis_instrumentation__.Notify(122494)

		right := mut(&n.children[i+1])
		child := mut(&n.children[i])
		xLa, grandChild := right.popFront()
		yLa := n.items[i]
		child.pushBack(yLa, grandChild)
		n.items[i] = xLa

		right.adjustUpperBoundOnRemoval(xLa, grandChild)
		child.adjustUpperBoundOnInsertion(yLa, grandChild)

	default:
		__antithesis_instrumentation__.Notify(122495)

		if i >= int(n.count) {
			__antithesis_instrumentation__.Notify(122500)
			i = int(n.count - 1)
		} else {
			__antithesis_instrumentation__.Notify(122501)
		}
		__antithesis_instrumentation__.Notify(122496)
		child := mut(&n.children[i])

		_ = mut(&n.children[i+1])
		mergeLa, mergeChild := n.removeAt(i)
		child.items[child.count] = mergeLa
		copy(child.items[child.count+1:], mergeChild.items[:mergeChild.count])
		if !child.leaf {
			__antithesis_instrumentation__.Notify(122502)
			copy(child.children[child.count+1:], mergeChild.children[:mergeChild.count+1])
		} else {
			__antithesis_instrumentation__.Notify(122503)
		}
		__antithesis_instrumentation__.Notify(122497)
		child.count += mergeChild.count + 1

		child.adjustUpperBoundOnInsertion(mergeLa, mergeChild)
		mergeChild.decRef(false)
	}
}

func (n *node) findUpperBound() keyBound {
	__antithesis_instrumentation__.Notify(122504)
	var max keyBound
	for i := int16(0); i < n.count; i++ {
		__antithesis_instrumentation__.Notify(122507)
		up := upperBound(n.items[i])
		if max.compare(up) < 0 {
			__antithesis_instrumentation__.Notify(122508)
			max = up
		} else {
			__antithesis_instrumentation__.Notify(122509)
		}
	}
	__antithesis_instrumentation__.Notify(122505)
	if !n.leaf {
		__antithesis_instrumentation__.Notify(122510)
		for i := int16(0); i <= n.count; i++ {
			__antithesis_instrumentation__.Notify(122511)
			up := n.children[i].max
			if max.compare(up) < 0 {
				__antithesis_instrumentation__.Notify(122512)
				max = up
			} else {
				__antithesis_instrumentation__.Notify(122513)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(122514)
	}
	__antithesis_instrumentation__.Notify(122506)
	return max
}

func (n *node) adjustUpperBoundOnInsertion(item *latch, child *node) bool {
	__antithesis_instrumentation__.Notify(122515)
	up := upperBound(item)
	if child != nil {
		__antithesis_instrumentation__.Notify(122518)
		if up.compare(child.max) < 0 {
			__antithesis_instrumentation__.Notify(122519)
			up = child.max
		} else {
			__antithesis_instrumentation__.Notify(122520)
		}
	} else {
		__antithesis_instrumentation__.Notify(122521)
	}
	__antithesis_instrumentation__.Notify(122516)
	if n.max.compare(up) < 0 {
		__antithesis_instrumentation__.Notify(122522)
		n.max = up
		return true
	} else {
		__antithesis_instrumentation__.Notify(122523)
	}
	__antithesis_instrumentation__.Notify(122517)
	return false
}

func (n *node) adjustUpperBoundOnRemoval(item *latch, child *node) bool {
	__antithesis_instrumentation__.Notify(122524)
	up := upperBound(item)
	if child != nil {
		__antithesis_instrumentation__.Notify(122527)
		if up.compare(child.max) < 0 {
			__antithesis_instrumentation__.Notify(122528)
			up = child.max
		} else {
			__antithesis_instrumentation__.Notify(122529)
		}
	} else {
		__antithesis_instrumentation__.Notify(122530)
	}
	__antithesis_instrumentation__.Notify(122525)
	if n.max.compare(up) == 0 {
		__antithesis_instrumentation__.Notify(122531)

		n.max = n.findUpperBound()
		return n.max.compare(up) != 0
	} else {
		__antithesis_instrumentation__.Notify(122532)
	}
	__antithesis_instrumentation__.Notify(122526)
	return false
}

type btree struct {
	root   *node
	length int
}

func (t *btree) Reset() {
	__antithesis_instrumentation__.Notify(122533)
	if t.root != nil {
		__antithesis_instrumentation__.Notify(122535)
		t.root.decRef(true)
		t.root = nil
	} else {
		__antithesis_instrumentation__.Notify(122536)
	}
	__antithesis_instrumentation__.Notify(122534)
	t.length = 0
}

func (t *btree) Clone() btree {
	__antithesis_instrumentation__.Notify(122537)
	c := *t
	if c.root != nil {
		__antithesis_instrumentation__.Notify(122539)

		c.root.incRef()
	} else {
		__antithesis_instrumentation__.Notify(122540)
	}
	__antithesis_instrumentation__.Notify(122538)
	return c
}

func (t *btree) Delete(item *latch) {
	__antithesis_instrumentation__.Notify(122541)
	if t.root == nil || func() bool {
		__antithesis_instrumentation__.Notify(122544)
		return t.root.count == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(122545)
		return
	} else {
		__antithesis_instrumentation__.Notify(122546)
	}
	__antithesis_instrumentation__.Notify(122542)
	if out, _ := mut(&t.root).remove(item); out != nilT {
		__antithesis_instrumentation__.Notify(122547)
		t.length--
	} else {
		__antithesis_instrumentation__.Notify(122548)
	}
	__antithesis_instrumentation__.Notify(122543)
	if t.root.count == 0 {
		__antithesis_instrumentation__.Notify(122549)
		old := t.root
		if t.root.leaf {
			__antithesis_instrumentation__.Notify(122551)
			t.root = nil
		} else {
			__antithesis_instrumentation__.Notify(122552)
			t.root = t.root.children[0]
		}
		__antithesis_instrumentation__.Notify(122550)
		old.decRef(false)
	} else {
		__antithesis_instrumentation__.Notify(122553)
	}
}

func (t *btree) Set(item *latch) {
	__antithesis_instrumentation__.Notify(122554)
	if t.root == nil {
		__antithesis_instrumentation__.Notify(122556)
		t.root = newLeafNode()
	} else {
		__antithesis_instrumentation__.Notify(122557)
		if t.root.count >= maxItems {
			__antithesis_instrumentation__.Notify(122558)
			splitLa, splitNode := mut(&t.root).split(maxItems / 2)
			newRoot := newNode()
			newRoot.count = 1
			newRoot.items[0] = splitLa
			newRoot.children[0] = t.root
			newRoot.children[1] = splitNode
			newRoot.max = newRoot.findUpperBound()
			t.root = newRoot
		} else {
			__antithesis_instrumentation__.Notify(122559)
		}
	}
	__antithesis_instrumentation__.Notify(122555)
	if replaced, _ := mut(&t.root).insert(item); !replaced {
		__antithesis_instrumentation__.Notify(122560)
		t.length++
	} else {
		__antithesis_instrumentation__.Notify(122561)
	}
}

func (t *btree) MakeIter() iterator {
	__antithesis_instrumentation__.Notify(122562)
	return iterator{r: t.root, pos: -1}
}

func (t *btree) Height() int {
	__antithesis_instrumentation__.Notify(122563)
	if t.root == nil {
		__antithesis_instrumentation__.Notify(122566)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(122567)
	}
	__antithesis_instrumentation__.Notify(122564)
	h := 1
	n := t.root
	for !n.leaf {
		__antithesis_instrumentation__.Notify(122568)
		n = n.children[0]
		h++
	}
	__antithesis_instrumentation__.Notify(122565)
	return h
}

func (t *btree) Len() int {
	__antithesis_instrumentation__.Notify(122569)
	return t.length
}

func (t *btree) String() string {
	__antithesis_instrumentation__.Notify(122570)
	if t.length == 0 {
		__antithesis_instrumentation__.Notify(122572)
		return ";"
	} else {
		__antithesis_instrumentation__.Notify(122573)
	}
	__antithesis_instrumentation__.Notify(122571)
	var b strings.Builder
	t.root.writeString(&b)
	return b.String()
}

func (n *node) writeString(b *strings.Builder) {
	__antithesis_instrumentation__.Notify(122574)
	if n.leaf {
		__antithesis_instrumentation__.Notify(122576)
		for i := int16(0); i < n.count; i++ {
			__antithesis_instrumentation__.Notify(122578)
			if i != 0 {
				__antithesis_instrumentation__.Notify(122580)
				b.WriteString(",")
			} else {
				__antithesis_instrumentation__.Notify(122581)
			}
			__antithesis_instrumentation__.Notify(122579)
			b.WriteString(n.items[i].String())
		}
		__antithesis_instrumentation__.Notify(122577)
		return
	} else {
		__antithesis_instrumentation__.Notify(122582)
	}
	__antithesis_instrumentation__.Notify(122575)
	for i := int16(0); i <= n.count; i++ {
		__antithesis_instrumentation__.Notify(122583)
		b.WriteString("(")
		n.children[i].writeString(b)
		b.WriteString(")")
		if i < n.count {
			__antithesis_instrumentation__.Notify(122584)
			b.WriteString(n.items[i].String())
		} else {
			__antithesis_instrumentation__.Notify(122585)
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
	__antithesis_instrumentation__.Notify(122586)
	if is.aLen == -1 {
		__antithesis_instrumentation__.Notify(122587)
		is.s = append(is.s, f)
	} else {
		__antithesis_instrumentation__.Notify(122588)
		if int(is.aLen) == len(is.a) {
			__antithesis_instrumentation__.Notify(122589)
			is.s = make([]iterFrame, int(is.aLen)+1, 2*int(is.aLen))
			copy(is.s, is.a[:])
			is.s[int(is.aLen)] = f
			is.aLen = -1
		} else {
			__antithesis_instrumentation__.Notify(122590)
			is.a[is.aLen] = f
			is.aLen++
		}
	}
}

func (is *iterStack) pop() iterFrame {
	__antithesis_instrumentation__.Notify(122591)
	if is.aLen == -1 {
		__antithesis_instrumentation__.Notify(122593)
		f := is.s[len(is.s)-1]
		is.s = is.s[:len(is.s)-1]
		return f
	} else {
		__antithesis_instrumentation__.Notify(122594)
	}
	__antithesis_instrumentation__.Notify(122592)
	is.aLen--
	return is.a[is.aLen]
}

func (is *iterStack) len() int {
	__antithesis_instrumentation__.Notify(122595)
	if is.aLen == -1 {
		__antithesis_instrumentation__.Notify(122597)
		return len(is.s)
	} else {
		__antithesis_instrumentation__.Notify(122598)
	}
	__antithesis_instrumentation__.Notify(122596)
	return int(is.aLen)
}

func (is *iterStack) reset() {
	__antithesis_instrumentation__.Notify(122599)
	if is.aLen == -1 {
		__antithesis_instrumentation__.Notify(122600)
		is.s = is.s[:0]
	} else {
		__antithesis_instrumentation__.Notify(122601)
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
	__antithesis_instrumentation__.Notify(122602)
	i.n = i.r
	i.pos = -1
	i.s.reset()
	i.o = overlapScan{}
}

func (i *iterator) descend(n *node, pos int16) {
	__antithesis_instrumentation__.Notify(122603)
	i.s.push(iterFrame{n: n, pos: pos})
	i.n = n.children[pos]
	i.pos = 0
}

func (i *iterator) ascend() {
	__antithesis_instrumentation__.Notify(122604)
	f := i.s.pop()
	i.n = f.n
	i.pos = f.pos
}

func (i *iterator) SeekGE(item *latch) {
	__antithesis_instrumentation__.Notify(122605)
	i.reset()
	if i.n == nil {
		__antithesis_instrumentation__.Notify(122607)
		return
	} else {
		__antithesis_instrumentation__.Notify(122608)
	}
	__antithesis_instrumentation__.Notify(122606)
	for {
		__antithesis_instrumentation__.Notify(122609)
		pos, found := i.n.find(item)
		i.pos = int16(pos)
		if found {
			__antithesis_instrumentation__.Notify(122612)
			return
		} else {
			__antithesis_instrumentation__.Notify(122613)
		}
		__antithesis_instrumentation__.Notify(122610)
		if i.n.leaf {
			__antithesis_instrumentation__.Notify(122614)
			if i.pos == i.n.count {
				__antithesis_instrumentation__.Notify(122616)
				i.Next()
			} else {
				__antithesis_instrumentation__.Notify(122617)
			}
			__antithesis_instrumentation__.Notify(122615)
			return
		} else {
			__antithesis_instrumentation__.Notify(122618)
		}
		__antithesis_instrumentation__.Notify(122611)
		i.descend(i.n, i.pos)
	}
}

func (i *iterator) SeekLT(item *latch) {
	__antithesis_instrumentation__.Notify(122619)
	i.reset()
	if i.n == nil {
		__antithesis_instrumentation__.Notify(122621)
		return
	} else {
		__antithesis_instrumentation__.Notify(122622)
	}
	__antithesis_instrumentation__.Notify(122620)
	for {
		__antithesis_instrumentation__.Notify(122623)
		pos, found := i.n.find(item)
		i.pos = int16(pos)
		if found || func() bool {
			__antithesis_instrumentation__.Notify(122625)
			return i.n.leaf == true
		}() == true {
			__antithesis_instrumentation__.Notify(122626)
			i.Prev()
			return
		} else {
			__antithesis_instrumentation__.Notify(122627)
		}
		__antithesis_instrumentation__.Notify(122624)
		i.descend(i.n, i.pos)
	}
}

func (i *iterator) First() {
	__antithesis_instrumentation__.Notify(122628)
	i.reset()
	if i.n == nil {
		__antithesis_instrumentation__.Notify(122631)
		return
	} else {
		__antithesis_instrumentation__.Notify(122632)
	}
	__antithesis_instrumentation__.Notify(122629)
	for !i.n.leaf {
		__antithesis_instrumentation__.Notify(122633)
		i.descend(i.n, 0)
	}
	__antithesis_instrumentation__.Notify(122630)
	i.pos = 0
}

func (i *iterator) Last() {
	__antithesis_instrumentation__.Notify(122634)
	i.reset()
	if i.n == nil {
		__antithesis_instrumentation__.Notify(122637)
		return
	} else {
		__antithesis_instrumentation__.Notify(122638)
	}
	__antithesis_instrumentation__.Notify(122635)
	for !i.n.leaf {
		__antithesis_instrumentation__.Notify(122639)
		i.descend(i.n, i.n.count)
	}
	__antithesis_instrumentation__.Notify(122636)
	i.pos = i.n.count - 1
}

func (i *iterator) Next() {
	__antithesis_instrumentation__.Notify(122640)
	if i.n == nil {
		__antithesis_instrumentation__.Notify(122644)
		return
	} else {
		__antithesis_instrumentation__.Notify(122645)
	}
	__antithesis_instrumentation__.Notify(122641)

	if i.n.leaf {
		__antithesis_instrumentation__.Notify(122646)
		i.pos++
		if i.pos < i.n.count {
			__antithesis_instrumentation__.Notify(122649)
			return
		} else {
			__antithesis_instrumentation__.Notify(122650)
		}
		__antithesis_instrumentation__.Notify(122647)
		for i.s.len() > 0 && func() bool {
			__antithesis_instrumentation__.Notify(122651)
			return i.pos >= i.n.count == true
		}() == true {
			__antithesis_instrumentation__.Notify(122652)
			i.ascend()
		}
		__antithesis_instrumentation__.Notify(122648)
		return
	} else {
		__antithesis_instrumentation__.Notify(122653)
	}
	__antithesis_instrumentation__.Notify(122642)

	i.descend(i.n, i.pos+1)
	for !i.n.leaf {
		__antithesis_instrumentation__.Notify(122654)
		i.descend(i.n, 0)
	}
	__antithesis_instrumentation__.Notify(122643)
	i.pos = 0
}

func (i *iterator) Prev() {
	__antithesis_instrumentation__.Notify(122655)
	if i.n == nil {
		__antithesis_instrumentation__.Notify(122659)
		return
	} else {
		__antithesis_instrumentation__.Notify(122660)
	}
	__antithesis_instrumentation__.Notify(122656)

	if i.n.leaf {
		__antithesis_instrumentation__.Notify(122661)
		i.pos--
		if i.pos >= 0 {
			__antithesis_instrumentation__.Notify(122664)
			return
		} else {
			__antithesis_instrumentation__.Notify(122665)
		}
		__antithesis_instrumentation__.Notify(122662)
		for i.s.len() > 0 && func() bool {
			__antithesis_instrumentation__.Notify(122666)
			return i.pos < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(122667)
			i.ascend()
			i.pos--
		}
		__antithesis_instrumentation__.Notify(122663)
		return
	} else {
		__antithesis_instrumentation__.Notify(122668)
	}
	__antithesis_instrumentation__.Notify(122657)

	i.descend(i.n, i.pos)
	for !i.n.leaf {
		__antithesis_instrumentation__.Notify(122669)
		i.descend(i.n, i.n.count)
	}
	__antithesis_instrumentation__.Notify(122658)
	i.pos = i.n.count - 1
}

func (i *iterator) Valid() bool {
	__antithesis_instrumentation__.Notify(122670)
	return i.pos >= 0 && func() bool {
		__antithesis_instrumentation__.Notify(122671)
		return i.pos < i.n.count == true
	}() == true
}

func (i *iterator) Cur() *latch {
	__antithesis_instrumentation__.Notify(122672)
	return i.n.items[i.pos]
}

type overlapScan struct {
	constrMinN       *node
	constrMinPos     int16
	constrMinReached bool

	constrMaxN   *node
	constrMaxPos int16
}

func (i *iterator) FirstOverlap(item *latch) {
	__antithesis_instrumentation__.Notify(122673)
	i.reset()
	if i.n == nil {
		__antithesis_instrumentation__.Notify(122675)
		return
	} else {
		__antithesis_instrumentation__.Notify(122676)
	}
	__antithesis_instrumentation__.Notify(122674)
	i.pos = 0
	i.o = overlapScan{}
	i.constrainMinSearchBounds(item)
	i.constrainMaxSearchBounds(item)
	i.findNextOverlap(item)
}

func (i *iterator) NextOverlap(item *latch) {
	__antithesis_instrumentation__.Notify(122677)
	if i.n == nil {
		__antithesis_instrumentation__.Notify(122679)
		return
	} else {
		__antithesis_instrumentation__.Notify(122680)
	}
	__antithesis_instrumentation__.Notify(122678)
	i.pos++
	i.findNextOverlap(item)
}

func (i *iterator) constrainMinSearchBounds(item *latch) {
	__antithesis_instrumentation__.Notify(122681)
	k := item.Key()
	j := sort.Search(int(i.n.count), func(j int) bool {
		__antithesis_instrumentation__.Notify(122683)
		return bytes.Compare(k, i.n.items[j].Key()) <= 0
	})
	__antithesis_instrumentation__.Notify(122682)
	i.o.constrMinN = i.n
	i.o.constrMinPos = int16(j)
}

func (i *iterator) constrainMaxSearchBounds(item *latch) {
	__antithesis_instrumentation__.Notify(122684)
	up := upperBound(item)
	j := sort.Search(int(i.n.count), func(j int) bool {
		__antithesis_instrumentation__.Notify(122686)
		return !up.contains(i.n.items[j])
	})
	__antithesis_instrumentation__.Notify(122685)
	i.o.constrMaxN = i.n
	i.o.constrMaxPos = int16(j)
}

func (i *iterator) findNextOverlap(item *latch) {
	__antithesis_instrumentation__.Notify(122687)
	for {
		__antithesis_instrumentation__.Notify(122688)
		if i.pos > i.n.count {
			__antithesis_instrumentation__.Notify(122693)

			i.ascend()
		} else {
			__antithesis_instrumentation__.Notify(122694)
			if !i.n.leaf {
				__antithesis_instrumentation__.Notify(122695)

				if i.o.constrMinReached || func() bool {
					__antithesis_instrumentation__.Notify(122696)
					return i.n.children[i.pos].max.contains(item) == true
				}() == true {
					__antithesis_instrumentation__.Notify(122697)
					par := i.n
					pos := i.pos
					i.descend(par, pos)

					if par == i.o.constrMinN && func() bool {
						__antithesis_instrumentation__.Notify(122700)
						return pos == i.o.constrMinPos == true
					}() == true {
						__antithesis_instrumentation__.Notify(122701)
						i.constrainMinSearchBounds(item)
					} else {
						__antithesis_instrumentation__.Notify(122702)
					}
					__antithesis_instrumentation__.Notify(122698)
					if par == i.o.constrMaxN && func() bool {
						__antithesis_instrumentation__.Notify(122703)
						return pos == i.o.constrMaxPos == true
					}() == true {
						__antithesis_instrumentation__.Notify(122704)
						i.constrainMaxSearchBounds(item)
					} else {
						__antithesis_instrumentation__.Notify(122705)
					}
					__antithesis_instrumentation__.Notify(122699)
					continue
				} else {
					__antithesis_instrumentation__.Notify(122706)
				}
			} else {
				__antithesis_instrumentation__.Notify(122707)
			}
		}
		__antithesis_instrumentation__.Notify(122689)

		if i.n == i.o.constrMaxN && func() bool {
			__antithesis_instrumentation__.Notify(122708)
			return i.pos == i.o.constrMaxPos == true
		}() == true {
			__antithesis_instrumentation__.Notify(122709)

			i.pos = i.n.count
			return
		} else {
			__antithesis_instrumentation__.Notify(122710)
		}
		__antithesis_instrumentation__.Notify(122690)
		if i.n == i.o.constrMinN && func() bool {
			__antithesis_instrumentation__.Notify(122711)
			return i.pos == i.o.constrMinPos == true
		}() == true {
			__antithesis_instrumentation__.Notify(122712)

			i.o.constrMinReached = true
		} else {
			__antithesis_instrumentation__.Notify(122713)
		}
		__antithesis_instrumentation__.Notify(122691)

		if i.pos < i.n.count {
			__antithesis_instrumentation__.Notify(122714)

			if i.o.constrMinReached {
				__antithesis_instrumentation__.Notify(122716)

				return
			} else {
				__antithesis_instrumentation__.Notify(122717)
			}
			__antithesis_instrumentation__.Notify(122715)
			if upperBound(i.n.items[i.pos]).contains(item) {
				__antithesis_instrumentation__.Notify(122718)
				return
			} else {
				__antithesis_instrumentation__.Notify(122719)
			}
		} else {
			__antithesis_instrumentation__.Notify(122720)
		}
		__antithesis_instrumentation__.Notify(122692)
		i.pos++
	}
}
