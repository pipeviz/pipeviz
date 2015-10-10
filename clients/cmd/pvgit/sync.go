package main

import (
	"encoding/hex"
	"log"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/clients/githelp"
	"github.com/tag1consulting/pipeviz/types/semantic"
	"gopkg.in/libgit2/git2go.v22"
)

type syncCmd struct {
	all bool
}

func syncCommand() *cobra.Command {
	sync := &syncCmd{}
	cmd := &cobra.Command{
		Use:   "sync [-a|--all] <repository>",
		Short: "Sends state of all local refs, and optionally all reachable local commits, to a pipeviz server",
		Run:   sync.run,
	}

	cmd.Flags().BoolVarP(&sync.all, "all", "a", false, "Send all commit history info in addition to all refs.")

	return cmd
}

func (s *syncCmd) run(cmd *cobra.Command, args []string) {
	syncHistory(getRepoOrExit(args...), s.all)
}

func syncHistory(repo *git.Repository, all bool) {
	var err error
	var ident string

	msgmap := map[string]interface{}{
		"commit-meta": make([]semantic.CommitMeta, 0),
	}

	if all {
		msgmap["commits"] = make([]semantic.Commit, 0)

		ident, err = githelp.GetRepoIdent(repo)
		if err != nil {
			log.Fatalf("Failed to retrieve a stable identifier for this repository; cannot formulate commits correctly. Aborting.")
		}
	}

	rvisited := make(map[git.Oid]struct{})
	cvisited := make(map[git.Oid]struct{})

	iter, err := repo.NewReferenceIterator()
	if err != nil {
		log.Fatalln("Error while creating reference iterator:", err)
	}

	// For simplicity, create a revwalker now even if we don't use it later
	w, err := repo.Walk()
	if err != nil {
		log.Fatalln("Could not create revwalker iterator:", err)
	}

	w.Sorting(git.SortTopological)
	defer w.Free()

	for ref, err := iter.Next(); err == nil; ref, err = iter.Next() {
		// in func for easy defer of Free()
		func(r *git.Reference, m map[string]interface{}) {
			defer r.Free()
			oid := r.Target()
			if !r.IsBranch() && !r.IsTag() {
				return
			}

			commit, err := repo.LookupCommit(oid)
			if err != nil {
				log.Fatalln("Failed to look up commit underlying ref target, err:", err.(git.GitError).Message)
			}
			defer commit.Free()

			// shorthand
			cms := m["commit-meta"].([]semantic.CommitMeta)

			// If we've already seen this commit, we need to add to the existing record
			if _, exists := rvisited[*oid]; exists {
				for k, cm := range cms {
					if cm.Sha1Str != hex.EncodeToString(oid[:]) {
						continue
					}

					if r.IsBranch() {
						cms[k].Branches = append(cms[k].Branches, r.Name())
					} else if r.IsTag() {
						cms[k].Tags = append(cms[k].Tags, r.Name())
					} else {
						log.Fatalf("Ref %s is neither branch nor tag - wtf\n", r.Name())
					}
					break
				}
			} else {
				// First time seeing a ref on this commit, add a new entry to commit-meta
				rvisited[*oid] = struct{}{}

				if r.IsBranch() {
					w.Push(oid)
					cms = append(cms, semantic.CommitMeta{
						Sha1Str:  hex.EncodeToString(oid[:]),
						Tags:     make([]string, 0),
						Branches: []string{r.Name()},
					})
				} else if r.IsTag() {
					w.Push(oid)
					cms = append(cms, semantic.CommitMeta{
						Sha1Str:  hex.EncodeToString(oid[:]),
						Tags:     []string{r.Name()},
						Branches: make([]string, 0),
					})
				} else {
					log.Fatalf("Ref %s is neither branch nor tag - wtf\n", r.Name())
				}
			}
			m["commit-meta"] = cms
		}(ref, msgmap)
	}
	iter.Free()

	if err != nil {
		if !git.IsErrorCode(err, git.ErrIterOver) {
			//if gerr, ok := err.(*git.GitError); !ok || gerr.Code != git.ErrIterOver {
			gerr := err.(*git.GitError)
			log.Fatalf("Iteration through repository refs terminated with unexpected error (code: %d, message: %q)\n", gerr.Code, gerr.Message)
		}
	}

	if all {
		commits := msgmap["commits"].([]semantic.Commit)

		w.Iterate(func(c *git.Commit) bool {
			defer c.Free()
			if _, exists := cvisited[*c.Id()]; exists {
				return false
			}

			cvisited[*c.Id()] = struct{}{}
			commits = append(commits, commitToSemanticForm(c, ident))
			return true
		})
		msgmap["commits"] = commits
	}
	//w.Free()

	sendMapToPipeviz(msgmap, repo)
}
