// just so my syntastic complains less
var React = React || {},
    _ = _ || {},
    d3 = d3 || {};

var data = new pvGraph(JSON.parse(document.getElementById("pipe-graph").innerHTML));

var Viz = React.createClass({
    displayName: "pipeviz-graph",
    getInitialState: function() {
        return {
            force: d3.layout.force()
            .charge(-4000)
            .chargeDistance(250)
            .size([this.props.width, this.props.height])
            .linkStrength(function(link) {
                if (link.source.Typ() == "environment") {
                    return 0.5;
                }
                if (link.source.Typ() === "anchor" || link.target.Typ() === "anchor") {
                    return 1;
                }
                return 0.3;
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
            vertices: {},
            edges: {},
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

        return this.graphRender(this.getDOMNode(), this.state, this.props);
    },
    graphRender: function(el, state, props) {
        var link = d3.select(el).selectAll('.link')
            .data(props.links, function(d) { return d.source.objid() + '-' + d.target.objid(); }),
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
            return 'node ' + d.vType();
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
        nodetext.append('tspan').text(function(d) { return d.name(); });
        nodetext.append('tspan').text(function(d) {
            // FIXME omg terrible
            if (d.Typ() !== "logic-state") {
                return '';
            }

            return d.ref().id.commit.slice(0, 7);
        })
        .attr('dy', "1.4em")
        .attr('x', 0)
        .attr('class', function(d) {
            if (d.Typ() !== "logic-state") {
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
    }
});

var VizPrep = React.createClass({
    getInitialState: function() {
        return {
            anchorL: new Anchor("L", 0, this.props.height/2),
            anchorR: new Anchor("R", this.props.width, this.props.height/2),
        };
    },
    shouldComponentUpdate: function(nextProps, nextState) {
        // In the graph object, state is invariant with respect to the message id.
        return nextProps.graph.mid !== this.props.graph.mid;
    },
    render: function() {
        return React.createElement(Viz, {width: this.props.width, height: this.props.height, graph: this.props.graph})
    },
});

var App = React.createClass({
    dispayName: "pipeviz",
    getInitialState: function() {
        return {
            graph: {},
        };
    },
    getDefaultProps: function() {
        return {
            vizWidth: window.innerWidth * 0.83,
            //vizWidth: window.innerWidth,
            vizHeight: window.innerHeight,
        };
    },
    render: function() {
        return React.createElement("div", {id: "pipeviz"},
                   React.createElement(VizPrep, {width: this.props.vizWidth, height: this.props.vizHeight, graph: this.state.graph})
              );
    },
});

React.render(React.createElement(App, {graph: data}), document.body)
