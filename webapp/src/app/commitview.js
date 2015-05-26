//define(function (require) {

    //var _ = require('lodash/dist/lodash'),
    //d3 = require('d3/d3'),
    //React = require('react/react-with-addons');
    //graphlib = require('graphlib/dist/graphlib.core');

    /* All the helper objects (just copied over for now) */
    var _objcounter = 0;

    function _sharedId() {}

    _sharedId.prototype.objid = function() {
        return this._objid;
    };

    _sharedId.prototype._nextId = function() {
        this._objid = _objcounter++;
    };

    function Anchor() {
        this.fixed = true; // tells d3 not to move it
        this.x = 0;
        this.y = 0;
        this._nextId();
    }

    Anchor.prototype = new _sharedId();

    Anchor.prototype.vType = function() {
        return 'sort-anchor';
    };

    Anchor.prototype.name = function() {
        return '';
    };

    function DataSet(obj, name, space) {
        _.assign(this, obj);
        this._name = name;
        this._space = space;
        this._nextId();
    }

    DataSet.prototype = new _sharedId();

    DataSet.prototype.vType = function() {
        return 'dataset';
    };

    DataSet.prototype.space = function() {
        return this._space;
    };

    DataSet.prototype.name = function() {
        return this._name;
    };

    function DataSpace(name, container) {
        this._sets = {}; // FIXME ugh circular
        this._name = name;
        this._container = container;
        this._nextId();
    }

    DataSpace.prototype = new _sharedId();

    DataSpace.prototype.vType = function() {
        return 'dataspace';
    };

    DataSpace.prototype.name = function() {
        return this._name;
    };

    DataSpace.prototype.dataSets = function() {
        return this._sets;
    };

    function LGroup(name, obj) {
        this._name = name;
        this._obj = obj;
    }

    LGroup.prototype.vType = function() {
        return 'lgroup';
    };

    LGroup.prototype.name = function() {
        return this._name;
    };

    LGroup.prototype.ref = function() {
        return this._obj;
    };

    LGroup.prototype.objid = function() {
        return this._obj.objid();
    };

    function LogicState(obj, path, container) {
        _.assign(this, obj);
        this._path = path;
        this._container = container;
        this._nextId();
    }

    LogicState.prototype = new _sharedId();

    LogicState.prototype.vType = function() {
        return 'logic';
    };

    LogicState.prototype.name = function() {
        if (_.has(this, 'nick')) {
            return this.nick;
        }

        var p = this._path.split('/');
        return p[p.length-1];
    };

    function Process(obj, container) {
        _.assign(this, obj);
        this._container = container;
        this._nextId();
    }

    Process.prototype = new _sharedId();

    Process.prototype.vType = function() {
        return 'process';
    };

    Process.prototype.name = function() {
        return _.reduce(this.logicStates(), function(accum, ls) {
            accum.push(ls.name());
            return accum;
        }, []).join('/');
    };

    Process.prototype.logicStates = function() {
        return _.pick(this._container.logicStates(), this['logic states']);
    };

    Process.prototype.dataSpaces = function() {
        return _.pick(this._container.dataSpaces(), this['data spaces']);
    };

    function Container(obj) {
        this.hostname = obj.hostname;
        this.type = obj.type;
        this.provider = obj.provider;
        if (obj.ipv4 !== undefined) {
            this.ipv4 = obj.ipv4;
        }

        this._nextId();

        var that = this;
        this._dataSpaces = _.has(obj, 'data spaces') ? _.mapValues(obj['data spaces'], function(space, id) {
            var ds = new DataSpace(id, that);
            ds._sets = _.mapValues(space, function(dc, set) {
                return new DataSet(dc, set, ds);
            });

            return ds;
        }) : {};

        this._logics = _.has(obj, 'logic states') ? _.mapValues(obj['logic states'], function(l, path) { return new LogicState(l, path, that); }) : {};
        this._processes = _.has(obj, 'processes') ? _.map(obj.processes, function(p) { return new Process(p, that); }) : {};
    }

    Container.prototype = new _sharedId();

    Container.prototype.vType = function() {
        return 'container';
    };

    Container.prototype.name = function() {
        return this.hostname;
    };

    Container.prototype.logicStates = function() {
        return this._logics;
    };

    Container.prototype.processes = function() {
        return this._processes;
    };

    Container.prototype.dataSpaces = function() {
        return this._dataSpaces;
    };

    Container.prototype.dataSets = function() {
        return _.flatten(_.map(this.dataSpaces(), function(space) {
            return _.values(space.dataSets());
        }));
    };

    Container.prototype.forInfo = function() {
        return {
            hostname: this.hostname,
            ipv4: this.ipv4.toString(),
            type: this.type,
        };
    };

    Container.prototype.findProcess = function(loc) {
        var f, found = false;
        if (loc.type === "unix") {
            f = function(proc) {
                return _.find(proc.listen, function(ingress) {
                    if (ingress.type !== 'unix') {return false;}
                    return loc.path === ingress.path;
                }) ? proc : false;
            };
        } else {
            f = function(proc) {
                return _.find(proc.listen, function(ingress) {
                    if (ingress.type !== 'net' && ingress.type !== 'port') {return false;}
                    if (loc.port !== ingress.number) {return false;}
                    if (_.isString(ingress.proto)) {
                        return loc.proto === ingress.proto;
                    }

                    return _.contains(ingress.proto, loc.proto);
                }) ? proc : false;
            };
        }

        _.each(this.processes(), function(p) {
            found = f(p);
            if (found) {return false;}
        });

        return found;
    };

    Container.prototype.findDataSet = function(gen) {
        var found = false;

        if (_.has(this._dataSpaces, gen['data space'])) {
            var ds = this._dataSpaces[gen['data space']];
            if (_.has(ds._sets, gen['data set'])) {
                found = ds._sets[gen['data set']];
            }
        }

        return found;
    };

    // PVD: PipeVizDatastore
    function PVD() {
        this._containers = {};
    }

    PVD.prototype.attachContainer = function(container) {
        // FIXME hostname is not a uuid
        this._containers[container.hostname] = container;
        return this;
    };

    /**
     * Attempt to locate a container by hostname.
     */
    PVD.prototype.byHostname = function(hostname) {
        if (_.has(this._containers, hostname)) {
            return this._containers[hostname];
        }
        return false;
    };

    PVD.prototype.eachContainer = function(lambda, thisArg) {
        return _.each(this._containers, lambda, thisArg);
    };

    /**
     * Finds all logical groups contained in this PVD.
     *
     * Returns a map keyed by group name with arrays of LGroup objects as values.
     */
    PVD.prototype.findLogicalGroups = function() {
        var lgroups = {};

        // FIXME just checks logical states. bc we're not a real graphdb (...yet)
        _.each(this._containers, function(container) {
            _.each(container.logicStates(), function(ls) {
                if (_.has(ls, 'lgroup')) {
                    // FIXME possible that there could be multiple...and that's a weird case to handle anyway
                    lgroups[ls.lgroup] = new LGroup(ls.lgroup, ls);
                }
            });
        });

        return lgroups;
    };

    /**
     * Returns a 2-tuple of the nodes and links present in this PVD.
     *
     * If provided, the nodeFilter and linkFilter functions will be called prior
     * to each object being added; objects will only be added if the filter func
     * returns true.
     *
     * If either filter function is not provided, a dummy is used instead that
     * unconditionally returns true (allows everything through).
     */
    PVD.prototype.nodesAndLinks = function(nodeFilter, linkFilter) {
        // if filter funcs were not provided, just attach one that always passes
        var pvd = this,
        nf = (nodeFilter instanceof Function) ? nodeFilter : function() { return true; },
        lf = (linkFilter instanceof Function) ? linkFilter : function() { return true; },
        nodes = [],
        links = [],
        maybeAddNode = function(node) {
            if (nf(node)) {
                nodes.push(node);
            }
        },
        maybeAddLink = function(link) {
            if (lf(link)) {
                links.push(link);
            }
        };

        // TODO this approach makes links entirely subordinate to nodes...fine for
        // now, maybe not the best in the long run though
        _.each(this._containers, function(container) {
            _.each(container.logicStates(), function(v) {
                maybeAddNode(v);
                maybeAddLink({source: container, target: v});
            });

            _.each(container.processes(), function(v) {
                maybeAddNode(v);
                maybeAddLink({source: container, target: v});
            });

            _.each(container.dataSpaces(), function(v) {
                maybeAddNode(v);
                maybeAddLink({source: container, target: v});
            });

            _.each(container.dataSets(), function(v) {
                maybeAddNode(v);
                maybeAddLink({source: container, target: v});
            });

            maybeAddNode(container);
        });

        // temp var - 'referred container'
        var rc;

        // All hierarchical data is processed; second pass for referential.
        _.forOwn(this._containers, function(cv) {
            // find logic refs to data
            _.each(cv.logicStates(), function(l) {
                if (_.has(l, 'datasets')) {
                    _.forOwn(l.datasets, function(dv, dk) {
                        var link = _.assign({source: l, name: dk}, dv);
                        var proc = false;
                        // check for hostname, else assume path
                        if (_.has(dv.loc, 'hostname')) {
                            rc = pvd.byHostname(dv.loc.hostname);
                            if (rc) {
                                proc = rc.findProcess(dv.loc);
                            }
                        } else {
                            proc = cv.findProcess(dv.loc);
                        }

                        if (proc) {
                            link.target = proc;
                            maybeAddLink(link);
                        }
                    });
                }
            });

            // next, link processes to logic states & data spaces
            _.each(cv.processes(), function(proc) {
                _.each(proc.logicStates(), function(ls) {
                    maybeAddLink({source: proc, target: ls});
                });

                _.each(proc.dataSpaces(), function(ds) {
                    maybeAddLink({source: proc, target: ds});
                });
            });

            _.each(cv.dataSpaces(), function(dg) {
                _.each(dg.dataSets(), function(ds) {
                    // direct tree/parentage info
                    maybeAddLink({source: dg, target: ds});
                    // add dataset genesis information
                    if (ds.genesis !== "α") {
                        var link = {source: ds};
                        var ods = false;
                        if (_.has(ds.genesis, 'hostname')) {
                            rc = pvd.byHostname(ds.genesis.hostname);
                            if (rc) {
                                ods = rc.findDataSet(ds.genesis);
                            }
                        } else {
                            ods = cv.findDataSet(ds.genesis);
                        }

                        if (ods) {
                            link.target = ods;
                            maybeAddLink(link);
                        }
                    }
                });
            });
        });

        return [nodes, links];
    };

    var graphRender = function(el, state, props) {
        var link = d3.select(el).selectAll('.link')
        .data(props.links, function(d) { return d.source.objid() + '-' + d.target.objid(); }),
        node = d3.select(el).selectAll('.node')
        .data(props.nodes, function(d) { return d.objid(); });

        link.enter().append('line')
        .attr('class', function(d) {
            if (d.source instanceof Anchor || d.target instanceof Anchor) {
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
            return 'node ' + d.vType();
        });

        nodeg.append('circle')
        .attr('x', 0)
        .attr('y', 0)
        .attr('r', function(d) {
            if (d instanceof LGroup) {
                return 45;
            }
            if (d instanceof Anchor) {
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
            if (d instanceof Anchor) {
                return;
            }

            // FIXME hahahahahhahahahahahahhaha hardcoded
            return 'assets/' + d.ref()._container.provider + '.svg';
        });

        var nodetext = nodeg.append('text');
        nodetext.append('tspan')
        .text(function(d) { return d.name(); });
        nodetext.append('tspan')
        .text(function(d) {
            // FIXME omg terrible
            if (d instanceof Anchor) {
                return '';
            }

            return d.ref().id.commit.slice(0, 7);
        })
        .attr('dy', "1.4em")
        .attr('x', 0)
        .attr('class', function(d) {
            if (d instanceof Anchor) {
                return;
            }

            var output = 'commit-subtext',
            commit = d.ref().id.commit;

            if (_.has(props.commitMeta, commit) &&
                _.has(props.commitMeta[commit], 'testState')) {
                output += ' commit-' + props.commitMeta[commit].testState;
            }

            return output;
        });

        node.exit().remove();
        link.exit().remove();

        state.force.on('tick', function() {
            link.attr("x1", function(d) { return d.source.x; })
            .attr("y1", function(d) { return d.source.y; })
            .attr("x2", function(d) { return d.target.x; })
            .attr("y2", function(d) { return d.target.y; });

            node.attr('transform', function(d) { return 'translate(' + d.x + ',' + d.y + ')'; });
        });

        state.force.start();
        return false;
    };

    var Viz = React.createClass({
        displayName: "pipeviz-graph",
        getInitialState: function() {
            return {
                force: d3.layout.force()
                .charge(-4000)
                .chargeDistance(250)
                .size([this.props.width, this.props.height])
                .linkStrength(function(link) {
                    if (link.source instanceof Container) {
                        return 0.5;
                    }
                    if (link.source instanceof Anchor || link.target instanceof Anchor) {
                        return 1;
                    }
                    return 0.3;
                })
                .linkDistance(function(link) {
                    if (link.source instanceof Anchor || link.target instanceof Anchor) {
                        return 25;
                    }
                    return 250;
                })
            };
        },
        getDefaultProps: function() {
            return {
                width: window.innerWidth,
                height: window.innerHeight,
                nodes: [],
                links: [],
                target: function() {}
            };
        },
        render: function() {
            return React.DOM.svg({
                className: "pipeviz",
                viewBox: "0 0 " + this.props.width + " " + this.props.height
            });
        },
        componentDidUpdate: function() {
            this.state.force.nodes(this.props.nodes);
            this.state.force.links(this.props.links);

            return graphRender(this.getDOMNode(), this.state, this.props);
        },
        shouldComponentUpdate: function(nextProps, nextState) {
            // FIXME totally cheating for now and just going by length.
            if (nextProps.nodes.length !== this.state.force.nodes().length) {
                return true;
            }
            if (nextProps.links.length !== this.state.force.links().length) {
                return true;
            }
            return false;
        }
    });

    var InfoBar = React.createClass({
        displayName: 'pipeviz-info',
        render: function() {
            var t = this.props.target,
            cmp = this;

            var outer = {
                id: "infobar",
                children: []
            };

            if (typeof t !== 'object') {
                outer.children = [React.DOM.p({}, "nothing selected")];
                // drop out early for the empty case
                return React.DOM.div(outer);
            }

            // find all linked envs for the logic state
            var linkedContainers = [t.ref()._container]; // parent env
            // env linked through dataset
            if (_.has(t.ref(), 'datasets')) {
                _.forOwn(t.ref().datasets, function(ds) {
                    if (_.has(ds.loc, 'hostname')) {
                        linkedContainers.push(cmp.props.pvd.byHostname(ds.loc.hostname));
                    }
                });
            }

            linkedContainers = _.uniq(linkedContainers);

            // title for containers
            outer.children.push(React.DOM.h3({}, t.name()));
            // list of containers
            outer.children.push(React.DOM.ul({children: [
                React.DOM.li({children: [
                'Comprises ' + linkedContainers.length + ' env(s), with hostnames:',
                React.DOM.ul({}, _.map(linkedContainers, function(d) {
                    return React.DOM.li({}, d.name());
                }))
            ]}),
            React.DOM.li({}, 'App path: ' + t.ref()._path)
            ]}));

            outer.children.push(React.DOM.h3({}, 'Active commit'));

            var commit = this.props.commits[t.ref().id.commit];
            var sha1line =  'sha1: ' + t.ref().id.commit.slice(0, 7);

            if (_.has(this.props.commitMeta, t.ref().id.commit) &&
                _.has(this.props.commitMeta[t.ref().id.commit], 'testState')) {
                sha1line += ' (' + this.props.commitMeta[t.ref().id.commit].testState + ')';
            }

            var items = [
                sha1line,
                commit.date,
                commit.author,
                '"' + commit.message + '"'
            ];

            outer.children.push(React.DOM.ul({
                children: items.map(function(d) {
                    return React.DOM.li({}, d);
                })
            }));
            return React.DOM.div(outer);
        }
    });

    var ControlBar = React.createClass({
        displayName: 'pipeviz-control',
        render: function() {
            var fc = this.props.filterChange;
            var boxes = this.props.filters.map(function(d) {
                return (React.createElement("input", {type: "checkbox", checked: d.selected, onChange: fc.bind(this, d.id)}, d.id));
            });

            return (
                React.createElement("div", {id: "controlbar"},
                                    "Filters: ", boxes,
                                    "Sort by: ", React.createElement("input", {type: "checkbox", checked: this.props.commitsort,
                                                                     onChange: this.props.csChange}, "commits")
                                   )
            );
        },
    });

    var App = React.createClass({
        displayName: 'pipeviz',
        getInitialState: function() {
            return {
                // FIXME having these in state is a bit fucked up...but consistency
                // in viz child class' state leaves us no choice for now
                anchorL: new Anchor(0, this.props.vizHeight/2),
                anchorR: new Anchor(this.props.vizWidth, this.props.vizHeight/2),
                gData: {},
                commits: [],
                commitsort: true,
                commitMeta: {},
                nodes: [],
                links: [],
                pvd: new PVD(),
                target: undefined,
                filters: Object.keys(this.filterFuncs).map(function(id) {
                    // TODO haha hardcoding
                    if (id === 'container') {
                        return {id: id, selected: false};
                    }
                    return {id: id, selected: true};
                })
            };
        },
        getDefaultProps: function() {
            return {
                vizWidth: window.innerWidth * 0.83,
                vizHeight: window.innerHeight
            };
        },
        filterChange: function(id) {
            var filters = this.state.filters.map(function(d) {
                return {
                    id: d.id,
                    selected: (d.id === id ? !d.selected : d.selected)
                };
            });

            this.setState({filters: filters});
        },
        toggleCommitSort: function() {
            this.setState({commitsort: !this.state.commitsort});
        },
        filterFuncs: {
            'container': function(node) {
                return !(node instanceof Container);
            },
            'process': function(node) {
                return !(node instanceof Process);
            },
            'logic': function(node) {
                return !(node instanceof LogicState);
            },
            'dataspace': function(node) {
                return !(node instanceof DataSpace);
            },
            'dataset': function(node) {
                return !(node instanceof DataSet);
            }
        },
        buildNodeFilter: function() {
            var check = _.filter(this.state.filters, function(d) {
                return d.selected;
            }).map(function(d) {
                return d.id;
            });

            var funcs = _.reduce(this.filterFuncs, function(accum, f, k) {
                if (_.contains(check, k)) {
                    accum.push(f);
                }
                return accum;
            }, []);

            // TODO just hacking this one on the end. super-hardcoding to our app
            funcs.push(function(node) {
                var filter = false;

                _.each(node.logicStates(), function(ls) {
                    if (ls.nick === 'ourapp') {
                        filter = true;
                        return false;
                    }
                });

                return filter;
            });

            if (funcs.length > 0) {
                // if any filter func returns false, we throw it out (logical OR)
                return function(node) {
                    var i;
                    for (i = 0; i < funcs.length; i++) {
                        if (!funcs[i](node)) {
                            return false;
                        }
                    }
                    return true;
                };
            }
            return false;
        },
        buildLinkFilter: function() {
            // TODO atm we have no direct link filtering, this just
            // filters links that are incident to filtered nodes
            var nf = this.buildNodeFilter();

            return nf ? function(link) {
                return nf(link.source) && nf(link.target);
            } : nf;
        },
        populatePVDFromJSON: function(pvd, containerData) {
            _.each(containerData, function(container) {
                pvd.attachContainer(new Container(container));
            });

            return pvd;
        },
        calculateCommitLinks: function(lgroups) {
            var g = new graphlib.Graph(),
            links = [],
            cmp = this,
            members = {};

            _.each(this.state.commits, function(cdatum, hash) {
                _.each(cdatum.parents, function(phash) {
                    g.setEdge(hash, phash);
                });
            });

            _.each(lgroups, function(lgroup) {
                var ls = lgroup.ref();
                if (ls.id && ls.id.commit && _.has(cmp.state.commits, ls.id.commit)) {
                    // FIXME this is the spot where we'd need to deal with multiple
                    // instances being on the same commit...only kinda doing it now
                    if (!_.has(members, ls.id.commit)) {
                        members[ls.id.commit] = [];
                    }
                    members[ls.id.commit].push({commit: ls.id.commit, obj: lgroup});
                }
            });

            // now traverse depth-first to figure out the overlaid edge structure
            var visited = [], // "black" list - vertices that have been visited
            path = [], // the current path of interstitial commits
            npath = [], // the current path, nodes only
            from, // head of the current exploration path
            v; // vertex (commit) currently being visited

            var walk = function(v) {
                // guaranteed acyclic, safe to skip grey/back-edge

                var pop_npath = false;
                // grab head of node path from stack
                from = npath[npath.length - 1];

                if (visited.indexOf(v) !== -1) {
                    // Vertex is black/visited; create link and return.
                    _.each(members[v], function(tgt) {
                        from.map(function(src) {
                            links.push({ source: src.obj, target: tgt.obj, path: path.slice(0) });
                        });
                    });
                    path = [];
                    return;
                }

                if (_.map(from, function(obj) { return obj.commit; }).indexOf(v) === -1) {
                    if (_.has(members, v)) {
                        // Found node point. Create a link
                        _.each(members[v], function(tgt) {
                            from.map(function(src) {
                                links.push({ source: src.obj, target: tgt.obj, path: path.slice(0) });
                            });
                        });

                        // Our exploration structure inherently guarantees a spanning
                        // tree, so we can safely discard historical path information ERR, NO IT DOESN'T
                        path = [];

                        // Push newly-found node point onto our npath, it's the new head
                        npath.push(members[v]);
                        // Correspondingly, indicate to pop the npath when exiting
                        pop_npath = true;
                    }
                    else {
                        // Not a node point and not self - push commit onto path
                        path.push(v);
                    }
                }

                // recursive call, the crux of this depth-first traversal
                g.successors(v).map(function(s) {
                    walk(s);
                });

                // Mark commit black/visited...but only if it's a member-associated
                // one. This trades CPU for memory, as it triggers a graph
                // re-traversal interstitial commits until one associated with an
                // instance is found. The alternative is keeping a map from ALL
                // interstitial commits to the eventual instance they arrive at...
                // and that's icky.
                if (_.has(members, v)) {
                    visited.push(v);
                }

                if (pop_npath) {
                    npath.pop();
                }
            };

            var stack = _.reduce(g.sources(), function(accum, commit) {
                // as long as we're in here, put the source anchor link in
                _.each(members[commit], function(member) {
                    links.push({ source: cmp.state.anchorL, target: member.obj });
                });

                // FIXME this assumes the sources of the commit graph we have happen to
                // align with commits we have in other logic states
                return _.has(members, commit) ? accum.concat(members[commit]) : accum;
            }, []);

            _.each(g.sinks(), function(commit) {
                _.each(members[commit], function(member) {
                    links.push({ source: member.obj, target: cmp.state.anchorR });
                });
            });

            // DF walk, working from source commit members
            while (stack.length !== 0) {
                v = stack.pop();
                npath.push(members[v.commit]);
                walk(v.commit);
            }

            return links;
        },
        targetNode: function(event) {
            this.setState({target: event});
        },
        render: function() {
            // FIXME can't afford to search the entire graph on every change, every
            // time in the long run
            var nf = this.buildNodeFilter();
            var nodes = this.state.pvd.findLogicalGroups();
            var graphData = [_.values(nodes), []];


            if (this.state.commitsort) {
                // FIXME wrong to change this state here like this, just making it work
                this.state.anchorL.x = 0;
                this.state.anchorL.y = this.props.vizHeight/2;
                this.state.anchorR.x = this.props.vizWidth;
                this.state.anchorR.y = this.props.vizHeight/2;

                graphData[0] = graphData[0].concat([this.state.anchorL, this.state.anchorR]);
                graphData[1] = graphData[1].concat(this.calculateCommitLinks(nodes));
            }

            return (
                React.createElement("div", {id: "pipeviz"},
                                    React.createElement(Viz, {width: this.props.vizWidth, height: this.props.vizHeight, nodes: graphData[0], links: graphData[1], target: this.targetNode, commitMeta: this.state.commitMeta}),
                                    React.createElement(InfoBar, {target: this.state.target, commits: this.state.commits, pvd: this.state.pvd, commitMeta: this.state.commitMeta})
                                   )
            );
        },
    });

    React.render(React.createElement(App, {gData: JSON.parse(document.getElementById("pipe-graph").innerHTML)}), document.body);
//});
