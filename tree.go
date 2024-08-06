package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const NotSelected = -1

type Tree struct {
	Root       *Node
	CurrentDir *Node
}

func (t *Tree) GetSelectedChild() *Node {
	return t.CurrentDir.Children[t.CurrentDir.Selected]
}
func (t *Tree) SelectNextChild() {
	if t.CurrentDir.Selected < len(t.CurrentDir.Children)-1 {
		t.CurrentDir.Selected += 1
		t.CurrentDir.smem += 1
	}
}
func (t *Tree) SelectPreviousChild() {
	if t.CurrentDir.Selected > 0 {
		t.CurrentDir.Selected -= 1
		t.CurrentDir.smem -= 1
	}
}
func (t *Tree) SetSelectedChildAsCurrent() error {
	selected := t.GetSelectedChild()
	if selected.Children == nil {
		err := selected.ReadChildren()
		if err != nil {
			return err
		}
	}
	if len(selected.Children) > 0 {
		t.CurrentDir = selected
		t.CurrentDir.Selected = t.CurrentDir.smem
	}
	return nil
}
func (t *Tree) SetParentAsCurrent() {
	if t.CurrentDir.Parent != nil {
		t.CurrentDir.Selected = NotSelected
		t.CurrentDir = t.CurrentDir.Parent
	}
}
func (t *Tree) CollapseOrExpandSelected() error {
	selectedNode := t.GetSelectedChild()
	if selectedNode.Children != nil {
		selectedNode.orphanChildren()
	} else {
		err := selectedNode.ReadChildren()
		if err != nil {
			return err
		}
	}
	return nil
}

func InitTree(dir string) (Tree, error) {
	var tree Tree
	rootInfo, err := os.Lstat(dir)
	if err != nil {
		return tree, err
	}
	if !rootInfo.IsDir() {
		return tree, fmt.Errorf("%s is not a directory", dir)
	}
	root := &Node{
		Path:     dir,
		Info:     rootInfo,
		Parent:   nil,
		Children: []*Node{},
	}

	err = root.ReadChildren()
	if err != nil {
		return tree, err
	}
	if len(root.Children) == 0 {
		return tree, fmt.Errorf("Can't initialize on empty directory '%s'", dir)
	}
	tree = Tree{Root: root, CurrentDir: root}
	return tree, nil
}

type Node struct {
	Path     string
	Info     fs.FileInfo
	Children []*Node // nil - not read or it's a file
	Parent   *Node
	Selected int
	smem     int
}

func (n *Node) ReadChildren() error {
	if !n.Info.IsDir() {
		return nil
	}
	children, err := os.ReadDir(n.Path)
	if err != nil {
		return err
	}
	chNodes := []*Node{}

	for _, ch := range children {
		ch := ch
		chInfo, err := ch.Info()
		if err != nil {
			return err
		}
		// Looking if child already exist
		var childToAdd *Node
		if n.Children != nil {
			for _, ech := range n.Children {
				if ech.Info.Name() == chInfo.Name() {
					childToAdd = ech
					break
				}
			}
		}
		if childToAdd == nil {
			childToAdd = &Node{
				Path:     filepath.Join(n.Path, chInfo.Name()),
				Info:     chInfo,
				Children: nil,
				Parent:   n,
				Selected: NotSelected,
			}
		}
		chNodes = append(chNodes, childToAdd)
	}
	n.Children = chNodes
	return nil
}
func (n *Node) orphanChildren() {
	n.Children = nil
	n.Selected = NotSelected
}
