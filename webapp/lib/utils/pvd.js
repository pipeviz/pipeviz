var _ = require('lodash');
var graphlib = require('graphlib');

var vertexProto = module.exports.vertexProto = {
  isVertex: function () { return true; },
  Typ: function () { return this.vertex.type; },
  prop: function (path) {
    if (_.has(this.vertex.properties, path)) {
      return this.vertex.properties[path];
    }
  },
  propv: function (path) {
    if (_.has(this.vertex.properties, path)) {
      return this.vertex.properties[path].value;
    }
  }
};

// vertex factory
var pvVertex = function (obj) {
  return _.assign(Object.create(vertexProto),
    obj, {
      outEdges: _.map(obj.outEdges, function (d) { return d.id; }),
      inEdges: _.map(obj.inEdges, function (d) { return d.id; })
    }
  );
};

var edgeProto = module.exports.edgeProto =  {
  isVertex: function () { return false; },
  Typ: function () { return this.etype; },
  prop: function (path) {
    if (_.has(this.properties, path)) {
      return this.properties[path];
    }
  },
  propv: function (path) {
    if (_.has(this.properties, path)) {
      return this.properties[path].value;
    }
  }
};

// edge factory
var pvEdge = function (obj) { return _.assign(Object.create(edgeProto), obj); };

var pvGraphProto = {
  get: function (id) {
    return this._objects[id];
  },
  // With no arguments, returns all vertices in the graph. If one argument
  // is passed, it is taken to be a filtering function, and each vertex will
  // be passed as a candidate for elimination.
  vertices: function () {
    if (arguments.length === 0) {
      return _.filter(this._objects, function (d) {
        return d.isVertex();
      });
    }

    var cf = arguments[0];
    return _.filter(this._objects, function (d) {
      return d.isVertex() && cf(d);
    });
  },
  verticesWithType: function (typ) {
    return _.filter(this._objects, function (d) {
      return d.isVertex() && d.vertex.type === typ;
    });
  },
  // Returns a graphlib.Graph object representing the graph(s) of all known
  // commit objects. If an argument is passed, it is assumed to be the name
  // of the repo to which results should be restricted.
  commitGraph: function () {
    var fltr;
    var g = new graphlib.Graph();
    var that = this;

    if (arguments.length === 0) {
      fltr = pq.and(filters.vertices, isType("commit"));
    }
    else {
      var repo = arguments[0];
      fltr = pq.and(filters.vertices, isType("commit"), function (c) { return c.propv("repository") === repo; });
    }

    _.each(_.filter(this._objects, fltr), function (commit) {
      g.setNode(commit.id);
      _.each(_.filter(_.map(commit.outEdges, function (edgeId) { return that.get(edgeId); }), isType("parent-commit")), function (edge) {
        g.setEdge(commit.id, edge.target, edge.prop("pnum").value);
      });
    });

    return g;
  }
};

module.exports.pvGraph = function (obj) {
  return _.assign(Object.create(pvGraphProto), (function () {
      var o = {
        _objects: {},
        mid: obj.id
      };

      _.each(obj.vertices, function (d) {
        o._objects[d.id] = pvVertex(d);
        _.each(d.outEdges, function (d2) { o._objects[d2.id] = pvEdge(d2); });
      });

      return o;
    }())
  );
};

var pq = module.exports.pq = {
  and: function () {
    var funcs = arguments;
    return function (d) {
      for (var i = 0; i < funcs.length; i++) {
        if (!funcs[i](d)) {
          return false;
        }
      }

      return true;
    };
  },
  or: function () {
    var funcs = arguments;
    return function (d) {
      for (var i = 0; i < funcs.length; i++) {
        if (funcs[i](d)) {
          return true;
        }
      }

      return false;
    };
  }
};

var filters = {
  vertices: function (d) {
    return d.isVertex();
  },
  edges: function (d) {
    return !d.isVertex();
  }
};

/**
 * Checks the provided object to see if it is one of our edge types.
 *
 * This is performed by checking the object prototype chain.
 */
module.exports.isEdge = function(obj) {
  return edgeProto.isPrototypeOf(obj);
};

/**
 * Checks the provided object to see if it is one of our vertex types.
 *
 * This is performed by checking the object prototype chain.
 */
module.exports.isVertex = function(obj) {
  return vertexProto.isPrototypeOf(obj);
};

var isType = module.exports.isType = function (typ) {
  if (arguments.length === 1) {
    return function (d) {
      return typ === d.Typ();
    };
  }

  var typs = arguments;
  return function (d) {
    var typ = d.Typ();
    for (var i = 0; i < typs.length; i++) {
      if (typs[i] === typ) {
        return true;
      }
    }

    return false;
  };
};
