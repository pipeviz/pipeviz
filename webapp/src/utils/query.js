var _ = require('lodash');
var pvd = require('./pvd');

// Searches the graph for all apps that derive from the given repo address.
module.exports.appsFromRepo = function (graph, repo) {
  return _.filter(graph.verticesWithType("logic-state"), function (v) {
    var vedges = _.filter(_.map(v.outEdges, function (edgeId) { return graph.get(edgeId); }), pvd.isType("version"));
    if (vedges.length === 0) {
      return false;
    }

    return graph.get(vedges[0]).prop("repository") === repo;
  });
};

module.exports.getCommit = function (g, d) {
  switch (d.Typ()) {
    case "logic-state":
    case "git-tag":
    case "git-branch":
      var vedges = _.filter(_.map(d.outEdges, function (edgeId) { return g.get(edgeId); }), pvd.isType("version"));
      if (vedges.length !== 0) {
        return g.get(vedges[0].target);
      }
      break;

    case "commit":
      return d;
  }
};

/**
 * Given a commit, or logic state (and a linked commit through its version identifier),
 * return the name of the repository associated with the commit.
 */
module.exports.getRepositoryName = function (g, d) {
  switch (d.Typ()) {
    case "logic-state":
    case "git-tag":
    case "git-branch":
      d = module.exports.getCommit(g, d);
      if (typeof d === 'undefined') {
        break;
      }
    // we have a commit now, so fall through to commit case
    case "commit":
      // TODO for now this is a prop on the commit, but it'll soon be its own vtx
      return d.propv("repository");
  }
};

/**
 * Try to find the containing environment for the given vertex.
 */
module.exports.getEnvironment = function (pvg, d) {
  var vedges = _.filter(_.map(d.outEdges, function (edgeId) { return pvg.get(edgeId); }), pvd.isType("envlink"));
  if (vedges.length !== 0) {
    return pvg.get(vedges[0].target);
  }
};

/**
 * Given an environment vertex, picks an appropriate name to represent it.
 * First tries hostname, then ipv4, then ipv6.
 */
module.exports.getEnvName = function (d) {
  return d.propv("hostname") || d.propv("ipv4") || d.propv("ipv6");
};

module.exports.getTestState = function (g, c) {
  var testState;
  _.each(g.verticesWithType("test-result"), function (d) {
    _.each(_.map(d.outEdges, function (edgeId) { return g.get(edgeId); }), function (d2) {
      if (d2.target === c.id) {
        testState = d.prop("result").value;
      }
    });
  });
  return testState;
};

/**
 * Given an object, returns a string appropriate for use as a label.
 *
 * This string may or may not be globally unique. Do not rely on it for that.
 */
module.exports.objectLabel = function (obj) {
  if (typeof obj.Typ === "function") {
    switch (obj.Typ()) {
      // first vertex types
      case "logic-state":
        return obj.propv("path");
      case "git-tag":
      case "git-branch":
      case "dataset":
      case "parent-dataset":
        return obj.propv("name");
      case "commit":
        return obj.propv("sha1").slice(0, 7);
      case "process":
        return obj.propv("pid");
      case "environment":
        return obj.propv("hostname") || obj.propv("ipv4") || obj.propv("ipv6");
      case "comm":
        return obj.propv("port") || obj.propv("path");
      // now, edges
      case "datalink":
        return obj.propv("name");
      case "envlink":
        return obj.propv("hostname") || obj.propv("ipv4") || obj.propv("ipv6") || obj.propv("nick");
      case "parent-commit":
        return (obj.propv("sha1") || '').slice(0, 7);
    }
  }
};
