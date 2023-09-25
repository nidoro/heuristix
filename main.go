package main

import (
    "math"
    "math/rand"
    "time"
    "github.com/davecgh/go-spew/spew"
    "fmt"
    //"flag"
)

type ImproveStrategy [T any] func (s *T) bool
type ImproveCallbackFunc [T any] func (s *T)

type ILSAlg [T any] struct {
    ImproveStrategies []ImproveStrategy [T]
    ImproveCallback ImproveCallbackFunc [T]
    
    MaxNonImprovingIter int
    Iter int
    Improvements int
    CurrentStrategy int
}

func ILS [T any] () ILSAlg[T] {
    return ILSAlg[T] {
        MaxNonImprovingIter: 5,
    }
}

func (ils *ILSAlg[T]) Improve(s *T) {
    nonImprovingIter := 0
    ils.Iter = 0
    
    for nonImprovingIter < ils.MaxNonImprovingIter {
        stg := 0
        improved := false
        for stg < len(ils.ImproveStrategies) {
            ils.CurrentStrategy = stg
            strategy := ils.ImproveStrategies[stg]
            if strategy(s) {
                improved = true
                ils.Improvements += 1
                ils.ImproveCallback(s)
                stg = 0
            } else {
                stg += 1
            }
        }
        
        if improved {
            nonImprovingIter = 0
        } else {
            nonImprovingIter += 1
        }
        
        ils.Iter += 1
    }
}

func (ils *ILSAlg [T]) SetMaxNonImprovingIter(value int) {
    ils.MaxNonImprovingIter = value
}

func (ils *ILSAlg [T]) AddImproveStrategy(strategy func (s *T) bool) {
    ils.ImproveStrategies = append(ils.ImproveStrategies, strategy)
}

func (ils *ILSAlg [T]) SetImproveCallback(callback func (s *T)) {
    ils.ImproveCallback = callback
}

type Node struct {
    X float64
    Y float64
}

type TSPData struct {
    N int
    Nodes [] Node
    Edges [][] float64
}

type Solution struct {
    Data *TSPData
    Order [] int
    Cost float64
}

func Swap(a *int, b *int) {
    *a, *b = *b, *a
}

func GetRandomNumber(min float64, max float64) float64 {
    return min + rand.Float64()*(max-min)
}

func GetRandomInt(min int, max int) int {
    return rand.Intn(max-min+1) + min
}

func GenRandomData(nodeCount int, regionWidth float64, regionHeight float64) TSPData {
    var d TSPData
    
    d.N = nodeCount
    d.Nodes = make([] Node, nodeCount)
    d.Edges = make([][] float64, nodeCount)
    
    for i := 0; i < nodeCount; i++ {
        node := Node {
            X: GetRandomNumber(0, regionWidth),
            Y: GetRandomNumber(0, regionHeight),
        }
        
        d.Nodes[i] = node
        d.Edges[i] = make([] float64, nodeCount)
    }
    
    for i := 0; i < nodeCount; i++ {
        a := d.Nodes[i]
        
        for j := i+1; j < nodeCount; j++ {
            b := d.Nodes[j]
            
            d.Edges[i][j] = math.Sqrt(math.Pow(a.X - b.X, 2) + math.Pow(a.Y - b.Y, 2))
            d.Edges[j][i] = d.Edges[i][j]
        }
    }
    
    return d
}

func GenRandomSolution(d *TSPData) Solution {
    var s Solution
    s.Data = d
    s.Order = make([] int, d.N+1)
    s.Cost = 0.0
    
    for i := 0; i < d.N; i++ {
        s.Order[i] = i
    }
    
    for i := 1; i < d.N; i++ {
        j := GetRandomInt(1, d.N-1)
        Swap(&s.Order[i], &s.Order[j])
    }
    
    for j := 1; j < len(s.Order); j++ {
        a := s.Order[j-1]
        b := s.Order[j]
        s.Cost += d.Edges[a][b]
    }
    
    return s
}
    
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

func ImproveCallback(s *Solution) {
    fmt.Println(s.Cost)
}

func main() {
    rand.Seed(time.Now().UnixNano())
    
    d := GenRandomData(10, 100, 100)
    s := GenRandomSolution(&d)
    
    spew.Dump(d)

    ils := ILS[Solution]()
    ils.AddImproveStrategy(ImproveByTwoOpt)
    ils.SetImproveCallback(ImproveCallback)
    
    ils.Improve(&s)
    
    spew.Dump(s.Order)
    spew.Dump(s.Cost)
}




