var width = window.innerWidth,
    height = window.innerHeight,
    color = d3.scale.category20();

var svg = d3.select('body').append('svg')
    .attr('width', width)
    .attr('height', height);

var force = d3.layout.force()
    .charge(-3000)
    .size([width, height]);

d3.json('/fixtures/ein/container.json', function(err, res) {
    // Selectors identifying various graph components
    var sel = {
        'container': svg.selectAll('.node.container'),
        'logic': svg.selectAll('.node.logic'),
        'process': svg.selectAll('.node.process'),
        'dspace': svg.selectAll('.node.dspace'),
        'dset': svg.selectAll('.node.dset'),
        'links': svg.selectAll('.link.commit'),
        'anchors': svg.selectAll('.node.anchor'),
        'anchorlinks': svg.selectAll('.link.anchor')
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
        _.each(c.logicStates(), function(l) {
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
    });

    // Populate nodes & links into graph
    force.nodes(allNodes).links(allLinks);
    bindings.container = sel.container.data(containers)
        .enter().append('g')
        .attr('class', 'node container');

    bindings.container.append("circle")
        .attr("x", 0)
        .attr("y", 0)
        .attr("r", 35);

    force.start();
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

Container.prototype.findProcess = function(loc) {
    var found = false;
    if (loc.type == "unix") {
        f = function(proc) {
            return _.find(proc.listen, function(ingress) {
                if (ingress.type != 'unix') return false;
                return loc.path == ingress.path
            });
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
            });
        }
    }

    _.each(this.processes(), function(proc) {
        f(proc);
        if (found) return false;
    });

    return found;
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

DataSet.prototype.vType = function() {
    return 'dataset';
}

DataSet.prototype.space = function() {
    return this._space;
}

DataSet.prototype.name = function() {
    return this._name;
}

