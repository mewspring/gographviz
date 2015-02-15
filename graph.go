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

import "fmt"

//The analysed representation of the Graph parsed from the DOT format.
type Graph struct {
	Attrs     Attrs
	Name      string
	Directed  bool
	Strict    bool
	Nodes     *Nodes
	Edges     *Edges
	SubGraphs *SubGraphs
	Relations *Relations
}

//Creates a new empty graph, ready to be populated.
func NewGraph() *Graph {
	return &Graph{
		Attrs:     make(Attrs),
		Name:      "",
		Directed:  false,
		Strict:    false,
		Nodes:     NewNodes(),
		Edges:     NewEdges(),
		SubGraphs: NewSubGraphs(),
		Relations: NewRelations(),
	}
}

// In returns the number of incoming edges to name in the graph.
func (g *Graph) In(name string) int {
	return len(g.Edges.DstToSrcs[name])
}

// Out returns the number of outgoing edges from name in the graph.
func (g *Graph) Out(name string) int {
	return len(g.Edges.SrcToDsts[name])
}

// HasEdge returns true if there exists a directed edge from src to dst.
func (g *Graph) HasEdge(src, dst string) bool {
	dsts, ok := g.Edges.SrcToDsts[src]
	if !ok {
		return false
	}
	_, ok = dsts[dst]
	return ok
}

// Merge merges the entry and exit node into a single node of the given name,
// with the incoming edges of entry and the outgoing edges of exit. All nodes
// and edges between entry and exit are removed in the process.
func (g *Graph) Merge(entry, exit *Node, name string) error {
	// Sanity checks.
	if entry == exit {
		return fmt.Errorf("invalid call to merge; entry and exit node are both %q", entry.Name)
	}

	if _, ok := g.Nodes.Lookup[name]; ok {
		return fmt.Errorf("invalid call to merge; node %q already present in graph", name)
	}
	// TODO: Verify that this sanity check always holds true.
	if entry != exit.Idom() {
		return fmt.Errorf("unable to merge entry %q and exit %q nodes; exit is not immediately dominated by entry", entry.Name, exit.Name)
	}

	// Create a new node of the given name, with incoming edges from entry and
	// outgoing edges from exit.
	// TODO: Create dedicated subgraph instead of node?
	g.AddNode(g.Name, name, nil)
	node := g.Nodes.Lookup[name]

	// Add edge from each predecessor to node.
	node.Preds = entry.Preds
	for _, pred := range entry.Preds {
		pred.Succs = append(pred.Succs, node)
		edge := g.Edges.SrcToDsts[pred.Name][entry.Name]
		g.AddEdge(pred.Name, name, true, edge.Attrs)
	}

	// Add edge from node to each successor.
	node.Succs = exit.Succs
	for _, succ := range exit.Succs {
		succ.Preds = append(succ.Preds, node)
		edge := g.Edges.DstToSrcs[succ.Name][exit.Name]
		g.AddEdge(name, succ.Name, true, edge.Attrs)
	}

	// Remove the nodes and edges between entry and exit, inclusively.
	g.removeUntil(entry, exit)

	// Recalculate the dominator tree.
	buildDomTree(g)

	return nil
}

// removeUntil removes the node, its edges and any of its successors recursively
// until the exit node is reached.
//
// NOTE: the dominator tree has to recalculated (e.g. buildDomTree) afterwards.
func (g *Graph) removeUntil(node, exit *Node) {
	if node != exit {
		// Remove any successors of the node.
		for _, succ := range node.Succs {
			g.removeUntil(succ, exit)
		}
	}

	// Remove the node and all of its edges.
	g.delNode(node)
}

// delNode deletes the node and all of its edges from the graph.
//
// NOTE: the dominator tree has to recalculated (e.g. buildDomTree) afterwards.
func (g *Graph) delNode(node *Node) {
	// Remove edges.
	for _, dst := range g.Edges.SrcToDsts[node.Name] {
		g.Edges.del(dst)
	}
	for _, src := range g.Edges.DstToSrcs[node.Name] {
		g.Edges.del(src)
	}

	// Remove node.
	g.Nodes.del(node)
}

//If the graph is strict then multiple edges are not allowed between the same pairs of nodes,
//see dot man page.
func (this *Graph) SetStrict(strict bool) {
	this.Strict = strict
}

//Sets whether the graph is directed (true) or undirected (false).
func (this *Graph) SetDir(dir bool) {
	this.Directed = dir
}

//Sets the graph name.
func (this *Graph) SetName(name string) {
	this.Name = name
}

//Adds an edge to the graph from node src to node dst.
//srcPort and dstPort are the port the node ports, leave as empty strings if it is not required.
//This does not imply the adding of missing nodes.
func (this *Graph) AddPortEdge(src, srcPort, dst, dstPort string, directed bool, attrs map[string]string) {
	this.Edges.Add(&Edge{src, srcPort, dst, dstPort, directed, attrs})
}

//Adds an edge to the graph from node src to node dst.
//This does not imply the adding of missing nodes.
func (this *Graph) AddEdge(src, dst string, directed bool, attrs map[string]string) {
	this.AddPortEdge(src, "", dst, "", directed, attrs)
}

//Adds a node to a graph/subgraph.
//If not subgraph exists use the name of the main graph.
//This does not imply the adding of a missing subgraph.
func (this *Graph) AddNode(parentGraph string, name string, attrs map[string]string) {
	node := &Node{
		Name:  name,
		Attrs: attrs,
	}
	this.Nodes.Add(node)
	this.Relations.Add(parentGraph, name)
}

func (this *Graph) getAttrs(graphName string) Attrs {
	if this.Name == graphName {
		return this.Attrs
	}
	g, ok := this.SubGraphs.SubGraphs[graphName]
	if !ok {
		panic("graph or subgraph " + graphName + " does not exist")
	}
	return g.Attrs
}

//Adds an attribute to a graph/subgraph.
func (this *Graph) AddAttr(parentGraph string, field string, value string) {
	this.getAttrs(parentGraph).Add(field, value)
}

//Adds a subgraph to a graph/subgraph.
func (this *Graph) AddSubGraph(parentGraph string, name string, attrs map[string]string) {
	this.SubGraphs.Add(name)
	for key, value := range attrs {
		this.AddAttr(name, key, value)
	}
}

func (this *Graph) IsNode(name string) bool {
	_, ok := this.Nodes.Lookup[name]
	return ok
}

func (this *Graph) IsSubGraph(name string) bool {
	_, ok := this.SubGraphs.SubGraphs[name]
	return ok
}
