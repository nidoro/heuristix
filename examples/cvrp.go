package main

import (
    "math"
    "math/rand"
    "time"
    //"github.com/davecgh/go-spew/spew"
    "fmt"
    "hx"
    
    "image/color"
    "gonum.org/v1/plot"
    "gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/vg"
)

type Node struct {
    X float64
    Y float64
    Demand int
}

type Nodes []Node

func (nodes Nodes) Len() int {
    return len(nodes)
}

func (nodes Nodes) XY(index int) (x float64, y float64) {
    return nodes[index].X, nodes[index].Y
}

type Data struct {
    N          int
    Nodes      Nodes
    Edges      [][] float64
    VehicleCap int
}

type Route struct {
    Id    int
    Order [] int
    Load  int
}

type Solution struct {
    Data      *Data
    Routes    [] *Route
    Cost float64
    NodeRoute [] int
}

func MakeRoute(n int) *Route {
    route := new(Route)
    route.Order = make([] int, 1, n+1)
    route.Order[0] = 0
    route.Load = 0
    return route
}

func Swap [T any] (a *T, b *T) {
    *a, *b = *b, *a
}

func Print(s Solution) {
    fmt.Printf("Cost: %.4f\n", s.Cost)
    
    for r, route := range s.Routes {
        fmt.Printf("Route %d: %d", r, route.Order[0])
        for i := 1; i < len(route.Order); i++ {
            fmt.Printf(" - %d", route.Order[i])
        }
        fmt.Println()
    }
}

func GetRandomNumber(min float64, max float64) float64 {
    return min + rand.Float64()*(max-min)
}

func GetRandomInt(min int, max int) int {
    return rand.Intn(max-min+1) + min
}

func GenRandomData(nodeCount int, regionWidth float64, regionHeight float64) Data {
    var d Data
    
    d.N = nodeCount
    d.Nodes = make([] Node, 0, nodeCount+1)
    d.Edges = make([][] float64, nodeCount+1)
    
    garage := Node {
        X: regionWidth/2,
        Y: regionHeight/2,
        Demand: 0,
    }
    
    d.Nodes = append(d.Nodes, garage)
    
    for i := 0; i < nodeCount; i++ {
        node := Node {
            X: GetRandomNumber(0, regionWidth),
            Y: GetRandomNumber(0, regionHeight),
            Demand: 1,
        }
        
        d.Nodes = append(d.Nodes, node)
    }
    
    for i := 0; i < nodeCount+1; i++ {
        d.Edges[i] = make([] float64, nodeCount+1)
    }
    
    for i := 0; i < nodeCount+1; i++ {
        a := d.Nodes[i]
        
        for j := 0; j < nodeCount+1; j++ {
            b := d.Nodes[j]
            
            edge := math.Sqrt(math.Pow(a.X - b.X, 2) + math.Pow(a.Y - b.Y, 2))
            d.Edges[i][j] = edge
            d.Edges[j][i] = d.Edges[i][j]
        }
    }
    
    return d
}

func GenRandomSolution(d *Data) Solution {
    var s Solution
    s.Data = d
    s.Routes = make([] *Route, 0, 1)
    s.Cost = 0.0
    
    nextRouteId := 0
    route := MakeRoute(d.N)
    route.Id = nextRouteId
    nextRouteId++
    s.Routes = append(s.Routes, route)
    
    for i := 1; i < d.N+1; i++ {
        demand := d.Nodes[i].Demand
        if route.Load + demand <= d.VehicleCap {
            route.Order = append(route.Order, i)
            route.Load += demand
            s.NodeRoute = append(s.NodeRoute, route.Id)
        } else {
            route.Order = append(route.Order, 0)
            route = MakeRoute(d.N)
            route.Id = nextRouteId
            nextRouteId++
            route.Order = append(route.Order, i)
            route.Load = demand
            s.Routes = append(s.Routes, route)
            s.NodeRoute = append(s.NodeRoute, route.Id)
        }
    }
    
    route.Order = append(route.Order, 0)
    
    for r := 0; r < len(s.Routes); r++ {
        route := s.Routes[r]
        
        for i := 0; i < len(route.Order)-1; i++ {
            s.Cost += d.Edges[route.Order[i]][route.Order[i+1]]
        }
    }
    
    return s
}

func RemoveIndexFromRoute(route *Route, index int) {
    before := route.Order[:index]
    after := route.Order[index+1:]
    route.Order = append(before, after...)
}

func InsertNodeIntoRoute(route *Route, nd int, index int) {
    newOrder := make([]int, len(route.Order)+1)
    copy(newOrder[:index], route.Order[:index])
    newOrder[index] = nd
    copy(newOrder[index+1:], route.Order[index:])
    route.Order = newOrder
}

func ImproveByReinsertingEx(s *Solution, heu hx.Heuristic[Solution]) float64 {
    // -> a -> b -> c
    // -> u -> b -> v
    d := s.Data
    
    for r1 := 0; r1 < len(s.Routes); r1++ {
        route1 := s.Routes[r1]
        
        for i := 1; i < len(route1.Order)-1; i++ {
            a := route1.Order[i-1]
            b := route1.Order[i]
            c := route1.Order[i+1]
            
            nodeB := d.Nodes[b]
            
            for r2 := 0; r2 < len(s.Routes); r2++ {
                route2 := s.Routes[r2]
                
                for j := 1; j < len(route2.Order)-1; j++ {
                    u := route2.Order[j]
                    v := route2.Order[j+1]
                    
                    if b == u || a == u {
                        continue
                    }
                    
                    if r1 != r2 && route2.Load + nodeB.Demand > d.VehicleCap {
                        continue
                    }
                    
                    abcCost := d.Edges[a][b] + d.Edges[b][c]
                    acCost  := d.Edges[a][c]
                    uvCost  := d.Edges[u][v]
                    ubvCost := d.Edges[u][b] + d.Edges[b][v]
                    
                    newCost := s.Cost - abcCost - uvCost + acCost + ubvCost
                    costDiff := newCost - s.Cost
                    
                    ok, sl := heu.AcceptCost(s, newCost)
                    
                    if (ok) {
                        route1 := sl.Routes[r1]
                        route2 := sl.Routes[r2]
                        
                        RemoveIndexFromRoute(route1, i)
                        k := 0
                        if i < j && route1 == route2 {
                            k = j
                        } else {
                            k = j+1
                        }
                        InsertNodeIntoRoute(route2, b, k)
                        if (r1 != r2) {
                            route2.Load += route2.Load + nodeB.Demand
                        }
                        sl.Cost = newCost
                        if (heu.AcceptSolution(&sl, newCost)) {
                            *s = sl
                            return costDiff
                        }
                    }
                }
            }
        }
    }
    
    return 0.0
}

func ImproveBy2OptEx(s *Solution, heu hx.Heuristic[Solution]) float64 {
    // Before:
    // -- a -- b -- ...> -- u -- v --
    // After:
    // -- a -- u -- <... -- b -- v --
    
    // i = [0, n-2]
    
    d := s.Data
    
    for r, route := range s.Routes {
        for i := 0; i < len(route.Order)-2; i++ {
            a := route.Order[i]
            b := route.Order[i+1]
            
            for j := i+2; j < len(route.Order)-1; j++ {
                u := route.Order[j]
                v := route.Order[j+1]
                
                if a == u || a == v || b == u || b == v {
                    continue
                }
                
                abCost := d.Edges[a][b]
                uvCost := d.Edges[u][v]
                auCost := d.Edges[a][u]
                bvCost := d.Edges[b][v]
                
                newCost := s.Cost - abCost - uvCost + auCost + bvCost
                costDiff := newCost - s.Cost
                
                ok, sl := heu.AcceptCost(s, newCost)
                
                if ok {
                    route := sl.Routes[r]
                    
                    // invert order between b and u
                    stretchNodes := j - i+1
                    for k := 0; k < int(stretchNodes/2); k++ {
                        orig := i+1+k
                        dest := j-k
                        Swap(&route.Order[orig], &route.Order[dest])
                    }
                    
                    sl.Cost = newCost
                    if heu.AcceptSolution(&sl, newCost) {
                        *s = sl
                        return costDiff
                    }
                }
            }
        }
    }
    
    return 0.0
}

func ImproveBySwapingAdjacent(s *Solution) float64 {
    // -> a -> b -> c -> d
    // -> a -> c -> b -> d
    data := s.Data
    
    for r := 0; r < len(s.Routes); r++ {
        route := s.Routes[r]
        
        for i := 1; i < len(route.Order)-2; i++ {
            a := route.Order[i-1]
            b := route.Order[i]
            c := route.Order[i+1]
            d := route.Order[i+2]
            
            abcdCost := data.Edges[a][b] + data.Edges[b][c] + data.Edges[c][d]
            acbdCost := data.Edges[a][c] + data.Edges[c][b] + data.Edges[b][d]
            
            newCost := s.Cost - abcdCost + acbdCost
            costDiff := newCost - s.Cost
            
            if newCost < s.Cost {
                Swap(&route.Order[i], &route.Order[i+1])
                s.Cost = newCost
                return costDiff
            }
        }
    }
    
    return 0.0
}

func DiversityBySwapingAdjacent(s *Solution) float64 {
    // -> a -> b -> c -> d
    // -> a -> c -> b -> d
    data := s.Data
    
    r := GetRandomInt(0, len(s.Routes)-1)
    route := s.Routes[r]
    
    i := GetRandomInt(1, len(route.Order)-3)
    a := route.Order[i-1]
    b := route.Order[i]
    c := route.Order[i+1]
    d := route.Order[i+2]
    
    abcdCost := data.Edges[a][b] + data.Edges[b][c] + data.Edges[c][d]
    acbdCost := data.Edges[a][c] + data.Edges[c][b] + data.Edges[b][d]
    
    newCost := s.Cost - abcdCost + acbdCost
    costDiff := newCost - s.Cost
    
    Swap(&route.Order[i], &route.Order[i+1])
    s.Cost = newCost
    
    return costDiff
}

func (s Solution) Copy() Solution {
    var result Solution
    result.Data = s.Data
    result.Cost = s.Cost
    copy(result.NodeRoute, s.NodeRoute)
    result.Routes = make([]*Route, len(s.Routes))
    for i, route := range s.Routes {
        result.Routes[i] = MakeRoute(s.Data.N)
        result.Routes[i].Order = make([] int, len(route.Order))
        copy(result.Routes[i].Order, route.Order)
        result.Routes[i].Id = route.Id
        result.Routes[i].Load = route.Load
    }
    
    return result
}

func DiversityByReinserting(s *Solution) float64 {
    d := s.Data
    
    for {
        r1 := GetRandomInt(0, len(s.Routes)-1)
        r2 := GetRandomInt(0, len(s.Routes)-1)
        
        route1 := s.Routes[r1]
        route2 := s.Routes[r2]
        
        i := GetRandomInt(1, len(route1.Order)-2)
        j := GetRandomInt(1, len(route2.Order)-2)
        
        a := route1.Order[i-1]
        b := route1.Order[i]
        c := route1.Order[i+1]
        
        nodeB := d.Nodes[b]
        
        u := route2.Order[j]
        v := route2.Order[j+1]
        
        if b == u || a == u {
            continue
        }
        
        if r1 != r2 && route2.Load + nodeB.Demand > d.VehicleCap {
            continue
        }
        
        abcCost := d.Edges[a][b] + d.Edges[b][c]
        acCost  := d.Edges[a][c]
        uvCost  := d.Edges[u][v]
        ubvCost := d.Edges[u][b] + d.Edges[b][v]
        
        newCost := s.Cost - abcCost - uvCost + acCost + ubvCost
        
        diff := newCost - s.Cost
        
        RemoveIndexFromRoute(route1, i)
        k := 0
        if i < j && route1 == route2 {
            k = j
        } else {
            k = j+1
        }
        InsertNodeIntoRoute(route2, b, k)
        if (r1 != r2) {
            route2.Load += route2.Load + nodeB.Demand
        }
        s.Cost = newCost
        
        return diff
    }
}

func ImprovementCallback(s *Solution, heu hx.Heuristic[Solution]) {
    fmt.Printf("INFO | Improved %-4d | Cost: %-14.4f | CurrentStrategy: %d\n", heu.GetImprovementsCount(), s.Cost, heu.GetCurrentStrategy())
    //Print(*s)
}

func PlotSolution(s Solution, filePath string) {
    d := s.Data
    
    plt := plot.New()
    plt.X.Label.Text = "X"
    plt.Y.Label.Text = "Y"
    plt.Add(plotter.NewGrid())
    
    scatter, _ := plotter.NewScatter(d.Nodes)
    scatter.GlyphStyle.Color = color.RGBA{R: 255, B: 128, A: 255}
    scatter.GlyphStyle.Radius = vg.Points(3)
    plt.Add(scatter)
    
    for _, route := range s.Routes {
        for i := 0; i < len(route.Order)-1; i++ {
            j := i+1
            coords := make(plotter.XYs, 2)
            coords[0].X = d.Nodes[route.Order[i]].X
            coords[0].Y = d.Nodes[route.Order[i]].Y
            coords[1].X = d.Nodes[route.Order[j]].X
            coords[1].Y = d.Nodes[route.Order[j]].Y
            line, _ := plotter.NewLine(coords)
            line.LineStyle.Width = vg.Points(1)
            line.LineStyle.Color = color.RGBA{B: 255, A: 255}
            plt.Add(line)
        }
    }
    
    _ = plt.Save(400, 400, filePath)
}

func (s Solution) GetCost() float64 {
    return s.Cost
}

func (s1 Solution) Compare(s2 Solution) bool {
    if len(s1.Routes) == len(s2.Routes) && s1.Cost == s2.Cost {
        for r, _ := range s1.Routes {
            if len(s1.Routes[r].Order) == len(s2.Routes[r].Order) && s1.Routes[r].Load == s2.Routes[r].Load {
                for i, _ := range s1.Routes {
                    if (s1.Routes[r].Order[i] != s2.Routes[r].Order[i]) {
                        return false
                    }
                }
                return true
            }
        }
    }
    
    return false
}

func main() {
    rand.Seed(time.Now().UnixNano())
    
    d := GenRandomData(40, 100, 100)
    d.VehicleCap = 40
    
    s0 := GenRandomSolution(&d)
    
    s := s0.Copy()
    
    vnd := hx.VND[Solution]()
    vnd.AddImprovingStrategy(ImproveBySwapingAdjacent)
    vnd.AddImprovingStrategyEx(ImproveByReinsertingEx)
    vnd.AddImprovingStrategyEx(ImproveBy2OptEx)
    vnd.Improve(&s, s.Cost)
    
    fmt.Println("VND Solution:")
    Print(s)
    PlotSolution(s, "s1.svg")
    
    s = s0.Copy()
    
    sa := hx.SA[Solution]()
    sa.AddImprovingStrategyEx(ImproveBy2OptEx)
    sa.AddDiversificationStrategy(DiversityByReinserting)
    sa.Improve(&s)
    
    fmt.Println("SA Solution:")
    Print(s)
    PlotSolution(s, "s2.svg")
    
    s = s0.Copy()
    
    ils := hx.ILS[Solution]()
    ils.MaxNonImprovingIter = 20
    ils.AddImprovingStrategy(ImproveBySwapingAdjacent)
    ils.AddImprovingStrategyEx(ImproveByReinsertingEx)
    ils.AddImprovingStrategyEx(ImproveBy2OptEx)
    ils.AddDiversificationStrategy(DiversityBySwapingAdjacent)
    ils.AddDiversificationStrategy(DiversityByReinserting)
    ils.Improve(&s)
    
    fmt.Println("ILS Solution:")
    Print(s)
    PlotSolution(s, "s3.svg")
    
    s = s0.Copy()
    
    ts := hx.TS[Solution]()
    ts.MaxNonImprovingIter = 100
    ts.TabuListMaxSize = 50
    ts.AddImprovingStrategyEx(ImproveByReinsertingEx)
    ts.AddImprovingStrategyEx(ImproveBy2OptEx)
    ts.Improve(&s)
    
    fmt.Println("TS Solution:")
    Print(s)
    PlotSolution(s, "s4.svg")
}




