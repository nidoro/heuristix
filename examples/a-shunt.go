// TODO: ler e utilizar "lengths"

// TODO: Otimizar checagem de caminhos que "atravessam" material rodante
// estacionado
// TODO: Construir matriz de custo não-unitária
package main

import (
    _ "math"
    _ "math/rand"
    _ "time"
    "sort"
    "github.com/davecgh/go-spew/spew"
    "fmt"
    _ "github.com/nidoro/heuristix"
    "encoding/json"
    "os"

    "image/color"
    "gonum.org/v1/plot"
    "gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/vg"
    "container/heap"
    "encoding/binary"
    "hash/fnv"
)

var _ = spew.Dump

type RollingStockRow struct {
    RollingStock    []int
    Positioning     []int
}

type State struct {
    SBs                 map[int][]int
    AssetLocation       []int
    Rows                []RollingStockRow
    Hash                uint64
}

type RollingStock struct {
    Id          int `json:"id"`
    HorsePower  int `json:"hp"`
}

type Config struct {
    Edges           [][]string          `json:"edges"`
    RollingStock    []RollingStock      `json:"rolling-stock"`
    //InitialState    map[string][]int    `json:"initial-state"`
    //TargetState     map[string][]int    `json:"target-state"`

    InitialState    []map[string][]int  `json:"initial-state"`
    TargetState     []map[string][]int  `json:"target-state"`
    Lengths         map[string]float64  `json:"lengths"`
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
    Length float64
}

func HashState(s State) uint64 {
    h := fnv.New64a()

    for _, row := range s.Rows {
        for _, v := range row.RollingStock {
            // convert int -> byte representation
            b := make([]byte, 8)
            binary.LittleEndian.PutUint64(b, uint64(v))
            h.Write(b)
        }
        // separator to avoid collisions between concatenated slices
        h.Write([]byte{0xff})

        for _, v := range row.Positioning {
            b := make([]byte, 8)
            binary.LittleEndian.PutUint64(b, uint64(v))
            h.Write(b)
        }
        h.Write([]byte{0xfe})
    }

    return h.Sum64()
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
    DistMatrix [][]float64
}

type Path struct {
    Nodes                   []string
    MaxCompositionLength    int
    OrientationChanges      int
    Cost                    float64
    Length                  float64
}

type ByCost []Path
func (a ByCost) Len() int           { return len(a) }
func (a ByCost) Less(i, j int) bool { return a[i].Length < a[j].Length }
func (a ByCost) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type Data struct {
    Graph           Graph
    RollingStock    []RollingStock
    InitialState    State
    TargetState     State
    DistMatrix      [][]float64
    Paths           [][][]Path
}

func Assert(cond bool, msg string) {
    if !cond {
        panic("Assertion failed: " + msg)
    }
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
            g.Nodes = append(g.Nodes, Node{Id: uParentId, ParentIndex: -1, AggregatorId: uParentId, AggregatorIndex: uParentIndex})
        }

        if vParentIndex < 0 {
            vParentIndex = len(g.Nodes)
            g.Nodes = append(g.Nodes, Node{Id: vParentId, ParentIndex: -1, AggregatorId: vParentId, AggregatorIndex: vParentIndex})
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

    for nodeId, length := range config.Lengths {
        node, _ := GetNode(g.Nodes, nodeId)
        node.Length = length
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

func CreateDistMatrix(g *Graph) {
    n := len(g.Nodes)
    dist := make([][]float64, n)

    // initialize distances
    for i := range dist {
        iNode := g.Nodes[i]

        dist[i] = make([]float64, n)
        for j := range dist[i] {
            jNode := g.Nodes[j]

            if i == j || ((iNode.Side == 0 && jNode.AggregatorIndex == i) || jNode.Side == 0 && iNode.AggregatorIndex == j) {
                dist[i][j] = 0
            } else if g.Edges[i][j] > 0 {
                dist[i][j] = 1 // edge cost
            } else {
                dist[i][j] = -1 // unreachable
            }
        }
    }

    // Floyd–Warshall relaxation
    for k := 0; k < n; k++ {
        for i := 0; i < n; i++ {
            if dist[i][k] == -1 {
                continue
            }

            for j := 0; j < n; j++ {
                if dist[k][j] == -1 {
                    continue
                }
                newDist := dist[i][k] + dist[k][j]
                if dist[i][j] == -1 || newDist < dist[i][j] {
                    dist[i][j] = newDist
                }
            }
        }
    }

    g.DistMatrix = dist
}

func CreateData(config Config) Data {
    var d Data

    d.Graph = CreateGraph(config)
    PrintEdges(d.Graph)

    d.RollingStock = config.RollingStock

    // d.InitialState.SBs = make(map[int][]int)
    // d.InitialState.AssetLocation = make([]int, len(d.RollingStock))
    // d.TargetState.SBs = make(map[int][]int)
    // d.TargetState.AssetLocation = make([]int, len(d.RollingStock))

    d.InitialState.Rows = make([]RollingStockRow, len(config.InitialState))
    d.TargetState.Rows = make([]RollingStockRow, len(config.TargetState))

    for k := range config.InitialState {
        d.InitialState.Rows[k].RollingStock = config.InitialState[k]["row"]
        d.InitialState.Rows[k].Positioning = make([]int, len(config.InitialState[k]["positioning"]))

        for p, nodeId := range config.InitialState[k]["positioning"] {
            d.InitialState.Rows[k].Positioning[p] = GetNodeIndex(d.Graph.Nodes, fmt.Sprintf("%d", nodeId))
        }
    }

    for k := range config.TargetState {
        d.TargetState.Rows[k].RollingStock = config.TargetState[k]["row"]
        d.TargetState.Rows[k].Positioning = make([]int, len(config.TargetState[k]["positioning"]))

        for p, nodeId := range config.TargetState[k]["positioning"] {
            d.TargetState.Rows[k].Positioning[p] = GetNodeIndex(d.Graph.Nodes, fmt.Sprintf("%d", nodeId))
        }
    }

    d.InitialState.Hash = HashState(d.InitialState)
    d.TargetState.Hash = HashState(d.TargetState)

    CreateDistMatrix(&d.Graph)

    // Todos os caminhos de todos para todos
    //-----------------------------------------
    d.Paths = make([][][]Path, len(d.Graph.Nodes))

    for i := range d.Graph.Nodes {
        d.Paths[i] = make([][]Path, len(d.Graph.Nodes))
        for j := range d.Graph.Nodes {
            d.Paths[i][j] = GetAllPaths(d.Graph, i, j)
        }
    }

    return d
}

func CopyState(s1 State) State {
    var s2 State

    // s2.SBs = make(map[int][]int)
    //
    // for k, v := range s1.SBs {
    //     s2.SBs[k] = make([]int, len(v))
    //     copy(s2.SBs[k], s1.SBs[k])
    // }
    //
    // s2.AssetLocation  = make([]int, len(s1.AssetLocation))
    // copy(s2.AssetLocation, s1.AssetLocation)

    s2.Rows = make([]RollingStockRow, len(s1.Rows))

    for i := range s1.Rows {
        s2.Rows[i].RollingStock = make([]int, len(s1.Rows[i].RollingStock))
        s2.Rows[i].Positioning = make([]int, len(s1.Rows[i].Positioning))
        copy(s2.Rows[i].RollingStock, s1.Rows[i].RollingStock)
        copy(s2.Rows[i].Positioning , s1.Rows[i].Positioning)
    }

    s2.Hash = s1.Hash

    return s2
}

func CopyRollingStockRow(row RollingStockRow) RollingStockRow {
    var result RollingStockRow

    result.RollingStock = make([]int, len(row.RollingStock))
    result.Positioning = make([]int, len(row.Positioning))
    copy(result.RollingStock, row.RollingStock)
    copy(result.Positioning, row.Positioning)

    return result
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

func MinFloat(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

func GetAllPaths(g Graph, sourceIndex int, targetIndex int) []Path {
    if sourceIndex < 0 || targetIndex < 0 {
        return nil
    }

    var paths []Path
    var path []int
    visited := make([]bool, len(g.Nodes))

    var dfs func(u int)
    dfs = func(u int) {
        path = append(path, u)
        visited[u] = true

        if u == targetIndex {
            pt := Path{}
            pt.Nodes = make([]string, len(path))

            for pos, nodeIndex := range path {
                pt.Nodes[pos] = g.Nodes[nodeIndex].Id
            }

            // Custo
            //--------------
            // for k := 0; k < len(path)-1; k++ {
            //     u := path[k]
            //     v := path[k+1]
            //     uNode := g.Nodes[u]
            //     vNode := g.Nodes[v]
            //
            //     if (uNode.Side != 0 && uNode.AggregatorIndex == v) || (uNode.Side == 0 && vNode.AggregatorIndex == u) {
            //         // no cost, dummy edge
            //     } else {
            //         pt.Cost += 1
            //     }
            // }

            // Comprimento do caminho
            //----------------------------------
            for k := 0; k < len(path); k++ {
                nd := g.Nodes[path[k]]
                pt.Length += nd.Length
                if nd.Side != 0 && k > 1 && k < len(path)-1 {
                    if GetNodeIndex(g.Nodes, g.Nodes[path[k-1]].Id) != nd.AggregatorIndex && GetNodeIndex(g.Nodes, g.Nodes[path[k+1]].Id) != nd.AggregatorIndex {
                        pt.Length += g.Nodes[nd.AggregatorIndex].Length
                    }
                }
            }

            // Mudanças de orientação
            //---------------------------
            for k := 1; k < len(path)-1; k++ {
                a := path[k-1]
                b := path[k]
                c := path[k+1]
                if g.Nodes[b].Side != 0 && a != g.Nodes[b].AggregatorIndex && c != g.Nodes[b].AggregatorIndex {
                    pt.OrientationChanges += 1
                }
            }

            paths = append(paths, pt)
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

    sort.Sort(ByCost(paths))
    return paths

    /*
    results := make([]Path, 0, len(paths))

    for _, path := range paths {
        skip := false

        for k := 0; k < len(path.Nodes)-2; k++ {
            kIndex := GetNodeIndex(g.Nodes, path.Nodes[k])
            lIndex := GetNodeIndex(g.Nodes, path.Nodes[k+2])

            uNode := g.Nodes[kIndex]
            vNode := g.Nodes[lIndex]

            _, ok := s.SBs[uNode.AggregatorIndex]

            if ok && len(s.SBs[uNode.AggregatorIndex]) > 0 {
                if uNode.Side != 0 && vNode.Side != 0 && uNode.AggregatorIndex == vNode.AggregatorIndex {
                    skip = true
                }

                if uNode.AggregatorIndex != sourceIndex {
                    path.MaxCompositionLength = Min(path.MaxCompositionLength, g.Nodes[uNode.AggregatorIndex].Length - len(s.SBs[uNode.AggregatorIndex]))
                }
            }
        }

        if !skip {
            results = append(results, path)
        }
    }

    return results*/
}

type Maneuver struct {
    Composition     []int
    Row             RollingStockRow
    Path            Path
    ManeuverCost    float64
    AccumCost       float64
    TotalCostEstimate float64
    EndState        State
    // HACK: para debug
    ExtraInfo       string

    // Árvore
    Children        []*Maneuver
    Parent          *Maneuver
}

type ManeuverHeap []*Maneuver

func (h ManeuverHeap) Len() int           { return len(h) }
func (h ManeuverHeap) Less(i, j int) bool { return h[i].TotalCostEstimate < h[j].TotalCostEstimate }
func (h ManeuverHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *ManeuverHeap) Push(x interface{}) {
    *h = append(*h, x.(*Maneuver))
}

func (h *ManeuverHeap) Pop() interface{} {
    old := *h
    n := len(old)
    maneuver := old[n-1]
    *h = old[0 : n-1]
    return maneuver
}

type CompositionRange struct {
    FirstAsset  int
    LastAsset   int
}

func GetDistanceToTargetState(d Data, s1 State) float64 {
    // fmt.Println("----------------------------")
    // fmt.Println("Distance:")
    // fmt.Println("From:", s1)
    // fmt.Println("To  :", d.TargetState)

    result := 0.0
    s2 := d.TargetState

    for r1, row1 := range s1.Rows {
        _ = r1
        a := row1.Positioning[0]
        b := row1.Positioning[len(row1.Positioning)-1]

        for r2, row2 := range s2.Rows {
            _ = r2
            for _, assetIndex := range row1.RollingStock {
                //fmt.Println(assetIndex, "vs", row2.RollingStock)
                if Contains(row2.RollingStock, assetIndex) {
                    u := row2.Positioning[0]
                    v := row2.Positioning[len(row2.Positioning)-1]

                    minCost := 9999.0
                    if a == u || a == v || b == u || b == v {
                        minCost = 0
                    } else {
                        if a != u {minCost = MinFloat(minCost, d.Paths[a][u][0].Length)}
                        if a != v {minCost = MinFloat(minCost, d.Paths[a][v][0].Length)}
                        if b != u {minCost = MinFloat(minCost, d.Paths[b][u][0].Length)}
                        if b != v {minCost = MinFloat(minCost, d.Paths[b][v][0].Length)}
                    }

                    //fmt.Println(r1, r2, minCost)

                    result += minCost

                    break
                }
            }
        }
    }

    //fmt.Println(result)

    return result
}

func Contains(slice []int, value int) bool {
    for _, v := range slice {
        if v == value {
            return true
        }
    }
    return false
}

func Reverse(nums []int) {
    for i, j := 0, len(nums)-1; i < j; i, j = i+1, j-1 {
        nums[i], nums[j] = nums[j], nums[i]
    }
}

func EqualSlices(a []int, b []int) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

func EqualRollingStockRows(r1 RollingStockRow, r2 RollingStockRow) bool {
    return EqualSlices(r1.RollingStock, r2.RollingStock) && EqualSlices(r1.Positioning, r2.Positioning)
}

func EqualStates(s1 State, s2 State) bool {
    if s1.Hash != s2.Hash {return false}

    if len(s1.Rows) != len(s2.Rows) {
        return false
    }

    for _, r1 := range s1.Rows {
        found := false
        for _, r2 := range s2.Rows {
            if EqualRollingStockRows(r1, r2) {
                found = true
                break
            }
        }

        if !found {
            return false
        }
    }

    return true
}

func GetRollingStockRow(state State, assetIndex int) *RollingStockRow {
    for r, row := range state.Rows {
        if Contains(row.RollingStock, assetIndex) {
            return &state.Rows[r]
        }
    }
    return nil
}

func EqualAssetLocations(s1 State, s2 State) bool {
    for _, row1 := range s1.Rows {
        for _, assetIndex := range row1.RollingStock {
            row2 := GetRollingStockRow(s2, assetIndex)
            ok := false
            for _, k := range row1.Positioning {
                if Contains(row2.Positioning, k) {
                    ok = true
                    break
                }
            }
            if !ok {return false}
        }
    }

    return true
}

func RepeatingState(maneuver *Maneuver) bool {
    anscestor := maneuver.Parent
    for anscestor != nil {
        if EqualStates(maneuver.EndState, anscestor.EndState) {
            return true
        }

        anscestor = anscestor.Parent
    }
    return false
}

func AppendReversed(dst []int, src []int) []int {
    for i := len(src) - 1; i >= 0; i-- {
        dst = append(dst, src[i])
    }
    return dst
}

func PrependReversed(dst []int, src []int) []int {
    result := make([]int, 0, len(dst)+len(src))

    for i := len(src) - 1; i >= 0; i-- {
        result = append(result, src[i])
    }

    result = append(result, dst...)

    return result
}

func PrependMany(slice []int, values ...int) []int {
    return append(values, slice...)
}

func GetManeuverWithLowestTotalCostEstimate(maneuvers []*Maneuver) (*Maneuver, int) {
    idx := 0
    result := maneuvers[idx]

    for i, maneuver := range maneuvers {
        if maneuver.TotalCostEstimate < result.TotalCostEstimate {
            result = maneuver
            idx = i
        }
    }

    return result, idx
}

func PopManeuverWithLowestTotalCostEstimate(maneuvers *ManeuverHeap) *Maneuver {
    result := heap.Pop(maneuvers).(*Maneuver)
    //result, idx := GetManeuverWithLowestTotalCostEstimate(*maneuvers)
    //*maneuvers = append((*maneuvers)[0:idx], (*maneuvers)[idx+1:]...)
    return result
}

func GetSideNode(d Data, aggIndex int, side byte) *Node {
    for k := range d.Graph.Nodes {
        if d.Graph.Nodes[k].AggregatorIndex == aggIndex && d.Graph.Nodes[k].Side == side {
            return &d.Graph.Nodes[k]
        }
    }

    return nil
}

func GetOppositeSide(d Data, agg1 int, agg2 int) *Node {
    path := d.Paths[agg1][agg2][0]
    //fmt.Println(agg1, agg2, path)
    wrongSide, _ := GetNode(d.Graph.Nodes, path.Nodes[1])
    side := 'A'
    if wrongSide.Side == 'A' {side = 'B'}
    return GetSideNode(d, agg1, byte(side))
}

func GetTotalLength(d Data, nodes []int) float64 {
    result := 0.0
    for _, nodeIndex := range nodes {
        result += d.Graph.Nodes[nodeIndex].Length
    }
    return result
}

func ExtendPath(d Data, path *Path, positioning []int) {
    // fmt.Println()
    // fmt.Println(path.Nodes)
    //
    // fmt.Print("[")
    // for _, k := range positioning {
    //     fmt.Printf(" %s", d.Graph.Nodes[k].Id)
    // }
    // fmt.Println(" ]")

    for _, v := range positioning {
        u := GetNodeIndex(d.Graph.Nodes, path.Nodes[len(path.Nodes)-1])
        if u == v {continue}

        extension := d.Paths[u][v][0]
        for _, k := range extension.Nodes[1:] {
            path.Nodes = append(path.Nodes, k)
            kNode, _ := GetNode(d.Graph.Nodes, k)
            path.Length += kNode.Length
        }
    }
    // fmt.Println(path)
}

func GetPossibleManeuvers(d Data, parent *Maneuver, visited map[uint64][]Maneuver) []*Maneuver {
    s := parent.EndState

    // Localizar fileiras móveis (fileiras com locomotivas)
    //---------------------------------------------------------
    movableRows := make([]*RollingStockRow, 0, len(s.Rows))

    for r, row := range s.Rows {
        for _, assetId := range row.RollingStock {
            // TODO: por enquanto, assetId == assetIndex
            if d.RollingStock[assetId].HorsePower > 0 {
                movableRows = append(movableRows, &s.Rows[r])
            }
        }
    }

    // Para cada fileira móvel, construir todos os caminhos possíveis
    // das pontas da fileira para outras SBs
    //--------------------------------------------------------------------------
    // TODO: Na verdade, é melhor isso já ficar pré-calculado
    // paths := make(map[int][][]Path)
    //
    // for sbIndex, _ := range locoLocations {
    //     paths[sbIndex] = make(map[int][]Path)
    //
    //     for j := range d.Graph.Nodes {
    //         iNode := d.Graph.Nodes[j]
    //         if j == sbIndex || iNode.Side != 0 {continue}
    //         paths[sbIndex][j] = GetAllPaths(d.Graph, s, sbIndex, j)
    //     }
    // }

    // Para cada fileira móvel, construir todas as composições possíveis
    //--------------------------------------------------------------------------
    possibleCompositions := make([][][]CompositionRange, len(movableRows))

    for r, row := range movableRows {
        possibleCompositions[r] = make([][]CompositionRange, 2)

        for pos, assetIndex := range row.RollingStock {
            if d.RollingStock[assetIndex].HorsePower == 0 {continue}

            comps := make([]CompositionRange, 0, 2*len(row.RollingStock))

            // Composições que incluem 1o material rodante
            for p := pos; p < len(row.RollingStock); p++ {
                compositionRange := CompositionRange{0, p}
                comps = append(comps, compositionRange)
            }

            possibleCompositions[r][0] = comps

            comps = make([]CompositionRange, 0, 2*len(row.RollingStock))

            // Composições que incluem o último material rodante
            for p := pos; p >= 0; p-- {
                compositionRange := CompositionRange{p, len(row.RollingStock)-1}
                comps = append(comps, compositionRange)
            }

            possibleCompositions[r][1] = comps
        }
    }

    // Para cada composição possível, construir todas as manobras possíveis
    //--------------------------------------------------------------------------
    SIDES := [2]byte{'A', 'B'}

    maneuvers := make([]*Maneuver, 0)

    for movableIndex := range possibleCompositions {
        row := movableRows[movableIndex]

        //fmt.Println(row)

        for direction := 0; direction <= 1; direction++ {
            fromIndex := row.Positioning[direction*(len(row.Positioning)-1)]

            for _, compositionRange := range possibleCompositions[movableIndex][direction] {
                composition := row.RollingStock[compositionRange.FirstAsset:compositionRange.LastAsset+1]

                rightStart := GetSideNode(d, row.Positioning[0], SIDES[direction])

                if len(row.Positioning) > 1 {
                    //fmt.Println("row.Positioning", row.Positioning)
                    //fmt.Println("direction", direction, direction*(len(possibleCompositions[movableIndex])-1) - 2*direction + 1)
                    wrongStart := row.Positioning[direction*(len(possibleCompositions[movableIndex])-1) - 2*direction + 1]
                    //fmt.Println("wrongStart", d.Graph.Nodes[wrongStart].Id)
                    rightStart = GetOppositeSide(d, fromIndex, wrongStart)
                    //fmt.Println("rightStart", rightStart.Id)
                }

                if rightStart == nil {continue}

                for toIndex := range d.Graph.Nodes {
                    if fromIndex == toIndex {continue}
                    if d.Graph.Nodes[toIndex].Side != 0 {continue}

                    for _, path := range d.Paths[fromIndex][toIndex] {
                        if path.OrientationChanges > 0 {continue}

                        // Checar se caminho "atravessa" material rodante estacionado
                        // e se composição termina engatando em outra fileira
                        //--------------------------------------------------------------
                        invalidPath := false
                        var mergedRow *RollingStockRow
                        mergeDirection := -1

                        for p, uNodeId := range path.Nodes {
                            if p == 0 {continue}
                            u := GetNodeIndex(d.Graph.Nodes, uNodeId)

                            for r, _ := range s.Rows {
                                if row == &s.Rows[r] {continue}

                                for p2, v := range s.Rows[r].Positioning {
                                    if v != u {continue}
                                    if p == len(path.Nodes)-1 {

                                        if len(s.Rows[r].Positioning) == 1 {
                                            if d.Graph.Nodes[GetNodeIndex(d.Graph.Nodes, path.Nodes[p-1])].Side == 'A' {
                                                mergedRow = &s.Rows[r]
                                                mergeDirection = 0
                                            } else {
                                                mergedRow = &s.Rows[r]
                                                mergeDirection = 1
                                            }
                                        } else if p2 == 0 {
                                            freeSide := GetOppositeSide(d, v, s.Rows[r].Positioning[p2+1])
                                            if path.Nodes[p-1] == freeSide.Id {
                                                mergedRow = &s.Rows[r]
                                                mergeDirection = 0
                                            } else {
                                                invalidPath = true
                                            }
                                        } else if p2 == len(s.Rows[r].Positioning)-1 {
                                            freeSide := GetOppositeSide(d, v, s.Rows[r].Positioning[p2-1])
                                            if path.Nodes[p-1] == freeSide.Id {
                                                mergedRow = &s.Rows[r]
                                                mergeDirection = 1
                                            } else {
                                                invalidPath = true
                                            }
                                        } else {
                                            invalidPath = true
                                        }
                                    } else {
                                        invalidPath = true
                                    }
                                    break
                                }

                                if invalidPath {break}
                            }

                            if invalidPath {break}
                        }

                        if invalidPath {continue}

                        // Checar se caminho começa indo para o lado errado
                        if path.Nodes[1] != rightStart.Id {continue}

                        // Criar manobra
                        maneuver := new(Maneuver)
                        maneuver.Parent = parent
                        maneuver.Row = CopyRollingStockRow(*row)
                        maneuver.Row.RollingStock = make([]int, len(composition))
                        copy(maneuver.Row.RollingStock, composition)
                        maneuver.Path = path
                        maneuver.ManeuverCost = path.Length
                        maneuver.AccumCost = parent.AccumCost + maneuver.ManeuverCost

                        // Estado final resultante da manobra
                        maneuver.EndState = CopyState(s)

                        //second, _ := GetNode(d.Graph.Nodes, path.Nodes[1])
                        secondLast, _ := GetNode(d.Graph.Nodes, path.Nodes[len(path.Nodes)-2])

                        // Inserir fileira movida no último nó do caminho e anteriores
                        destEntrance := secondLast
                        _ = destEntrance
                        comp := make([]int, 0, len(composition))

                        // Colocar material rodante na ordem em que vão chegar no último nó

                        if direction == 0 {
                            if mergedRow != nil {
                                endPositioning := make([]int, len(mergedRow.Positioning))
                                copy(endPositioning, mergedRow.Positioning)

                                if mergeDirection == 0 {
                                    comp = AppendReversed(comp, mergedRow.RollingStock)
                                } else if mergeDirection == 1 {
                                    comp = append(comp, mergedRow.RollingStock...)
                                    Reverse(endPositioning)
                                }

                                ExtendPath(d, &path, endPositioning)
                            }
                            comp = append(comp, composition...)
                        } else {
                            comp = append(comp, composition...)
                            Reverse(comp)

                            if mergedRow != nil {
                                endPositioning := make([]int, len(mergedRow.Positioning))
                                copy(endPositioning, mergedRow.Positioning)

                                if mergeDirection == 0 {
                                    comp = PrependReversed(comp, mergedRow.RollingStock)
                                } else if mergeDirection == 1 {
                                    comp = PrependMany(comp, mergedRow.RollingStock...)
                                    Reverse(endPositioning)
                                }

                                ExtendPath(d, &path, endPositioning)
                            }
                        }

                        // if mergedRow != nil {
                        //     if mergeDirection == 0 {
                        //         if direction == 0 {
                        //             comp = append(comp, mergedRow.RollingStock...)
                        //             comp = PrependReversed(comp, composition)
                        //         } else {
                        //             comp = append(comp, composition...)
                        //             comp = append(comp, mergedRow.RollingStock...)
                        //         }
                        //         ExtendPath(d, &path, mergedRow.Positioning)
                        //     } else {
                        //         if direction == 1 {
                        //             comp = append(comp, mergedRow.RollingStock...)
                        //             comp = AppendReversed(comp, composition)
                        //         } else {
                        //             comp = append(comp, mergedRow.RollingStock...)
                        //             comp = append(comp, composition...)
                        //         }
                        //         revPositioning := make([]int, len(mergedRow.Positioning))
                        //         copy(revPositioning, mergedRow.Positioning)
                        //         Reverse(revPositioning)
                        //         ExtendPath(d, &path, revPositioning)
                        //     }
                        // } else {
                        //     comp = append(comp, composition...)
                        // }
                        //
                        // if direction == 0 {
                        //     Reverse(comp)
                        // }

                        var newRow RollingStockRow
                        newRow.RollingStock = make([]int, len(comp))
                        copy(newRow.RollingStock, comp)

                        remainingCompositionLength := float64(len(comp))
                        newRow.Positioning = make([]int, 0, len(path.Nodes))

                        for k := len(path.Nodes)-1; k >= 1; k-- {
                            node, nodeIndex := GetNode(d.Graph.Nodes, path.Nodes[k])
                            if node.Side != 0 {continue}

                            newRow.Positioning = append(newRow.Positioning, nodeIndex)
                            destEntrance, _ = GetNode(d.Graph.Nodes, path.Nodes[k-1])
                            remainingCompositionLength -= node.Length
                            if remainingCompositionLength <= 0 {break}
                        }

                        if remainingCompositionLength > 0 {continue}

                        if len(newRow.Positioning) == 1 && secondLast.Side == 'A' {
                            Reverse(newRow.RollingStock)
                        }

                        maneuver.ExtraInfo += fmt.Sprintf("direction: %d\n", direction)
                        maneuver.ExtraInfo += fmt.Sprintf("secondLast.Side: %c", secondLast.Side)

                        maneuver.EndState.Rows = append(maneuver.EndState.Rows, newRow)

                        // if mergedRow == nil {
                        //     var newRow RollingStockRow
                        //     newRow.RollingStock = make([]int, len(comp))
                        //     copy(newRow.RollingStock, comp)
                        //
                        //     remainingCompositionLength := float64(len(comp))
                        //     newRow.Positioning = make([]int, 0, len(path.Nodes))
                        //
                        //     destEntrance := secondLast
                        //
                        //     for k := len(path.Nodes)-1; k >= 1; k-- {
                        //         node, nodeIndex := GetNode(d.Graph.Nodes, path.Nodes[k])
                        //         if node.Side != 0 {continue}
                        //
                        //         newRow.Positioning = append(newRow.Positioning, nodeIndex)
                        //         destEntrance, _ = GetNode(d.Graph.Nodes, path.Nodes[k-1])
                        //         remainingCompositionLength -= node.Length
                        //         if remainingCompositionLength <= 0 {break}
                        //     }
                        //
                        //     if remainingCompositionLength > 0 {continue}
                        //
                        //     Reverse(newRow.Positioning)
                        //
                        //     if SIDES[direction] == destEntrance.Side {
                        //         Reverse(newRow.RollingStock)
                        //     }
                        //
                        //     maneuver.EndState.Rows = append(maneuver.EndState.Rows, newRow)
                        // } else {
                        //     var newRow RollingStockRow
                        //
                        //     comp := make([]int, len(composition))
                        //     copy(comp, composition)
                        //     if len(row.Positioning) == 1 && second.Side == secondLast.Side {
                        //         Reverse(comp)
                        //     }
                        //
                        //     newRow.RollingStock = make([]int, 0, len(comp) + len(mergedRow.RollingStock))
                        //     //fmt.Println("mergeDirection", mergeDirection)
                        //     if mergeDirection == 0 {
                        //         newRow.RollingStock = append(newRow.RollingStock, comp...)
                        //         newRow.RollingStock = append(newRow.RollingStock, mergedRow.RollingStock...)
                        //     } else {
                        //         newRow.RollingStock = append(newRow.RollingStock, mergedRow.RollingStock...)
                        //         newRow.RollingStock = append(newRow.RollingStock, comp...)
                        //     }
                        //
                        //     availableSpace := GetTotalLength(d, mergedRow.Positioning) - float64(len(mergedRow.RollingStock))
                        //     remainingCompositionLength := float64(len(comp)) - availableSpace
                        //     newRow.Positioning = make([]int, 0, len(mergedRow.Positioning) + len(path.Nodes))
                        //     newRow.Positioning = append(newRow.Positioning, mergedRow.Positioning...)
                        //
                        //     positioningExtension := make([]int, 0, len(path.Nodes))
                        //
                        //     if remainingCompositionLength > 0 {
                        //         // Até o penúltimo nó do caminho
                        //         for k := len(path.Nodes)-2; k >= 1; k-- {
                        //             node, nodeIndex := GetNode(d.Graph.Nodes, path.Nodes[k])
                        //             if node.Side != 0 {continue}
                        //
                        //             positioningExtension = append(positioningExtension, nodeIndex)
                        //             remainingCompositionLength -= node.Length
                        //             if remainingCompositionLength <= 0 {break}
                        //         }
                        //     }
                        //
                        //     if remainingCompositionLength > 0 {continue}
                        //
                        //     //fmt.Println("positioningExtension", positioningExtension)
                        //
                        //     if mergeDirection == 0 {
                        //         Reverse(positioningExtension)
                        //         newRow.Positioning = append(positioningExtension, newRow.Positioning...)
                        //     } else {
                        //         newRow.Positioning = append(newRow.Positioning, positioningExtension...)
                        //     }
                        //
                        //     //fmt.Println("newRow.RollingStock", newRow.RollingStock)
                        //     //fmt.Println("newRow.Positioning", newRow.Positioning)
                        //
                        //     maneuver.EndState.Rows = append(maneuver.EndState.Rows, newRow)
                        // }

                        // Adicionar fileira não movimentada
                        if len(composition) < len(row.RollingStock) {
                            nonMoved := row.RollingStock[0:compositionRange.FirstAsset]

                            if compositionRange.FirstAsset == 0 {
                                nonMoved = row.RollingStock[compositionRange.LastAsset+1:]
                            }

                            var newRow RollingStockRow
                            newRow.RollingStock = make([]int, len(nonMoved))
                            copy(newRow.RollingStock, nonMoved)

                            remainingCompositionLength := float64(len(nonMoved))
                            newRow.Positioning = make([]int, 0, len(row.Positioning))

                            if compositionRange.FirstAsset == 0 {
                                for k := len(row.Positioning) - 1; k >= 0; k-- {
                                    nodeIndex := row.Positioning[k]
                                    remainingCompositionLength -= d.Graph.Nodes[nodeIndex].Length
                                    newRow.Positioning = append(newRow.Positioning, row.Positioning[k])
                                    if remainingCompositionLength <= 0 {break}
                                }
                                Reverse(newRow.Positioning)
                            } else {
                                for k := 0; k < len(row.Positioning); k++ {
                                    nodeIndex := row.Positioning[k]
                                    remainingCompositionLength -= d.Graph.Nodes[nodeIndex].Length
                                    newRow.Positioning = append(newRow.Positioning, row.Positioning[k])
                                    if remainingCompositionLength <= 0 {break}
                                }
                            }

                            maneuver.EndState.Rows = append(maneuver.EndState.Rows, newRow)
                        }

                        // Remover do estado final a fileira movida
                        f := 0
                        for f = 0; f < len(maneuver.EndState.Rows); f++ {
                            if EqualRollingStockRows(*row, maneuver.EndState.Rows[f]) {
                                break
                            }
                        }

                        maneuver.EndState.Rows = append(maneuver.EndState.Rows[:f], maneuver.EndState.Rows[f+1:]...)

                        // Remover do estado final a fileira mesclada
                        if mergedRow != nil {
                            f := 0
                            for f = 0; f < len(maneuver.EndState.Rows); f++ {
                                if EqualRollingStockRows(*mergedRow, maneuver.EndState.Rows[f]) {
                                    break
                                }
                            }

                            maneuver.EndState.Rows = append(maneuver.EndState.Rows[:f], maneuver.EndState.Rows[f+1:]...)
                        }

                        skip := false

                        maneuver.EndState.Hash = HashState(maneuver.EndState)

                        maybeEqual, _ := visited[maneuver.EndState.Hash]

                        for _, vis := range maybeEqual {
                            if EqualStates(maneuver.EndState, vis.EndState) {
                                skip = true
                                break
                            }
                        }

                        if skip {continue}

                        // Calcular TotalCostEstimate
                        //------------------------------
                        maneuver.TotalCostEstimate = maneuver.AccumCost + GetDistanceToTargetState(d, maneuver.EndState)

                        if !ValidState(d, maneuver.EndState) {
                            fmt.Println("------------------------------")
                            fmt.Println("Parent:")
                            PrintState(d, s)
                            fmt.Println("Maneuver:")
                            PrintManeuver(d, maneuver)
                            fmt.Println("Merge direction", mergeDirection)
                            fmt.Println("Composition range", compositionRange)
                            panic("Invalid State!")
                        }

                        maneuvers = append(maneuvers, maneuver)
                    }
                }
            }
        }
    }

    return maneuvers
}

func PrintRollingStockRow(d Data, row RollingStockRow) {
    fmt.Print(row.RollingStock, " : [")
    for i := range row.Positioning {
        fmt.Printf(" %s", d.Graph.Nodes[row.Positioning[i]].Id)
    }
    fmt.Println(" ]")
}

func PrintState(d Data, s State) {
    // g := d.Graph
    // for z, assets := range state.SBs {
    //     if len(assets) == 0 {continue}
    //     fmt.Printf("  %-3s: ", g.Nodes[z].Id)
    //     fmt.Println(assets)
    // }

    for _, row := range s.Rows {
        fmt.Print("  ")
        PrintRollingStockRow(d, row)
    }
}

func PrintManeuver(d Data, maneuver *Maneuver) {
    fmt.Printf ("Row               :  ")
    PrintRollingStockRow(d, maneuver.Row)
    fmt.Println("Path              : ", maneuver.Path.Nodes)
    //fmt.Println("Composition       : ", maneuver.Composition)
    fmt.Println("ManeuverCost      : ", maneuver.ManeuverCost)
    fmt.Println("AccumCost         : ", maneuver.AccumCost)
    fmt.Println("TotalCostEstimate : ", maneuver.TotalCostEstimate)
    fmt.Println("ExtraInfo         : ")
    fmt.Println(maneuver.ExtraInfo)
    fmt.Println("EndState          : ")
    PrintState(d, maneuver.EndState)
}

func PrintManeuverSequence(d Data, maneuver *Maneuver) {
    sequence := make([]*Maneuver, 0)
    sequence = append(sequence, maneuver)
    maneuver = maneuver.Parent
    for maneuver != nil {
        sequence = append(sequence, maneuver)
        maneuver = maneuver.Parent
    }

    for i := len(sequence)-1; i >= 0; i-- {
        PrintManeuver(d, sequence[i])
    }
}

func ValidState(d Data, s State) bool {
    for i := range d.RollingStock {
        count := 0
        for _, row := range s.Rows {
            for _, a := range row.RollingStock {
                if a == i {
                    count += 1
                }
            }
        }

        if count != 1 {fmt.Println("A"); return false}
    }

    for k := range d.Graph.Nodes {
        if d.Graph.Nodes[k].Side != 0 {continue}
        count := 0
        for _, row := range s.Rows {
            for _, u := range row.Positioning {
                if u == k {
                    count += 1
                }
            }
        }

        if count > 1 {fmt.Println("B"); return false}
    }

    for _, row := range s.Rows {
        if len(row.RollingStock) == 0 {return false}
        if len(row.Positioning) == 0 {return false}

        if len(row.Positioning) > 1 {
            a := row.Positioning[0]
            b := row.Positioning[len(row.Positioning)-1]

            pathFound := false

            for _, path := range d.Paths[a][b] {
                pathMainNodes := make([]int, 0, len(path.Nodes))
                for _, k := range path.Nodes {
                    u := GetNodeIndex(d.Graph.Nodes, k)
                    if d.Graph.Nodes[u].Side == 0 {
                        pathMainNodes = append(pathMainNodes, u)
                    }
                }
                if EqualSlices(pathMainNodes, row.Positioning) {
                    pathFound = true
                    break
                }
            }

            if !pathFound {fmt.Println("C"); return false}
        }
    }

    // HACK: Asserção temporária de que o número de SBs ocupadas não pode
    // ser maior que o número de materiais rodantes
    // for _, row := range s.Rows {
    //     if len(row.Positioning) > len(row.RollingStock) {return false}
    // }

    return true
}

func main() {
    configFile, _ := os.Open("local/config5.json")
    defer configFile.Close()

    config := Config{}
    json.NewDecoder(configFile).Decode(&config)

    d := CreateData(config)
    Assert(ValidState(d, d.InitialState), "Invalid initial state!")
    Assert(ValidState(d, d.TargetState), "Invalid target state!")

    fmt.Println("InitialState:")
    PrintState(d, d.InitialState)

    fmt.Println("TargetState:")
    PrintState(d, d.TargetState)

    visited := make(map[uint64][]Maneuver)

    unvisited := &ManeuverHeap{}
    *unvisited = make(ManeuverHeap, 0, 500_000)
    heap.Init(unvisited)

    m0 := new(Maneuver)
    m0.EndState = CopyState(d.InitialState)
    m0.Children = GetPossibleManeuvers(d, m0, visited)

    visited[m0.EndState.Hash] = append(visited[m0.EndState.Hash], *m0)

    for _, child := range m0.Children {
        heap.Push(unvisited, child)
    }

    var bestLeaf *Maneuver
    maxIterations := 20_000_000
    iter := 0

    for unvisited.Len() > 0 && iter <= maxIterations {
        maneuver := PopManeuverWithLowestTotalCostEstimate(unvisited)
        fmt.Printf("\rIteration: %-8d  | unvisited: %-8d | current maneuver heuristic: %-8.0f", iter, unvisited.Len(), maneuver.AccumCost)

        if EqualAssetLocations(maneuver.EndState, d.TargetState) {
            bestLeaf = maneuver
            break
        }

        maneuver.Children = GetPossibleManeuvers(d, maneuver, visited)
        for _, child := range maneuver.Children {
            heap.Push(unvisited, child)
        }

        visited[maneuver.EndState.Hash] = append(visited[maneuver.EndState.Hash], *maneuver)

        iter += 1
    }

    fmt.Println()

    fmt.Println("---------------------------------------------------")
    fmt.Println(" SOLUTION")
    fmt.Println("---------------------------------------------------")

    if bestLeaf != nil {
        PrintManeuverSequence(d, bestLeaf)
    }
}







