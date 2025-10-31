package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	svg "github.com/ajstarks/svgo"
)

func fixName(name string) string {
	return strings.ReplaceAll(name, ".json", ".svg")
}

func main() {
	// create dirs if not existing
	err := os.MkdirAll("svg/atlas", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("svg/passives", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("json/atlas", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("json/passives", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	// loop over all json files in atlastree and generate svg files
	entries, err := os.ReadDir("atlastree")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range entries {
		if strings.HasSuffix(file.Name(), ".json") {
			fmt.Printf("Generating SVG for %s\n", file.Name())
			DrawTree("atlastree/"+file.Name(), "svg/atlas/"+fixName(file.Name()))
			SaveCompactJson("atlastree/"+file.Name(), "json/atlas/"+file.Name())
		}
	}

	entries, err = os.ReadDir("skilltree")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range entries {
		if strings.HasSuffix(file.Name(), ".json") {
			fmt.Printf("Generating SVG for %s\n", file.Name())
			DrawTree("skilltree/"+file.Name(), "svg/passives/"+fixName(file.Name()))
			SaveCompactJson("skilltree/"+file.Name(), "json/passives/"+file.Name())
		}
	}

}

type TreeDrawer struct {
	s    *svg.SVG
	Tree Tree
}

func InitTreeDrawer(w http.ResponseWriter) *TreeDrawer {
	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)

	file, err := os.Open("atlastree/3.25.0.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	tree := Tree{}
	err = json.NewDecoder(file).Decode(&tree)
	if err != nil {
		log.Fatal(err)
	}
	MoveAscendancyTrees(&tree)
	s.Start(1000, 1000, fmt.Sprintf("viewBox=\"%d %d %d %d\"", tree.MinX, tree.MinY, tree.MaxX-tree.MinX, tree.MaxY-tree.MinY))

	s.Def()
	s.Style("text/css", `
		/* Node styles */
		circle {
			fill: #3e3e3e;
			stroke: #8b8b8b;
			stroke-width: 2;
		}

		circle.keystone {
			fill: #b8860b;
			stroke: #ffd700;
			stroke-width: 3;
		}

		circle.mastery {
			fill: #4169e1;
			stroke: #87ceeb;
			stroke-width: 2;
		}

		circle.isolated {
			fill: #ff6b6b;
			stroke: #ff4757;
			stroke-width: 2;
		}

		circle.ascendancy {
			fill: #9932cc;
			stroke: #ba55d3;
			stroke-width: 2;
		}

		/* Connection styles */
		line, path {
			stroke: #666666;
			stroke-width: 2;
			fill: none;
		}

		line.ascendancy, path.ascendancy {
			stroke: #9932cc;
			stroke-width: 3;
		}

		/* Hover effects */
		circle:hover {
			stroke-width: 4;
			filter: brightness(1.2);
		}

		line:hover, path:hover {
			stroke-width: 4;
			filter: brightness(1.2);
		}

		/* Group orbit styles */
		circle[fill="none"] {
			stroke: #228b22;
			stroke-width: 1;
			stroke-dasharray: 5,5;
		}

		/* Mastery image styles */
		image.mastery {
			filter: brightness(1.0);
			border: 2px solid #4169e1;
			border-radius: 50%;
		}

		image.mastery:hover {
			filter: brightness(1.3);
			border-width: 3px;
		}
	`)
	s.DefEnd()

	return &TreeDrawer{
		s:    s,
		Tree: tree,
	}
}

func HasOverlap[T comparable](x, y []T) bool {
	set := make(map[T]struct{})
	for _, item := range x {
		set[item] = struct{}{}
	}
	for _, item := range y {
		if _, exists := set[item]; exists {
			return true
		}
	}
	return false
}

func Intersect[T comparable](x, y []T) []T {
	set := make(map[T]struct{})
	for _, item := range x {
		set[item] = struct{}{}
	}
	intersection := make([]T, 0)
	for _, item := range y {
		if _, exists := set[item]; exists {
			intersection = append(intersection, item)
		}
	}
	return intersection
}

func MoveAscendancyTrees(Tree *Tree) {
	ascendancyStarts := make([]string, 0)
	bloodlineStarts := make([]string, 0)
	ascendancyMap := make(map[string][]string)
	bloodlineMap := make(map[string][]string)
	minx, miny, maxx, maxy := 0, 0, 0, 0
	for nodeid, node := range Tree.Nodes {
		if node.AscendancyName != nil {
			if node.IsBloodline {
				bloodlineMap[*node.AscendancyName] = append(bloodlineMap[*node.AscendancyName], nodeid)
			} else {
				ascendancyMap[*node.AscendancyName] = append(ascendancyMap[*node.AscendancyName], nodeid)
			}
		}
		if node.IsAscendancyStart {
			if node.IsBloodline {
				bloodlineStarts = append(bloodlineStarts, nodeid)
			} else {
				ascendancyStarts = append(ascendancyStarts, nodeid)
			}
		}
		if node.ShouldDraw() && node.AscendancyName == nil {
			x, y, err := GetCoordinates(node, *Tree)
			if err != nil {
				continue
			}
			if x < minx {
				minx = x
			}
			if x > maxx {
				maxx = x
			}
			if y < miny {
				miny = y
			}
			if y > maxy {
				maxy = y
			}
		}

	}
	Tree.MinX = minx - 200
	Tree.MinY = miny - 200
	Tree.MaxX = maxx + 200
	Tree.MaxY = maxy + 200

	ascendancyCenterGroups := make([]string, 0)
	bloodlineCenterGroups := make([]string, 0)
	ascendancyToGroups := make(map[string][]string)
	bloodlineToGroups := make(map[string][]string)
	for groupId, group := range Tree.Groups {
		for ascendancyName, nodeids := range ascendancyMap {
			if HasOverlap(group.Nodes, nodeids) {
				ascendancyToGroups[ascendancyName] = append(ascendancyToGroups[ascendancyName], groupId)
				if HasOverlap(group.Nodes, ascendancyStarts) {
					ascendancyCenterGroups = append(ascendancyCenterGroups, groupId)
				}
			}
		}
		for bloodlineName, nodeids := range bloodlineMap {
			if HasOverlap(group.Nodes, nodeids) {
				bloodlineToGroups[bloodlineName] = append(bloodlineToGroups[bloodlineName], groupId)
				if HasOverlap(group.Nodes, bloodlineStarts) {
					bloodlineCenterGroups = append(bloodlineCenterGroups, groupId)
				}
			}
		}
	}
	dist := 1000.0
	centerX, centerY := float64(Tree.MaxX)-dist, -float64(Tree.MaxY)+dist
	for _, groupIds := range ascendancyToGroups {
		largest := Intersect(groupIds, ascendancyCenterGroups)[0]
		largestGroup := Tree.Groups[largest]
		for _, groupId := range groupIds {
			group := Tree.Groups[groupId]
			group.X = centerX + (group.X - largestGroup.X)
			group.Y = centerY + (group.Y - largestGroup.Y)
			Tree.Groups[groupId] = group
		}
	}
	centerX, centerY = float64(Tree.MinX)+dist, -float64(Tree.MaxY)+dist
	for _, groupIds := range bloodlineToGroups {
		largest := Intersect(groupIds, bloodlineCenterGroups)[0]
		largestGroup := Tree.Groups[largest]
		for _, groupId := range groupIds {
			group := Tree.Groups[groupId]
			group.X = centerX + (group.X - largestGroup.X)
			group.Y = centerY + (group.Y - largestGroup.Y)
			Tree.Groups[groupId] = group
		}
	}
}

func GetCoordinates(node Node, tree Tree) (int, int, error) {
	if node.Group == 0 {
		return 0, 0, fmt.Errorf("node has no group")
	}
	group, exists := tree.Groups[fmt.Sprintf("%d", node.Group)]
	if !exists {
		log.Printf("Group %d does not exist for node %s", node.Group, *node.Name)
		return 0, 0, fmt.Errorf("group does not exist")
	}
	radius := tree.Constants.OrbitRadii[node.Orbit]
	angle := GetOrbitAngle(node.OrbitIndex, tree.Constants.SkillsPerOrbit[node.Orbit])
	x := int(group.X + float64(radius)*math.Sin(angle))
	y := int(group.Y - float64(radius)*math.Cos(angle))

	return x, y, nil
}

func (d *TreeDrawer) DrawNode(node Node) {
	if !node.ShouldDraw() {
		return
	}
	classes := []string{}
	extras := []string{}
	if !node.HasConnections() {
		classes = append(classes, "isolated")
	}
	if node.AscendancyName != nil {
		classes = append(classes, "ascendancy")
		extras = append(extras, *node.AscendancyName)
	}
	if node.IsBloodline {
		classes = append(classes, "bloodline")
		extras = append(extras, *node.AscendancyName)
	}
	if node.IsNotable {
		d.DrawPassive(node, 50, classes, extras)
	} else if node.IsKeystone || node.IsWormhole {
		classes = append(classes, "keystone")
		d.DrawPassive(node, 80, classes, extras)
	} else if node.IsMastery {
		classes = append(classes, "mastery")
		extras = append(extras, *node.Name)
		d.DrawPassive(node, 30, classes, extras)
	} else {
		d.DrawPassive(node, 30, classes, extras)
	}
}

func (d *TreeDrawer) DrawPassive(node Node, radius int, cls []string, extras []string) {
	x, y, err := d.GetCoordinates(node)
	if err != nil {
		return
	}
	attr := fmt.Sprintf("id=\"n-%d\"", node.Skill)
	if len(cls) > 0 {
		attr += fmt.Sprintf(" class=\"%s\"", strings.Join(cls, " "))
	}
	if len(extras) > 0 {
		attr += fmt.Sprintf(" data-extras=\"%s\"", strings.Join(extras, ","))
	}
	d.s.Circle(x, y, radius, attr)
}

func (d *TreeDrawer) DrawConnections(node Node) {
	if node.GrantedPassivePoints == 2 {
		return
	}
	for _, neighbourId := range node.Out {
		neighbour := d.Tree.Nodes[neighbourId]
		d.DrawConnection(node, neighbour)
	}
}

func (d *TreeDrawer) DrawConnection(node1 Node, node2 Node) {
	if !node1.ShouldDraw() || !node2.ShouldDraw() || !node1.ShouldConnectTo(node2) {
		return
	}
	attr := fmt.Sprintf("id=\"c-%d-%d\"", node1.Skill, node2.Skill)
	if node1.AscendancyName != nil {
		attr += fmt.Sprintf(" class=\"ascendancy %s\"", *node1.AscendancyName)
	}
	if node1.Group == node2.Group && node1.Orbit == node2.Orbit {
		d.DrawArc(node1, node2, attr)
	} else {
		d.DrawLine(node1, node2, attr)
	}
}

func (d *TreeDrawer) DrawLine(node1 Node, node2 Node, attr string) {
	x1, y1, err := d.GetCoordinates(node1)
	if err != nil {
		return
	}
	x2, y2, err := d.GetCoordinates(node2)
	if err != nil {
		return
	}
	d.s.Line(x1, y1, x2, y2, attr)
}

func (d *TreeDrawer) DrawArc(node1 Node, node2 Node, attr string) {
	x1, y1, err := d.GetCoordinates(node1)
	if err != nil {
		return
	}
	x2, y2, err := d.GetCoordinates(node2)
	if err != nil {
		return
	}

	radius := d.Tree.Constants.OrbitRadii[node1.Orbit]
	node1Angle := GetOrbitAngle(node1.OrbitIndex, d.Tree.Constants.SkillsPerOrbit[node1.Orbit])
	node2Angle := GetOrbitAngle(node2.OrbitIndex, d.Tree.Constants.SkillsPerOrbit[node2.Orbit])

	angleDiff := node2Angle - node1Angle
	if angleDiff > math.Pi {
		angleDiff -= 2 * math.Pi
	} else if angleDiff < -math.Pi {
		angleDiff += 2 * math.Pi
	}

	largeArc := math.Abs(angleDiff) > math.Pi
	sweep := angleDiff > 0
	d.s.Arc(x1, y1, radius, radius, 0, largeArc, sweep, x2, y2, attr)
}

func (d *TreeDrawer) GetCoordinates(node Node) (int, int, error) {
	return GetCoordinates(node, d.Tree)
}

func (d *TreeDrawer) DrawGroup(group Group) {
	for _, orbit := range group.Orbits {
		radius := d.Tree.Constants.OrbitRadii[orbit]
		d.s.Circle(int(group.X), int(group.Y), radius, "fill:none;stroke:green")
	}
}

func SaveCompactJson(fileName string, outFileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	compactTree := CompactTree{}
	err = json.NewDecoder(file).Decode(&compactTree)
	if err != nil {
		log.Fatal(err)
	}
	outFile, err := os.Create(outFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "")
	err = encoder.Encode(compactTree)
	if err != nil {
		log.Fatal(err)
	}
}

func DrawTree(fileName string, out string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	tree := Tree{}
	err = json.NewDecoder(file).Decode(&tree)
	if err != nil {
		log.Fatal(err)
	}
	MoveAscendancyTrees(&tree)

	outFile, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	s := svg.New(outFile)
	s.Startraw(fmt.Sprintf("viewBox=\"%d %d %d %d\"", tree.MinX, tree.MinY, tree.MaxX-tree.MinX, tree.MaxY-tree.MinY))

	drawer := &TreeDrawer{
		s:    s,
		Tree: tree,
	}

	nodeids := make([]string, 0, len(drawer.Tree.Nodes))
	for nodeid := range drawer.Tree.Nodes {
		nodeids = append(nodeids, nodeid)
	}
	sort.Slice(nodeids, func(i, j int) bool {
		intI, _ := strconv.Atoi(nodeids[i])
		intJ, _ := strconv.Atoi(nodeids[j])
		return intI < intJ
	})
	drawer.s.Gid("connections")
	for _, nodeid := range nodeids {
		node := drawer.Tree.Nodes[nodeid]
		drawer.DrawConnections(node)
	}
	drawer.s.Gend()
	drawer.s.Gid("nodes")
	for _, nodeid := range nodeids {
		node := drawer.Tree.Nodes[nodeid]
		drawer.DrawNode(node)
	}
	drawer.s.Gend()
	drawer.s.End()
}
