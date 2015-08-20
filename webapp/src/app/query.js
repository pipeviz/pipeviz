// Searches the graph for all apps that derive from the given repo address.
var appsFromRepo = function(graph, repo) {
    return _.filter(graph.verticesWithType("logic-state"), function(v) {
        var vedges = _.filter(_.map(v.outEdges, function(edgeId) { return graph.get(edgeId); }), isType("version"));
        if (vedges.length === 0) {
            return false;
        }

        return graph.get(vedges[0]).prop("repository") === repo;
    });
};

var getCommit = function(g, d) {
    switch (d.Typ()) {
        case "logic-state":
        case "git-tag":
        case "git-branch":
            var vedges = _.filter(_.map(d.outEdges, function(edgeId) { return g.get(edgeId); }), isType("version"));
            if (vedges.length !== 0) {
                return g.get(vedges[0].target);
            }
            break;

        case "commit":
            return d;
    }
    return;
};

/**
 * Given a commit, or logic state (and a linked commit through its version identifier),
 * return the name of the repository associated with the commit.
 */
var getRepositoryName = function(g, d) {
    switch (d.Typ()) {
        case "logic-state":
        case "git-tag":
        case "git-branch":
            d = getCommit(g, d);
            if (d === undefined) {
                break;
            }
            // we have a commit now, so fall through to commit case
        case "commit":
            // TODO for now this is a prop on the commit, but it'll soon be its own vtx
            return d.propv("repository");
    }
    return;
};

/**
 * Try to find the containing environment for the given vertex.
 */
var getEnvironment = function(pvg, d) {
    var vedges = _.filter(_.map(d.outEdges, function(edgeId) { return pvg.get(edgeId); }), isType("envlink"));
    if (vedges.length !== 0) {
        return pvg.get(vedges[0].target);
    }
    return;
};

/**
 * Given an environment vertex, picks an appropriate name to represent it.
 * First tries hostname, then ipv4, then ipv6.
 */
var getEnvName = function(d) {
    return d.propv("hostname") || d.propv("ipv4") || d.propv("ipv6");
};

var getTestState = function(g, c) {
    var testState;
    _.each(g.verticesWithType("test-result"), function(d) {
        _.each(_.map(d.outEdges, function(edgeId) { return g.get(edgeId); }), function(d2) {
            if (d2.target === c.id) {
                testState = d.prop("result").value;
            }
        });
    });
    return testState;
};
