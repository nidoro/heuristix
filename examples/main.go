package main

import (
    "math"
    "math/rand"
    "time"
    "github.com/davecgh/go-spew/spew"
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

type Data struct {
    N          int
    Nodes      [] Node
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
    Cost      float64
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

func ImproveByReinserting(s *Solution) bool {
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
                    
                    if newCost < s.Cost {
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
                        return true
                    }
                }
            }
        }
    }
    
    return false
}

/*
func ImproveByTwoOpt(s *Solution) bool {
    // Before:
    // -- a -- b -- ...> -- u -- v --
    // After:
    // -- a -- u -- <... -- b -- v --
    
    d := s.Data
    
    // i = [0, n-2]
    for i := 0; i < d.N-2; i++ {
        a := s.Order[i]
        b := s.Order[i+1]
        
        for j := i+2; j < d.N; j++ {
            u := s.Order[j]
            v := s.Order[j+1]
            
            if a == u || a == v || b == u || b == v {
                continue
            }
            
            abCost := d.Edges[a][b]
            uvCost := d.Edges[u][v]
            auCost := d.Edges[a][u]
            bvCost := d.Edges[b][v]
            
            newCost := s.Cost - abCost - uvCost + auCost + bvCost
            
            if newCost < s.Cost {
                // invert order between b and u
                stretchNodes := j - i+1
                for k := 0; k < int(stretchNodes/2); k++ {
                    orig := i+1+k
                    dest := j-k
                    Swap(&s.Order[orig], &s.Order[dest])
                }
                
                s.Cost = newCost
                return true
            }
        }
    }
    
    return false
}
*/

func ImproveBy2Opt(s *Solution) bool {
    // Before:
    // -- a -- b -- ...> -- u -- v --
    // After:
    // -- a -- u -- <... -- b -- v --
    
    // i = [0, n-2]
    
    d := s.Data
    
    for _, route := range s.Routes {
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
                
                if newCost < s.Cost {
                    // invert order between b and u
                    stretchNodes := j - i+1
                    for k := 0; k < int(stretchNodes/2); k++ {
                        orig := i+1+k
                        dest := j-k
                        Swap(&route.Order[orig], &route.Order[dest])
                    }
                    
                    s.Cost = newCost
                    return true
                }
            }
        }
    }
    
    return false
}
    
func ImproveCallback(ils *hx.ILSAlg[Solution], s *Solution) {
    fmt.Println(s.Cost, ils.CurrentStrategy)
}

func main() {
    rand.Seed(time.Now().UnixNano())
    
    d := GenRandomData(20, 100, 100)
    d.VehicleCap = 20
    
    s := GenRandomSolution(&d)
    
    spew.Dump(d)
    spew.Dump(s.Routes)
    
    ils := hx.ILS[Solution]()
    ils.AddImproveStrategy(ImproveByReinserting)
    ils.AddImproveStrategy(ImproveBy2Opt)
    ils.SetImproveCallback(ImproveCallback)
    ils.Improve(&s)
    
    spew.Dump(s.Routes)
    spew.Dump(s.Cost)
    
    scatterData := make(plotter.XYs, d.N+1)
    
    for i, node := range d.Nodes {
        scatterData[i].X = node.X
        scatterData[i].Y = node.Y
    }
    
    plt := plot.New()
    plt.X.Label.Text = "X"
    plt.Y.Label.Text = "Y"
    plt.Add(plotter.NewGrid())
    
    scatter, _ := plotter.NewScatter(scatterData)
    scatter.GlyphStyle.Color = color.RGBA{R: 255, B: 128, A: 255}
    scatter.GlyphStyle.Radius = vg.Points(3)
    plt.Add(scatter)
    
    for _, route := range s.Routes {
        for i := 0; i < len(route.Order)-1; i++ {
            j := i+1
            pair := make(plotter.XYs, 2)
            pair[0].X = d.Nodes[route.Order[i]].X
            pair[0].Y = d.Nodes[route.Order[i]].Y
            pair[1].X = d.Nodes[route.Order[j]].X
            pair[1].Y = d.Nodes[route.Order[j]].Y
            line, _ := plotter.NewLine(pair)
            line.LineStyle.Width = vg.Points(1)
            line.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}
            line.LineStyle.Color = color.RGBA{B: 255, A: 255}
            plt.Add(line)
        }
    }
    
    _ = plt.Save(200, 200, "scatter.png")
}




