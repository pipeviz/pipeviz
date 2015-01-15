var _ = require('../bower_components/lodash/dist/lodash');

function DataSet(obj, name, space) {
    _.assign(this, obj);
    this._name = name;
    this._space = space;
}

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
