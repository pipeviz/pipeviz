var width = window.innerWidth,
    height = window.innerHeight,
    color = d3.scale.category20();

var svg = d3.select('body').append('svg')
    .attr('width', width)
    .attr('height', height);

var force = d3.layout.force()
    .charge(-3000)
    .chargeDistance(250)
    .linkStrength(function(link) {
        if (link.source instanceof Container) {
            return 1;
        }
        return 0.5;
    })
    .size([width, height]);

d3.json('/fixtures/ein/container.json', function(err, res) {
    // Selectors identifying various graph components
    var sel = {
        'nodes': svg.selectAll('.node'),
        'container': svg.selectAll('.node.container'),
        'logic': svg.selectAll('.node.logic'),
        'process': svg.selectAll('.node.process'),
        'dataspace': svg.selectAll('.node.dataspace'),
        'dataset': svg.selectAll('.node.dataset'),
        'links': svg.selectAll('.link')
    },
    containers = {},
    allNodes = [],
    allLinks = [],
    bindings = {};

    _.each(res, function(cnt) {
        c = new Container(cnt);
        allNodes.push(c);

        _.each(c.logicStates(), function(v) {
            allNodes.push(v);
            allLinks.push({source: c, target: v})
        });

        _.each(c.processes(), function(v) {
            allNodes.push(v)
            allLinks.push({source: c, target: v})
        });

        _.each(c.dataSpaces(), function(v) {
            allNodes.push(v)
            allLinks.push({source: c, target: v})
        });

        _.each(c.dataSets(), function(v) {
            allNodes.push(v)
            allLinks.push({source: c, target: v})
        });

        // FIXME assumes hostname is uuid
        containers[cnt.hostname] = c;
    });

    // All hierarchical data is processed; second pass for referential.
    _.forOwn(containers, function(cv, ck) {
        // find logic refs to data
        _.each(cv.logicStates(), function(l) {
            if (_.has(l, 'datasets')) {
                _.forOwn(l.datasets, function(dv, dk) {
                    var link = _.assign({source: l, name: dk}, dv);
                    var proc = false;
                    // check for hostname, else assume path
                    if (_.has(dv.loc, 'hostname')) {
                        if (_.has(containers[dv.loc.hostname])) {
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

        // next, link processes to logic states
        _.each(cv.processes(), function(proc) {
            _.each(proc.logicStates(), function(ls) {
                allLinks.push({source: proc, target: ls});
            })
        })
    });

    // Populate nodes & links into graph
    force.nodes(allNodes).links(allLinks);

    bindings.link = sel.links.data(allLinks)
        .enter().append('line')
        .attr('class', 'link');

    bindings.node = sel.nodes.data(allNodes)
        .enter().append('g')
        .attr('class', function(d) {
            return 'node ' + d.vType();
        });

    bindings.node.append('circle')
        .attr('x', 0)
        .attr('y', 0)
        .attr('r', function(d) {
            if (d instanceof Container) {
                return 45;
            } else if (d instanceof LogicState) {
                return 30;
            } else if (d instanceof DataSet) {
                return 30;
            } else if (d instanceof DataSpace) {
                return 30;
            } else if (d instanceof Process) {
                return 37;
            }
        });

    bindings.node.append('text')
        .text(function(d) { return d.name() });

    force.on('tick', function() {
        bindings.link.attr("x1", function(d) { return d.source.x; })
        .attr("y1", function(d) { return d.source.y; })
        .attr("x2", function(d) { return d.target.x; })
        .attr("y2", function(d) { return d.target.y; });

        bindings.node.attr('transform', function(d) { return 'translate(' + d.x + ',' + d.y + ')'});
    });

    force.start();
});

function Container(obj) {
    this.hostname = obj.hostname;
    this.ipv4 = obj.ipv4;
    this.type = obj.type;

    var that = this;
    this._logics = _.has(obj, 'logic states') ? _.mapValues(obj['logic states'], function(l, path) { return new LogicState(l, path, that) }) : {};
    this._processes = _.has(obj, 'processes') ? _.map(obj['processes'], function(p) { return new Process(p, that) }) : {};
    this._dataSpaces = _.has(obj, 'data spaces') ? _.mapValues(obj['data spaces'], function(space, id) {
        return new DataSpace(_.mapValues(space, function(dc, set) {
            return new DataSet(dc, set, space);
        }), id, that);
    }) : {};
}

Container.prototype.vType = function() {
    return 'container';
}

Container.prototype.name = function() {
    return this.hostname;
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
    return _.flatten(_.map(this.dataSpaces(), function(space) {
        return _.values(space.dataSets())
    }));
}

Container.prototype.findProcess = function(loc) {
    var f, found = false;
    if (loc.type == "unix") {
        f = function(proc) {
            return _.find(proc.listen, function(ingress) {
                if (ingress.type != 'unix') return false;
                return loc.path == ingress.path
            }) ? proc : false;
        }
    } else {
        f = function(proc) {
            return _.find(proc.listen, function(ingress) {
                if (ingress.type != 'net' && ingress.type != 'port') return false;
                if (loc.port != ingress.number) return false;
                if (_.isString(ingress.proto)) {
                    return loc.proto == ingress.proto;
                } else {
                    return _.contains(ingress.proto, loc.proto);
                }
            }) ? proc : false;
        }
    }

    _.each(this.processes(), function(p) {
        found = f(p);
        if (found) return false;
    });

    return found;
}

function LogicState(obj, path, container) {
    _.assign(this, obj);
    this._path = path;
    this._container = container;
}

LogicState.prototype.vType = function() {
    return 'logic';
}

LogicState.prototype.name = function() {
    if (_.has(this, 'nick')) {
        return this.nick;
    } else {
        p = this._path.split('/');
        return p[p.length-1];
    }
}

function Process(obj, container) {
    _.assign(this, obj);
    this._container = container;
}

Process.prototype.vType = function() {
    return 'process';
}

Process.prototype.name = function() {
    return _.reduce(this.logicStates(), function(accum, ls) {
        accum.push(ls.name());
        return accum;
    }, []).join('/');
}

Process.prototype.logicStates = function() {
    return _.pick(this._container.logicStates(), this['logic states'])
}

function DataSpace(sets, name, container) {
    this._sets = sets;
    this._name = name;
    this._container = container;
}

DataSpace.prototype.vType = function() {
    return 'dataspace';
}

DataSpace.prototype.name = function() {
    return this._name;
}

DataSpace.prototype.dataSets = function() {
    return this._sets;
}

function DataSet(obj, name, space) {
    _.assign(this, obj);
    this._name = name;
    this._space = space;
}

DataSet.prototype.vType = function() {
    return 'dataset';
}

DataSet.prototype.space = function() {
    return this._space;
}

DataSet.prototype.name = function() {
    return this._name;
}

