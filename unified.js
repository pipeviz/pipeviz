var width = window.innerWidth,
    height = window.innerHeight,
    color = d3.scale.category20(),
    allNodes = [],
    allLinks = [];

var svg = d3.select('body').append('svg')
    .attr('width', width)
    .attr('height', height);

var force = d3.layout.force()
    .charge(-3000)
    .size([width, height]);

var containers = {};

// Selectors identifying various graph components
var sel = {
    'instances': svg.selectAll('.node.container'),
    'commits': svg.selectAll('.node.logic-state'),
    'links': svg.selectAll('.link.commit'),
    'anchors': svg.selectAll('.node.anchor'),
    'anchorlinks': svg.selectAll('.link.anchor')
}


d3.json('/fixtures/ein/containers.json', function(err, res) {
    _.each(res, function(cnt) {
        c = new Container(cnt);
        allNodes.push(c);

        _.each(c.logicStates(), function(v) {
            allNodes.push(v);
        });

        _.each(c.processes(), function(v) {
            allNodes.push(v)
        })

        _.each(c.dataSpaces(), function(v) {
            allNodes.push(v)
        })

        _.each(c.dataSets(), function(v) {
            allNodes.push(v)
        })

        // FIXME assumes hostname is uuid
        containers[cnt.hostname] = c;
    });

    // Drop all nodes into the graph
    force.nodes(allNodes);

});

function Container(obj) {
    this.hostname = obj.hostname;
    this.ipv4 = obj.ipv4;
    this.type = obj.type;

    this._logics = _.has(obj, 'logic states') ? _.mapValues(obj['logic states'], function(l) { return new LogicState(l) }) : {};
    this._processes = _.has(obj, 'processes') ? obj['processes'] : {};
    this._dataSpaces = _.has(obj, 'data spaces') ? _.mapValues(obj['data spaces'], function(space) {
        return _.mapValues(space, function(dc) {
            return new DataSet(dc, space);
        });
    }) : {};
}

Container.prototype.vType = function() {
    return 'container';
}

Container.prototype.logicStates = function() {
    return this._logics;
}

Container.prototype.processes = function() {
    return this._processes;
}

Container.prototype.dataSpaces = function() {
    return this._dataSpaces;
}

Container.prototype.dataSets = function() {
    return _.flatten(_.each(this.dataSpaces(), function(space) {
        return _.values(space)
    }))
}

function LogicState(obj) {
    _.assign(this, obj);
}

LogicState.prototype.vType = function() {
    return 'logic';
}

function DataSet(obj, name, space) {
    _.assign(this, obj);
    this._name = name;
    this._space = space;
}

Dataset.prototype.vType = function() {
    return 'dataset';
}

Dataset.prototype.space = function() {
    return this._space;
}

Dataset.prototype.name = function() {
    return this._name;
}

