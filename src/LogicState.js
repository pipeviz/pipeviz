var _ = require('../bower_components/lodash/dist/lodash');

var _sharedId = require('./_sharedId');

function LogicState(obj, path, container) {
    _.assign(this, obj);
    this._path = path;
    this._container = container;
    this._nextId();
}

LogicState.prototype = new _sharedId();

LogicState.prototype.vType = function() {
    return 'logic';
};

LogicState.prototype.name = function() {
    if (_.has(this, 'nick')) {
        return this.nick;
    }

    var p = this._path.split('/');
    return p[p.length-1];
};

module.exports = LogicState;
