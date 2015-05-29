// just so my syntastic complains less
var _ = _ || {};

// TODO get rid of this anchor shit, replace with well-formed tree rendering
function Anchor(id, x, y) {
    this.fixed = true; // tells d3 not to move it
    this.x = x || 0;
    this.y = y || 0;
    this.id = id;
}

Anchor.prototype.name = function() {
    return '';
};

Anchor.prototype.Typ = function() {
    return 'anchor';
};

Anchor.prototype.isVertex = function() {
    return true;
};

Anchor.prototype.prop = function() {
    return;
};

var vertexProto = {
    isVertex: function() { return true; },
    Typ: function() { return this.vertex.type; },
    prop: function(path) {
        if (_.has(this.vertex.properties, path)) {
            return this.vertex.properties[path];
        }
    }
};

// vertex factory
var Vertex = function(obj) {
    return _.assign(Object.create(vertexProto),
        obj, { outEdges: _.map(obj.outEdges, function(d) { return d.id; })}
    );
};

var edgeProto = {
    isVertex: function() { return false; },
    Typ: function() { return this.etype; },
    prop: function(path) {
        if (_.has(this.properties, path)) {
            return this.properties[path];
        }
    }
};

// edge factory
var Edge = function(obj) { return _.assign(Object.create(edgeProto), obj); };

var pvGraphProto = {
    get: function(id) {
        return this._objects[id];
    },
    // With no arguments, returns all vertices in the graph. If one argument
    // is passed, it is taken to be a filtering function, and each vertex will
    // be passed as a candidate for elimination.
    vertices: function() {
        if (arguments.length === 1) {
            return _.filter(this._objects, function(d) {
                return d.isVertex();
            });
        }

        var cf = arguments[1];
        return _.filter(this._objects, function(d) {
            return d.isVertex() && cf(d);
        });
    },
    verticesWithType:  function(typ) {
        return _.filter(this._objects, function(d) {
            return  d.isVertex() && d.vertex.type === typ;
        });
    },
    // Returns a graphlib.Graph object representing the graph(s) of all known
    // commit objects.
    commitGraph: function() {
        var g = new graphlib.Graph();
        var that = this;

        //_.each(_.filter(this._objects, filters.vertices), function(vertex) {
        _.each(_.filter(this._objects, function(d) { return filters.vertices(d) && isType("commit")(d); }), function(commit) {
            g.setNode(commit.id);
            _.each(_.filter(_.map(commit.outEdges, function(edgeId) { return that.get(edgeId); }), isType("version")), function (edge) {
                g.setEdge(commit.id, edge.target);
            });
        });

        return g;
    }
},
// TODO this pretty much mirrors what we have serverside...for now. ugh.
// pipeviz graph datastore factory
pvGraph = function(obj) {
    return _.assign(Object.create(pvGraphProto), (function() {
        var o = {
            _objects: {},
            mid: obj.id
        };

        _.each(obj.vertices, function(d) {
            o._objects[d.id] = Vertex(d);
            _.each(d.outEdges, function(d2) { o._objects[d2.id] = Edge(d2); });
        });

        return o;
    }())
    );
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
