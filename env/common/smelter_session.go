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
// projects rooted at p. Slot 0 lands on the primary (editor) workspace, slot 1
// on the partner (terminal) workspace.
func SmelterSessionTemplates(p string) []env.SessionTemplate {
	shell := func(id, dir string, slot int) env.SessionTask {
		return env.SessionTask{ID: id, Cwd: path.Join(p, dir), Args: []string{"zsh"}, Slot: slot}
	}
	pnpmDev := func(id, dir string, slot int) env.SessionTask {
		return env.SessionTask{ID: id, Cwd: path.Join(p, dir), Args: []string{"pnpm", "dev"}, Slot: slot}
	}
	// A terminal running `claude --ide` on the partner workspace (slot 1, the
	// secondary output) so each smelter session comes up with Claude attached.
	claudeIde := func(dir string) env.SessionTask {
		return env.SessionTask{ID: "claude", Cwd: path.Join(p, dir), Args: []string{"claude", "--ide"}, Slot: 1}
	}
	// Submodule-populate command for a fresh worktree. When the primary
	// checkout's object store is present, source the snapshot submodule's
	// ~470MB from it (--reference) and copy them in (--dissociate) so the clone
	// is a standalone full clone fetched off local disk rather than re-downloaded;
	// otherwise fall back to a plain network clone.
	submoduleUpdate := []string{"git", "submodule", "update", "--init", "--checkout"}
	if ref := path.Join(p, "smelter", ".git", "modules", "snapshots"); file.Exists(ref) {
		submoduleUpdate = append(submoduleUpdate, "--reference", ref, "--dissociate")
	}
	submoduleUpdate = append(submoduleUpdate, "integration-tests/snapshots")
	templates := []env.SessionTemplate{
		{ID: "smelter", Name: "smelter", Tasks: []env.SessionTask{
			shell("shell", "smelter", 0),
			shell("shell-partner", "smelter", 1),
			claudeIde("smelter"),
		}},
		{
			// Like "smelter (core)" but in a fresh worktree of the smelter repo
			// on branch @wkozyra95/<project-name>, created from the prompted name.
			ID:      "smelter-worktree",
			Name:    "smelter worktree",
			Prepare: smelterWorktreePrepare(p),
			Tasks: []env.SessionTask{
				{ID: "shell", Args: []string{"zsh"}, Slot: 0},
				{ID: "shell-partner", Args: []string{"zsh"}, Slot: 1},
				// `claude --ide` on the partner workspace (secondary output); Cwd
				// is the worktree, filled in by Prepare/overrideCwd.
				{ID: "claude", Slot: 1, Args: []string{"claude", "--ide"}},
				// Populate the snapshot submodule in its own terminal on the
				// partner workspace; Cwd is the worktree, filled in by Prepare.
				{ID: "submodules", Slot: 1, Args: submoduleUpdate},
			},
		},
		{ID: "smelter-website", Name: "smelter website", Tasks: []env.SessionTask{
			shell("shell", "smelter-website", 0),
			pnpmDev("dev", "smelter-website", 1),
			claudeIde("smelter-website"),
		}},
		{ID: "smelter-tools", Name: "smelter tools", Tasks: []env.SessionTask{
			shell("shell", "tools", 0),
			pnpmDev("dev", "tools", 1),
			claudeIde("tools"),
		}},
		{ID: "smelter-skills", Name: "smelter skills", Tasks: []env.SessionTask{
			shell("shell", "skills", 0),
			claudeIde("skills"),
		}},
		{ID: "smelter-examples", Name: "smelter examples", Tasks: []env.SessionTask{
			shell("shell", "examples", 0),
			claudeIde("examples"),
		}},
	}
	// One template per existing worktree of the smelter repo (resolved
	// dynamically), behaving like "smelter (core)" but in that worktree.
	return append(templates, smelterWorktreeTemplates(p)...)
}

// smelterWorktreeTemplates builds a "smelter (core)"-style template for each
// existing linked worktree of the smelter repo (the primary checkout is skipped
// since "smelter (core)" already covers it).
func smelterWorktreeTemplates(root string) []env.SessionTemplate {
	repo := path.Join(root, "smelter")
	result := []env.SessionTemplate{}
	for _, wt := range listSmelterWorktrees(repo) {
		if wt == repo {
			continue
		}
		base := path.Base(wt)
		result = append(result, env.SessionTemplate{
			ID:   "worktree-" + base,
			Name: "smelter worktree: " + base,
			Tasks: []env.SessionTask{
				{ID: "shell", Cwd: wt, Args: []string{"zsh"}, Slot: 0},
				{ID: "shell-partner", Cwd: wt, Args: []string{"zsh"}, Slot: 1},
				{ID: "claude", Cwd: wt, Args: []string{"claude", "--ide"}, Slot: 1},
			},
		})
	}
	return result
}

// listSmelterWorktrees returns the working-tree paths registered for the repo.
func listSmelterWorktrees(repo string) []string {
	var stdout, stderr bytes.Buffer
	_ = exec.Command().WithBufout(&stdout, &stderr).WithCwd(repo).
		Args("git", "worktree", "prune").Run()
	if err := exec.Command().WithBufout(&stdout, &stderr).WithCwd(repo).
		Args("git", "worktree", "list", "--porcelain").Run(); err != nil {
		return nil
	}
	paths := []string{}
	for _, line := range strings.Split(stdout.String(), "\n") {
		if p, ok := strings.CutPrefix(line, "worktree "); ok {
			paths = append(paths, p)
		}
	}
	return paths
}

// smelterWorktreePrepare returns a Prepare hook that creates (or reuses) a git
// worktree of the smelter repo at <root>/smelter-<name> on branch
// @wkozyra95/<name>, and returns that worktree path.
func smelterWorktreePrepare(root string) func(string) (string, error) {
	return func(name string) (string, error) {
		repo := path.Join(root, "smelter")
		worktree := path.Join(root, "smelter-"+name)
		branch := "@wkozyra95/" + name

		switch {
		case worktreeForBranch(repo, branch) != "":
			// Branch already checked out in a worktree (a branch can only be
			// checked out once); reuse it.
			worktree = worktreeForBranch(repo, branch)
		case file.Exists(worktree):
			// Target path already exists; reuse it.
		default:
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
		}
		return worktree, nil
	}
}

// branchExists reports whether the given local branch exists in the repo.
func branchExists(repo, branch string) bool {
	var stdout, stderr bytes.Buffer
	return exec.Command().WithBufout(&stdout, &stderr).WithCwd(repo).
		Args("git", "rev-parse", "--verify", "--quiet", "refs/heads/"+branch).Run() == nil
}

// worktreeForBranch returns the path of the worktree that currently has the
// given branch checked out, or "" if none.
func worktreeForBranch(repo, branch string) string {
	var stdout, stderr bytes.Buffer
	_ = exec.Command().WithBufout(&stdout, &stderr).WithCwd(repo).
		Args("git", "worktree", "prune").Run()
	if err := exec.Command().WithBufout(&stdout, &stderr).WithCwd(repo).
		Args("git", "worktree", "list", "--porcelain").Run(); err != nil {
		return ""
	}
	target := "refs/heads/" + branch
	current := ""
	for _, line := range strings.Split(stdout.String(), "\n") {
		if p, ok := strings.CutPrefix(line, "worktree "); ok {
			current = p
		} else if b, ok := strings.CutPrefix(line, "branch "); ok && b == target {
			return current
		}
	}
	return ""
}
