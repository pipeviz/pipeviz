/**
 * Instantiates a new ReachCount object with fresh memoization caches.
 * The memoization is not robust - it *assumes* that you use the same graph
 * for all calls to each reach func, but does not verify this (the overhead
 * would reduce the benefit of memoization). So make sure you do, or the
 * results will be wildly incorrect.
 */
var reachCounter = function() {
    // set up caches out here, then attach them to each recursive folder
    // within the funs.

    // mpff - memoized predecessor foldl cache
    var mpfc = _.memoize(function(accum, value) {}),
    // msfc - memoized successor foldl cache
    msfc = _.memoize(function(accum, value) {});

    return {
        /**
         * Counts the number of vertices reachable from the given vertex in the
         * given graph by traversing through predecessors. If a set of vertices
         * is provided by which to filter as the third parameter, only vertices
         * in that set will be used to create the final reach count, though all
         * predecessors the graph makes available will still be traversed.
         */
        throughPredecessors: function(g, v, filter) {
            var pred = g.predecessors(v),
            visited = {},
            r = _.memoize(function(accum, value) {
                visited[value] = true;
                return accum.concat(_.foldl(_.filter(g.predecessors(value), function(v) { return !_.has(visited,v); }), r, [value]));
            }, function(accum, value) { return value; });
            r.cache = mpfc.cache;

            if (pred === undefined || pred.length === 0) {
                return 0;
            } else if (filter === undefined) {
                return _.uniq(_.foldl(pred, r, [v])).length;
            } else {
                // TODO this intersection is still quite expensive; moar memoization
                return _.uniq(_.intersection(_.foldl(pred, r, [v]), filter)).length;
            }
        },
        /**
         * Counts the number of vertices reachable from the given vertex in the
         * given graph by traversing through successors. If a set of vertices
         * is provided by which to filter as the third parameter, only vertices
         * in that set will be used to create the final reach count, though all
         * successors the graph makes available will still be traversed.
         */
        throughSuccessors: function(g, v, filter) {
            var succ = g.successors(v),
            visited = {},
            r = _.memoize(function(accum, value) {
                visited[value] = true;
                return accum.concat(_.foldl(_.filter(g.successors(value), function(v) { return !_.has(visited,v); }), r, [value]));
            }, function(accum, value) { return value; });
            r.cache = msfc.cache;

            if (succ === undefined) {
                return 0;
            } else if (filter === undefined) {
                return _.uniq(_.foldl(succ, r, [v])).length;
            } else {
                // TODO this intersection is still quite expensive; moar memoization
                return _.uniq(_.intersection(_.foldl(succ, r, [v]), filter)).length;
            }
        }
    };
};

// Bitmask constants representing the different ways a commit can operate within the pipeline viz algo.
// Focals are usually noelide and boundaries are usually focal and noelide, but this is not necessarily
// the case, so the mask keeps each property separate.
// TODO declare as const w/ES6

    // Commit is used as a boundary in establishing the range of commits to visualize.
var V_BOUNDARY = 0x08,
    // Consider this commit when performing reachability counts (for segment spread ranking).
    V_REACHRANK = 0x04,
    // If the commit is in range, the reduced commit tree should include a path to it.
    V_FOCAL = 0x02,
    // If the commit is in range, don't elide it - there's something interesting there.
    V_INTERESTING = 0x01,
    // Uninteresting, won't show up unless something else draws it in. This is default.
    V_UNINTERESTING = 0x00;

var vizExtractor = {
    mostCommonRepo: function(pvg) {
        return _(pvg.verticesWithType("logic-state"))
            .filter(function(v) {
                var vedges = _.filter(_.map(v.outEdges, function(edgeId) { return pvg.get(edgeId); }), isType("version"));
                return vedges.length !== 0;
            })
            .countBy(function(v) {
                return pvg.get(_.filter(_.map(v.outEdges, function(edgeId) { return pvg.get(edgeId); }), isType("version"))[0].target).propv("repository");
            })
            .reduce(function(accum, count, repo) {
                return count < accum[1] ? accum : [repo, count];
            }, ["", 0])[0];
    },
    // Searches the pvg for commits that will play various roles in determining the boundaries
    // and spatial positions in the pipeline visualization.
    findGuideCommits: function(pvg, repo, branchlvl, taglvl) {
        var guide = {};

        // FIXME note that the old method dup'd the vertex object - will need to do that somewhere else now
        // TODO doing it all in one pass b/c pvg is dumb and it's O(n) to get vertex subsets
        _.each(pvg.vertices(pq.or(isType("git-branch"), isType("git-tag"), isType("logic-state"))), function(v) {
            var level = V_UNINTERESTING,
                commit = getCommit(pvg, v);

            if (commit !== undefined && commit.propv("repository") === repo) {
                switch (v.Typ()) {
                    case "logic-state":
                        level = V_BOUNDARY | V_FOCAL | V_INTERESTING | V_REACHRANK;
                    break;
                    case "git-tag":
                        level = taglvl;
                    break;
                    case "git-branch":
                        level = branchlvl;
                    break;
                }

                if (!_.has(guide, commit.id)) {
                    guide[commit.id] = {
                        level: 0,
                        assoc: [],
                    };
                }

                // levels are designed to be additive, so we always OR them together here
                guide[commit.id].level |= level;
                guide[commit.id].assoc.push(v);
            }
        });

        return guide;
    },
    focalLogicStateByRepo: function(pvg, repo) {
        var focalCommits = {},
        focal = _.filter(_.map(pvg.verticesWithType("logic-state"), function(d) { return _.create(vertexProto, d); }), function(v) {
            var vedges = _.filter(_.map(v.outEdges, function(edgeId) { return pvg.get(edgeId); }), isType("version"));
            if (vedges.length === 0) {
                return false;
            }

            if (pvg.get(vedges[0].target).propv("repository") === repo) {
                if (!_.has(focalCommits, vedges[0].target)) {
                    focalCommits[vedges[0].target] = [];
                }
                focalCommits[vedges[0].target].push(v);
                return true;
            }
        });

        return [focal, focalCommits];
    },
    // FIXME note that the sense of "focal" here is from before the refactoring that split the focal concept into the bitmask
    focalTransposedGraph: function(cg, focalCommits) {
        var fg = new graphlib.Graph(), // graph with all non-focal vertices contracted and edges transposed (direction reversed)
        visited = {},
        fgwalk = function(v, fpath) { // Depth-first walker, builds the focal graph
            var last = fpath.length === 0 ? [] : fpath[fpath.length - 1],
            pop_fpath = false;

            // If vertex is already focal-visited, return early. We only record focal vertices
            // in this list, so there will be some double-traversal, but that's acceptable.
            if (_.has(visited, v)) {
                // Create edges from every focal on current vertex to every focal
                // from last. No op if last is empty.
                _.each(last, function(lvtx) {
                    _.each(focalCommits[v], function(vtx) {
                        fg.setEdge(vtx.id, lvtx.id);
                    });
                });
                return;
            }

            if (_.has(focalCommits, v)) {
                // This commit has one or more of our focal vertices. Create an
                // edge if fpath is non-empty, push focals onto fpath, and set
                // var to pop fpath later.
                pop_fpath = true;

                // Create edges from every focal on current vertex to every focal
                // from last. No op if last is empty.
                _.each(last, function(lvtx) {
                    _.each(focalCommits[v], function(vtx) {
                        fg.setEdge(vtx.id, lvtx.id);
                    });
                });

                fpath.push(focalCommits[v]);
            }

            _.each(cg.successors(v), function(s) {
                fgwalk(s, fpath);
            });

            if (pop_fpath) {
                visited[v] = true;
                fpath.pop();
            }
        };

        _.each(focalCommits, function(d, k) {
            fgwalk(k, []);
        });

        return fg;
    },
    treeAndRoot: function(cg, boundaryCommits, focalCommits) {
        var tree = new graphlib.Graph(), // A tree, almost an induced subgraph, of first-parent paths in the commit graph
        visited = {},
        candidates = [];

        // To build the tree, we start from each boundary commit and walk back up the
        // commit history, following ONLY first-parent. We carefully track every
        // time the walker revisits a vertex, as each of these points is a candidate
        // for being the root of the tree.
        var treewalk = function(v, last) {
            // We always want to record the (reversed) edge, unless last does not exist
            if (last !== undefined) {
                tree.setEdge(v, last);
            }

            // If vertex is already visited, it's a candidate for being the root
            if (_.has(visited, v)) {
                candidates.push(v);
                return;
            }

            // Only visit first parent; that's the path that matters to our subgraph/tree
            // We also discount successors where the current commit is not the first parent,
            // as otherwise we'd still form a graph, not a tree.
            var succ = _.filter(cg.successors(v) || [], function(s) {
                return cg.edge(v, s) === 1;
            });
            if (succ.length > 0) {
                treewalk(succ[0], v);
            }

            visited[v] = true;
        };

        // Always walk the boundary commits, b/c just walking the commit graph's source vertices
        // can miss them if they're on a path that is only reachable from a sink along a n>1
        // parent of a merge.
        _.each(boundaryCommits, function(d, k) {
            treewalk(k);
        });

        // Sort the candidates list so we can use binary search.
        candidates.sort(function(a, b) { return a - b; });

        // Now we have to find the topologically greatest common root among all candidates.
        var root,
        oob = {}; // Tracks "out-of-bound" commits; those between the natural root and the earliest candidate

        // Things are a little weird if there's only one boundary vertex - for
        // one, the treewalk won't find any candidates. In that case, set the
        // candidates list to the single vertex itself.
        if (_.size(boundaryCommits) === 1) {
            candidates.push(_.keys(boundaryCommits)[0]);
        }

        // This identifies the shared root (in git terms, the merge base) from all
        // candidates by walking down the reversed subgraph/tree until we find a
        // vertex in the candidate list. The first one we find is guaranteed to
        // be the root.
        var rootfind = function(v) {
            if (_.indexOf(candidates, v, true) !== -1) {
                root = v;
                return;
            }

            // Record this as an "out-of-bound" vertex
            oob[v] = true;

            var succ = tree.successors(v) || [];
            if (succ.length > 0) {
                rootfind(succ[0]);
            }
        };
        rootfind(tree.sources()[0]);

        // With the root found using the boundary commits, we now must do another walk
        // through the commit graph using the focal commits as starting points, to see if
        // they add any additional paths to the tree that weren't already covered by the
        // boundary commits.
        var focalwalk = function(init, v, path, last) {
            // If there's a last visited, build up our list of edges to be optionally
            // added to the tree
            if (last !== undefined) {
                path.push([v, last]);
            }

            // We infer from the treewalk's 'visited' map. If it's visited but oob,
            // then we bail on the whole path; if it's visited and not oob, we incorporate
            // the new path into the tree.
            if (_.has(visited, v)) {
                if (!_.has(oob, v)) {
                    _.each(path, function(pair) {
                        tree.setEdge(pair[0], pair[1]);
                    });
                }
                return;
            }

            var succ = _.filter(cg.successors(v) || [], function(s) {
                return cg.edge(v, s) === 1;
            });
            if (succ.length > 0) {
                focalwalk(succ[0], v);
            }

            visited[v] = true;
        };

        _.each(focalCommits, function(d, k) {
            focalwalk(k, k, []);
        });

        return [tree, root];
    },
    extractTree: function(cg, tree, root, focalCommits, interestingCommits, rankCommits) {
        var vmeta = {}, // metadata we build for each vertex. keyed by vertex id
        protolinks = [], // we can start figuring out some links in the next walk
        segments = 0, // total number of divergent segment paths. starts at 0, increases as needed
        diameter = 0, // maximum depth reached. Useful info later that we can avoid recalculating
        idepths = [], // list of interesting depths (so, no elision), to avoid another walk later
        rc = reachCounter(), // create a new memoizing reach counter
        rankIds = _.keys(rankCommits), // just to avoid repeating this in the walk
        mainwalk = function(v, path, segment, pseg) {
            // tree, so zero possibility of revisiting any vtx; no "visited" checks needed
            var succ = tree.successors(v) || [];

            // every commit in the tree makes it into the vmeta
            vmeta[v] = {
                // distance from root
                depth: path.length,
                // count of reachable ranking commits in original commit graph
                reach: rc.throughPredecessors(cg, v, rankIds),
                // count of reachable ranking commits in the tree/almost-induced subgraph
                treach: rc.throughSuccessors(tree, v, rankIds),
                // the segment to which the commit belongs
                segment: segment,
                // the commit's parent segment
                pseg: pseg
            };

            // now we decide if the commit is eligible for elision
            if (_.has(focalCommits, v) || _.has(interestingCommits, v)) {
                idepths.push(path.length);
            } else {
                // otherwise, only safe from elision if the commit or its parent has multiple successors
                var psucc = path.length === 0 ? [] : tree.successors(path[path.length -1]);
                if (succ.length > 1 || psucc.length > 1) {
                    // also push it onto the idepths list
                    idepths.push(path.length);
                }
            }

            path.push(v);
            _.each(succ, function(d) {
                if (succ.length === 1) {
                    mainwalk(d, path, segment, pseg);
                } else {
                    // if more than one succ, always increment to delineate segments
                    mainwalk(d, path, ++segments, segment);
                    // we know this'll be included, so add to protolinks right now
                    protolinks.push([v, d]);
                }
            });
            path.pop();
            // diameter is calculated AFTER path is popped because it is the
            // number of hops back to root. A hop goes from one vertex to
            // another, which means the number of hops is always one less than
            // the number of vertices (it is non-inclusive of self).
            diameter = Math.max(diameter, path.length);
        };

        // we only need to enter at root to get everything
        mainwalk(root, [], 0, 0);

        return [
            vmeta,
            protolinks,
            diameter,
            // sort idepths so binary search works later
            _.uniq(idepths).sort(function(a, b) { return a - b; })
        ];
    }
};

/*
 * The main magic function for processing a pv graph into the base state usable
 * for the app/commit viz.
 *
 */
function extractVizGraph(pvg, cg, guideCommits, elide) {
//function extractVizGraph(pvg, repo, elide) {
    var boundaryCommits = _.pick(guideCommits, function(gc) { return gc.level & V_BOUNDARY; }),
        focalCommits = _.pick(guideCommits, function(gc) { return gc.level & V_FOCAL; }),
        interestingCommits = _.pick(guideCommits, function(gc) { return gc.level & V_INTERESTING; }),
        rankCommits = _.pick(guideCommits, function(gc) { return gc.level & V_REACHRANK; });

    // We must have boundary commits for the algorithm to do anything. If there are none, bail.
    if (boundaryCommits.length === 0) {
        return;
    }

        // TODO this is commented b/c there's something horribly non-performant in the fgwalk impl atm. also, it hasn't been refactored for the commit role tiers
        //fg = vizExtractor.focalTransposedGraph(cg, focalCommits),
    var tr = vizExtractor.treeAndRoot(cg, boundaryCommits, focalCommits),
        tree = tr[0],
        root = tr[1];

    var main = vizExtractor.extractTree(cg, tree, root, focalCommits, interestingCommits, rankCommits),
        vmeta = main[0],
        protolinks = main[1],
        diameter = main[2],
        idepths = main[3];

    var elidable, elranges, ediam;
    if (elide) {
        // now we have all the base meta; construct elidables lists and segment rankings
        elidable = _.filter(_.range(diameter+1), function(depth) { return _.indexOf(idepths, depth, true) === -1; });
        // also compute the minimum set of contiguous elidable depths
        elranges = _.reduce(elidable, function(accum, v, k, coll) {
            if (coll[k-1] === v-1) {
                // contiguous section, push onto last series
                accum[accum.length - 1].push(v);
            } else {
                // non-contiguous, start a new series
                accum.push([v]);
            }

            return accum;
        }, []);
        // we also need the elided diameter
        ediam = diameter - elidable.length;
    } else {
        elidable = [];
        elranges = [];
        ediam = diameter;
    }

    // A 'segment' is an unbroken chain of commits that have only one parent commit
    // (of degree 0 or 1 in the simplified commit tree)
    var segmentinfo = _(vmeta)
        .mapValues(function(v, k) {
            return {
                segment: v.segment,
                pseg: v.pseg,
                reach: v.reach,
                treach: v.treach,
                id: parseInt(k)
            };
        })
        .groupBy(function(v) { return v.segment; }) // collects vertices on same segment into a single array
        .mapValues(function(v) {
            return {
                ids: _.map(v, function(v2) { return v2.id; }),
                pseg: v[0].pseg, // parent seg (pseg) must be the same for all vertices in a seg
                maxreach: _.max(v, 'reach').reach,
                maxtreach: _.max(v, 'treach').treach,
                rank: 0,
            };
        }).value(); // object keyed by segment number w/segment info

    // Next, we determine rankings for each segment.
    //
    // Begin by grouping vertex metadata by their depth in the tree, then iterate across
    // the groups. It is inherently guaranteed that each segment has at most one member
    // in each depth group.
    _.each(_.groupBy(vmeta, function(v) { return v.depth; }), function(depthgroup) {
        // Up to four-dimensional sort of depth group. When this is complete, the depth
        // group is sorted by appropriate rank.
        //
        depthgroup.sort(function(a, b) {
            // for shorthand
            var ab = segmentinfo[a.segment],
                bb = segmentinfo[b.segment];

            // Multi-dimensional sort. First is tree-reach ("treach"), which is the number
            // of focal commits reachable in the tree extracted from the commit graph. We
            // compare the maximum value in each segment.
            if (ab.maxtreach !== bb.maxtreach) {
                return bb.maxtreach - ab.maxtreach;

            // Second is reach, which is the number of focal commits reachable in the
            // original commit graph. Again, we compare the maximum value in the
            // segment, not the reach of the individual vertex.
            } else if (ab.maxreach !== bb.maxreach) {
                return bb.maxreach - ab.maxreach;

            // If both of those are the same, then we prefer the longer segment.
            } else if (ab.ids.length !== bb.ids.length) {
                return bb.ids.length - ab.ids.length;

            // If all of these are equal, check to see if different ranks have already
            // been assigned for segments, and rely on those.
            } else if (ab.rank !== bb.rank) {
                return ab.rank - bb.rank;

            // If all of these are equal, we have to make an arbitrary decision,
            // which is persisted to the segment's rank.
            } else {
                // TODO could this change compare order results? doing so makes sort's behavior undefined
                // TODO do we need a flag to indicate it's set this way?
                bb.rank += 1;
                return -1;
            }
        });

        // With the depth group sorted, walk across it and ensure each segment's rank is at
        // least that of its index within the sorted group.
        _.each(depthgroup, function(meta, rank) {
            segmentinfo[meta.segment].rank = Math.max(segmentinfo[meta.segment].rank, rank);
        });
    });

    // FINALLY, assign x and y coords to all visible vertices
    var vertices;
    if (elide) {
        vertices = _(vmeta)
            .pick(function(v) { return _.indexOf(elidable, v.depth, true) === -1; })
            .mapValues(function(v, k) {
                return _.assign({
                    ref: _.has(focalCommits, k) ? focalCommits[k][0] : pvg.get(k), // TODO handle multiple on same commit
                    x: v.depth - _.sortedIndex(elidable, v.depth), // x is depth, less preceding elided x-positions
                    y: segmentinfo[v.segment].rank // y is just the segment rank TODO alternate up/down projection
                }, v);
            }).value();
    } else {
        vertices = _(vmeta)
            .mapValues(function(v, k) {
                return _.assign({
                    ref: _.has(focalCommits, k) ? focalCommits[k][0] : pvg.get(k), // TODO handle multiple on same commit
                    x: v.depth, // without elision, depth is x
                    y: segmentinfo[v.segment].rank // y is just the segment rank TODO alternate up/down projection
                }, v);
            }).value();
    }

    // Build up the list of links
    var links = [], // all the links we'll ultimately return
        xmap = {}; // a map of x-positions to labels, for use on a commit axis

    // transform the list of protolinks compiled during mainwalk into real
    // links, now that the vertices list is assembled and ready.
    _.each(protolinks, function(d) { links.push([vertices[d[0]], vertices[d[1]]]); });

    // collect the vertices together by segment in a way that it's easy to see
    // where connections are needed to cross elision ranges
    var vtxbyseg = _.merge(
        // make sure we get all segments, even empties
        _.mapValues(segmentinfo, function() { return []; }),
        _.groupBy(vertices, function(v) { return v.segment; })
    );

    if (elide) {
        // build an offset map telling us how much a segment's internal offset should
        // be increased by to give the real x-position
        var segoffset = _.mapValues(segmentinfo, function(seg, k) {
            // Recursive function to count length of parent segments
            var r = _.memoize(function(id, rseg) {
                // pseg === id IFF we're on segment 0, which is the base
                return rseg.pseg === id ? vtxbyseg[id].length : vtxbyseg[id].length + r(id, segmentinfo[id]);
            });

            // Return total length of parent segments, but not self length. Return 0 if first segment (no parents)
            return k === "0" ? 0 : r(seg.pseg, segmentinfo[seg.pseg]);
        });

        // then walk through the vertices in order and make the links (elision or no)
        _.each(vtxbyseg, function(vtxs) {
            _.each(_.values(vtxs).sort(function(a, b) { return a.depth - b.depth; }), function(v, k, coll) {
                xmap[v.depth] = segoffset[v.segment] + k;
                if (k === 0) {
                    // all other ops require looking back at previous item, but if
                    // k === 0 then there is no previous item. so, bail out
                    return;
                }

                // if this is true, it means there's an elided range between these two elements
                if (v.depth !== coll[k-1].depth + 1) {
                    xmap[(coll[k-1].depth + 1) + ' - ' + (v.depth - 1)] = segoffset[v.segment] + k-0.5;
                }

                // whether or not there's elision, adjacent vertices in this list need a link
                links.push([coll[k-1], v]);
            });
        });
    } else {
        // If we're not doing elision, things are a whole lot simpler
        _.each(vtxbyseg, function(vtxs) {
            _.each(_.values(vtxs).sort(function(a, b) { return a.depth - b.depth; }), function(v, k, coll) {
                if (k === 0) {
                    return;
                }
                links.push([coll[k-1], v]);
            });
        });
        xmap = _.zipObject(_.unzip([_.range(diameter+1), _.range(diameter+1)]));
    }

    // TODO branches/tags

    return {
        vertices: _.values(vertices),
        links: links,
        elidable: elidable,
        elranges: elranges,
        xmap: xmap,
        rangepos: _.map(elranges, function(range) { return range[0] - _.sortedIndex(elidable, range[0]-1); }),
        g: tree,
        diameter: diameter,
        ediam: ediam,
        root: root,
        segments: segmentinfo,
        dump: function() {
            var flattenVertex = function(v) {
                return _.assign({id: v.id, vtype: v.vertex.type},
                    {outEdges: _.map(v.outEdges, function(eid) {
                        var e = pvg.get(eid);
                        return _.defaults({target: pvg.get(e.target)}, e);
                    })},
                    _.zipObject(_.map(v.vertex.properties, function(p, k) {
                        return [k, p.value];
                    }))
                );
            };
            console.log(_.defaults({
                mid: pvg.mid,
                vertices: _.map(this.vertices, function(v) {
                    return _.defaults({ref: flattenVertex(v.ref)}, v);
                }),
                segments: _.mapValues(this.segments, function(b) {
                    return {
                        ids: _.map(b.ids, function(c) {
                            return pvg.get(c).propv('sha1');
                        }),
                        rank: b.rank
                    };
                })
            }, this));
        }
    };
}

/*
 * Given a viewport width and height, a diameter (number of x positions)
 * and a deviation (max y-deviation from the middle), returns
 * a set of functions for transforming x and y coordinates into the
 * numbers appropriate for the target user coordinate space.
 */
function createTransforms(vpwidth, vpheight, diameter, deviation, revx) {
    var ar = vpwidth / vpheight,
        unit = vpwidth / (diameter + 1); // the base grid unit size, presumably in px
    // TODO different paths depending on whether x or y exceeds ratio
    return {
        unit: function() { return unit; },
        x: function(x) {
            return revx ?
                Math.abs((diameter+1)*unit - (0.5*unit + x*unit))
                : 0.5*unit + x*unit;
        },
        y: function(y) {
            return 0.5*vpheight + y*unit;
        },
    };
}

function dumpGraph(g, rg, mapper) {
    console.log({
        nodes: _.map(g.nodes(), function(n) {
            return mapper(rg.get(n));
        }),
        edges: _.map(g.edges(), function(e) {
            return [mapper(rg.get(e.v)), mapper(rg.get(e.w))];
        }),
    });
}

function clMapper(v) {
    _.assign(v, v.vertex);
    switch (v.vertex.type) {
        case 'logic-state':
            v.lgroup = v.propv('lgroup');
            break;
        case 'commit':
            v.sha1 = v.propv('sha1');
            break;
    }

    return v;
}
