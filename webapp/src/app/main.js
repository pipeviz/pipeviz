var Viz = React.createClass({
    displayName: "pipeviz-graph",
    render: function() {
        return React.DOM.svg({
            className: "pipeviz",
            width: "83%",
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
            tf = createTransforms(props.width, props.height - 30, props.vizdata.ediam, props.vizdata.segments.length, props.opts.revx.flag);

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
            .style('opacity', 0)
            // attach click handler for infobar
            .on('click', props.selected);
        selections.veg.append('circle');
        selections.nte = selections.veg.append('text');
        selections.nte.append('tspan') // add vertex label tspan on enter
            .attr('class', 'vtx-label');
        selections.nte.append('tspan') // add commit info tspan on enter and position it
            .attr('dy', "1.4em")
            .attr('x', 0)
            .attr('class', function(d) {
                if (d.ref.Typ() !== "logic-state") {
                    return "";
                }
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
            opts: {},
        };
    },
    shouldComponentUpdate: function(nextProps) {
        // In the graph object, state is invariant with respect to the message id.
        return nextProps.graph.mid !== this.props.graph.mid ||
            JSON.stringify(this.props.opts) !== JSON.stringify(nextProps.opts);
    },
    render: function() {
        return React.createElement(Viz, {
            width: this.props.width,
            height: this.props.height,
            graph: this.props.graph,
            vizdata: extractVizGraph(this.props.graph, this.props.focalRepo, this.props.opts.noelide.flag),
            opts: this.props.opts,
            selected: this.props.selected,
        });
    },
});

var InfoBar = React.createClass({
    displayName: 'pipeviz-info',
    render: function() {
        var t = this.props.selected,
            cmp = this,
            pvg = this.props.pvg;

        var outer = {
            id: "infobar",
            children: []
        };

        if (_.isUndefined(t)) {
            outer.children = [React.DOM.p({}, "nothing selected")];
            // drop out early for the empty case
            return React.DOM.div(outer);
        }

        // TODO right now we only have two possibilities - commit or logic-state - but this will need to become its WHOLE own subsystem
        if (t.vertex.type === "logic-state") {
            // First, pick a title. Start with the nick
            var infotitle = "Instance of ",
                listitems = [];
            if (!_.isUndefined(t.propv('nick'))) {
                // TODO nick is not great since it's actually set per logic-state
                infotitle += "'" + t.propv('nick') + "'";
                // since we're not showing the repo addr in the title, put it in here
                listitems.push(React.DOM.li({}, "From repository " + getRepositoryName(pvg, t)));
            } else {
                // No nick, so grab the repo addr
                infotitle += "'" + getRepositoryName(pvg, t) + "'";
            }
            outer.children.push(React.DOM.h3({}, infotitle));

            var env = getEnvironment(pvg, t),
                commit = getCommit(pvg, t);

            listitems.push(React.DOM.li({}, "Located on " + getEnvName(env) + " at " + t.propv("path")));
            listitems.push(React.DOM.li({}, "Commit info:"));
            listitems.push(React.DOM.ul({children: [
                React.DOM.li({}, "Sha1: " + commit.propv("sha1")),
                React.DOM.li({}, "Author: " + commit.propv("author")),
                React.DOM.li({}, "Date: " + commit.propv("date")),
                React.DOM.li({}, "Subject: " + commit.propv("subject")),
            ]}));

            outer.children.push(React.DOM.ul({children: listitems}));
        } else { // can only be a commit, for now
            outer.children.push(React.DOM.h3({}, "Commit from " + getRepositoryName(pvg, t)));
            outer.children.push(React.DOM.ul({children: [
                React.DOM.li({}, "Sha1: " + t.propv("sha1")),
                React.DOM.li({}, "Author: " + t.propv("author")),
                React.DOM.li({}, "Date: " + t.propv("date")),
                React.DOM.li({}, "Subject: " + t.propv("subject")),
            ]}));
        }

        return React.DOM.div(outer);
    }
});

var ControlBar = React.createClass({
    displayName: 'pipeviz-control',
    render: function() {
        var oc = this.props.changeOpts;
            var boxes = _.map(this.props.opts, function(v, opt) {
            return (React.createElement("input", {
                key: opt,
                type: "checkbox",
                checked: v.flag,
                onChange: oc.bind(this, opt, v)
            }, v.label));
        });

        return (
            React.createElement("div", {id: "controlbar"},
                "Options: ", boxes
            )
        );
    },
});

var App = React.createClass({
    dispayName: "pipeviz",
    getInitialState: function() {
        return {
            selected: undefined,
            opts: {
                revx: {label: "Reverse x positions", flag: false},
                noelide: {label: "No commit elision", flag: false},
            },
        };
    },
    changeOpts: function(opt, v) {
        v.flag = !v.flag;
        this.setState({opts: _.merge(this.state.opts, _.zipObject([[opt, v]]))});
    },
    setSelected: function(tgt) {
        this.setState({selected: tgt.ref});
    },
    getDefaultProps: function() {
        return {
            // TODO uggghhh lol hardcoding
            vizWidth: window.innerWidth * 0.83,
            vizHeight: window.innerHeight - 30,
            graph: pvGraph({id: 0, vertices: []}),
        };
    },
    render: function() {
        return React.createElement("div", {id: "pipeviz"},
            React.createElement(ControlBar, {opts: this.state.opts, changeOpts: this.changeOpts}),
            React.createElement(VizPrep, {
                width: this.props.vizWidth,
                height: this.props.vizHeight,
                graph: this.props.graph,
                focalRepo: vizExtractor.mostCommonRepo(this.props.graph),
                opts: JSON.parse(JSON.stringify(this.state.opts)),
                selected: this.setSelected,
            }),
            React.createElement(InfoBar, {
                selected: this.state.selected,
                pvg: this.props.graph,
            })
        );
    },
});

var e = React.render(React.createElement(App), document.body);
var genesis = new WebSocket("ws://" + window.location.hostname + ":" + window.location.port + "/sock");
var lastg;
genesis.onmessage = function(m) {
    lastg = pvGraph(JSON.parse(m.data));
    e.setProps({graph: lastg});
};
