var _sharedId = require('./_sharedId');

function Anchor() {
    this.fixed = true; // tells d3 not to move it
    this.x = 0;
    this.y = 0;
    this._nextId();
}

Anchor.prototype = new _sharedId();

Anchor.prototype.vType = function() {
    return 'sort-anchor';
};

Anchor.prototype.name = function() {
    return '';
}

module.exports = Anchor;
