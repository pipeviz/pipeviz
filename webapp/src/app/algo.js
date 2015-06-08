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

    var cg = g.commitGraph(), // TODO narrow to only commits in repo
    fg = new graphlib.Graph(), // graph with all non-focal vertices contracted and edges transposed (direction reversed)
    isg = new graphlib.Graph(), // the induced subgraph; actually a tree, for now
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
            _.each(last, function(fvs) {
                _.each(fvs, function(fv) {
                    fg.setEdge(v, fv);
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
            _.each(last, function(fvs) {
                _.each(fvs, function(fv) {
                    fg.setEdge(v, fv);
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

    _.each(focalCommits, function(d) {
        fgwalk(d, []);
    });

    // TODO exception handling for case where there's multiple roots...should that be impossible?

    // Nearly the same walk, but only follow first parent
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

        // Only visit first parent; that's the path that matters to our subgraph
        var succ = cg.successors(v);
        if (succ.length > 0) {
            isgwalk(succ[0], v);
        }

        visited[v] = true;
    };
}

// TODO memoize this
var reachCount = function(g, v) {
    var succ = g.successors(v),
    r = function(accum, value) {
        accum.push(_.foldl(g.successors(value), r, []));
        return accum;
    };

    return succ === undefined ? 0 : _.flatten(_.foldl(succ, r, [])).length + 1;
};
