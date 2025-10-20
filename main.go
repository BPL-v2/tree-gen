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

func main() {
	http.Handle("/", http.HandlerFunc(circle))
	err := http.ListenAndServe(":2003", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

type TreeDrawer struct {
	s    *svg.SVG
	Tree Tree
}

func InitTreeDrawer(w http.ResponseWriter) *TreeDrawer {
	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)

	file, err := os.Open("atlas-tree.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	tree := Tree{}
	err = json.NewDecoder(file).Decode(&tree)
	if err != nil {
		log.Fatal(err)
	}
	FixTree(&tree)
	fmt.Printf("Tree size: %d x %d\n", tree.MaxX-tree.MinX, tree.MaxY-tree.MinY)
	s.Start(1000, 1000, fmt.Sprintf("viewBox=\"%d %d %d %d\"", tree.MinX, tree.MinY, tree.MaxX-tree.MinX, tree.MaxY-tree.MinY))
	s.Rect(tree.MinX, tree.MinY, tree.MaxX-tree.MinX, tree.MaxY-tree.MinY, "fill:black")
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

func FixTree(Tree *Tree) {
	ascendancyStarts := make([]string, 0)
	ascendancyMap := make(map[string][]string)
	minx, miny, maxx, maxy := 0, 0, 0, 0
	for nodeid, node := range Tree.Nodes {
		if node.AscendancyName != nil {
			ascendancyMap[*node.AscendancyName] = append(ascendancyMap[*node.AscendancyName], nodeid)
		}
		if node.IsAscendancyStart {
			ascendancyStarts = append(ascendancyStarts, nodeid)
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

	centerGroups := make([]string, 0)
	ascendancyToGroups := make(map[string][]string)
	for groupId, group := range Tree.Groups {
		for ascendancyName, nodeids := range ascendancyMap {
			if HasOverlap(group.Nodes, nodeids) {
				ascendancyToGroups[ascendancyName] = append(ascendancyToGroups[ascendancyName], groupId)
				if HasOverlap(group.Nodes, ascendancyStarts) {
					centerGroups = append(centerGroups, groupId)
				}
			}
		}
	}
	dist := 1000.0
	centerX, centerY := float64(Tree.MaxX)-dist, -float64(Tree.MaxY)+dist
	for _, groupIds := range ascendancyToGroups {
		largest := Intersect(groupIds, centerGroups)[0]
		largestGroup := Tree.Groups[largest]
		for _, groupId := range groupIds {
			group := Tree.Groups[groupId]
			group.X = centerX + (group.X - largestGroup.X)
			group.Y = centerY + (group.Y - largestGroup.Y)
			Tree.Groups[groupId] = group
		}
	}

	fmt.Printf("Tree size: %d x %d\n", Tree.MaxX-Tree.MinX, Tree.MaxY-Tree.MinY)

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

func (d *TreeDrawer) End() {
	d.s.End()
}

func (d *TreeDrawer) DrawNode(node Node) {
	if !node.ShouldDraw() {
		return
	}
	classes := []string{}
	if !node.HasConnections() {
		classes = append(classes, "isolated")
	}
	if node.AscendancyName != nil {
		classes = append(classes, "ascendancy")
		classes = append(classes, *node.AscendancyName)
	}
	if node.IsNotable {
		d.DrawPassive(node, 50, classes)
	} else if node.IsKeystone {
		classes = append(classes, "keystone")
		d.DrawPassive(node, 80, classes)
	} else if node.IsMastery {
		classes = append(classes, "mastery")
		d.DrawPassive(node, 40, classes)
	} else {
		d.DrawPassive(node, 30, classes)
	}
}

func (d *TreeDrawer) DrawPassive(node Node, radius int, cls []string) {
	x, y, err := d.GetCoordinates(node)
	if err != nil {
		return
	}
	attr := fmt.Sprintf("id=\"n-%d\"", node.Skill)
	if len(cls) > 0 {
		attr += fmt.Sprintf(" class=\"%s\"", strings.Join(cls, " "))
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
	if !node1.ShouldDraw() || !node2.ShouldDraw() || !node1.ShouldConnect() || !node2.ShouldConnect() {
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

	// Determine arc direction (choose shorter path)
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

func circle(w http.ResponseWriter, req *http.Request) {
	drawer := InitTreeDrawer(w)
	nodeids := make([]string, 0, len(drawer.Tree.Nodes))
	for nodeid := range drawer.Tree.Nodes {
		nodeids = append(nodeids, nodeid)
	}
	sort.Slice(nodeids, func(i, j int) bool {
		intI, _ := strconv.Atoi(nodeids[i])
		intJ, _ := strconv.Atoi(nodeids[j])
		return intI < intJ
	})
	drawer.s.Gstyle("stroke:white;stroke-width:4;fill:none")
	for _, nodeid := range nodeids {
		node := drawer.Tree.Nodes[nodeid]
		// if node.Skill != 50340 && node.Skill != 5065 {
		// 	continue
		// }

		drawer.DrawConnections(node)
	}
	drawer.s.Gend()
	drawer.s.Gstyle("fill:grey")
	for _, nodeid := range nodeids {
		node := drawer.Tree.Nodes[nodeid]
		// if node.Skill != 50340 && node.Skill != 5065 {
		// 	continue
		// }
		drawer.DrawNode(node)

	}
	drawer.s.Gend()

	drawer.End()
}
