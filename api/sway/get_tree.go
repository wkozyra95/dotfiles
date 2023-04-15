package sway

import (
	"bytes"
	"encoding/json"

	"github.com/wkozyra95/dotfiles/utils/exec"
)

type TreeNode struct {
	Type    string     `json:"type"`
	Name    string     `json:"name"`
	PID     int        `json:"pid"`
	Focused bool       `json:"focused"`
	Visible bool       `json:"visible"`
	AppID   string     `json:"app_id"`
	Nodes   []TreeNode `json:"nodes"`
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
	err := exec.Command().WithBufout(&stdout, &stderr).Run("swaymsg", "-t", "get_tree", "-r")
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
