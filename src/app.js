var d3 = require('../bower_components/d3/d3'),
    queue = require('../bower_components/queue-async/queue'),
    _ = require('../bower_components/lodash/dist/lodash'),
    React = require('../bower_components/react/react.js');

var Container = require('./Container');
var LogicState = require('./LogicState');
var Process = require('./Process');
var DataSpace = require('./DataSpace.js');
var DataSet = require('./DataSet');

var force = d3.layout.force();

var processJson = function(err, res) {
    var allLinks = [],
        allNodes = [],
        containers = {};

    _.each(res, function(cnt) {
        var c = new Container(cnt);
        allNodes.push(c);

        _.each(c.logicStates(), function(v) {
            allNodes.push(v);
            allLinks.push({source: c, target: v});
        });

        _.each(c.processes(), function(v) {
            allNodes.push(v);
            allLinks.push({source: c, target: v});
        });

        _.each(c.dataSpaces(), function(v) {
            allNodes.push(v);
            allLinks.push({source: c, target: v});
        });

        _.each(c.dataSets(), function(v) {
            allNodes.push(v);
            allLinks.push({source: c, target: v});
        });

        // FIXME assumes hostname is uuid
        containers[cnt.hostname] = c;
    });

    // All hierarchical data is processed; second pass for referential.
    _.forOwn(containers, function(cv) {
        // find logic refs to data
        _.each(cv.logicStates(), function(l) {
            if (_.has(l, 'datasets')) {
                _.forOwn(l.datasets, function(dv, dk) {
                    var link = _.assign({source: l, name: dk}, dv);
                    var proc = false;
                    // check for hostname, else assume path
                    if (_.has(dv.loc, 'hostname')) {
                        if (_.has(containers, dv.loc.hostname)) {
                            proc = containers[dv.loc.hostname].findProcess(dv.loc);
                        }
                    } else {
                        proc = cv.findProcess(dv.loc);
                    }

                    if (proc) {
                        link.target = proc;
                        allLinks.push(link);
                    }
                });
            }
        });

        // next, link processes to logic states & data spaces
        _.each(cv.processes(), function(proc) {
            _.each(proc.logicStates(), function(ls) {
                allLinks.push({source: proc, target: ls});
            });

            _.each(proc.dataSpaces(), function(ds) {
                allLinks.push({source: proc, target: ds});
            });
        });

        _.each(cv.dataSpaces(), function(dg) {
            _.each(dg.dataSets(), function(ds) {
                // direct tree/parentage info
                allLinks.push({source: dg, target: ds});
                // add dataset genesis information
                if (ds.genesis !== "α") {
                    var link = {source: ds};
                    var ods = false;
                    if (_.has(ds.genesis, 'hostname')) {
                        if (_.has(containers, ds.genesis.hostname)) {
                            ods = containers[ds.genesis.hostname].findDataSet(ds.genesis);
                        }
                    } else {
                        ods = cv.findDataSet(ds.genesis);
                    }

                    if (ods) {
                        link.target = ods;
                        allLinks.push(link);
                    }
                }
            });
        });
    });

    return {links: allLinks, nodes: allNodes, containers: containers};
};

var graphRender = function(el, state) {
    var link = d3.select(el).selectAll('.link'),
    node = d3.select(el).selectAll('.node');

    link = link.data(state.links);
    node = node.data(state.nodes);

    link.enter().append('line')
    .attr('class', 'link');

    var nodeg = node.enter().append('g')
    .attr('class', function(d) {
        return 'node ' + d.vType();
    });

    nodeg.append('circle')
    .attr('x', 0)
    .attr('y', 0)
    .attr('r', function(d) {
        if (d instanceof Container) {
            return 45;
        }
        if (d instanceof LogicState) {
            return 30;
        }
        if (d instanceof DataSet) {
            return 30;
        }
        if (d instanceof DataSpace) {
            return 30;
        }
        if (d instanceof Process) {
            return 37;
        }
    });

    nodeg.append('text')
    .text(function(d) { return d.name(); });

    node.exit().remove();
    link.exit().remove();

    state.force.start();
    return false;
};

var App = React.createClass({ displayName: "Pipeviz",
    getInitialState: function() {
        return {
            force: d3.layout.force()
                .charge(-3000)
                .chargeDistance(250)
                .linkStrength(function(link) {
                    if (link.source instanceof Container) {
                        return 1;
                    }
                    return 0.5;
                }),
            links: [],
            nodes: [],
            containers: {}
        };
    },
    getDefaultProps: function() {
        return {
            width: window.innerWidth,
            height: window.innerHeight
        };
    },
    render: function() {
        return false;
    },
    componentWillMount: function() {
        var cmp = this;

        // load up all our data
        queue()
            .defer(d3.json, '/fixtures/ein/container.json', processJson)
            .await(function(err, res) {
                cmp.setState(res);
            });
    },
    componentDidMount: function() {
        var el = this.getDOMNode();
        d3.select(el).append('svg')
            .attr('width', this.props.width)
            .attr('height', this.props.height);

        graphRender(el, this.state);
    },
    componentDidUpdate: function() {
        return graphRender(this.getDOMNode(), this.state);
    },
    shouldComponentUpdate: function(nextProps, nextState) {
        // TODO this needs to just be the first check - do another
        // if this is true that sees if our links or nodes have changed
        return nextState.force.alpha <= 0;
    }
});

React.renderComponent(App(), document.body);
