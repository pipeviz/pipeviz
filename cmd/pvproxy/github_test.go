package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/tag1consulting/pipeviz/types/semantic"
)

// Real/sample data from github v3 api
const ghPushPayload = `{
  "ref": "refs/heads/master",
  "before": "5627b4bf954465918bd9ede94a2484be03ddb44b",
  "after": "4d59fb584b15a94d7401e356d2875c472d76ef45",
  "created": false,
  "deleted": false,
  "forced": false,
  "base_ref": null,
  "compare": "https://github.com/sdboyer/testrepo/compare/5627b4bf9544...4d59fb584b15",
  "commits": [
    {
      "id": "b75da01c073384926e782a4371195c851f45b20f",
      "distinct": true,
      "message": "one commit, will it show?",
      "timestamp": "2015-10-12T21:39:52-04:00",
      "url": "https://github.com/sdboyer/testrepo/commit/b75da01c073384926e782a4371195c851f45b20f",
      "author": {
        "name": "Sam Boyer",
        "email": "notareal@email.com",
        "username": "sdboyer"
      },
      "committer": {
        "name": "Sam Boyer",
        "email": "notareal@email.com",
        "username": "sdboyer"
      },
      "added": [

      ],
      "removed": [

      ],
      "modified": [
        "README"
      ]
    },
    {
      "id": "4d59fb584b15a94d7401e356d2875c472d76ef45",
      "distinct": true,
      "message": "second commit, should show both",
      "timestamp": "2015-10-12T21:40:18-04:00",
      "url": "https://github.com/sdboyer/testrepo/commit/4d59fb584b15a94d7401e356d2875c472d76ef45",
      "author": {
        "name": "Sam Boyer",
        "email": "notareal@email.com",
        "username": "sdboyer"
      },
      "committer": {
        "name": "Sam Boyer",
        "email": "notareal@email.com",
        "username": "sdboyer"
      },
      "added": [

      ],
      "removed": [

      ],
      "modified": [
        "README",
        "tmp2"
      ]
    }
  ],
  "head_commit": {
    "id": "4d59fb584b15a94d7401e356d2875c472d76ef45",
    "distinct": true,
    "message": "second commit, should show both",
    "timestamp": "2015-10-12T21:40:18-04:00",
    "url": "https://github.com/sdboyer/testrepo/commit/4d59fb584b15a94d7401e356d2875c472d76ef45",
    "author": {
      "name": "Sam Boyer",
      "email": "notareal@email.com",
      "username": "sdboyer"
    },
    "committer": {
      "name": "Sam Boyer",
      "email": "notareal@email.com",
      "username": "sdboyer"
    },
    "added": [

    ],
    "removed": [

    ],
    "modified": [
      "README",
      "tmp2"
    ]
  },
  "repository": {
    "id": 44144360,
    "name": "testrepo",
    "full_name": "sdboyer/testrepo",
    "owner": {
      "name": "sdboyer",
      "email": "notareal@email.com"
    },
    "private": false,
    "html_url": "https://github.com/sdboyer/testrepo",
    "description": null,
    "fork": false,
    "url": "https://github.com/sdboyer/testrepo",
    "forks_url": "https://api.github.com/repos/sdboyer/testrepo/forks",
    "keys_url": "https://api.github.com/repos/sdboyer/testrepo/keys{/key_id}",
    "collaborators_url": "https://api.github.com/repos/sdboyer/testrepo/collaborators{/collaborator}",
    "teams_url": "https://api.github.com/repos/sdboyer/testrepo/teams",
    "hooks_url": "https://api.github.com/repos/sdboyer/testrepo/hooks",
    "issue_events_url": "https://api.github.com/repos/sdboyer/testrepo/issues/events{/number}",
    "events_url": "https://api.github.com/repos/sdboyer/testrepo/events",
    "assignees_url": "https://api.github.com/repos/sdboyer/testrepo/assignees{/user}",
    "branches_url": "https://api.github.com/repos/sdboyer/testrepo/branches{/branch}",
    "tags_url": "https://api.github.com/repos/sdboyer/testrepo/tags",
    "blobs_url": "https://api.github.com/repos/sdboyer/testrepo/git/blobs{/sha}",
    "git_tags_url": "https://api.github.com/repos/sdboyer/testrepo/git/tags{/sha}",
    "git_refs_url": "https://api.github.com/repos/sdboyer/testrepo/git/refs{/sha}",
    "trees_url": "https://api.github.com/repos/sdboyer/testrepo/git/trees{/sha}",
    "statuses_url": "https://api.github.com/repos/sdboyer/testrepo/statuses/{sha}",
    "languages_url": "https://api.github.com/repos/sdboyer/testrepo/languages",
    "stargazers_url": "https://api.github.com/repos/sdboyer/testrepo/stargazers",
    "contributors_url": "https://api.github.com/repos/sdboyer/testrepo/contributors",
    "subscribers_url": "https://api.github.com/repos/sdboyer/testrepo/subscribers",
    "subscription_url": "https://api.github.com/repos/sdboyer/testrepo/subscription",
    "commits_url": "https://api.github.com/repos/sdboyer/testrepo/commits{/sha}",
    "git_commits_url": "https://api.github.com/repos/sdboyer/testrepo/git/commits{/sha}",
    "comments_url": "https://api.github.com/repos/sdboyer/testrepo/comments{/number}",
    "issue_comment_url": "https://api.github.com/repos/sdboyer/testrepo/issues/comments{/number}",
    "contents_url": "https://api.github.com/repos/sdboyer/testrepo/contents/{+path}",
    "compare_url": "https://api.github.com/repos/sdboyer/testrepo/compare/{base}...{head}",
    "merges_url": "https://api.github.com/repos/sdboyer/testrepo/merges",
    "archive_url": "https://api.github.com/repos/sdboyer/testrepo/{archive_format}{/ref}",
    "downloads_url": "https://api.github.com/repos/sdboyer/testrepo/downloads",
    "issues_url": "https://api.github.com/repos/sdboyer/testrepo/issues{/number}",
    "pulls_url": "https://api.github.com/repos/sdboyer/testrepo/pulls{/number}",
    "milestones_url": "https://api.github.com/repos/sdboyer/testrepo/milestones{/number}",
    "notifications_url": "https://api.github.com/repos/sdboyer/testrepo/notifications{?since,all,participating}",
    "labels_url": "https://api.github.com/repos/sdboyer/testrepo/labels{/name}",
    "releases_url": "https://api.github.com/repos/sdboyer/testrepo/releases{/id}",
    "created_at": 1444700054,
    "updated_at": "2015-10-13T01:34:14Z",
    "pushed_at": 1444700432,
    "git_url": "git://github.com/sdboyer/testrepo.git",
    "ssh_url": "git@github.com:sdboyer/testrepo.git",
    "clone_url": "https://github.com/sdboyer/testrepo.git",
    "svn_url": "https://github.com/sdboyer/testrepo",
    "homepage": null,
    "size": 0,
    "stargazers_count": 0,
    "watchers_count": 0,
    "language": null,
    "has_issues": true,
    "has_downloads": true,
    "has_wiki": true,
    "has_pages": false,
    "forks_count": 0,
    "mirror_url": null,
    "open_issues_count": 0,
    "forks": 0,
    "open_issues": 0,
    "watchers": 0,
    "default_branch": "master",
    "stargazers": 0,
    "master_branch": "master"
  },
  "pusher": {
    "name": "sdboyer",
    "email": "notareal@email.com"
  },
  "sender": {
    "login": "sdboyer",
    "id": 21599,
    "avatar_url": "https://avatars.githubusercontent.com/u/21599?v=3",
    "gravatar_id": "",
    "url": "https://api.github.com/users/sdboyer",
    "html_url": "https://github.com/sdboyer",
    "followers_url": "https://api.github.com/users/sdboyer/followers",
    "following_url": "https://api.github.com/users/sdboyer/following{/other_user}",
    "gists_url": "https://api.github.com/users/sdboyer/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/sdboyer/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/sdboyer/subscriptions",
    "organizations_url": "https://api.github.com/users/sdboyer/orgs",
    "repos_url": "https://api.github.com/users/sdboyer/repos",
    "events_url": "https://api.github.com/users/sdboyer/events{/privacy}",
    "received_events_url": "https://api.github.com/users/sdboyer/received_events",
    "type": "User",
    "site_admin": false
  }
}`

func TestPushToMessageMap(t *testing.T) {
	gpe := githubPushEvent{}
	json.Unmarshal([]byte(ghPushPayload), &gpe)
	// FIXME mock this somehow; we don't want to actually reach out
	m := gpe.ToMessage("")

	if len(m.C) != 2 {
		t.Errorf("Commits array should have two elements, got %d\n", len(m.C))
	}

	assert.Equal(t, m.C[0], semantic.Commit{
		Sha1Str:    "b75da01c073384926e782a4371195c851f45b20f",
		Subject:    "one commit, will it show?",
		Author:     "\"Sam Boyer\" <notareal@email.com>",
		Date:       "Mon Oct 12 2015 21:39:52 -0400",
		Repository: "https://github.com/sdboyer/testrepo",
		ParentsStr: []string{
			"5627b4bf954465918bd9ede94a2484be03ddb44b",
		},
	})

	assert.Equal(t, m.C[1], semantic.Commit{
		Sha1Str:    "4d59fb584b15a94d7401e356d2875c472d76ef45",
		Subject:    "second commit, should show both",
		Author:     "\"Sam Boyer\" <notareal@email.com>",
		Date:       "Mon Oct 12 2015 21:40:18 -0400",
		Repository: "https://github.com/sdboyer/testrepo",
		ParentsStr: []string{
			"b75da01c073384926e782a4371195c851f45b20f",
		},
	})

	if len(m.Cm) != 1 {
		t.Errorf("Should be one item in commit meta, found %d\n", len(m.Cm))
	}

	if !reflect.DeepEqual(m.Cm[0], semantic.CommitMeta{
		Sha1Str:  "4d59fb584b15a94d7401e356d2875c472d76ef45",
		Branches: []string{"master"},
	}) {
		t.Error("Commit meta not as expected")
	}
}
