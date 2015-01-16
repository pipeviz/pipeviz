var d3 = require('../bower_components/d3/d3'),
    queue = require('../bower_components/queue-async/queue'),
    _ = require('../bower_components/lodash/dist/lodash'),
    React = require('../bower_components/react/react.js');

var Container = require('./Container'),
    LogicState = require('./LogicState'),
    Process = require('./Process'),
    DataSpace = require('./DataSpace.js'),
    DataSet = require('./DataSet'),
    PVD = require('./PVD');

var graphRender = function(el, state, props) {
    var link = d3.select(el).selectAll('.link')
            .data(state.force.links()),
        node = d3.select(el).selectAll('.node')
            .data(state.force.nodes());

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
        })
        .on('click', props.target);

    nodeg.append('text')
        .text(function(d) { return d.name(); });

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
                .charge(-3000)
                .chargeDistance(250)
                .size([this.props.width, this.props.height])
                .linkStrength(function(link) {
                    if (link.source instanceof Container) {
                        return 1;
                    }
                    return 0.5;
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
        // TODO probably a suboptimal way to do this compare...?
        if (nextProps.nodes !== this.state.force.nodes()) {
            return true;
        }
        if (nextProps.links !== this.state.force.links()) {
            return true;
        }
        return false;
    }
});

var InfoBar = React.createClass({
    displayName: 'pipeviz-info',
    render: function() {
        var t = this.props.target;
        var outer = {
            id: "infobar",
            children: []
        };

        if (typeof t !== 'object') {
            outer.children = [React.DOM.p({}, "nothing selected")];
            // drop out early for the empty case
            return React.DOM.div(outer);
        }

        var items = [],
            title;
        if (t instanceof Container) {
            title = 'Container';
            items = [
                'hostname: ' + t.hostname,
                'type: ' + t.type,
                Object.keys(t._logics).length + ' logic state(s)',
                t._processes.length + ' running process(es)',
                t.dataSets().length + ' data set(s)'
            ];

            if (t.hasOwnProperty('ipv4')) {
                items.push('ipv4: ' + t.ipv4.toString());
            }
        } else if (t instanceof LogicState) {
            title = 'Logic';
            items = [
                'path: ' + t._path,
                'type: ' + t.type,
            ];

            if (t.hasOwnProperty('nick')) {
                items.push('nick: ' + t.nick);
            }

            if (t.hasOwnProperty('id')) {
                if (t.id.hasOwnProperty('commit')) {
                    items.push('id by commit: ' + t.id.commit.slice(0, 7));
                } else {
                    items.push('id by version: ' + t.id.version);
                }
            } else {
                // FIXME stupid libraries
                items.push('id by version: ' + t.version);
            }

            if (t.hasOwnProperty('datasets')) {
                items.push('connected to ' + Object.keys(t.datasets).length + ' datasets');
            }
        } else if (t instanceof DataSet) {
            title = 'Data set';
            items = [
                "set name: " + t.name(),
                "in space: " + t.space().name()
            ];

            if (t.genesis === "α") {
                items.push('has no upstream (α)');
            } else {
                items.push('created from ' + t.genesis.hostname + ':' + t.genesis['data space'] + ':' + t.genesis['data set'] + ' at ' + new Date(t.genesis.at * 1000).toLocaleString());
            }
        } else if (t instanceof DataSpace) {
            title = 'Data space';
            items = [
                'space identifier: ' + t.name(),
                'sets contained: ' + Object.keys(t.dataSets()).toString()
            ];
        } else if (t instanceof Process) {
            title = 'Process';
            items = [
                'hey its a proc'
            ];
        }

        outer.children.push(React.DOM.h3({}, title));
        outer.children.push(React.DOM.ul({
            children: items.map(function(d) {
                return React.DOM.li({}, d);
            })
        }));
        return React.DOM.div(outer);
    }
});

var App = React.createClass({
    displayName: 'pipeviz',
    getInitialState: function() {
        return {
            pvd: new PVD(),
            target: undefined
        };
    },
    populatePVDFromJSON: function(pvd, containerData) {
        _.each(containerData, function(container) {
            pvd.attachContainer(new Container(container));
        });

        return pvd;
    },
    targetNode: function(event) {
        this.setState({target: event});
    },
    render: function() {
        var graphData = this.state.pvd.nodesAndLinks();
        return (
            <div id="pipeviz">
                <Viz width={window.innerWidth * 0.83} height={window.innerHeight} nodes={graphData[0]} links={graphData[1]} target={this.targetNode}/>
                <InfoBar target={this.state.target}/>
            </div>
        );
    },
    componentDidMount: function() {
        var cmp = this;

        // TODO this whole retrieval/population pattern will all change
        d3.json('fixtures/ein/container.json', function(err, res) {
            if (err) {
                return;
            }

            cmp.setState({pvd: cmp.populatePVDFromJSON(cmp.state.pvd, res)});
        });
    }
});

React.renderComponent(App(), document.body);
