var _ = require('bower_components/lodash/dist/lodash');

var _sharedId = require('./_sharedId');

function DataSpace(name, container) {
    this._sets = {}; // FIXME ugh circular
    this._name = name;
    this._container = container;
    this._nextId();
}

DataSpace.prototype = new _sharedId();

DataSpace.prototype.vType = function() {
    return 'dataspace';
};

DataSpace.prototype.name = function() {
    return this._name;
};

DataSpace.prototype.dataSets = function() {
    return this._sets;
};

module.exports = DataSpace;
