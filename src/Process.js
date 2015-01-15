var _ = require('../bower_components/lodash/dist/lodash');

function Process(obj, container) {
    _.assign(this, obj);
    this._container = container;
}

Process.prototype.vType = function() {
    return 'process';
};

Process.prototype.name = function() {
    return _.reduce(this.logicStates(), function(accum, ls) {
        accum.push(ls.name());
        return accum;
    }, []).join('/');
};

Process.prototype.logicStates = function() {
    return _.pick(this._container.logicStates(), this['logic states']);
};

Process.prototype.dataSpaces = function() {
    return _.pick(this._container.dataSpaces(), this['data spaces']);
};

module.exports = Process;
