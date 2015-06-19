var Viz = React.createClass({
    displayName: "pipeviz-graph",
    render: function() {
        return React.DOM.svg({
            className: "pipeviz",
            width: "100%",
            children: [React.DOM.g({
                id: 'commit-pipeline',
                children: [React.DOM.g({
                    id: 'commitview-edges'
                }), React.DOM.g({
                    id: 'commit-axis',
                    width: this.props.width,
                    height: 30,
                    transform: 'translate(0,' + (this.props.height - 30) + ')'
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

        // Vertices
        selections.vertices = selections.outerg.selectAll('.node')
            .data(props.vizdata.vertices, function(d) { return d.ref.id; });

        selections.vertices.exit().transition().remove(); // exit removes vertex
        // tons of stuff to do on enter
        selections.veg = selections.vertices.enter().append('g') // store the enter group and build it up
            .attr('class', function(d) { return 'node ' + d.ref.Typ(); })
            // so we don't transition from 0,0
            .attr('transform', function(d) { return 'translate(' + tf.x(d.x) + ',' + tf.y(d.y) + ')'; })
            // and start from invisible
            .style('opacity', 0);
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

        // now, update
        selections.vertices.transition() // update assigns the position via transform
            .attr('transform', function(d) { return 'translate(' + tf.x(d.x) + ',' + tf.y(d.y) + ')'; })
            .style('opacity', 1);

        // now work within the g for each vtx
        selections.vertices.select('circle').transition()
            .attr('r', function(d) { return d.ref.Typ() === "commit" ? tf.unit()*0.03 : tf.unit()*0.3; });

        // and the info text
        selections.nodetext = selections.vertices.select('text').transition();
        selections.nodetext.select('.vtx-label')
            .text(function(d) { return d.ref.propv("lgroup"); }); // set text value to data from lgroup
        selections.nodetext.select('.commit-subtext') // set the commit text on update
            .text(function(d) { return getCommit(props.graph, d.ref).propv("sha1").slice(0, 7); });

        // Links
        selections.links = selections.outerg.select('#commitview-edges').selectAll('.link')
            .data(props.vizdata.links, function(d) {
                return d[0].ref.id + '-' +  d[1].ref.id;
            });

        selections.links.exit().transition()
            .style('opacity', 0)
            .remove(); // exit removes line
        selections.links.enter().append('line') // enter appends a line
            .attr('class', 'link')
            .style('opacity', 0)
            // set all these initially so that we don't transition from 0,0,0,0
            .attr('x1', function(d) { return tf.x(d[0].x); })
            .attr('y1', function(d) { return tf.y(d[0].y); })
            .attr('x2', function(d) { return tf.x(d[1].x); })
            .attr('y2', function(d) { return tf.y(d[1].y); });

        selections.links.transition() // update sets the line's x and y positions
            .attr('x1', function(d) { return tf.x(d[0].x); })
            .attr('y1', function(d) { return tf.y(d[0].y); })
            .attr('x2', function(d) { return tf.x(d[1].x); })
            .attr('y2', function(d) { return tf.y(d[1].y); })
            .style('opacity', 1);

        // commit elision markers
        selections.elisions = selections.outerg.selectAll('.elision-bar')
            .data(_.map(props.vizdata.elranges, function(range) {
                // creates the same string as what's used in vizdata.xmap
                return range[0] + ' - ' + range[range.length - 1];
            }), _.identity);

        selections.elisions.exit().transition().remove(); // remove on exit
        selections.elisions.enter().append('line')
            .attr('class', 'elision-bar')
            .style('opacity', 0)
            // set all these initially so that we don't transition from 0,0,0,0
            .attr('x1', function(d) { return tf.x(props.vizdata.xmap[d]); })
            .attr('y1', 0)
            .attr('x2', function(d) { return tf.x(props.vizdata.xmap[d]); })
            .attr('y2', props.height - 30);

        selections.elisions.transition() // update sets x position
            .style('opacity', 1)
            .attr('x1', function(d) { return tf.x(props.vizdata.xmap[d]); })
            .attr('y1', 0)
            .attr('x2', function(d) { return tf.x(props.vizdata.xmap[d]); })
            .attr('y2', props.height - 30)
            .attr('stroke-width', tf.unit() * 0.007);

        // Commit distance axis
        var xlbls = _.map(_.pairs(props.vizdata.xmap).sort(function(a, b) { return a[1] - b[1]; }), _.head),
            xscale = d3.scale.ordinal()
                .domain(xlbls)
                .range(_.map(xlbls, function(label) {
                    return tf.x(props.vizdata.xmap[label]);
                })),
            xaxis = d3.svg.axis()
                .scale(xscale)
                .orient('bottom')
                .ticks(props.vizdata.ediam);

        d3.select('#commit-axis').transition()
            .call(xaxis);
            //.append('text')
                //.attr('transform', 'translate(' + tf.x(0) + ',-5)')
                //.attr('text-anchor', 'start')
                //.text('distance to root');
    },
});

var VizPrep = React.createClass({
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
