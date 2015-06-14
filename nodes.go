//Copyright 2013 Vastech SA (PTY) LTD
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package dot

import (
	"sort"
)

//Represents a Node.
type Node struct {
	Name         string
	Attrs        Attrs
	Index        int     // index of this node within Graph.Nodes.Nodes of its parent.
	Preds, Succs []*Node // predecessors and successors
	dom          domInfo // dominator tree info
}

func (node *Node) String() string {
	return node.Name
}

// HasPred reports whether m is a predecessor of n.
func (n *Node) HasPred(m *Node) bool {
	for _, pred := range n.Preds {
		if pred == m {
			return true
		}
	}
	return false
}

// HasSucc reports whether m is a successor of n.
func (n *Node) HasSucc(m *Node) bool {
	for _, succ := range n.Succs {
		if succ == m {
			return true
		}
	}
	return false
}

// addEdge adds a control-flow graph edge from from to to.
func addEdge(from, to *Node) {
	from.Succs = append(from.Succs, to)
	to.Preds = append(to.Preds, from)
}

//Represents a set of Nodes.
type Nodes struct {
	Lookup map[string]*Node
	Nodes  []*Node
}

//Creates a new set of Nodes.
func NewNodes() *Nodes {
	return &Nodes{make(map[string]*Node), make([]*Node, 0)}
}

//Adds a Node to the set of Nodes, ammending the attributes of an already existing node.
func (this *Nodes) Add(node *Node) {
	n, ok := this.Lookup[node.Name]
	if ok {
		n.Attrs.Ammend(node.Attrs)
		return
	}
	this.Lookup[node.Name] = node
	node.Index = len(this.Nodes)
	this.Nodes = append(this.Nodes, node)
}

// del deletes the node from the set of nodes.
//
// NOTE: calls to Nodes.del must be complemented with corresponding calls to
// Edges.del.
//
// NOTE: the dominator tree has to recalculated (e.g. buildDomTree) afterwards.
func (nodes *Nodes) del(node *Node) {
	// Remove node from lookup.
	delete(nodes.Lookup, node.Name)

	// Remove node from the successor list of each predecessor node.
	for _, pred := range node.Preds {
		var succs []*Node
		for _, succ := range pred.Succs {
			if succ == node {
				continue
			}
			succs = append(succs, succ)
		}
		pred.Succs = succs
	}

	// Remove node from the predecessor list of each successor node.
	for _, succ := range node.Succs {
		var preds []*Node
		for _, pred := range succ.Preds {
			if pred == node {
				continue
			}
			preds = append(preds, pred)
		}
		succ.Preds = preds
	}

	// Remove node from nodes list.
	var ns []*Node
	for _, n := range nodes.Nodes {
		if n == node {
			continue
		}
		n.Index = len(ns)
		ns = append(ns, n)
	}
	nodes.Nodes = ns

	// Clear deleted node to help with GC.
	*node = Node{}
}

//Returns a sorted list of nodes.
func (this Nodes) Sorted() []*Node {
	keys := make([]string, 0, len(this.Lookup))
	for key := range this.Lookup {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	nodes := make([]*Node, len(keys))
	for i := range keys {
		nodes[i] = this.Lookup[keys[i]]
	}
	return nodes
}

// DomSorted returns a sorted list of nodes in dominance order.
func (this Nodes) DomSorted() []*Node {
	nodes := make([]*Node, len(this.Nodes))
	copy(nodes, this.Nodes)
	sort.Sort(domOrder(nodes))
	return nodes
}

// domOrder attaches the methods of sort.Interface to []*Node, sorting in
// dominance order.
type domOrder []*Node

func (ns domOrder) Less(i, j int) bool {
	// The "entry" node is always first.
	if ns[i].Attrs != nil && ns[i].Attrs["label"] == "entry" {
		return true
	}
	if ns[j].Attrs != nil && ns[j].Attrs["label"] == "entry" {
		return false
	}

	if ns[i].Dominates(ns[j]) {
		return true
	} else if ns[j].Dominates(ns[i]) {
		return false
	}
	return ns[i].Name < ns[j].Name
}

func (ns domOrder) Len() int {
	return len(ns)
}

func (ns domOrder) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}
