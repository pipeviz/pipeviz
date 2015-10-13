package ingest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tag1consulting/pipeviz/types/semantic"
)

type githubPushEvent struct {
	Ref        string            `json:"ref"`
	Head       string            `json:"head"`
	Before     string            `json:"before"`
	Commits    []githubCommitObj `json:"commits"`
	Repository struct {
		Ident string `json:"url"`
	} `json:"repository"`
	HeadCommit githubCommitObj `json:"head_commit"`
}

type githubCommitObj struct {
	Sha     string `json:"id"`
	Message string `json:"message"`
	Author  struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	URL       string `json:"url"`
	Distinct  bool   `json:"distinct"`
	Timestamp string `json:"timestamp"`
}

func githubIngestor(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// TODO this is all pretty sloppy
	bod, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Too long, or otherwise malformed request body
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	gpe := githubPushEvent{}
	err = json.Unmarshal(bod, &gpe)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	m := gpe.ToMsgMap()
	tf, err := json.Marshal(m)

	client := http.Client{Timeout: 2 * time.Second}
	client.Post("http://localhost:2309", "application/json", bytes.NewReader(tf))

	// tell github it's all OK
	w.WriteHeader(202)
}

func (gpe githubPushEvent) ToMsgMap() map[string]interface{} {
	msgmap := make(map[string]interface{})

	commits := make([]semantic.Commit, 0)

	// for reuse in all commits...ugh
	emptyparents := make([]string, 0)
	for _, c := range gpe.Commits {
		// don't include commits we know not to be new - make that someone else's job
		if !c.Distinct {
			continue
		}

		// take up to 50 bytes for subject
		subjlen := len(c.Message)
		if subjlen > 50 {
			subjlen = 50
		}

		commits = append(commits, semantic.Commit{
			Sha1Str: c.Sha,
			Subject: c.Message[:subjlen],
			Author:  fmt.Sprintf("%q <%s>", c.Author.Name, c.Author.Email),
			// TODO fix this once other branch is merged in - reuse date fmt
			Date:       c.Timestamp,
			Repository: gpe.Repository.Ident,
			// FIXME soooo github doesn't include the parents list in the push payload. wtf to do
			ParentsStr: emptyparents,
		})
	}
	if len(commits) > 0 {
		msgmap["commits"] = commits
	}

	if gpe.Ref[:11] == "refs/heads/" {
		msgmap["commit-meta"] = []semantic.CommitMeta{{
			Sha1Str:  gpe.HeadCommit.Sha,
			Branches: []string{gpe.Ref[11:]},
		}}
	} else if gpe.Ref[:10] == "refs/tags/" {
		msgmap["commit-meta"] = []semantic.CommitMeta{{
			Sha1Str: gpe.HeadCommit.Sha,
			Tags:    []string{gpe.Ref[10:]},
		}}
	}

	return msgmap
}
