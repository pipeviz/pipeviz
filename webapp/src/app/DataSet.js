var _ = require('bower_components/lodash/dist/lodash');

var _sharedId = require('./_sharedId');

function DataSet(obj, name, space) {
    _.assign(this, obj);
    this._name = name;
    this._space = space;
    this._nextId();
}

DataSet.prototype = new _sharedId();

DataSet.prototype.vType = function() {
    return 'dataset';
};

DataSet.prototype.space = function() {
    return this._space;
};

DataSet.prototype.name = function() {
    return this._name;
};

module.exports = DataSet;
