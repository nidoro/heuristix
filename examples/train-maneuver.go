package main

import (
    _ "math"
    _ "math/rand"
    _ "time"
    _ "sort"
    _ "github.com/davecgh/go-spew/spew"
    "fmt"
    _ "github.com/nidoro/heuristix"
    "encoding/json"
    "os"

    "image/color"
    "gonum.org/v1/plot"
    "gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/vg"
    "strconv"
)

type State map[int][]int

type RollingStockAsset struct {
    Id          int `json:"id"`
    HorsePower  int `json:"hp"`
}

type Config struct {
    Edges           [][]string          `json:"edges"`
    RollingStock    []RollingStockAsset `json:"rolling-stock"`
    InitialState    map[string][]int    `json:"initial-state"`
    TargetState     map[string][]int    `json:"target-state"`
}

type Node struct {
    Id string
    Side byte
    ParentId string
    ParentIndex int
    X float64
    Y float64
    Length float64
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
    RollingStock    []RollingStockAsset
    InitialState    map[int][]int
    TargetState     map[int][]int
}

type Route struct {
    Id         int
    Order      [] int
    Load       int
    Cost       float64
    VehicleCap int
}

type Solution struct {
    Graph      *Graph
    Routes    [] *Route
    Cost      float64
    NodeRoute [] int
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
            g.Nodes = append(g.Nodes, Node{Id: uParentId, Length: 5})
        }

        if vParentIndex < 0 {
            vParentIndex = len(g.Nodes)
            g.Nodes = append(g.Nodes, Node{Id: vParentId, Length: 5})
        }

        if u == nil {
            g.Nodes = append(g.Nodes, Node{Id: uId, Side: uSide, ParentId: uParentId, ParentIndex: uParentIndex})
        }

        if v == nil {
            g.Nodes = append(g.Nodes, Node{Id: vId, Side: vSide, ParentId: vParentId, ParentIndex: vParentIndex})
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
        sb, _ := strconv.Atoi(k)
        d.InitialState[sb] = v
    }

    for k, v := range config.TargetState {
        sb, _ := strconv.Atoi(k)
        d.TargetState[sb] = v
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

func main() {
    configFile, _ := os.Open("local/config.json")
    defer configFile.Close()

    config := Config{}
    json.NewDecoder(configFile).Decode(&config)

    d := CreateData(config)
    s0 := CopyState(d.InitialState)

    fmt.Println(s0)

    fmt.Println(ShortestPaths(d.Graph, "2"))

    _ = d
    _ = s0
}







