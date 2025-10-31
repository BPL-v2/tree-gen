package main

import "math"

type Tree struct {
	Tree            string                       `json:"tree"`
	Classes         []Classes                    `json:"classes"`
	Groups          map[string]Group             `json:"groups"`
	Nodes           map[string]Node              `json:"nodes"`
	ExtraImages     map[string]ExtraImage        `json:"extraImages"`
	JewelSlots      []int                        `json:"jewelSlots"`
	MinX            int                          `json:"min_x"`
	MaxX            int                          `json:"max_x"`
	MinY            int                          `json:"min_y"`
	MaxY            int                          `json:"max_y"`
	Constants       Constants                    `json:"constants"`
	Sprites         map[string]map[string]Sprite `json:"sprites"`
	ImageZoomLevels []float64                    `json:"imageZoomLevels"`
	Points          PassivePoints                `json:"points"`
}

type CompactTree struct {
	Groups map[string]CompactGroup `json:"groups"`
	Nodes  map[string]CompactNode  `json:"nodes"`
}

type CompactGroup struct {
	Nodes []string `json:"nodes"`
}

type CompactNode struct {
	Name        *string  `json:"name,omitempty"`
	Stats       []string `json:"stats,omitempty"`
	IsMastery   bool     `json:"isMastery,omitempty"`
	IsNotable   bool     `json:"isNotable,omitempty"`
	IsKeystone  bool     `json:"isKeystone,omitempty"`
	IsBloodline bool     `json:"isBloodline,omitempty"`
}

type Classes struct {
	Name         string       `json:"name"`
	BaseStr      int          `json:"base_str"`
	BaseDex      int          `json:"base_dex"`
	BaseInt      int          `json:"base_int"`
	Ascendancies []Ascendancy `json:"ascendancies"`
}

type Ascendancy struct {
	Id                string           `json:"id"`
	Name              string           `json:"name"`
	FlavourText       *string          `json:"flavourText,omitempty"`
	FlavourTextColour *string          `json:"flavourTextColour,omitempty"`
	FlavourTextRect   *FlavourTextRect `json:"flavourTextRect,omitempty"`
}

type FlavourTextRect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Group struct {
	X          float64    `json:"x"`
	Y          float64    `json:"y"`
	Orbits     []int      `json:"orbits"`
	Background Background `json:"background"`
	Nodes      []string   `json:"nodes"`
	IsProxy    bool       `json:"isProxy"`
}

type Background struct {
	Image       string `json:"image"`
	IsHalfImage bool   `json:"isHalfImage"`
}

type Node struct {
	Group                  int             `json:"group"`
	Orbit                  int             `json:"orbit"`
	OrbitIndex             int             `json:"orbitIndex"`
	Out                    []string        `json:"out"`
	In                     []string        `json:"in"`
	Skill                  int             `json:"skill,omitempty"`
	Name                   *string         `json:"name,omitempty"`
	Icon                   *string         `json:"icon,omitempty"`
	AscendancyName         *string         `json:"ascendancyName,omitempty"`
	Stats                  []string        `json:"stats,omitempty"`
	ReminderText           []string        `json:"reminderText,omitempty"`
	FlavourText            []string        `json:"flavourText,omitempty"`
	ExpansionJewel         *ExpansionJewel `json:"expansionJewel,omitempty"`
	GrantedPassivePoints   int             `json:"grantedPassivePoints,omitempty"`
	GrantedIntelligence    int             `json:"grantedIntelligence,omitempty"`
	GrantedStrength        int             `json:"grantedStrength,omitempty"`
	GrantedDexterity       int             `json:"grantedDexterity,omitempty"`
	Recipe                 []string        `json:"recipe,omitempty"`
	IsNotable              bool            `json:"isNotable,omitempty"`
	IsProxy                bool            `json:"isProxy,omitempty"`
	IsBlighted             bool            `json:"isBlighted,omitempty"`
	IsMultipleChoiceOption bool            `json:"isMultipleChoiceOption,omitempty"`
	IsMultipleChoice       bool            `json:"isMultipleChoice,omitempty"`
	IsJewelSocket          bool            `json:"isJewelSocket,omitempty"`
	IsAscendancyStart      bool            `json:"isAscendancyStart,omitempty"`
	IsMastery              bool            `json:"isMastery,omitempty"`
	IsKeystone             bool            `json:"isKeystone,omitempty"`
	IsWormhole             bool            `json:"isWormhole,omitempty"`
	IsBloodline            bool            `json:"isBloodline,omitempty"`
	InactiveIcon           *string         `json:"inactiveIcon,omitempty"`
	ActiveIcon             *string         `json:"activeIcon,omitempty"`
	ActiveEffectImage      *string         `json:"activeEffectImage,omitempty"`
	MasteryEffects         []MasteryEffect `json:"masteryEffects,omitempty"`
	ClassStartIndex        *int            `json:"classStartIndex,omitempty"`
}

func (n Node) ShouldDraw() bool {
	return n.ClassStartIndex == nil && !n.IsProxy && !(n.ExpansionJewel != nil && n.ExpansionJewel.Size < 2)
}

func (n1 Node) ShouldConnectTo(n2 Node) bool {
	return !n1.IsMastery && n1.ClassStartIndex == nil && !n2.IsMastery && n2.ClassStartIndex == nil && (!n1.IsWormhole || !n2.IsWormhole)
}

func (n Node) HasConnections() bool {
	return len(n.Out) > 0 || len(n.In) > 0
}

type ExpansionJewel struct {
	Size   int    `json:"size"`
	Index  int    `json:"index"`
	Proxy  string `json:"proxy"`
	Parent string `json:"parent"`
}

type MasteryEffect struct {
	Effect int      `json:"effect"`
	Stats  []string `json:"stats"`
}

type ExtraImage struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Image string  `json:"image"`
}

type Constants struct {
	// Classes              map[string]int `json:"classes"`
	// CharacterAttributes  map[string]int `json:"characterAttributes"`
	PSSCentreInnerRadius int   `json:"pssCentreInnerRadius"`
	SkillsPerOrbit       []int `json:"skillsPerOrbit"`
	OrbitRadii           []int `json:"orbitRadii"`
}

type Sprite struct {
	FileName string                  `json:"fileName"`
	W        int                     `json:"w"`
	H        int                     `json:"h"`
	Coords   map[string]SpriteCoords `json:"coords"`
}

type SpriteCoords struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type PassivePoints struct {
	TotalPoints      int `json:"totalPoints"`
	AscendancyPoints int `json:"ascendancyPoints"`
}

func GetOrbitAngle(index int, total int) float64 {
	i := math.Pi / 180
	switch total {
	case 40:
		switch index {
		case 0:
			return GetOrbitAngle(0, 12)
		case 1:
			return GetOrbitAngle(0, 12) + 10*i
		case 2:
			return GetOrbitAngle(0, 12) + 20*i
		case 3:
			return GetOrbitAngle(1, 12)
		case 4:
			return GetOrbitAngle(1, 12) + 10*i
		case 5:
			return GetOrbitAngle(1, 12) + 15*i
		case 6:
			return GetOrbitAngle(1, 12) + 20*i
		case 7:
			return GetOrbitAngle(2, 12)
		case 8:
			return GetOrbitAngle(2, 12) + 10*i
		case 9:
			return GetOrbitAngle(2, 12) + 20*i
		case 10:
			return GetOrbitAngle(3, 12)
		case 11:
			return GetOrbitAngle(3, 12) + 10*i
		case 12:
			return GetOrbitAngle(3, 12) + 20*i
		case 13:
			return GetOrbitAngle(4, 12)
		case 14:
			return GetOrbitAngle(4, 12) + 10*i
		case 15:
			return GetOrbitAngle(4, 12) + 15*i
		case 16:
			return GetOrbitAngle(4, 12) + 20*i
		case 17:
			return GetOrbitAngle(5, 12)
		case 18:
			return GetOrbitAngle(5, 12) + 10*i
		case 19:
			return GetOrbitAngle(5, 12) + 20*i
		case 20:
			return GetOrbitAngle(6, 12)
		case 21:
			return GetOrbitAngle(6, 12) + 10*i
		case 22:
			return GetOrbitAngle(6, 12) + 20*i
		case 23:
			return GetOrbitAngle(7, 12)
		case 24:
			return GetOrbitAngle(7, 12) + 10*i
		case 25:
			return GetOrbitAngle(7, 12) + 15*i
		case 26:
			return GetOrbitAngle(7, 12) + 20*i
		case 27:
			return GetOrbitAngle(8, 12)
		case 28:
			return GetOrbitAngle(8, 12) + 10*i
		case 29:
			return GetOrbitAngle(8, 12) + 20*i
		case 30:
			return GetOrbitAngle(9, 12)
		case 31:
			return GetOrbitAngle(9, 12) + 10*i
		case 32:
			return GetOrbitAngle(9, 12) + 20*i
		case 33:
			return GetOrbitAngle(10, 12)
		case 34:
			return GetOrbitAngle(10, 12) + 10*i
		case 35:
			return GetOrbitAngle(10, 12) + 15*i
		case 36:
			return GetOrbitAngle(10, 12) + 20*i
		case 37:
			return GetOrbitAngle(11, 12)
		case 38:
			return GetOrbitAngle(11, 12) + 10*i
		case 39:
			return GetOrbitAngle(11, 12) + 20*i
		}
	case 16:
		switch index {
		case 0:
			return 0
		case 1:
			return 30 * i
		case 2:
			return 45 * i
		case 3:
			return 60 * i
		case 4:
			return 90 * i
		case 5:
			return 120 * i
		case 6:
			return 135 * i
		case 7:
			return 150 * i
		case 8:
			return math.Pi
		case 9:
			return 210 * i
		case 10:
			return 225 * i
		case 11:
			return 240 * i
		case 12:
			return math.Pi * 3 / 2
		case 13:
			return 300 * i
		case 14:
			return 315 * i
		case 15:
			return 330 * i
		}
	}
	return 2 * math.Pi * float64(index) / float64(total)
}
