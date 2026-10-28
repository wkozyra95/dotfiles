package common

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/wkozyra95/dotfiles/env"
	"github.com/wkozyra95/dotfiles/utils/exec"
	"github.com/wkozyra95/dotfiles/utils/file"
)

// SmelterSessionTemplates returns the session-manager templates for the smelter
// projects, whose repos are cloned under root (e.g. ~/smelter). Slot 0 is the
// primary (editor) workspace, slot 1 the partner workspace. A task without Args
// is a plain shell.
func SmelterSessionTemplates(root string) []env.SessionTemplate {
	var (
		smelterDir        = path.Join(root, "smelter")
		websiteDir        = path.Join(root, "smelter-website")
		toolsDir          = path.Join(root, "tools")
		skillsDir         = path.Join(root, "skills")
		examplesDir       = path.Join(root, "examples")
		membranePluginDir = path.Join(homeDir, "membrane", "membrane_smelter_plugin")
	)

	// Populates the snapshot submodule of a fresh worktree. When the primary
	// checkout's object store is present, the submodule's ~470MB is sourced from
	// it (--reference) and copied in (--dissociate), so the clone is a
	// standalone full clone fetched off local disk rather than re-downloaded;
	// otherwise it falls back to a plain network clone.
	submoduleUpdate := []string{"git", "submodule", "update", "--init", "--checkout"}
	if ref := path.Join(smelterDir, ".git", "modules", "snapshots"); file.Exists(ref) {
		submoduleUpdate = append(submoduleUpdate, "--reference", ref, "--dissociate")
	}
	submoduleUpdate = append(submoduleUpdate, "integration-tests/snapshots")

	templates := []env.SessionTemplate{}
	// One template per existing linked worktree of the smelter repo, like
	// "smelter" but in that worktree, named after the worktree directory
	// (smelter-<name>). The primary checkout is "smelter" itself.
	for _, wt := range listWorktrees(smelterDir) {
		if wt.path == smelterDir {
			continue
		}
		templates = append(templates, env.SessionTemplate{
			Name: path.Base(wt.path),
			Tasks: []env.SessionTask{
				{Slot: 0, Cwd: wt.path, Args: []string{"nvim", "."}},
				{Slot: 1, Cwd: wt.path, Args: []string{"zsh"}},
				{Slot: 1, Cwd: wt.path, Args: []string{"claude", "--ide"}},
			},
		})
	}
	return append(templates,
		env.SessionTemplate{
			Name: "smelter",
			Tasks: []env.SessionTask{
				{Slot: 0, Cwd: smelterDir, Args: []string{"nvim", "."}},
				{Slot: 1, Cwd: smelterDir, Args: []string{"zsh"}},
				{Slot: 1, Cwd: smelterDir, Args: []string{"claude", "--ide"}},
			},
		},
		env.SessionTemplate{
			// Like "smelter" but in a fresh worktree on branch
			// @wkozyra95/<project-name>, created from the prompted name. Cwd is
			// left empty: every task runs in the worktree Prepare returns.
			Name:    "smelter-worktree",
			Prepare: smelterWorktreePrepare(root),
			Tasks: []env.SessionTask{
				{Slot: 0, Args: []string{"nvim", "."}},
				{Slot: 1, Args: []string{"zsh"}},
				{Slot: 1, Args: []string{"claude", "--ide"}},
				{Slot: 1, Args: submoduleUpdate},
			},
		},
		env.SessionTemplate{
			Name: "smelter-website",
			Tasks: []env.SessionTask{
				{Slot: 0, Cwd: websiteDir, Args: []string{"nvim", "."}},
				{Slot: 1, Cwd: websiteDir, Args: []string{"pnpm", "dev"}},
				{Slot: 1, Cwd: websiteDir, Args: []string{"claude", "--ide"}},
			},
		},
		env.SessionTemplate{
			Name: "smelter-tools",
			Tasks: []env.SessionTask{
				{Slot: 0, Cwd: toolsDir, Args: []string{"nvim", "."}},
				{Slot: 1, Cwd: toolsDir, Args: []string{"pnpm", "dev"}},
				{Slot: 1, Cwd: toolsDir, Args: []string{"claude", "--ide"}},
			},
		},
		env.SessionTemplate{
			Name: "smelter-skills",
			Tasks: []env.SessionTask{
				{Slot: 0, Cwd: skillsDir, Args: []string{"nvim", "."}},
				{Slot: 1, Cwd: skillsDir, Args: []string{"claude", "--ide"}},
			},
		},
		env.SessionTemplate{
			Name: "smelter-examples",
			Tasks: []env.SessionTask{
				{Slot: 0, Cwd: examplesDir, Args: []string{"nvim", "."}},
				{Slot: 1, Cwd: examplesDir, Args: []string{"claude", "--ide"}},
			},
		},
		env.SessionTemplate{
			Name: "membrane-smelter-plugin",
			Tasks: []env.SessionTask{
				{Slot: 0, Cwd: membranePluginDir, Args: []string{"nvim", "."}},
				{Slot: 1, Cwd: membranePluginDir, Args: []string{"zsh"}},
				{Slot: 1, Cwd: membranePluginDir, Args: []string{"claude", "--ide"}},
			},
		},
	)
}

// smelterWorktreePrepare returns a Prepare hook that creates (or reuses) a git
// worktree of the smelter repo at <root>/smelter-<name> on branch
// @wkozyra95/<name>, and returns that worktree path.
func smelterWorktreePrepare(root string) func(string) (string, error) {
	return func(name string) (string, error) {
		repo := path.Join(root, "smelter")
		worktree := path.Join(root, "smelter-"+name)
		branch := "@wkozyra95/" + name

		if existing := worktreeForBranch(repo, branch); existing != "" {
			// Branch already checked out in a worktree (a branch can only be
			// checked out once); reuse it.
			return existing, nil
		}
		if file.Exists(worktree) {
			// Target path already exists; reuse it.
			return worktree, nil
		}
		// Create the worktree: check out the branch if it already exists,
		// otherwise create it with -b.
		args := []string{"git", "worktree", "add", worktree}
		if branchExists(repo, branch) {
			args = append(args, branch)
		} else {
			args = append(args, "-b", branch)
		}
		if err := exec.Command().WithCwd(repo).Args(args...).Run(); err != nil {
			return "", fmt.Errorf("git worktree add %s: %w", branch, err)
		}
		return worktree, nil
	}
}

// gitWorktree is one entry of `git worktree list --porcelain`.
type gitWorktree struct {
	path   string
	branch string // full ref (refs/heads/...), empty when detached
}

// listWorktrees returns the worktrees registered for repo, after pruning stale
// entries, or nil if git fails.
func listWorktrees(repo string) []gitWorktree {
	var stdout, stderr bytes.Buffer
	_ = exec.Command().WithBufout(&stdout, &stderr).WithCwd(repo).
		Args("git", "worktree", "prune").Run()
	if err := exec.Command().WithBufout(&stdout, &stderr).WithCwd(repo).
		Args("git", "worktree", "list", "--porcelain").Run(); err != nil {
		return nil
	}
	worktrees := []gitWorktree{}
	for _, line := range strings.Split(stdout.String(), "\n") {
		if p, ok := strings.CutPrefix(line, "worktree "); ok {
			worktrees = append(worktrees, gitWorktree{path: p})
		} else if b, ok := strings.CutPrefix(line, "branch "); ok && len(worktrees) > 0 {
			worktrees[len(worktrees)-1].branch = b
		}
	}
	return worktrees
}

// worktreeForBranch returns the path of the worktree that currently has the
// given branch checked out, or "" if none.
func worktreeForBranch(repo, branch string) string {
	for _, wt := range listWorktrees(repo) {
		if wt.branch == "refs/heads/"+branch {
			return wt.path
		}
	}
	return ""
}

// branchExists reports whether the given local branch exists in the repo.
func branchExists(repo, branch string) bool {
	var stdout, stderr bytes.Buffer
	return exec.Command().WithBufout(&stdout, &stderr).WithCwd(repo).
		Args("git", "rev-parse", "--verify", "--quiet", "refs/heads/"+branch).Run() == nil
}
