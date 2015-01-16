var _objcounter = 0;

function _sharedId() {}

_sharedId.prototype.objid = function() {
    return this._objid;
};

_sharedId.prototype._nextId = function() {
    this._objid = _objcounter++;
};

module.exports = _sharedId;
