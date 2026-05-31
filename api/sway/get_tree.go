package sway

import (
	"bytes"
	"encoding/json"

	"github.com/wkozyra95/dotfiles/utils/exec"
)

type TreeNode struct {
	ID            int64      `json:"id"`
	Type          string     `json:"type"`
	Name          string     `json:"name"`
	Num           int        `json:"num"`
	PID           int        `json:"pid"`
	Focused       bool       `json:"focused"`
	Visible       bool       `json:"visible"`
	AppID         string     `json:"app_id"`
	Nodes         []TreeNode `json:"nodes"`
	FloatingNodes []TreeNode `json:"floating_nodes"`
}

// OutputForCon walks the sway tree and returns the name of the output
// that contains the given con_id, or "" if not found.
func OutputForCon(conID int64) string {
	tree, err := GetTree()
	if err != nil {
		return ""
	}
	for _, output := range tree.Nodes {
		if output.Type == "output" && containsCon(output, conID) {
			return output.Name
		}
	}
	return ""
}

func containsCon(node TreeNode, conID int64) bool {
	if node.ID == conID {
		return true
	}
	for _, child := range node.Nodes {
		if containsCon(child, conID) {
			return true
		}
	}
	return false
}

func FindContainer(tree TreeNode, matchFn func(TreeNode) bool) *TreeNode {
	for _, node := range tree.Nodes {
		if matchFn(node) {
			nodeCopy := node
			return &nodeCopy
		}
		resultNode := FindContainer(node, matchFn)
		if resultNode != nil {
			return resultNode
		}
	}
	return nil
}

func GetTree() (TreeNode, error) {
	var stdout, stderr bytes.Buffer
	err := exec.Command().WithBufout(&stdout, &stderr).Args("swaymsg", "-t", "get_tree", "-r").Run()
	if err != nil {
		return TreeNode{}, err
	}
	return parseGetTreeResult(stdout.Bytes())
}

func parseGetTreeResult(rawContent []byte) (TreeNode, error) {
	tree := TreeNode{}
	if err := json.Unmarshal(rawContent, &tree); err != nil {
		return TreeNode{}, err
	}

	return tree, nil
}
