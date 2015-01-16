var _ = require('../bower_components/lodash/dist/lodash');

var LogicState = require('./LogicState');
var DataSpace = require('./DataSpace');
var DataSet = require('./DataSet');
var Process = require('./Process');

function Container(obj) {
    this.hostname = obj.hostname;
    this.type = obj.type;
    if (obj.ipv4 !== undefined) {
        this.ipv4 = obj.ipv4;
    }

    var that = this;
    this._dataSpaces = _.has(obj, 'data spaces') ? _.mapValues(obj['data spaces'], function(space, id) {
        var ds = new DataSpace(id, that);
        ds._sets = _.mapValues(space, function(dc, set) {
            return new DataSet(dc, set, ds);
        });

        return ds;
    }) : {};
    this._logics = _.has(obj, 'logic states') ? _.mapValues(obj['logic states'], function(l, path) { return new LogicState(l, path, that) }) : {};
    this._processes = _.has(obj, 'processes') ? _.map(obj.processes, function(p) { return new Process(p, that) }) : {};
}

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
    }
}

Container.prototype.findProcess = function(loc) {
    var f, found = false;
    if (loc.type === "unix") {
        f = function(proc) {
            return _.find(proc.listen, function(ingress) {
                if (ingress.type !== 'unix') return false;
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

module.exports = Container;
