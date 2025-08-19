// TODO: Otimizar checagem de caminhos que "atravessam" material rodante
// estacionado
// TODO: Construir matriz de custo não-unitária
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
    "container/heap"
)

var _ = spew.Dump

type State struct {
    SBs             map[int][]int
    AssetLocation  []int
}

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
    DistMatrix [][]float64
}

type Data struct {
    Graph           Graph
    RollingStock    []RollingStock
    InitialState    State
    TargetState     State
    DistMatrix      [][]float64
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
            g.Nodes = append(g.Nodes, Node{Id: uParentId, Length: 3, ParentIndex: -1, AggregatorId: uParentId, AggregatorIndex: uParentIndex})
        }

        if vParentIndex < 0 {
            vParentIndex = len(g.Nodes)
            g.Nodes = append(g.Nodes, Node{Id: vParentId, Length: 3, ParentIndex: -1, AggregatorId: vParentId, AggregatorIndex: vParentIndex})
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

    d.InitialState.SBs = make(map[int][]int)
    d.InitialState.AssetLocation = make([]int, len(d.RollingStock))
    d.TargetState.SBs = make(map[int][]int)
    d.TargetState.AssetLocation = make([]int, len(d.RollingStock))

    for k, assets := range config.InitialState {
        sbIndex := GetNodeIndex(d.Graph.Nodes, k)
        d.InitialState.SBs[sbIndex] = assets

        for _, assetIndex := range assets {
            // TODO: Assumimos que o id é igual ao índice na lista RollingStock, mas
            // isso deve ser mudado no futuro
            d.InitialState.AssetLocation[assetIndex] = sbIndex
        }
    }

    for k, assets := range config.TargetState {
        sbIndex := GetNodeIndex(d.Graph.Nodes, k)
        d.TargetState.SBs[sbIndex] = assets

        for _, assetIndex := range assets {
            // TODO: Assumimos que o id é igual ao índice na lista RollingStock, mas
            // isso deve ser mudado no futuro
            d.TargetState.AssetLocation[assetIndex] = sbIndex
        }
    }

    CreateDistMatrix(&d.Graph)

    return d
}

func CopyState(s1 State) State {
    var s2 State

    s2.SBs = make(map[int][]int)

    for k, v := range s1.SBs {
        s2.SBs[k] = make([]int, len(v))
        copy(s2.SBs[k], s1.SBs[k])
    }

    s2.AssetLocation  = make([]int, len(s1.AssetLocation))
    copy(s2.AssetLocation, s1.AssetLocation)

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
    OrientationChanges      int
    Cost                    float64
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

    var paths []Path
    var path []int
    visited := make([]bool, len(g.Nodes))

    var dfs func(u int, capSoFar int)
    dfs = func(u int, capSoFar int) {
        path = append(path, u)
        visited[u] = true

        if g.Nodes[u].Side != 0 {
            agg := g.Nodes[u].AggregatorIndex
            free := g.Nodes[agg].Length - len(s.SBs[agg])
            if free < capSoFar { capSoFar = free }
        }

        if u == targetIndex {
            pt := Path{MaxCompositionLength: 999}
            pt.Nodes = make([]string, len(path))

            for pos, nodeIndex := range path {
                pt.Nodes[pos] = g.Nodes[nodeIndex].Id
            }

            // Custo
            //--------------
            for k := 0; k < len(path)-1; k++ {
                u := path[k]
                v := path[k+1]
                uNode := g.Nodes[u]
                vNode := g.Nodes[v]

                if (uNode.Side != 0 && uNode.AggregatorIndex == v) || (uNode.Side == 0 && vNode.AggregatorIndex == u) {
                    // no cost, dummy edge
                } else {
                    pt.Cost += 1
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
                    dfs(v, capSoFar)
                }
            }
        }

        path = path[:len(path)-1]
        visited[u] = false
    }

    dfs(sourceIndex, 9999)

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

    return results
}

type Maneuver struct {
    Composition     []int
    Path            Path
    ManeuverCost    float64
    PartialCost     float64
    TotalCostEstimate float64
    EndState        State

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
    s2 := d.TargetState

    paths := make(map[int]map[int]float64)
    var ok bool

    for assetIndex := range s2.AssetLocation {
        if s1.AssetLocation[assetIndex] != s2.AssetLocation[assetIndex] {
            z1 := s1.AssetLocation[assetIndex]
            z2 := s2.AssetLocation[assetIndex]

            _, ok = paths[z1]
            if !ok {
                paths[z1] = make(map[int]float64)
            }

            _, ok = paths[z1][z2]
            if !ok {
                paths[z1][z2] = d.Graph.DistMatrix[z1][z2]
            }
        }
    }

    result := 0.0
    minTrips := 0

    for z, _ := range paths {
        minTrips += len(paths[z])
    }

    multiplier := 2.0
    if minTrips == 1 {
        multiplier = 1.0
    }

    for z1, _ := range paths {
        for z2, _ := range paths[z1] {
            result += multiplier*paths[z1][z2]
        }
    }

    return result
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

func EqualStates(s1 State, s2 State) bool {
    for z := range s1.SBs {
        _, exists := s2.SBs[z]
        if len(s1.SBs[z]) == 0 {
            if exists && len(s2.SBs[z]) > 0 {
                return false
            }
        } else if exists && !EqualSlices(s1.SBs[z], s2.SBs[z]) {
            return false
        }
    }

    for z := range s2.SBs {
        if len(s2.SBs[z]) > 0 {
            _, exists := s1.SBs[z]
            if !exists {
                return false
            }
        }
    }

    return true
}

func EqualAssetLocations(s1 State, s2 State) bool {
    return EqualSlices(s1.AssetLocation, s2.AssetLocation)
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

func GetPossibleManeuvers(d Data, parent *Maneuver) []*Maneuver {
    s := parent.EndState

    // Localizar locomotivas (sb e posição na sb)
    //-------------------------------------------------
    locoLocations := make(map[int][]int)

    for sbIndex, rollingStock := range s.SBs {
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

    for sbIndex, _ := range locoLocations {
        paths[sbIndex] = make(map[int][]Path)

        for j := range d.Graph.Nodes {
            iNode := d.Graph.Nodes[j]
            if j == sbIndex || iNode.Side != 0 {continue}
            paths[sbIndex][j] = GetAllPaths(d.Graph, s, sbIndex, j)
        }
    }

    // Para cada locomotiva em cada sb, construir todas as composições possíveis
    //--------------------------------------------------------------------------
    possibleCompositions := make(map[int][]CompositionRange)

    for sbIndex, locoPositions := range locoLocations {
        possibleCompositions[sbIndex] = make([]CompositionRange, 0)

        for _, pos := range locoPositions {
            // Composições que incluem 1o material rodante
            for p := pos; p < len(s.SBs[sbIndex]); p++ {
                compositionRange := CompositionRange{0, p}
                possibleCompositions[sbIndex] = append(possibleCompositions[sbIndex], compositionRange)
            }

            // Composições que incluem o último material rodante (mas não o primeiro, já coletado acima)
            for p := pos; p >= 1; p-- {
                compositionRange := CompositionRange{p, len(s.SBs[sbIndex])-1}
                possibleCompositions[sbIndex] = append(possibleCompositions[sbIndex], compositionRange)
            }
        }
    }

    // Para cada composição possível em cada sb, construir todas as manobras
    // possíveis
    //--------------------------------------------------------------------------
    maneuvers := make([]*Maneuver, 0)

    for sbIndex, compList := range possibleCompositions {
        for _, compositionRange := range compList {
            for targetIndex, pathsToTarget := range paths[sbIndex] {
                for _, path := range pathsToTarget {
                    compositionSize := compositionRange.LastAsset - compositionRange.FirstAsset + 1
                    if compositionSize > path.MaxCompositionLength {continue}

                    firstNodeId := path.Nodes[1]
                    firstNode, _ := GetNode(d.Graph.Nodes, firstNodeId)
                    if firstNode.Side == 'A' {
                        if compositionRange.FirstAsset > 0 {
                            continue
                        }
                    } else {
                        if compositionRange.LastAsset < len(s.SBs[sbIndex])-1 {
                            continue
                        }
                    }

                    maneuver := new(Maneuver)
                    maneuver.Parent = parent
                    maneuver.Composition = make([]int, compositionSize)
                    copy(maneuver.Composition, s.SBs[sbIndex][compositionRange.FirstAsset:compositionRange.LastAsset+1])
                    maneuver.Path = path
                    maneuver.ManeuverCost = path.Cost
                    maneuver.PartialCost = parent.PartialCost + maneuver.ManeuverCost

                    maneuver.EndState = CopyState(s)

                    // maneuver.EndState.SBs[sbIndex] = append(
                    //     s.SBs[sbIndex][0:compositionRange.FirstAsset],
                    //     s.SBs[sbIndex][compositionRange.LastAsset+1:len(s.SBs[sbIndex])]...
                    // )

                    maneuver.EndState.SBs[sbIndex] = append(
                        maneuver.EndState.SBs[sbIndex][0:compositionRange.FirstAsset],
                        maneuver.EndState.SBs[sbIndex][compositionRange.LastAsset+1:len(maneuver.EndState.SBs[sbIndex])]...
                    )

                    secondLastNodeId := path.Nodes[len(path.Nodes)-2]
                    secondLastNode, _ := GetNode(d.Graph.Nodes, secondLastNodeId)

                    aux := 0
                    if firstNode.Side == secondLastNode.Side {
                        aux = 1
                    }
                    finalOrientation := (aux + path.OrientationChanges) % 2

                    if finalOrientation == 0 {
                        if secondLastNode.Side == 'A' {
                            maneuver.EndState.SBs[targetIndex] = append(maneuver.Composition, maneuver.EndState.SBs[targetIndex]...)
                        } else {
                            maneuver.EndState.SBs[targetIndex] = append(maneuver.EndState.SBs[targetIndex], maneuver.Composition...)
                        }
                    } else {
                        if secondLastNode.Side == 'A' {
                            maneuver.EndState.SBs[targetIndex] = PrependReversed(maneuver.EndState.SBs[targetIndex], maneuver.Composition)
                        } else {
                            maneuver.EndState.SBs[targetIndex] = AppendReversed(maneuver.EndState.SBs[targetIndex], maneuver.Composition)
                        }
                    }

                    for pos := compositionRange.FirstAsset; pos <= compositionRange.LastAsset; pos++ {
                        assetIndex := s.SBs[sbIndex][pos]
                        maneuver.EndState.AssetLocation[assetIndex] = targetIndex
                    }

                    if RepeatingState(maneuver) {
                        continue
                    }

                    // Calcular TotalCostEstimate
                    //------------------------------
                    maneuver.TotalCostEstimate = maneuver.PartialCost + GetDistanceToTargetState(d, maneuver.EndState)

                    maneuvers = append(maneuvers, maneuver)
                }
            }
        }
    }

    return maneuvers
}

func PrintState(d Data, state State) {
    g := d.Graph
    for z, assets := range state.SBs {
        if len(assets) == 0 {continue}
        fmt.Printf("  %-3s: ", g.Nodes[z].Id)
        fmt.Println(assets)
    }
}

func PrintManeuver(d Data, maneuver *Maneuver) {
    fmt.Println("Path              : ", maneuver.Path.Nodes)
    fmt.Println("OrientationChanges: ", maneuver.Path.OrientationChanges)
    fmt.Println("Composition       : ", maneuver.Composition)
    fmt.Println("ManeuverCost      : ", maneuver.ManeuverCost)
    fmt.Println("PartialCost       : ", maneuver.PartialCost)
    fmt.Println("TotalCostEstimate : ", maneuver.TotalCostEstimate)
    fmt.Println("EndState    : ")
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

func main() {
    configFile, _ := os.Open("local/config2.json")
    //configFile, _ := os.Open("local/config.json")
    defer configFile.Close()

    config := Config{}
    json.NewDecoder(configFile).Decode(&config)

    d := CreateData(config)
    node, _ := GetNode(d.Graph.Nodes, "0")
    node.Length = 10
    node, _ = GetNode(d.Graph.Nodes, "1")
    node.Length = 10
    node, _ = GetNode(d.Graph.Nodes, "2")
    node.Length = 10
    node, _ = GetNode(d.Graph.Nodes, "3")
    node.Length = 10
    node, _ = GetNode(d.Graph.Nodes, "4")
    node.Length = 10
    node, _ = GetNode(d.Graph.Nodes, "5")
    node.Length = 10

    fmt.Println("InitialState:")
    PrintState(d, d.InitialState)

    fmt.Println("TargetState:")
    PrintState(d, d.TargetState)

    //unvisited := make([]*Maneuver, 0)
    unvisited := &ManeuverHeap{}
    heap.Init(unvisited)

    m0 := new(Maneuver)
    m0.EndState = CopyState(d.InitialState)
    m0.Children = GetPossibleManeuvers(d, m0)

    for _, child := range m0.Children {
        heap.Push(unvisited, child)
    }
    //unvisited = append(unvisited, m0.Children...)

    var bestLeaf *Maneuver
    maxIterations := 100_000
    iter := 0

    for unvisited.Len() > 0 && iter <= maxIterations {
        fmt.Printf("\rIteration: %-8d  | unvisited: %-8d", iter, unvisited.Len())
        maneuver := PopManeuverWithLowestTotalCostEstimate(unvisited)

        if EqualAssetLocations(maneuver.EndState, d.TargetState) {
            bestLeaf = maneuver
            break
        }

        maneuver.Children = GetPossibleManeuvers(d, maneuver)
        for _, child := range maneuver.Children {
            heap.Push(unvisited, child)
        }
        //unvisited = append(unvisited, maneuver.Children...)

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







