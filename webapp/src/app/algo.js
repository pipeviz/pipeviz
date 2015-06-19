// TODO memoize this
var reachCounts = function(g, v, filter) {
    var succ = g.successors(v),
    // Recursive folder to create a flattened array with all successors.
    r = function(accum, value) {
        return accum.concat(_.foldl(g.successors(value), r, [value]));
    };


    if (succ === undefined) {
        return 0;
    } else if (filter === undefined) {
        return _.uniq(_.foldl(succ, r, [v])).length;
    } else {
        return _.uniq(_.intersection(_.foldl(succ, r, [v]), filter)).length;
    }
};

var reachCountp = function(g, v, filter) {
    var pred = g.predecessors(v),
    // Recursive folder to create a flattened array with all predecessors.
    r = function(accum, value) {
        return accum.concat(_.foldl(g.predecessors(value), r, [value]));
    };

    if (pred === undefined || pred.length === 0) {
        return 0;
    } else if (filter === undefined) {
        return _.uniq(_.foldl(pred, r, [v])).length;
    } else {
        return _.uniq(_.intersection(_.foldl(pred, r, [v]), filter)).length;
    }
};

//var vizExtractorProto = {
    //apply: function(pvg) {
        //this._pvg = pvg;
    //}
//},
//vizExtractor = function(pvg) {
    //var ve = Object.create(vizExtractorProto);
    //ve.apply(pvg);
    //return ve;
//}
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
};

/*
 * The main magic function for processing a pv graph into the base state usable
 * for the app/commit viz.
 *
 */
function extractVizGraph(g, repo) {
    var focalCommits = {},
    focal = _.filter(_.map(g.verticesWithType("logic-state"), function(d) { return _.create(vertexProto, d); }), function(v) {
        var vedges = _.filter(_.map(v.outEdges, function(edgeId) { return g.get(edgeId); }), isType("version"));
        if (vedges.length === 0) {
            return false;
        }

        if (g.get(vedges[0].target).propv("repository") === repo) {
            if (!_.has(focalCommits, vedges[0].target)) {
                focalCommits[vedges[0].target] = [];
            }
            focalCommits[vedges[0].target].push(v);
            return true;
        }
    });

    if (focal.length === 0) {
        return;
    }

    var cg = g.commitGraph(), // the git commit graph TODO narrow to only commits in repo
    fg = new graphlib.Graph(), // graph with all non-focal vertices contracted and edges transposed (direction reversed)
    isg = new graphlib.Graph(), // A tree, almost an induced subgraph, representing first-parent commit graph paths
    visited = {},
    candidates = [];

    // Depth-first walk to build the focal graph and find root candidates
    var fgwalk = function(v, fpath) {
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

    // TODO this is commented b/c there's something horribly non-performant in the fgwalk impl atm
    //_.each(focalCommits, function(d, k) {
        //fgwalk(k, []);
    //});

    // Nearly the same walk, but only follow first parent. Builds root candidate list
    // AND build the first-parent tree (almost an induced subgraph, but not quite) at
    // the same time.
    var isgwalk = function(v, last) {
        // We always want to record the (reversed) edge, unless last does not exist
        if (last !== undefined) {
            isg.setEdge(v, last);
        }

        // If vertex is already visited, it's a candidate for being the root
        if (_.has(visited, v)) {
            candidates.push(v);
            return;
        }

        // Only visit first parent; that's the path that matters to our subgraph/tree
        var succ = cg.successors(v);
        if (succ !== undefined && succ.length > 0) {
            isgwalk(succ[0], v);
        }

        visited[v] = true;
    };

    // reset visited list
    visited = {};

    _.each(focalCommits, function(d, k) {
        isgwalk(k);
    });

    // Now we have to find the topologically largest common root among all candidates.
    var root;

    // If there's only one focal vertex then things are a little weird - for
    // one, the isgwalk won't find any candidates. In that case, set the
    // candidates list to the single vertex itself.
    if (_.size(focalCommits) === 1) {
    //if (focal.length === 1) { TODO i think this will be correct once there's multi-focus per commit handling
        root = _.keys(focalCommits)[0];
    } else {
        // This identifies the shared root (in git terms, the merge base) from all
        // candidates by walking down the reversed subgraph/tree until we find a
        // vertex in the candidate list. The first one we find is guaranteed to
        // be the root.
        var rootfind = function(v) {
            if (candidates.indexOf(v) !== -1) {
                root = v;
                return;
            }

            var succ = isg.successors(v) || [];
            if (succ.length > 0) {
                rootfind(succ[0]);
            }
        };
        rootfind(isg.sources()[0]);
    }

    var vmeta = {}, // metadata we build for each vertex. keyed by vertex id
    protolinks = [], // we can start figuring out some links in the next walk
    branches = 0, // total number of divergent branch paths. starts at 0, increases as needed
    mainwalk = function(v, path, branch) {
        // tree, so zero possibility of revisiting any vtx; no "visited" checks needed
        var succ = isg.successors(v) || [];
        if (_.has(focalCommits, v)) {
            vmeta[v] = {
                depth: path.length, // distance from root
                interesting: true, // all focal commits are interesting
                reach: reachCountp(cg, v, _.keys(focalCommits)),
                treach: reachCounts(isg, v, _.keys(focalCommits)),
                branch: branch
            };
        } else {
            var psucc = path.length === 0 ? [] : isg.successors(path[path.length -1]);
            vmeta[v] = {
                depth: path.length,
                interesting: succ.length > 1 || psucc.length > 1, // interesting only if has multiple successors, or parent did
                reach: reachCountp(cg, v, _.keys(focalCommits)),
                treach: reachCounts(isg, v, _.keys(focalCommits)),
                branch: branch
            };
        }

        path.push(v);
        _.each(succ, function(d, i) {
            if (succ.length > 1) {
                // we know this'll be included, so add to protolinks right now
                protolinks.push([v, d]);
            }

            if (i === 0) {
                // if first, consider it the same branch
                mainwalk(d, path, branch);
            } else {
                // otherwise, increment then pass branches counter
                mainwalk(d, path, ++branches);
            }
        });
        path.pop();
    };

    // we only need to enter at root to get everything
    mainwalk(root, [], 0);

    // now we have all the base meta; construct elidables lists and branch rankings
    var elidable = _(vmeta)
        .filter(function(v) { return v.interesting === false; })
        .map(function(v) { return v.depth; })
        .uniq()
        .value()
        .sort(function(a, b) { return a - b; }), // TODO bubble sort, meh
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
    }, []),
    branchinfo = _(vmeta)
        .mapValues(function(v, k) { return { branch: v.branch, id: parseInt(k) }; })
        .groupBy(function(v) { return v.branch; }) // collects vertices on same branch into a single array
        .mapValues(function(v) {
            return {
                ids: _.map(v, function(v2) { return v2.id; }),
                rank: 0,
            };
        })
        .value(); // object keyed by branch number w/branch info

    _(vmeta).groupBy(function(v) { return v.depth; })
        .each(function(metas, x) {
            // TODO this needs to get much, much smarter in order to weave smartly as unsigned y
            // sort first by reach, then by tree-reach. if those end up equal, fuck it, good nuf
            _(metas).sortBy(["reach", "treach"])
                .reverse()
                .each(function(meta, rank) {
                    branchinfo[meta.branch].rank = Math.max(branchinfo[meta.branch].rank, rank);
                });
        });

    // FINALLY, assign x and y coords to all visible vertices
    var vertices = _(vmeta)
        .pick(function(v) { return _.indexOf(elidable, v.depth, true) === -1; })
        .mapValues(function(v, k) {
            return _.assign({
                ref: _.has(focalCommits, k) ? focalCommits[k][0] : g.get(k), // TODO handle multiple on same commit
                x: v.depth - _.sortedIndex(elidable, v.depth), // x is depth, less preceding elided x-positions
                y: branchinfo[v.branch].rank // y is just the branch rank TODO alternate up/down projection
            }, v);
        }).value(),
    diameter = _.max(_.map(focalCommits, function(v, k) { return vmeta[k].depth; })),
    ediam = diameter - elidable.length;

    // Build up the list of links
    var links = [], // all the links we'll ultimately return
        xmap = {}; // a map of x-positions to labels, for use on a commit axis

    // transform the list of protolinks compiled during mainwalk into real
    // links, now that the vertices list is assembled and ready.
    _.each(protolinks, function(d) { links.push([vertices[d[0]], vertices[d[1]]]); });

    // collect the vertices together by branch in a way that it's easy to see
    // where connections are needed to cross elision ranges, then walk through
    // the vertices in order and make the links (elision or no)
    _.each(_.groupBy(vertices, function(v) { return v.branch; }), function(vtxs) {
        _.each(_.values(vtxs).sort(function(a, b) { return a.depth - b.depth; }), function(v, k, coll) {
            if (v.branch === 0) {
                // the k-count is only correct for the xmap on the first branch,
                // because other branches are guaranteed not to have their first
                // member be at the 0 x-position
                xmap[v.depth] = k;
            }
            if (k === 0) {
                // all other ops require looking back at previous item, but if
                // k === 0 then there is no previous item. so, bail out
                return;
            }

            // if this is true, it means there's an elided range between these two elements
            if (v.depth !== coll[k-1].depth + 1) {
                if (v.branch === 0) { // same reasoning as above
                    xmap[(coll[k-1].depth + 1) + ' - ' + (v.depth - 1)] = k-0.5;
                }
            }

            // whether or not there's elision, adjacent vertices in this list need a link
            links.push([coll[k-1], v]);
        });
    });

    // TODO branches/tags

    return {
        vertices: _.values(vertices),
        links: links,
        elidable: elidable,
        elranges: elranges,
        xmap: xmap,
        rangepos: _.map(elranges, function(range) { return range[0] - _.sortedIndex(elidable, range[0]-1); }),
        g: isg,
        diameter: diameter,
        ediam: ediam,
        root: root,
        branches: branchinfo,
    };
}

/*
 * Given a viewport width and height, a diameter (number of x positions)
 * and a deviation (max y-deviation from the middle), returns
 * a set of functions for transforming x and y coordinates into the
 * numbers appropriate for the target user coordinate space.
 */
function createTransforms(vpwidth, vpheight, diameter, deviation) {
    var ar = vpwidth / vpheight,
        unit = vpwidth / (diameter + 1); // the base grid unit size, presumably in px
    // TODO different paths depending on whether x or y exceeds ratio
    return {
        unit: function() { return unit; },
        x: function(x) {
            return Math.abs((diameter+1)*unit - (0.5*unit + x*unit));
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
