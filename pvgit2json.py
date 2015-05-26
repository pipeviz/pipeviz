#!/usr/bin/env python3
import pygit2
import json
import argparse
import datetime
import sys

def grab_args():
    parser = argparse.ArgumentParser(description='Options for converting git repo data into Pipeviz JSON.')
    parser.add_argument('-r', '--repo', default='.', help='repo directory to work with')
    parser.add_argument('-p', '--pretty', default=None, help='Output human readable JSON.', 
        action='store_const', const=4)
    parser.add_argument('-o', '--output', default='-', help='Destination file to write JSON. Default - to STDOUT')
    args = parser.parse_args()
    return args

def jsonify(repo, destination=False):
    last = repo[repo.head.target]
    first = True
    intro = "["
    if args.pretty:
        intro += "\n"
    if destination:
        destination.write(intro)
    else:
        sys.stdout.write(intro)
    for commit in repo.walk(last.id, pygit2.GIT_SORT_TIME):
        output = {}
        output['sha1'] = str(commit.tree_id)
        output['subject'] = shorten(commit.message)
        output['author'] = '"{}" <{}>'.format(commit.author.name, commit.author.email)
        output['date'] = datetime.datetime.fromtimestamp(commit.commit_time).strftime('%c') + " {:=02d}{:02d}".format(commit.commit_time_offset//60, commit.commit_time_offset % 60)
        output['repository'] = repo.remotes["origin"].url
        output['parents'] = list(map(str, commit.parent_ids))
        if destination:
            if not first:
                destination.write(",")
                if args.pretty:
                    destination.write("\n")
            json.dump(output, destination, indent=args.pretty)
        else:
            if not first:
                sys.stdout.write(",")
                if args.pretty:
                    sys.stdout.write("\n")
            sys.stdout.write(json.dumps(output, indent=args.pretty))
        first = False
    conclude = "]"
    if args.pretty:
        conclude = "\n]"
    if destination:
        destination.write(conclude)
    else:
        sys.stdout.write(conclude)

def shorten(message, lines=1, length=50):
    keep_lines = '\n'.join(message.split('\n')[0:lines])
    if len(keep_lines) > length:
        keep_lines = keep_lines[0:length]
    return keep_lines
    

if __name__ == "__main__":
    args = grab_args()
    repo = pygit2.Repository(pygit2.discover_repository(args.repo))
    if args.output == '-':
        jsonify(repo)
    else:
        with open(args.output, 'w') as f:
            jsonify(repo, f)
# vim: tabstop=8 expandtab shiftwidth=4 softtabstop=4
