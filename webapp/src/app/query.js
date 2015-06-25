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
    if (d.Typ() === "logic-state") {
        var vedges = _.filter(_.map(d.outEdges, function(edgeId) { return g.get(edgeId); }), isType("version"));
        if (vedges.length !== 0) {
            return g.get(vedges[0].target);
        }
    } else if (d.Typ() === "commit") {
        return d;
    }
    return;
};

/**
 * Given a commit, or logic state (and a linked commit through its version identifier),
 * return the name of the repository associated with the commit.
 */
var getRepositoryName = function(g, d) {
    if (d.Typ() === "logic-state") {
        d = getCommit(d);
        if (_.isUndefined(d)) {
            return;
        }
    }
    // TODO for now this is a prop on the commit, but it'll soon be its own vtx
    return d.propv("repository");
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
