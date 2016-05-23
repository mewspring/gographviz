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
	"fmt"

	"github.com/mewspring/dot/ast"
)

//Creates a Graph structure by analysing an Abstract Syntax Tree representing a parsed graph.
func NewAnalysedGraph(graph *ast.Graph) *Graph {
	g := NewGraph()
	Analyse(graph, g)

	// TODO: Make deterministic.

	// Add edges between each node for the dominator tree construction.
	for src, dsts := range g.Edges.SrcToDsts {
		for dst := range dsts {
			// Add edges between each node for the dominator tree construction.
			from, ok := g.Nodes.Lookup[src]
			if ok {
				addEdges(g, from, dst)
			} else if g.IsSubGraph(src) {
				// Child nodes of the src SubGraph.
				for srcNode := range g.Relations.ParentToChildren[src] {
					from := g.Nodes.Lookup[srcNode]
					addEdges(g, from, dst)
				}
			} else {
				panic(fmt.Sprintf("unable to add edge from src %v", src))
			}
		}
	}

	// Make sure that the "entry" node is first in the list.
	for index, node := range g.Nodes.Nodes {
		if node.Attrs != nil && node.Attrs["label"] == "entry" {
			if index != 0 {
				// Swap.
				g.Nodes.Nodes[0], g.Nodes.Nodes[index] = g.Nodes.Nodes[index], g.Nodes.Nodes[0]
				g.Nodes.Nodes[0].Index = 0
				g.Nodes.Nodes[index].Index = index
			}
			break
		}
	}

	// Calculate the dominator tree.
	buildDomTree(g)
	return g
}

func addEdges(graph *Graph, from *Node, dst string) {
	to, ok := graph.Nodes.Lookup[dst]
	if ok {
		addEdge(from, to)
	} else if graph.IsSubGraph(dst) {
		// Child nodes of the dst SubGraph.
		for dstNode := range graph.Relations.ParentToChildren[dst] {
			to := graph.Nodes.Lookup[dstNode]
			addEdge(from, to)
		}
	} else {
		panic(fmt.Sprintf("unable to add edge to dst %v", dst))
	}
}

//Analyses an Abstract Syntax Tree representing a parsed graph into a newly created graph structure Interface.
func Analyse(graph *ast.Graph, g Interface) {
	graph.Walk(&graphVisitor{g})
}

type nilVisitor struct {
}

func (this *nilVisitor) Visit(v ast.Elem) ast.Visitor {
	return this
}

type graphVisitor struct {
	g Interface
}

func (this *graphVisitor) Visit(v ast.Elem) ast.Visitor {
	graph, ok := v.(*ast.Graph)
	if !ok {
		return this
	}
	this.g.SetStrict(graph.Strict)
	this.g.SetDir(graph.Type == ast.DIGRAPH)
	graphName := graph.Id.String()
	this.g.SetName(graphName)
	return newStmtVisitor(this.g, graphName)
}

func newStmtVisitor(g Interface, graphName string) *stmtVisitor {
	return &stmtVisitor{g, graphName, make(Attrs), make(Attrs), make(Attrs)}
}

type stmtVisitor struct {
	g                 Interface
	graphName         string
	currentNodeAttrs  Attrs
	currentEdgeAttrs  Attrs
	currentGraphAttrs Attrs
}

func (this *stmtVisitor) Visit(v ast.Elem) ast.Visitor {
	switch s := v.(type) {
	case ast.NodeStmt:
		return this.nodeStmt(s)
	case ast.EdgeStmt:
		return this.edgeStmt(s)
	case ast.NodeAttrs:
		return this.nodeAttrs(s)
	case ast.EdgeAttrs:
		return this.edgeAttrs(s)
	case ast.GraphAttrs:
		return this.graphAttrs(s)
	case *ast.SubGraph:
		return this.subGraph(s)
	case *ast.Attr:
		return this.attr(s)
	case ast.AttrList:
		return &nilVisitor{}
	default:
		//fmt.Fprintf(os.Stderr, "unknown stmt %T\n", v)
	}
	return this
}

func ammend(attrs Attrs, add Attrs) Attrs {
	for key, value := range add {
		if _, ok := attrs[key]; !ok {
			attrs[key] = value
		}
	}
	return attrs
}

func overwrite(attrs Attrs, overwrite Attrs) Attrs {
	for key, value := range overwrite {
		attrs[key] = value
	}
	return attrs
}

func (this *stmtVisitor) nodeStmt(stmt ast.NodeStmt) ast.Visitor {
	attrs := Attrs(stmt.Attrs.GetMap())
	attrs = ammend(attrs, this.currentNodeAttrs)
	this.g.AddNode(this.graphName, stmt.NodeId.String(), attrs)
	return &nilVisitor{}
}

func (this *stmtVisitor) edgeStmt(stmt ast.EdgeStmt) ast.Visitor {
	attrs := stmt.Attrs.GetMap()
	attrs = ammend(attrs, this.currentEdgeAttrs)
	src := stmt.Source.GetId()
	srcName := src.String()
	if stmt.Source.IsNode() {
		this.g.AddNode(this.graphName, srcName, this.currentNodeAttrs.Copy())
	}
	srcPort := stmt.Source.GetPort()
	for i := range stmt.EdgeRHS {
		directed := bool(stmt.EdgeRHS[i].Op)
		dst := stmt.EdgeRHS[i].Destination.GetId()
		dstName := dst.String()
		if stmt.EdgeRHS[i].Destination.IsNode() {
			this.g.AddNode(this.graphName, dstName, this.currentNodeAttrs.Copy())
		}
		dstPort := stmt.EdgeRHS[i].Destination.GetPort()
		this.g.AddPortEdge(srcName, srcPort.String(), dstName, dstPort.String(), directed, attrs)
		src = dst
		srcPort = dstPort
		srcName = dstName
	}
	return this
}

func (this *stmtVisitor) nodeAttrs(stmt ast.NodeAttrs) ast.Visitor {
	this.currentNodeAttrs = overwrite(this.currentNodeAttrs, ast.AttrList(stmt).GetMap())
	return &nilVisitor{}
}

func (this *stmtVisitor) edgeAttrs(stmt ast.EdgeAttrs) ast.Visitor {
	this.currentEdgeAttrs = overwrite(this.currentEdgeAttrs, ast.AttrList(stmt).GetMap())
	return &nilVisitor{}
}

func (this *stmtVisitor) graphAttrs(stmt ast.GraphAttrs) ast.Visitor {
	attrs := ast.AttrList(stmt).GetMap()
	for key, value := range attrs {
		this.g.AddAttr(this.graphName, key, value)
	}
	this.currentGraphAttrs = overwrite(this.currentGraphAttrs, attrs)
	return &nilVisitor{}
}

func (this *stmtVisitor) subGraph(stmt *ast.SubGraph) ast.Visitor {
	subGraphName := stmt.Id.String()
	this.g.AddSubGraph(this.graphName, subGraphName, this.currentGraphAttrs)
	return newStmtVisitor(this.g, subGraphName)
}

func (this *stmtVisitor) attr(stmt *ast.Attr) ast.Visitor {
	this.g.AddAttr(this.graphName, stmt.Field.String(), stmt.Value.String())
	return this
}
