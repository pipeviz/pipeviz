var _ = require('../bower_components/lodash/dist/lodash');

function DataSpace(name, container) {
    this._sets = {}; // FIXME ugh circular
    this._name = name;
    this._container = container;
}

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
