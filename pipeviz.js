var width = window.innerWidth
    height = window.innerHeight;

var svg = d3.select('body').append('svg')
    .attr('width', width)
    .attr('height', height);

var nlist = [
    {index: 0, type: "graph-anchor", id: "sink", fixed: true, y: height / 2, x: width - 20},
    {index: 1, type: "graph-anchor", id: "source", fixed: true, y: height / 2, x: 20},
    {index: 2, id: "prod"},
    {index: 3, id: "stage"},
    {index: 4, id: "qa"},
    {index: 5, id: "dev1"},
    {index: 6, id: "dev2"}
];

var links = [
    {source: nlist[2], target: nlist[0]},
    {source: nlist[3], target: nlist[2]},
    {source: nlist[4], target: nlist[3]},
    {source: nlist[5], target: nlist[4]},
    {source: nlist[6], target: nlist[3]},
    {source: nlist[1], target: nlist[5]},
    {source: nlist[1], target: nlist[6]}
];

var color = d3.scale.category20();

var force = d3.layout.force()
    .nodes(nlist)
    .links(links)
    .charge(-3000)
    .linkDistance(150)
    .gravity(0)
    .size([width, height])
    .start();

// Capture the vertex and edge set as top-level vars
var n = svg.selectAll(".node")
    l = svg.selectAll(".links")

var link = l.data(links)
    .enter().append("line")
    .attr("class", "link")
    .style("stroke-width", function(d) {
        return (d.target == nlist[0] || d.source == nlist[1]) ? 0 : Math.sqrt(d.value);
    });


var nodes = n.data(nlist, function(d, i) { return d.index; })
    .enter().append("g")
    .attr("class", "node")
    .call(force.drag);


nodes.append("circle")
    .attr("x", 0)
    .attr("y", 0)
    .attr("r", function(d) {
        return d.type ? 1 : 35;
    })
    .attr("fill", color(2));

// Put a label on non-anchor vars
nodes.filter(function(d) {
    return !d.hasOwnProperty('type');
}).append("text").text(function(d) { return d.id });

force.on("tick", function() {
    link.attr("x1", function(d) { return d.source.x; })
        .attr("y1", function(d) { return d.source.y; })
        .attr("x2", function(d) { return d.target.x; })
        .attr("y2", function(d) { return d.target.y; });

    nodes.attr("transform", function(d) { return "translate(" + d.x + "," + d.y + ")"; });
});
