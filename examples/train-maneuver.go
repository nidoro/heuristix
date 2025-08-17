// TODO: Otimizar checagem de caminhos que "atravessam" material rodante
// estacionado
// TODO: Construir ManeuverTree
package main

import (
    _ "math"
    _ "math/rand"
    _ "time"
    _ "sort"
    "github.com/davecgh/go-spew/spew"
    "fmt"
    _ "github.com/nidoro/heuristix"
    "encoding/json"
    "os"

    "image/color"
    "gonum.org/v1/plot"
    "gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/vg"
)

var _ = spew.Dump

type State map[int][]int

type RollingStock struct {
    Id          int `json:"id"`
    HorsePower  int `json:"hp"`
}

type Config struct {
    Edges           [][]string          `json:"edges"`
    RollingStock    []RollingStock      `json:"rolling-stock"`
    InitialState    map[string][]int    `json:"initial-state"`
    TargetState     map[string][]int    `json:"target-state"`
}

type Node struct {
    Id string
    Side byte
    ParentId string
    ParentIndex int
    AggregatorId string
    AggregatorIndex int
    X float64
    Y float64
    Length int
}

type Nodes []Node

func (nodes Nodes) Len() int {
    return len(nodes)
}

func (nodes Nodes) XY(index int) (x float64, y float64) {
    return nodes[index].X, nodes[index].Y
}

type Graph struct {
    N          int
    Nodes      Nodes
    Edges      [][]int
}

type Data struct {
    Graph           Graph
    RollingStock    []RollingStock
    InitialState    map[int][]int
    TargetState     map[int][]int
}

func GetNodeIndex(nodes []Node, id string) int {
    for i := range nodes {
        if id == nodes[i].Id {
            return i
        }
    }
    return -1
}

func GetNode(nodes []Node, id string) (*Node, int) {
    index := GetNodeIndex(nodes, id)
    if index >= 0 {
        return &nodes[index], index
    }
    return nil, -1
}

func CreateGraph(config Config) Graph {
    var g Graph
    g.Nodes = make([]Node, 0, 4*len(config.Edges))

    for e := range config.Edges {
        uId := config.Edges[e][0]
        vId := config.Edges[e][1]

        uParentId := uId[0:len(uId)-1]
        uSide := uId[len(uId)-1]

        vParentId := vId[0:len(vId)-1]
        vSide := vId[len(vId)-1]

        u, _ := GetNode(g.Nodes, uId)
        v, _ := GetNode(g.Nodes, vId)

        uParentIndex := GetNodeIndex(g.Nodes, uParentId)
        vParentIndex := GetNodeIndex(g.Nodes, vParentId)

        if uParentIndex < 0 {
            uParentIndex = len(g.Nodes)
            g.Nodes = append(g.Nodes, Node{Id: uParentId, Length: 5, ParentIndex: -1, AggregatorId: uParentId, AggregatorIndex: uParentIndex})
        }

        if vParentIndex < 0 {
            vParentIndex = len(g.Nodes)
            g.Nodes = append(g.Nodes, Node{Id: vParentId, Length: 5, ParentIndex: -1, AggregatorId: vParentId, AggregatorIndex: vParentIndex})
        }

        if u == nil {
            g.Nodes = append(g.Nodes, Node{Id: uId, Side: uSide, ParentId: uParentId, ParentIndex: uParentIndex, AggregatorId: uParentId, AggregatorIndex: uParentIndex})
        }

        if v == nil {
            g.Nodes = append(g.Nodes, Node{Id: vId, Side: vSide, ParentId: vParentId, ParentIndex: vParentIndex, AggregatorId: vParentId, AggregatorIndex: vParentIndex})
        }
    }

    g.Edges = make([][]int, len(g.Nodes))

    for i := range g.Edges {
        g.Edges[i] = make([]int, len(g.Nodes))

        for j := range g.Edges[i] {
            g.Edges[i][j] = 0
        }
    }

    for e := range config.Edges {
        uId := config.Edges[e][0]
        vId := config.Edges[e][1]

        uParentId := uId[0:len(uId)-1]
        vParentId := vId[0:len(vId)-1]

        uIndex := GetNodeIndex(g.Nodes, uId)
        vIndex := GetNodeIndex(g.Nodes, vId)

        uParentIndex := GetNodeIndex(g.Nodes, uParentId)
        vParentIndex := GetNodeIndex(g.Nodes, vParentId)

        g.Edges[uIndex][uParentIndex] = 1
        g.Edges[uParentIndex][uIndex] = 1

        g.Edges[vIndex][vParentIndex] = 1
        g.Edges[vParentIndex][vIndex] = 1

        g.Edges[uIndex][vIndex] = 1
        g.Edges[vIndex][uIndex] = 1
    }

    return g
}

func PlotGraph(g Graph, filePath string) {
    plt := plot.New()
    plt.X.Label.Text = "X"
    plt.Y.Label.Text = "Y"
    plt.Add(plotter.NewGrid())

    scatter, _ := plotter.NewScatter(g.Nodes)
    scatter.GlyphStyle.Color = color.RGBA{R: 255, B: 128, A: 255}
    scatter.GlyphStyle.Radius = vg.Points(3)
    plt.Add(scatter)

    for i := range g.Nodes {
        for j := range g.Nodes {
            edge := g.Edges[i][j]

            if edge >= 0 {
                coords := make(plotter.XYs, 2)

                coords[0].X = g.Nodes[i].X
                coords[0].Y = g.Nodes[i].Y
                coords[1].X = g.Nodes[j].X
                coords[1].Y = g.Nodes[j].Y

                line, _ := plotter.NewLine(coords)
                line.LineStyle.Width = vg.Points(1)
                line.LineStyle.Color = color.RGBA{B: 255, A: 255}
                plt.Add(line)
            }
        }
    }

    _ = plt.Save(400, 400, filePath)
}

func PrintEdges(g Graph) {
    fmt.Printf("%4s", "")

    for i := range g.Nodes {
        fmt.Printf("%4s", g.Nodes[i].Id)
    }

    fmt.Println()

    for i := range g.Nodes {
        fmt.Printf("%4s", g.Nodes[i].Id)
        for j := range g.Nodes {
            fmt.Printf("%4d", g.Edges[i][j])
        }
        fmt.Println()
    }
}

func CreateData(config Config) Data {
    var d Data

    d.Graph = CreateGraph(config)
    PrintEdges(d.Graph)

    d.RollingStock = config.RollingStock

    for i := range d.RollingStock {
        fmt.Println(d.RollingStock[i])
    }

    d.InitialState = make(map[int][]int)
    d.TargetState = make(map[int][]int)

    for k, v := range config.InitialState {
        sbIndex := GetNodeIndex(d.Graph.Nodes, k)
        d.InitialState[sbIndex] = v
    }

    for k, v := range config.TargetState {
        sbIndex := GetNodeIndex(d.Graph.Nodes, k)
        d.TargetState[sbIndex] = v
    }

    fmt.Println("InitialState", d.InitialState)
    fmt.Println("TargetState", d.TargetState)

    return d
}

func CopyState(s1 State) State {
    var s2 State

    s2 = make(State)

    for k, v := range s1 {
        s2[k] = make([]int, len(v))
        copy(s2[k], s1[k])
    }

    return s2
}

// ShortestPaths runs BFS from a source node and returns shortest paths
// as a map[targetNodeIndex] = path (slice of node indices).
func ShortestPaths(g Graph, sourceId string) map[string][]string {
    source := GetNodeIndex(g.Nodes, sourceId)

    n := g.N
    if n == 0 {
        n = len(g.Nodes)
    }

    // Distances initialized to -1 (unvisited)
    dist := make([]int, n)
    for i := range dist {
        dist[i] = -1
    }

    // Parents for path reconstruction
    parent := make([]int, n)
    for i := range parent {
        parent[i] = -1
    }

    // BFS queue
    queue := []int{source}
    dist[source] = 0

    for len(queue) > 0 {
        u := queue[0]
        queue = queue[1:]

        for v := 0; v < n; v++ {
            if g.Edges[u][v] > 0 && dist[v] == -1 {
                dist[v] = dist[u] + 1
                parent[v] = u
                queue = append(queue, v)
            }
        }
    }

    // Build paths
    paths := make(map[string][]string)
    for v := 0; v < n; v++ {
        if dist[v] >= 0 && v != source {
            // Reconstruct path from source to v
            path := []string{}
            for x := v; x != -1; x = parent[x] {
                path = append([]string{g.Nodes[x].Id}, path...) // prepend
            }
            paths[g.Nodes[v].Id] = path
        }
    }

    return paths
}

type Path struct {
    Nodes                   []string
    MaxCompositionLength    int
}

func Max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func Min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func GetAllPaths(g Graph, s State, sourceIndex int, targetIndex int) []Path {
    if sourceIndex < 0 || targetIndex < 0 {
        return nil
    }

    var results []Path
    var path []int
    pathMaxCompositionLength := 9999
    visited := make([]bool, len(g.Nodes))

    var dfs func(u int)
    dfs = func(u int) {
        path = append(path, u)
        visited[u] = true

        uAggIndex := g.Nodes[u].AggregatorIndex
        if uAggIndex != sourceIndex {
            uCap := g.Nodes[uAggIndex].Length - len(s[uAggIndex])
            pathMaxCompositionLength = Min(pathMaxCompositionLength, uCap)
        }

        if u == targetIndex {
            // Checar se material rodante é atravessado
            // TODO: Otimizar - esta checagem pode ser feita antes da construção
            // completa do caminho.
            for sbIndex := range s {
                if len(s[sbIndex]) > 0 {
                    for _, k := range path {
                        for _, l := range path {
                            if k == l {continue}
                            if g.Nodes[k].AggregatorIndex == sbIndex && g.Nodes[k].AggregatorIndex == g.Nodes[l].AggregatorIndex && g.Nodes[k].Side != 0 && g.Nodes[l].Side != 0 {
                                pathMaxCompositionLength = 9999
                                return
                            }
                        }
                    }
                }
            }

            pt := Path{MaxCompositionLength: pathMaxCompositionLength}
            pt.Nodes = make([]string, len(path))

            for i, idx := range path {
                pt.Nodes[i] = g.Nodes[idx].Id
            }
            results = append(results, pt)
            pathMaxCompositionLength = 9999
        } else {
            for v := 0; v < len(g.Nodes); v++ {
                if g.Edges[u][v] > 0 && !visited[v] {
                    dfs(v)
                }
            }
        }

        path = path[:len(path)-1]
        visited[u] = false
    }

    dfs(sourceIndex)
    return results
}

type Maneuver struct {
    Composition     []int
    Path            Path
    Cost            float64
    EndState        State
}

type CompositionRange struct {
    FirstAsset  int
    LastAsset   int
}

func GetPossibleManeuvers(d Data, s State) []Maneuver {
    // Localizar locomotivas (sb e posição na sb)
    //-------------------------------------------------
    locoLocations := make(map[int][]int)

    for sbIndex, rollingStock := range s {
        for pos, asset := range rollingStock {
            if d.RollingStock[asset].HorsePower > 0 {
                if locoLocations[sbIndex] == nil {
                    locoLocations[sbIndex] = make([]int, 0, len(rollingStock))
                }
                locoLocations[sbIndex] = append(locoLocations[sbIndex], pos)
            }
        }
    }

    // Para cada sb com locomotiva, construir todos os caminhos possíveis
    // partindo daquela sb com outras sbs como destino
    //--------------------------------------------------------------------------
    paths := make(map[int]map[int][]Path)

    fmt.Println("locoLocations: ", locoLocations)

    for sbIndex, _ := range locoLocations {
        paths[sbIndex] = make(map[int][]Path)

        for j := range d.Graph.Nodes {
            iNode := d.Graph.Nodes[j]
            if j == sbIndex || iNode.Side != 0 {continue}
            paths[sbIndex][j] = GetAllPaths(d.Graph, s, sbIndex, j)
        }
    }

    //spew.Dump(paths)

    // Para cada locomotiva em cada sb, construir todas as composições possíveis
    //--------------------------------------------------------------------------
    possibleCompositions := make(map[int][]CompositionRange)

    for sbIndex, locoPositions := range locoLocations {
        possibleCompositions[sbIndex] = make([]CompositionRange, 0)

        for _, pos := range locoPositions {
            // Composições que incluem 1o material rodante
            for p := pos; p < len(s[sbIndex]); p++ {
                compositionRange := CompositionRange{0, p}
                possibleCompositions[sbIndex] = append(possibleCompositions[sbIndex], compositionRange)
            }

            // Composições que incluem o último material rodante
            for p := pos; p >= 0; p-- {
                compositionRange := CompositionRange{p, len(s[sbIndex])-1}
                possibleCompositions[sbIndex] = append(possibleCompositions[sbIndex], compositionRange)
            }
        }
    }

    fmt.Println("possibleCompositions", possibleCompositions)

    // Para cada composição possível em cada sb, construir todas as manobras
    // possíveis
    //--------------------------------------------------------------------------
    maneuvers := make([]Maneuver, 0)

    for sbIndex, compList := range possibleCompositions {
        for _, compositionRange := range compList {
            for targetIndex, pathsToTarget := range paths[sbIndex] {
                for _, path := range pathsToTarget {
                    compositionAssets := compositionRange.LastAsset - compositionRange.FirstAsset + 1
                    if compositionAssets > path.MaxCompositionLength {continue}

                    firstNodeId := path.Nodes[1]
                    firstNode, _ := GetNode(d.Graph.Nodes, firstNodeId)
                    if firstNode.Side == 'A' {
                        if compositionRange.FirstAsset > 0 {
                            continue
                        }
                    } else {
                        if compositionRange.LastAsset < len(s[sbIndex])-1 {
                            continue
                        }
                    }

                    maneuver := Maneuver{}
                    maneuver.Composition = s[sbIndex][compositionRange.FirstAsset:compositionRange.LastAsset+1]
                    maneuver.Path = path
                    maneuver.Cost = float64(len(path.Nodes))

                    maneuver.EndState = CopyState(s)

                    maneuver.EndState[sbIndex] = append(
                        maneuver.EndState[sbIndex][0:compositionRange.FirstAsset],
                        maneuver.EndState[sbIndex][compositionRange.LastAsset+1:len(maneuver.EndState[sbIndex])]...)

                    secondLastNodeId := path.Nodes[len(path.Nodes)-2]
                    secondLastNode, _ := GetNode(d.Graph.Nodes, secondLastNodeId)

                    if secondLastNode.Side == 'A' {
                        maneuver.EndState[targetIndex] = append(maneuver.Composition, maneuver.EndState[targetIndex]...)
                    } else {
                        maneuver.EndState[targetIndex] = append(maneuver.EndState[targetIndex], maneuver.Composition...)
                    }

                    maneuvers = append(maneuvers, maneuver)
                }
            }
        }
    }

    spew.Dump(maneuvers)

    return maneuvers
}

func main() {
    configFile, _ := os.Open("local/config.json")
    defer configFile.Close()

    config := Config{}
    json.NewDecoder(configFile).Decode(&config)

    d := CreateData(config)
    s0 := CopyState(d.InitialState)

    fmt.Println(s0)

    fmt.Println(d.RollingStock)

    GetPossibleManeuvers(d, s0)

    _ = d
    _ = s0
}







