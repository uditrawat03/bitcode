package treeview

import (
	"os"
	"path/filepath"
)

// Node represents a file/folder
type Node struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*Node
	Expanded bool
	Parent   *Node
}

// TreeView manages the node tree
type TreeView struct {
	Root       *Node
	Nodes      []*Node       // flattened visible nodes
	NodeLevels map[*Node]int // indentation levels
}

// NewTreeView creates a tree view from a folder
func NewTreeView(rootPath string) *TreeView {
	tv := &TreeView{
		NodeLevels: map[*Node]int{},
	}
	tv.Load(rootPath)
	return tv
}

// Load loads directory recursively
func (tv *TreeView) Load(rootPath string) {
	tv.Root = &Node{
		Name:     filepath.Base(rootPath),
		Path:     rootPath,
		IsDir:    true,
		Expanded: true,
	}
	tv.buildTree(tv.Root)
	tv.flattenVisible()
}

// ReloadChildren refreshes the children of a node without rebuilding the entire tree
func (tv *TreeView) ReloadChildren(node *Node) {
	if node == nil || !node.IsDir {
		return
	}

	// Clear existing children
	node.Children = nil

	// Rebuild only this node's children
	tv.buildTree(node)

	// Refresh flattened visible nodes
	tv.flattenVisible()
}

// isHidden returns true if the file/folder is hidden
func isHidden(path string, info os.FileInfo) bool {
	name := info.Name()
	if name == "." || name == ".." {
		return true
	}
	// skip hidden directories only
	return info.IsDir() && name[0] == '.'
}

// Recursively build tree
func (tv *TreeView) buildTree(node *Node) {
	entries, err := os.ReadDir(node.Path)
	if err != nil {
		return
	}
	for _, entry := range entries {
		childPath := filepath.Join(node.Path, entry.Name())
		info, err := entry.Info()
		if err != nil || isHidden(childPath, info) {
			continue
		}

		child := &Node{
			Name:   entry.Name(),
			Path:   childPath,
			IsDir:  entry.IsDir(),
			Parent: node,
		}
		node.Children = append(node.Children, child)
		if entry.IsDir() {
			child.Expanded = false
			tv.buildTree(child)
		}
	}
}

// Flatten tree to visible nodes
func (tv *TreeView) flattenVisible() {
	tv.Nodes = []*Node{}
	tv.NodeLevels = map[*Node]int{}
	tv.flattenNode(tv.Root, 0)
}

func (tv *TreeView) flattenNode(node *Node, level int) {
	tv.Nodes = append(tv.Nodes, node)
	tv.NodeLevels[node] = level
	if node.IsDir && node.Expanded {
		for _, child := range node.Children {
			tv.flattenNode(child, level+1)
		}
	}
}

// Toggle node expand/collapse
func (tv *TreeView) Toggle(node *Node) {
	if node != nil && node.IsDir {
		node.Expanded = !node.Expanded
		tv.flattenVisible()
	}
}

// Get node level
func (tv *TreeView) Level(node *Node) int {
	if l, ok := tv.NodeLevels[node]; ok {
		return l
	}
	return 0
}
