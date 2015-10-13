package ingest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/types/semantic"
)

type githubPushEvent struct {
	Ref        string            `json:"ref"`
	Head       string            `json:"head"`
	Before     string            `json:"before"`
	Commits    []githubCommitObj `json:"commits"`
	Repository struct {
		Ident         string `json:"url"`
		GitCommitsURL string `json:"git_commits_url"`
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

	m := gpe.ToMessage()
	tf, err := json.Marshal(m)

	client := http.Client{Timeout: 2 * time.Second}
	// TODO inject local url somehow, don't hardcode
	client.Post("http://localhost:2309", "application/json", bytes.NewReader(tf))

	// tell github it's all OK
	w.WriteHeader(202)
}

func (gpe githubPushEvent) ToMessage() *Message {
	msg := new(Message)
	client := http.Client{Timeout: 2 * time.Second}

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

		// github doesn't include parent commit list in push payload (UGHHHH). so, call out for it.
		resp, err := client.Get(strings.Replace(gpe.Repository.GitCommitsURL, "{/sha}", "/"+c.Sha, 1))
		if err != nil {
			// just drop the problematic commit
			logrus.WithFields(logrus.Fields{
				"system": "ingestor",
				"err":    err,
				"sha1":   c.Sha,
			}).Warn("Request to github to retrieve commit parent info failed; commit dropped.")
			continue
		}

		// skip err here, it's effectively caught by the json unmarshaler
		bod, _ := ioutil.ReadAll(resp.Body)

		var jmap interface{}
		err = json.Unmarshal(bod, &jmap)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"system": "ingestor",
				"err":    err,
				"sha1":   c.Sha,
			}).Warn("Bad JSON response from github when requesting commit parent info; commit dropped.")
			continue
		}

		// reset to empty slice, but reuse same array
		parents := make([]string, 0)
		for _, iparent := range jmap.(map[string]interface{})["parents"].([]interface{}) {
			parent := iparent.(map[string]interface{})
			parents = append(parents, parent["sha"].(string))
		}

		msg.Add(semantic.Commit{
			Sha1Str: c.Sha,
			Subject: c.Message[:subjlen],
			Author:  fmt.Sprintf("%q <%s>", c.Author.Name, c.Author.Email),
			// TODO fix this once other branch is merged in - reuse date fmt
			Date:       c.Timestamp,
			Repository: gpe.Repository.Ident,
			ParentsStr: parents,
		})
	}

	if gpe.Ref[:11] == "refs/heads/" {
		msg.Add(semantic.CommitMeta{
			Sha1Str:  gpe.HeadCommit.Sha,
			Branches: []string{gpe.Ref[11:]},
		})
	} else if gpe.Ref[:10] == "refs/tags/" {
		msg.Add(semantic.CommitMeta{
			Sha1Str: gpe.HeadCommit.Sha,
			Tags:    []string{gpe.Ref[10:]},
		})
	}

	return msg
}
