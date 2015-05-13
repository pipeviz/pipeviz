function pvVertex(obj, g) {
    this.id = obj.id;
    this.vertex = obj.vertex;

    // strip out the actual edge objects
    this.outEdges = _.map(obj.outEdges, function(d) { return d.id; });
    this._g = g;
}

pvVertex.prototype.isVertex = function() {
    return true;
};

pvVertex.prototype.Typ = function() {
    return this.vertex.type;
};

function pvEdge(obj, g) {
    _.assign(this, obj);
    this._g = g;
}

pvEdge.prototype.isVertex = function() {
    return false;
};

pvEdge.prototype.Typ = function() {
    return this.etype;
};

// TODO this pretty much mirrors what we have serverside...for now. ugh.
// pipeviz datastore
function pvGraph(gdata) {
    // contains all objects, vertices and edges, keyed by objid
    this._objects = {};

    var that = this;
    var o = this._objects;
    _.each(gdata, function(d) {
        // capture vertex
        o[d.id] = new pvVertex(d, that);
        // and its out-edges
        _.each(d.outEdges, function(d2) { o[d2.id] = new pvEdge(d2, that); });
    });
}

pvGraph.prototype.get = function(id) {
    return this._objects[id];
};

pvGraph.prototype.verticesWithType = function(typ) {
    return _.filter(this._objects, function(d) {
        return  d.isVertex() && d.vertex.type === typ;
    });
};

var filters = {
    vertices: function(d) {
        return d.isVertex();
    },
    edges: function(d) {
        return !d.isVertex();
    }
};

var isType = function(typ) {
    return function(d) {
        return typ === d.Typ();
    };
};

// Returns a graphlib.Graph object representing the graph(s) of all known
// commit objects.
pvGraph.prototype.commitGraph = function() {
    var g = new graphlib.Graph();
    var that = this;

    //_.each(_.filter(this._objects, filters.vertices), function(vertex) {
    _.each(_.filter(this._objects, function(d) { return filters.vertices(d) && isType("commit")(d); }), function(commit) {
        g.setNode(commit.id);
        _.each(_.filter(_.map(commit.outEdges, function(edgeId) { return that.get(edgeId); }), isType("version")), function (edge) {
            g.setEdge(commit.id, edge.id);
        });
    });

    return g;
};

