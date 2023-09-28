package hx

import (
    "fmt"
    "math"
    "math/rand"
)

const ZERO = 0.00000001

func GetRandomInt(min int, max int) int {
    return rand.Intn(max-min+1) + min
}

// AlgState struct and interface
//--------------------------------
type ImprovementStrategy[T any] func (s *T) float64
type ImprovementStrategyEx[T any] func (s *T, heu Heuristic[T]) float64
type ImprovementCallback[T any] func (s *T, heu Heuristic[T])

type AlgState[T any] struct {
    ImproveStrategiesEx [] ImprovementStrategyEx[T]
    OnImprovement       ImprovementCallback[T]
    
    Improvements int
    CurrentStrategy int
    CurrentCost float64
    NewCost float64
    BestCost float64
    Verbose bool
}

func CreateAlgState[T any]() AlgState[T] {
    return AlgState[T] {
        OnImprovement: func (s *T, heu Heuristic[T]) {},
        Verbose: true,
    }
}

type AlgStateInterface[T any] interface {
    Improve(solution *T)
    AcceptSolution(s *T, cost float64) bool
    AcceptCost(s *T, newCost float64) (bool, T)
    AddImprovingStrategy(strategy ImprovementStrategy[T])
    LogImprovement(s *T)
}

func (alg *AlgState[T]) AddImprovingStrategy(strategy ImprovementStrategy[T]) {
    alg.ImproveStrategiesEx = append(alg.ImproveStrategiesEx, func(s *T, heu Heuristic[T]) float64 {return strategy(s)})
}

func (alg *AlgState[T]) AddImprovingStrategyEx(strategy ImprovementStrategyEx[T]) {
    alg.ImproveStrategiesEx = append(alg.ImproveStrategiesEx, strategy)
}

func (alg AlgState[T]) LogImprovement() {
    if alg.Verbose {
        fmt.Printf("Improvement %-4d | Cost: %-14.4f\n", alg.Improvements, alg.BestCost)
    }
}

func (alg AlgState[T]) LogInitialSolution(cost float64) {
    if alg.Verbose {
        fmt.Printf("Initial Solution | Cost: %-14.4f\n", cost)
    }
}

func (alg AlgState[T]) AcceptSolution(s *T, cost float64) bool {
    return true
}

func (alg *AlgState[T]) AcceptCost(s *T, newCost float64) (bool, T) {
    if newCost < alg.CurrentCost {
        alg.NewCost = newCost
        return true, *s
    } else {
        return false, *s
    }
}

func (alg AlgState[T]) GetImprovementsCount() int {
    return alg.Improvements
}

func (alg AlgState[T]) GetCurrentStrategy() int {
    return alg.CurrentStrategy
}

// VNDAlg struct
//---------------
type VNDAlg [T any] struct {
    AlgState[T]
}

// Constructor
func VND [T any] () VNDAlg[T] {
    return VNDAlg[T] {
        AlgState: CreateAlgState[T](),
    }
}

func SetStrategiesEx [T any] (vnd *VNDAlg[T], strategies [] ImprovementStrategyEx[T]) {
    vnd.ImproveStrategiesEx = make([] ImprovementStrategyEx[T], len(strategies))
    copy(vnd.ImproveStrategiesEx, strategies)
}

func (vnd *VNDAlg[T]) Improve(s *T, cost float64) bool {
    stg := 0
    improved := false
    vnd.CurrentCost = cost
    
    vnd.LogInitialSolution(cost)
    
    for stg < len(vnd.ImproveStrategiesEx) {
        vnd.CurrentStrategy = stg
        strategy := vnd.ImproveStrategiesEx[stg]
        costDiff := strategy(s, &vnd.AlgState)
        
        if costDiff < 0.0 {
            improved = true
            vnd.CurrentCost += costDiff
            vnd.BestCost = vnd.CurrentCost
            vnd.Improvements += 1
            stg = 0
            vnd.OnImprovement(s, vnd)
            vnd.LogImprovement()
        } else {
            stg += 1
        }
    }
    
    return improved
}

// User interface
//----------------
type Heuristic[T any] interface {
    AcceptSolution(s *T, cost float64) bool
    AcceptCost(s *T, newCost float64) (bool, T)
    GetImprovementsCount() int
    GetCurrentStrategy() int
}

// Heuristic
//------------
type Solution[T any] interface {
    GetCost() float64
    Copy() T
}

type ComparableSolution[T any] interface {
    Solution[T]
    Compare(s T) bool
}

type DiversificationStrategy[T Solution[T]] func (s *T) float64

type HeuristicBase[T Solution[T]] struct {
    AlgState[T]
    DiversificationStrategies [] DiversificationStrategy[T]
}

type HeuristicInterface interface {
    AddDiversificationStrategy()
}

func (h *HeuristicBase[T]) AddDiversificationStrategy(method DiversificationStrategy[T]) {
    h.DiversificationStrategies = append(h.DiversificationStrategies, method)
}

func CreateHeuristicBase[T Solution[T]]() HeuristicBase[T] {
    return HeuristicBase[T]{
        AlgState: CreateAlgState[T](),
    }
}

// ILSAlg struct
//--------------------------------
type ILSAlg [T Solution[T]] struct {
    HeuristicBase[T]
    MaxNonImprovingIter int
}

// Constructor
func ILS [T Solution[T]]() ILSAlg[T] {
    return ILSAlg[T] {
        HeuristicBase: CreateHeuristicBase[T](),
        MaxNonImprovingIter: 5,
    }
}

func (ils *ILSAlg[T]) Improve(s *T) {
    ils.LogInitialSolution((*s).GetCost())
    
    vnd := VND[T]()
    vnd.Verbose = false
    SetStrategiesEx(&vnd, ils.ImproveStrategiesEx)
    
    nonImprovingIter := 0
    
    best := (*s).Copy()
    
    for nonImprovingIter <= ils.MaxNonImprovingIter {
        for p := 0; p < nonImprovingIter; p++ {
            m := GetRandomInt(0, len(ils.DiversificationStrategies)-1)
            ils.DiversificationStrategies[m](s)
        }
        
        _ = vnd.Improve(s, (*s).GetCost())
        
        if best.GetCost() - (*s).GetCost() >= ZERO {
            ils.Improvements++
            best = (*s).Copy()
            ils.BestCost = best.GetCost()
            nonImprovingIter = 1
            ils.OnImprovement(s, ils)
            ils.LogImprovement()
        } else {
            *s = best.Copy()
            nonImprovingIter++
        }
    }
}

// SAAlg struct and interface
//--------------------------------
type SAAlg [T Solution[T]] struct {
    HeuristicBase[T]
    IterationsEachTemperature int
    InitialTemperature float64
    MinTemperature float64
    CoolingRate float64
}

func SA [T Solution[T]] () SAAlg[T] {
    return SAAlg[T] {
        HeuristicBase: CreateHeuristicBase[T](),
        IterationsEachTemperature: 1,
        InitialTemperature: 1000,
        MinTemperature: 0.001,
        CoolingRate: 0.999,
    }
}

func (sa *SAAlg[T]) Improve(s *T) {
    sa.LogInitialSolution((*s).GetCost())
    
    temperature := sa.InitialTemperature
    best := (*s).Copy()

    for temperature > sa.MinTemperature {
        for i := 0; i < sa.IterationsEachTemperature; i++ {
            candidate := (*s).Copy()
            
            sa.CurrentStrategy = GetRandomInt(0, len(sa.DiversificationStrategies)-1)
            costDiff := sa.DiversificationStrategies[sa.CurrentStrategy](&candidate)
            
            if costDiff < 0.0 || rand.Float64() < math.Exp(-costDiff/temperature) {
                *s = candidate
            }
            
            if ((*s).GetCost() < best.GetCost()) {
                sa.Improvements++
                best = *s
                sa.BestCost = best.GetCost()
                sa.OnImprovement(s, sa)
                sa.LogImprovement()
            }
        }
        
        temperature *= sa.CoolingRate
    }
    
    *s = best
    
    vnd := VND[T]()
    vnd.Verbose = false
    SetStrategiesEx(&vnd, sa.ImproveStrategiesEx)
    vnd.Improve(s, (*s).GetCost())
}

// TabuSearch
//-------------------------------
type TSAlg[T ComparableSolution[T]] struct {
    HeuristicBase[T]
    TabuListMaxSize     int
    MaxNonImprovingIter int
    TabuList            [] ComparableSolution[T]
    
    BestSolution        T
    CurrentSolution     T
    BestNeighbor        T
    BestNeighborCost    float64
}

func TS[T ComparableSolution[T]]() TSAlg[T] {
    return TSAlg[T] {
        HeuristicBase: CreateHeuristicBase[T](),
        TabuListMaxSize: 20,
        MaxNonImprovingIter: 10,
    }
}

func (ts *TSAlg[T]) AcceptCost(s *T, newCost float64) (bool, T) {
    return true, (*s).Copy()
}

func (ts *TSAlg[T]) AcceptSolution(sl *T, cost float64) bool {
    isTabu := false
    
    for _, s2 := range ts.TabuList {
        if s2.Compare(*sl) {
            isTabu = true
            break
        }
    }
    
    if !isTabu && (*sl).GetCost() < ts.BestNeighborCost {
        ts.BestNeighbor = (*sl).Copy()
        ts.BestNeighborCost = (*sl).GetCost()
    }
    
    return false
}

func (ts *TSAlg[T]) Improve(s *T) {
    ts.LogInitialSolution((*s).GetCost())
    
    vnd := VND[T]()
    vnd.Verbose = false
    SetStrategiesEx(&vnd, ts.ImproveStrategiesEx)
    
    vnd.Improve(s, (*s).GetCost())
    
    nonImprovingIter := 0
    
    ts.BestSolution = (*s).Copy()
    
    
    for nonImprovingIter < ts.MaxNonImprovingIter {
        ts.CurrentSolution = (*s).Copy()
        ts.BestNeighborCost = math.Inf(1)
        
        for _, strategy := range ts.ImproveStrategiesEx {
            _ = strategy(s, ts)
        }
        
        if ts.BestNeighborCost == math.Inf(1) {
            break
        }
        
        *s = ts.BestNeighbor.Copy()
        
        if (ts.BestSolution.GetCost() - (*s).GetCost() >= ZERO) {
            ts.Improvements++
            ts.BestSolution = (*s).Copy()
            nonImprovingIter = 0
            ts.BestCost = ts.BestSolution.GetCost()
            ts.OnImprovement(s, ts)
            ts.LogImprovement()
        } else {
            nonImprovingIter++
        }
        
        ts.TabuList = append(ts.TabuList, *s)
        if len(ts.TabuList) > ts.TabuListMaxSize {
            ts.TabuList = ts.TabuList[1:]
        }
    }
    
    *s = ts.BestSolution
}



