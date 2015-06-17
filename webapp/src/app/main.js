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
            children: [React.DOM.g({
                id: 'commit-pipeline',
                children: [React.DOM.g({
                    id: 'commitview-edges'
                })]
            })]
        });
    },
    shouldComponentUpdate: function(nextProps, prevProps) {
        return nextProps.vizdata !== undefined;
    },
    componentDidUpdate: function() {
        // x-coordinate space is the elided diameter as a factor of viewport width
        var selections = {},
            props = this.props,
            tf = createTransforms(props.width, props.height - 30, props.vizdata.ediam, props.vizdata.branches.length);

        // Outer g first
        selections.outerg = d3.select(this.getDOMNode()).select('#commit-pipeline');

        // Now links
        selections.links = selections.outerg.select('#commitview-edges').selectAll('.link')
            .data(props.vizdata.links, function(d) {
                return d[0].ref.id + '-' +  d[1].ref.id;
            });

        selections.links.exit().remove(); // exit removes line
        selections.links.enter().append('line')
            .attr('class', 'link'); // enter appends a line
        selections.links // update sets the line's x and y positions
            .attr('x1', function(d) { return tf.x(d[0].x); })
            .attr('y1', function(d) { return tf.y(d[0].y); })
            .attr('x2', function(d) { return tf.x(d[1].x); })
            .attr('y2', function(d) { return tf.y(d[1].y); });

        selections.vertices = selections.outerg.selectAll('.node')
            .data(props.vizdata.vertices, function(d) { return d.ref.id; });

        selections.vertices.exit().remove(); // exit removes vertex
        selections.veg = selections.vertices.enter().append('g') // store the enter group and build it up
            .attr('class', function(d) { return 'node ' + d.ref.Typ(); });
        selections.veg.append('circle');
        selections.nte = selections.veg.append('text');
        selections.nte.append('tspan') // add vertex label tspan on enter
            .attr('class', 'vtx-label');
        selections.nte.append('tspan') // add commit info tspan on enter and position it
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


        selections.vertices // update assigns the position via transform
            .attr('transform', function(d) { return 'translate(' + tf.x(d.x) + ',' + tf.y(d.y) + ')'; });

        // now work within the g for each vtx
        selections.vertices.select('circle')
            .attr('r', function(d) { return d.ref.Typ() === "commit" ? tf.unit()*0.03 : tf.unit()*0.3; });
            //.attr('cx', function(d) { return tf.x(d.x); })
            //.attr('cy', function(d) { return tf.y(d.y); })

        // and the info text
        selections.nodetext = selections.vertices.select('text');
        selections.nodetext.select('.vtx-label')
            .text(function(d) { return d.ref.propv("lgroup"); }); // set text value to data from lgroup
        selections.nodetext.select('.commit-subtext') // set the commit text on update
            .text(function(d) { return getCommit(props.graph, d.ref).propv("sha1").slice(0, 7); });

        // Axes last, always on top
        var xposmap = _.uniq(
                _.map(props.vizdata.vertices, function(d) { return d.depth; })
                .sort(function(a, b) { return a - b; }));
            xscale = d3.scale.ordinal()
                .domain(xposmap)
                .range(_.map(xposmap, function(orig, x) { return tf.x(x); }));

        var xaxis = d3.svg.axis()
                .scale(xscale)
                .orient('bottom')
                .ticks(props.vizdata.ediam);

        // TODO just remove and rerender each time, for now
        d3.select(this.getDOMNode()).select('g.commit-axis').remove();
        selections.axes = d3.select(this.getDOMNode()).append('g')
            .attr('class', 'commit-axis')
            .attr('width', props.width)
            .attr('height', 30)
            .append('g')
                .attr('transform', 'translate(0,' + (props.height - 30) + ')')
                .call(xaxis)
                .append('text')
                    .attr('transform', 'translate(' + tf.x(0) + ',-5)')
                    .attr('text-anchor', 'start')
                    .text('distance to root');
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
