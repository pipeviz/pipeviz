var _ = require('bower_components/lodash/dist/lodash');

function LGroup(name, obj) {
    this._name = name;
    this._obj = obj;
}

LGroup.prototype.vType = function() {
    return 'lgroup';
};

LGroup.prototype.name = function() {
    return this._name;
};

LGroup.prototype.ref = function() {
    return this._obj;
};

LGroup.prototype.objid = function() {
    return this._obj.objid();
};

module.exports = LGroup;
