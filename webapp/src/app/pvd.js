function pvVertex(obj) {
    _.assign(this, obj)
}

pvVertex.prototype.isVertex = function() {
    return true;
};

function pvVertex(obj) {
    _.assign(this, obj)
}

pvVertex.prototype.isVertex = function() {
    return true;
};

function pvEdge(obj) {
    _.assign(this, obj)
}

pvEdge.prototype.isEdge = function() {
    return false;
};

// this pretty much mirrors what we have serverside...for now
// pipeviz datastore
function pvd() {
    // contains all objects, vertices and edges, keyed by objid
    this._objects = {};
}

pvd.prototype.get = function(id) {
    return this._objects[id];
};

pvd.prototype.verticesWithType = function(typ) {
    _.filter(this._objects, function(d) {
        return  d.isVertex() && d.vertex.type === typ;
    });
};
