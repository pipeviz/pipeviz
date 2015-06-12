// just so my syntastic complains less
var React = React || {},
    _ = _ || {},
    d3 = d3 || {};

var Viz = React.createClass({
    displayName: "pipeviz-graph",
    getInitialState: function() {
        return {
            force: d3.layout.force()
            .charge(-4000)
            .chargeDistance(250)
            .size([this.props.width, this.props.height])
            .linkStrength(function(link) {
                if (link.source.Typ() == "logic-state") {
                    return 0.5;
                }
                if (link.source.Typ() === "anchor" || link.target.Typ() === "anchor") {
                    return 1;
                }
                return 0.1;
            })
            .linkDistance(function(link) {
                if (link.source.Typ() === "anchor" || link.target.Typ() === "anchor") {
                    return 25;
                }
                return 250;
            })
        };
    },
    getDefaultProps: function() {
        return {
            nodes: [],
            links: [],
            labels: [],
        };
    },
    render: function() {
        // TODO terrible hack this way
        var vd = this.props.vizdata || { ediam: 0, branches: []};
        return React.DOM.svg({
            className: "pipeviz",
            width: "100%",
            //viewBox: '0 0 ' + (vd.ediam + 1) + ' ' + (this.props.height / this.props.width * (vd.ediam + 1))
        });
    },
    shouldComponentUpdate: function(nextProps, prevProps) {
        return nextProps.vizdata !== undefined;
    },
    componentDidUpdate: function() {
        // x-coordinate space is the elided diameter as a factor of viewport width
        var selections = {},
            props = this.props,
            tf = createTransforms(props.width, props.height, props.vizdata.ediam, props.vizdata.branches.length);

        //selections.links = d3.select(el).selectAll('.edge')
            //.data(function() {

            //});

        selections.outerg = d3.select(this.getDOMNode()).append('g');
        selections.outerg
            .attr('id', 'commit-pipeline');
            //.attr('transform', 'translate(0.5, 0.5)');
            //.attr('width', '100%').attr('viewBox', '0 0 ' + (this.props.vizdata.ediam + 2) + ' ' + _.size(this.props.vizdata.branches));
        selections.vertices = selections.outerg.selectAll('.node')
        //selections.vertices = d3.select(this.getDOMNode()).selectAll('.node')
            .data(props.vizdata.vertices, function(d) { return d.ref.id; });

        selections.nodes = selections.vertices.enter().append('g')
            .attr('class', function(d) { return 'node ' + d.ref.Typ(); })
            .attr('transform', function(d) { return 'translate(' + tf.x(d.x) + ',' + tf.y(d.y) + ')'; });
        selections.nodes.append('circle')
            //.attr('cx', function(d) { return tf.x(d.x); })
            //.attr('cy', function(d) { return tf.y(d.y); })
            .attr('r', function(d) { return d.ref.Typ() === "commit" ? tf.unit()*0.03 : tf.unit()*0.3; });

        selections.nodetext = selections.nodes.append('text');
        selections.nodetext.append('tspan').text(function(d) { return d.ref.propv("lgroup"); });
        selections.nodetext.append('tspan').text(function(d) {
            return getCommit(props.graph, d.ref).propv("sha1").slice(0, 7);
        })
        .attr('dy', "1.4em")
        .attr('x', 0)
        .attr('class', function(d) {
            var output = 'commit-subtext',
                commit = getCommit(props.graph, d.ref),
                testState = getTestState(props.graph, commit);
            if (testState !== undefined) {
                output += ' commit-' + testState;
            }

            return output;
        });
    },
    graphRender: function(el, state, props) {
        var link = d3.select(el).selectAll('.link')
            .data(props.links, function(d) { return d.source.id + '-' + d.target.id; }),
        node = d3.select(el).selectAll('.node')
            .data(props.nodes, function(d) { return d.id; });

        link.enter().append('line')
        .attr('class', function(d) {
            if (d.source.Typ() === "anchor" || d.target.Typ() === "anchor") {
                return 'link anchor';
            }
            if (_.has(d, 'path')) {
                return 'link link-commit';
            }
            return 'link';
        })
        .style('stroke-width', function(d) {
            return (d.path && d.path.length > 0) ? 1.5 * Math.sqrt(d.path.length) : 1;
        });

        var nodeg = node.enter().append('g')
        .attr('class', function(d) {
            return 'node ' + d.Typ();
        });

        nodeg.append('circle')
        .attr('x', 0)
        .attr('y', 0)
        .attr('r', function(d) {
            if (d.Typ() === "logic-state") {
                return 45;
            }
            if (d.Typ() === "commit") {
                return 3;
            }
            if (d.Typ() === "anchor") {
                return 0;
            }
        })
        .on('click', props.target);

        nodeg.append('image')
        .attr('class', 'provider-logo')
        .attr('height', 22)
        .attr('width', 22)
        .attr('y', '-37')
        .attr('x', '-10')
        .attr('xlink:href', function(d) {
            if (d.Typ() === "logic-state" && _.has(d.vertex, "provider")) {
                return "assets/" + d.vertex.provider + ".svg";
            }
        });

        var nodetext = nodeg.append('text');
        //nodetext.append('tspan').text(function(d) { return d.name(); });
        nodetext.append('tspan').text(function(d) {
            var v = d.prop("lgroup");
            if (v !== undefined) {
                return v.value;
            }
            return;
        });
        nodetext.append('tspan').text(function(d) {
            // FIXME omg terrible
            if (d.Typ() !== "logic-state") {
                return '';
            }

            return getCommit(props.graph, d).propv("sha1").slice(0, 7);
        })
        .attr('dy', "1.4em")
        .attr('x', 0)
        .attr('class', function(d) {
            if (d.Typ() !== "logic-state") {
                return;
            }

            var output = 'commit-subtext',
                commit = getCommit(props.graph, d),
                testState = getTestState(props.graph, commit);
            if (testState !== undefined) {
                output += ' commit-' + testState;
            }

            return output;
        });

        // Attach the vcs labels
        var label = d3.select(el).selectAll('.vcs-label')
            .data(props.labels, function(d) { return d.id; });
        var labelg = label.enter().append("g")
            .attr("class", "vcs-label");

        // ...ugh scaling around text
        labelg.append("rect").attr("width", 80).attr("height", 20).attr("x", -40).attr("y", -15);
        labelg.append("text")
            .text(function(d) {
                return props.graph.get(d.id).propv("name");
            });

        state.force.on('tick', function() {
            link.attr("x1", function(d) { return d.source.x; })
            .attr("y1", function(d) { return d.source.y; })
            .attr("x2", function(d) { return d.target.x; })
            .attr("y2", function(d) { return d.target.y; });

            node.attr('transform', function(d) { return 'translate(' + d.x + ',' + d.y + ')'; });
            label.attr("transform", function(d) {
                if (d.pos === 0) {
                    return "translate(" + d.l.x + "," + (d.l.y - 65) + ") rotate(90)";
                }
                return "translate(" + (d.l.x + ((d.r.x - d.l.x) * d.pos)) + "," + (d.l.y + ((d.r.y - d.l.y) * d.pos) - 20) + ") rotate(90)";
            });
        });

        node.exit().remove();
        link.exit().remove();
        label.exit().remove();

        state.force.start();

        return false;
    }
});

var VizPrep = React.createClass({
    getInitialState: function() {
        return {
            anchorL: new Anchor("L", 0, this.props.height/2),
            anchorR: new Anchor("R", this.props.width, this.props.height/2),
        };
    },
    extractVizGraph: function(repo) {
        var g = this.props.graph.commitGraph(),
        links = [],
        cmp = this,
        members = {},
        clabels = {},
        nodes = _.filter(_.map(this.props.graph.verticesWithType("logic-state"), function(d) { return _.create(vertexProto, d); }), function(v) {
            var vedges = _.filter(_.map(v.outEdges, function(edgeId) { return cmp.props.graph.get(edgeId); }), isType("version"));
            if (vedges.length === 0) {
                return false;
            }

            if (cmp.props.graph.get(vedges[0].target).propv("repository") === repo) {
                if (!_.has(members, vedges[0].target)) {
                    members[vedges[0].target] = [v];
                    return true;
                }
                // TODO temporary, until commit multitenancy is worked out
                //members[vedges[0].target].push(v);
                //return true;
            }
            return false;
        });

        _.each(this.props.graph.vertices(isType("git-tag", "git-branch")), function(l) {
            var vedges = _.filter(_.map(l.outEdges, function(edgeId) { return cmp.props.graph.get(edgeId); }), isType("version"));

            if (cmp.props.graph.get(vedges[0].target).propv("repository") === repo) {
                if (!_.has(clabels, vedges[0].target)) {
                    clabels[vedges[0].target] = [];
                }
                clabels[vedges[0].target].push(_.create(vertexProto, l));
            }
        });

        // walker to discover 'joints' - otherwise uninteresting commits with
        // multiple descendents that will necessitate creating a node later. also
        // identifies the 'interesting' sources and sinks (apps) that should actually
        // be used, rather than sinks/sources of the commit graph on its own.
        var visited = [], // "black" list - vertices that have been visited
        deepestApp, // var to hold the deepest app we've visited (for sink tracking)
        npath = [], // the current path, but only with logic states
        isources = [], // interesting sources
        isinks = [], // interesting sinks
        path = [], // the current path of interstitial commits
        sinkmap = {}, // map that tracks eventual sink target; prevents duplicate traversal
        prepwalk = function(v) {
            // DAG, so skip grey/back-edge

            var has_app = false;
            // check if this commit has an app. don't check if white b/c sink calc necessitates we
            // recheck this every time through.
            if (_.has(members, v) && _.filter(members[v], isType("logic-state")).length !== 0) {
                if (npath.length === 0) {
                    // nothing in the npath, this *could* be a source.
                    isources.push(v);
                }
                // so we know to pop npath stack and check if source later
                has_app = true;
                npath.push(v);
                // track that this is now our deepest app
                deepestApp = v;
            }
            // if black, and nothing in members about the commit, it's a joint
            if (visited.indexOf(v) !== -1 && !_.has(members, v)) {
                members[v] = [_.create(vertexProto, cmp.props.graph.get(v))];
                // if the sinkmap knows about v, set that as the deepest app
                if (_.has(sinkmap, v)) {
                    deepestApp = sinkmap[v];
                }
                return;
            }

            path.push(v);
            // If v is a merge commit, it will have multiple successors, and we must implement
            // first-parent-like handling: only include the n>1 parents if there is an
            // interesting commit along that path.
            // TODO this logic will probably backfire on some reasonable graph structures
            var succ = g.successors(v);
            if (succ.length > 0) {
                //prepwalk(succ[0]);
                _.each(succ, function(s) {
                    prepwalk(s);
                });
                // TODO experiment with a method where we only walk first parent, but use as
                // starting point all the interesting vertices. this should guarantee we get
                // only the necessary joints, but may require an optional extra pass to discover
                // necessary connections between interesting vtx. This approach means what we're
                // *really* doing here is forming an induced subgraph using only paths we know
                // will be interesting.
            }
            visited.push(v);
            path.pop();

            if (has_app) {
                npath.pop();
                // if deepestApp is self, it's because we found no apps further down the dag,
                // so this is effectively a sink
                if (deepestApp === v) {
                    isinks.push(v);
                    // write this vertex into the sinkmap for all current paths
                    _.merge(sinkmap, _.zipObject(_.zip(path, _.map(_.range(path.length), function() { return v; })))); // _.fill comes in lodash v.3.8.0
                }
            }
        };

        // now traverse depth-first to figure out the overlaid edge structure
        var labels = [], // the set of non-app-coinciding labels and their relative positions
        lpnodes = {}, // the set of non-app-coinciding labels that have found their origin, but not their target
        walk = function(v) { // main depth-first walker
            // Each top-level entry into this DF search is from a virtual source found
            // by the first walker

            // The npath is a subset of the current visit path, and contains *only*
            // commits that have something interesting there (a logic-state instance, or a joint).
            // Thus, we have to keep a locally-scoped variable to indicate whether, on returning
            // from this level the search, the npath should be popped off (aka, the current level
            // is one of these interesting commits).
            var pop_npath = false;

            // For easy reference, we keep the head of the npath in, again, a locally-scoped var
            // called 'from'. Thus, for this level of descent, the contents of from are the last
            // array of "interesting" things we visited in our search. If the npath is empty -
            // which will occur iff v is a source vertex/we're at the top of a descent path,
            // then we initialize it to an empty array.
            var from = npath.length === 0 ? [] : npath[npath.length - 1];

            // Ordinarily a depth-first search would check here for a back-edge, or "grey" vertex;
            // however, git guarantees commit graph is acyclic. A back-edge, by definition, identifies
            // a cycle, which we know we don't have, so we can skip straight ahead to the "black"
            // check - see if we've already completed a visit to this vertex. We bend the definition
            // a little, though - we only ever mark interesting nodes as black, as we need to always
            // walk the uninteresting commit path for other algorithimic purposes.
            if (visited.indexOf(v) !== -1) {
                // Vertex is black/visited, so we need to create links between instances in
                // our 'from' and instances in the visited vertex.
                _.each(members[v], function(tgt) {
                    _.each(from, function(src) {
                        // slice guarantees path array is copied
                        links.push({ source: src, target: tgt, path: path.slice(0) });
                    });
                });

                // process all label nodes we have waiting around
                _.forOwn(lpnodes, function(llen, id) {
                    // need to push the actual objects on so the viz can cheat and track x/y props
                    labels.push({id: id, l: from[0], r: clabels[v][0], pos: llen / (path.length+1)});
                });

                // zero out lpnodes set
                lpnodes = {};

                // Don't need to pop path or npath b/c we haven't pushed the black vertex onto either.
                return;
            }

            // see if this is a commit that's interesting (has an app instance or label)
            if (_.has(members, v) || _.has(clabels, v)) {
                // At this point, it's possible that we've found any of the following possible combinations:
                // 1. An instance
                // 2. A branch/tag
                // 3. A joint
                // 4. An instance and a branch/tag
                // 5. A joint and a branch/tag
                //
                // As well as any N>1 of any type in any of those possibilities. Labels are the subordinate
                // case; we handle them one way if a commit/joint is present, and a different way if not.
                // And so, we have the following if/then structure:
                if (_.has(members, v)) {
                    // has at least one app, or is a commit joint. create link from each last thing to
                    // each interesting target.
                    // TODO this is the first spot where where we need to collapse multiple things on the same commit down into one
                    _.each(members[v], function(tgt) {
                        _.each(from, function(src) {
                            links.push({ source: src, target: tgt, path: path.slice(0) });
                            if (tgt.Typ() === "commit") {
                                // push the joint onto the node list
                                nodes.push(tgt);
                            }
                        });
                    });

                    // Since we've reached a properly interesting point, we need to process all the labels we discovered
                    // between here and the last interesting point, as they need to appear on that edge.
                    _.forOwn(lpnodes, function(llen, id) {
                        // need to push the actual objects on so the viz can cheat and track x/y props
                        labels.push({id: id, l: from[0], r: members[v][0], pos: llen / (path.length+1)});
                    });

                    // We've processed all the interstitial labels, so we empty the lpnodes set
                    lpnodes = {};

                    // There also could be some labels directly on this commit. This will add them.
                    _.each(clabels[v], function(lbl) {
                        labels.push({id: lbl.id, l: members[v][0], r: members[v][0], pos: 0});
                    });

                    // We've found a new point of interest, so the commit path tracker, which is
                    // only responsible for tracking the path BETWEEN interesting points, not the
                    // entire path, is emptied.
                    path = [];
                    // Push newly-found instances or joints onto the npath; it will become the new 'from' on the next level.
                    npath.push(members[v]);
                    // correspondingly, indicate to pop the npath when exiting
                    pop_npath = true;
                } else {
                    // Not a real node point, just a label, so need to push the commit onto the path
                    path.push(v);

                    // For each label found, record the length of the path walked so far.
                    _.each(clabels[v], function(lbl) {
                        lpnodes[lbl.id] = path.length;
                    });
                }
            }
            else {
                // not a node point and not self - push commit onto path
                path.push(v);
            }

            // recursive call, the crux of this depth-first traversal. but we
            // skip it if the first search proved this to be a sink
            if (isinks.indexOf(v) === -1) {
                g.successors(v).map(function(s) {
                    walk(s);
                });
            }
            // Recursion is done for this vertex; now it's cleanup time.

            // Mark commit black/visited...but only if it's an interesting one.
            // This is OK to do because we already traversed once and found all
            // the joints, and those are guaranteed to be members.
            if (_.has(members, v)) {
                visited.push(v);
            }

            if (pop_npath) {
                npath.pop();
            }
            // pop current commit off the visit path
            path.pop();

            // decrement lpnodes values, and filter out any that reach 0;
            // they're not between any apps and thus won't be displayed.
            // this SHOULD be unnecessary b/c we preselected good sinks and
            // sources, but just in case.
            lpnodes = _(lpnodes)
                .mapValues(function(v) { return v-1; })
                .filter(function(v) { return v > 0; }) // > 1?
                .value();
        };

        var stack = _.filter(g.sources(), function(d) {
            return cmp.props.graph.get(d).propv("repository") === repo;
        });
        // DF walk, working from source commit members
        while (stack.length !== 0) {
            prepwalk(stack.pop());
        }

        // traversal pattern almost guarantees duplicate sinks
        isinks = _.uniq(isinks);

        // put the source anchor links in
        _.each(isources, function(commit) {
            _.each(_.filter(members[commit], isType("logic-state")), function(member) {
                links.push({ source: cmp.state.anchorL, target: member });
            });
        });

        // and the sink anchor links
        _.each(isinks, function(commit) {
            _.each(_.filter(members[commit], isType("logic-state")), function(member) {
                links.push({ source: member, target: cmp.state.anchorR });
            });
        });

        // reset search path vars
        npath = [];
        visited = [];
        path = [];

        while (isources.length !== 0) {
            walk(isources.pop());
        }

        return [nodes, links, labels];
    },
    getDefaultProps: function() {
        return {
            width: 0,
            height: 0,
            graph: pvGraph({id: 0, vertices: []}),
            focalRepo: "",
        };
    },
    shouldComponentUpdate: function(nextProps) {
        // In the graph object, state is invariant with respect to the message id.
        return nextProps.graph.mid !== this.props.graph.mid;
    },
    render: function() {
        //var vizdata = this.extractVizGraph(this.props.focalRepo);
        //return React.createElement(Viz, {width: this.props.width, height: this.props.height, graph: this.props.graph, nodes: vizdata[0].concat(this.state.anchorL, this.state.anchorR), links: vizdata[1], labels: vizdata[2]});
        return React.createElement(Viz, {width: this.props.width, height: this.props.height, graph: this.props.graph, vizdata: extractVizGraph(this.props.graph, this.props.focalRepo)});
    },
});

var App = React.createClass({
    dispayName: "pipeviz",
    getDefaultProps: function() {
        return {
            vizWidth: window.innerWidth,
            vizHeight: window.innerHeight,
            graph: pvGraph({id: 0, vertices: []}),
        };
    },
    mostCommonRepo: function(g) {
        return _.reduce(_.countBy(_.filter(g.verticesWithType("logic-state"), function(v) {
            var vedges = _.filter(_.map(v.outEdges, function(edgeId) { return g.get(edgeId); }), isType("version"));
            return vedges.length !== 0;
        }), function(v) {
            return g.get(_.filter(_.map(v.outEdges, function(edgeId) { return g.get(edgeId); }), isType("version"))[0].target).propv("repository");
        }), function(accum, count, repo) {
            return count < accum[1] ? accum : [repo, count];
        }, ["", 0])[0];
    },
    render: function() {
        return React.createElement("div", {id: "pipeviz"},
                   React.createElement(VizPrep, {width: this.props.vizWidth, height: this.props.vizHeight, graph: this.props.graph, focalRepo: this.mostCommonRepo(this.props.graph)})
              );
    },
});

var e = React.render(React.createElement(App), document.body);
var genesis = new WebSocket("ws://" + window.location.hostname + ":" + window.location.port + "/sock");
genesis.onmessage = function(m) {
    //console.log(m);
    e.setProps({graph: pvGraph(JSON.parse(m.data))});
};
