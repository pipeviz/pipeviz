var _ = require('../bower_components/lodash/dist/lodash');

// PVD: PipeVizDatastore
function PVD() {
    this._containers = {};
}

PVD.prototype.attachContainer = function(container) {
    // FIXME hostname is not a uuid
    this._containers[container.hostname] = container;
    return this;
};

/**
 * Attempt to locate a container by hostname.
 */
PVD.prototype.byHostname = function(hostname) {
    if (_.has(this._containers, hostname)) {
        return this._containers[hostname];
    }
    return false;
};

PVD.prototype.eachContainer = function(lambda, thisArg) {
    return _.each(this._containers, lambda, thisArg);
};

/**
 * Finds all logical groups contained in this PVD.
 *
 * Returns a map keyed by group name where the values are an array of all domain
 * objects that are marked with that logical group.
 */
PVD.prototype.findLogicalGroups = function() {
    var lgroups = {};

    // FIXME just checks logical states. bc we're not a real graphdb (...yet)
    _.each(this._containers, function(container) {
        _.each(container.logicStates(), function(ls) {
            if (_.has(ls, 'lgroup')) {
                if (!_.has(lgroups, ls.lgroup)) {
                    lgroups[ls.lgroup] = [];
                }
                lgroups[ls.lgroup].push(ls);
            }
        });
    });

    return lgroups;
};

/**
 * Returns a 2-tuple of the nodes and links present in this PVD.
 *
 * If provided, the nodeFilter and linkFilter functions will be called prior
 * to each object being added; objects will only be added if the filter func
 * returns true.
 *
 * If either filter function is not provided, a dummy is used instead that
 * unconditionally returns true (allows everything through).
 */
PVD.prototype.nodesAndLinks = function(nodeFilter, linkFilter) {
    // if filter funcs were not provided, just attach one that always passes
    var pvd = this,
        nf = (nodeFilter instanceof Function) ? nodeFilter : function() { return true; },
        lf = (linkFilter instanceof Function) ? linkFilter : function() { return true; },
        nodes = [],
        links = [],
        maybeAddNode = function(node) {
            if (nf(node)) {
                nodes.push(node);
            }
        },
        maybeAddLink = function(link) {
            if (lf(link)) {
                links.push(link);
            }
        };

    // TODO this approach makes links entirely subordinate to nodes...fine for
    // now, maybe not the best in the long run though
    _.each(this._containers, function(container) {
        _.each(container.logicStates(), function(v) {
            maybeAddNode(v);
            maybeAddLink({source: container, target: v});
        });

        _.each(container.processes(), function(v) {
            maybeAddNode(v);
            maybeAddLink({source: container, target: v});
        });

        _.each(container.dataSpaces(), function(v) {
            maybeAddNode(v);
            maybeAddLink({source: container, target: v});
        });

        _.each(container.dataSets(), function(v) {
            maybeAddNode(v);
            maybeAddLink({source: container, target: v});
        });

        maybeAddNode(container);
    });

    // temp var - 'referred container'
    var rc;

    // All hierarchical data is processed; second pass for referential.
    _.forOwn(this._containers, function(cv) {
        // find logic refs to data
        _.each(cv.logicStates(), function(l) {
            if (_.has(l, 'datasets')) {
                _.forOwn(l.datasets, function(dv, dk) {
                    var link = _.assign({source: l, name: dk}, dv);
                    var proc = false;
                    // check for hostname, else assume path
                    if (_.has(dv.loc, 'hostname')) {
                        rc = pvd.byHostname(dv.loc.hostname);
                        if (rc) {
                            proc = rc.findProcess(dv.loc);
                        }
                    } else {
                        proc = cv.findProcess(dv.loc);
                    }

                    if (proc) {
                        link.target = proc;
                        maybeAddLink(link);
                    }
                });
            }
        });

        // next, link processes to logic states & data spaces
        _.each(cv.processes(), function(proc) {
            _.each(proc.logicStates(), function(ls) {
                maybeAddLink({source: proc, target: ls});
            });

            _.each(proc.dataSpaces(), function(ds) {
                maybeAddLink({source: proc, target: ds});
            });
        });

        _.each(cv.dataSpaces(), function(dg) {
            _.each(dg.dataSets(), function(ds) {
                // direct tree/parentage info
                maybeAddLink({source: dg, target: ds});
                // add dataset genesis information
                if (ds.genesis !== "α") {
                    var link = {source: ds};
                    var ods = false;
                    if (_.has(ds.genesis, 'hostname')) {
                        rc = pvd.byHostname(ds.genesis.hostname);
                        if (rc) {
                            ods = rc.findDataSet(ds.genesis);
                        }
                    } else {
                        ods = cv.findDataSet(ds.genesis);
                    }

                    if (ods) {
                        link.target = ods;
                        maybeAddLink(link);
                    }
                }
            });
        });
    });

    return [nodes, links];
};

module.exports = PVD;
