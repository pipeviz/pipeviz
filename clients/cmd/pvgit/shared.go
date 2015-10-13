package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/tag1consulting/pipeviz/clients/githelp"
	"github.com/tag1consulting/pipeviz/ingest"
	"github.com/tag1consulting/pipeviz/types/semantic"

	"gopkg.in/libgit2/git2go.v22"
)

func getRepoOrExit(paths ...string) *git.Repository {
	var path string
	var err error

	if len(paths) > 0 {
		path = paths[0]
	} else if path, err = os.Getwd(); err != nil {
		log.Fatalln("Error getting cwd:", err)
	}

	repostr, err := git.Discover(path, false, []string{"/"})
	if err != nil {
		log.Fatalln(err)
	}

	repo, err := git.OpenRepository(repostr)
	if err != nil {
		log.Fatalf("Error opening repo at %s: %s", repostr+"/.git", err)
	}

	return repo
}

func commitToSemanticForm(in *git.Commit, ident string) (out semantic.Commit) {
	authsig := in.Author()
	out = semantic.Commit{
		Sha1Str: hex.EncodeToString(in.Id()[:]),
		// FIXME JSON marshaling seems to encode <> weirdly
		Author:     fmt.Sprintf("%q <%s>", authsig.Name, authsig.Email),
		Date:       authsig.When.Format(githelp.GitDateFormat),
		Subject:    in.Summary(),
		Repository: ident,
		ParentsStr: make([]string, 0),
	}

	// Parent list is base 0, though 0 is really "first parent"
	for i := uint(0); i < in.ParentCount(); i++ {
		out.ParentsStr = append(out.ParentsStr, hex.EncodeToString(in.ParentId(i)[:]))
	}

	return
}

func recordHead(m *ingest.Message, repo *git.Repository) {
	head, err := repo.Head()
	if err != nil {
		log.Fatalf("Could not get repo HEAD")
	}
	oid := head.Target()

	hn, err := os.Hostname()
	if err != nil {
		log.Fatalln("Could not determine hostname for environment")
	}

	// be extra safe about cleaning the path wrt trailing slashes
	p := repo.Path()
	if strings.LastIndex(p, string(os.PathSeparator)) == len(p)-1 {
		p = path.Dir(path.Dir(p))
	} else {
		p = path.Dir(p)
	}

	m.Add(semantic.LogicState{
		Environment: semantic.EnvLink{
			Address: semantic.Address{
				Hostname: hn,
			},
		},
		Path: p,
		ID: semantic.LogicIdentiifer{
			CommitStr: hex.EncodeToString(oid[:]),
		},
	})

	if head.IsBranch() {
		b := head.Branch()
		bn, _ := b.Name()

		m.Add(semantic.CommitMeta{
			Sha1Str:  hex.EncodeToString(oid[:]),
			Tags:     make([]string, 0),
			Branches: []string{bn},
		})
	}
}

func sendMapToPipeviz(m *ingest.Message, r *git.Repository) {
	addr, err := githelp.GetTargetAddr(r)
	// Find the target pipeviz instance from the git config
	if err != nil {
		log.Fatalln("Could not find target address in config:", err)
	}

	msg, err := json.Marshal(m)
	if err != nil {
		log.Fatalln("JSON encoding failed with err:", err)
	}
	//dump, _ := json.MarshalIndent(m, "", "    ")
	//fmt.Println(string(dump))

	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(addr, "application/json", bytes.NewReader(msg))
	if err != nil {
		log.Fatalln("Error on sending to server:", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bod, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Fatalf(err.Error())
		}

		log.Fatalln("Message rejected by pipeviz server: %v", string(bod))
	}
}
