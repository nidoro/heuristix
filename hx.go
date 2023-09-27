package hx

import (
    //"fmt"
    "math"
    "math/rand"
)

func GetRandomInt(min int, max int) int {
    return rand.Intn(max-min+1) + min
}

// AlgState struct and interface
//--------------------------------
type ImproveStrategy [T any] func (s *T) float64
type ImproveStrategyEx [T any] func (s *T, alg *AlgState[T]) float64
type ImproveCallbackFunc [T any] func (s *T, alg AlgState[T])

type AlgState [T any] struct {
    ImproveStrategiesEx []ImproveStrategyEx [T]
    ImproveCallback ImproveCallbackFunc [T]
    AcceptFunc func(alg *AlgState[T], s *T) bool
    
    Improvements int
    CurrentStrategy int
}

func CreateAlgState [T any] () AlgState[T] {
    return AlgState[T] {
        ImproveCallback: func (s *T, alg AlgState[T]) {},
        AcceptFunc: func(alg *AlgState[T], s *T) bool {return true},
    }
}

type AlgStateInterface [T any] interface {
    Improve(solution *T)
    Accept(s *T)
    AddImprovingStrategy(strategy ImproveStrategy [T])
    SetImproveCallback(callback ImproveCallbackFunc [T])
}

func (alg *AlgState [T]) AddImprovingStrategy(strategy ImproveStrategy [T]) {
    alg.ImproveStrategiesEx = append(alg.ImproveStrategiesEx, func(s *T, alg *AlgState[T]) float64 {return strategy(s)})
}

func (alg *AlgState [T]) AddImprovingStrategyEx(strategy ImproveStrategyEx [T]) {
    alg.ImproveStrategiesEx = append(alg.ImproveStrategiesEx, strategy)
}

func (alg *AlgState [T]) SetImproveCallback(callback ImproveCallbackFunc[T]) {
    alg.ImproveCallback = callback
}

func (alg *AlgState [T]) Accept(s *T) bool {
    return alg.AcceptFunc(alg, s)
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

func SetStrategiesEx [T any] (vnd *VNDAlg[T], strategies [] ImproveStrategyEx[T]) {
    vnd.ImproveStrategiesEx = make([] ImproveStrategyEx[T], len(strategies))
    copy(vnd.ImproveStrategiesEx, strategies)
}

func (vnd *VNDAlg[T]) Improve(s *T) bool {
    stg := 0
    improved := false
    
    for stg < len(vnd.ImproveStrategiesEx) {
        vnd.CurrentStrategy = stg
        strategy := vnd.ImproveStrategiesEx[stg]
        if strategy(s, &vnd.AlgState) < 0.0 {
            improved = true
            vnd.Improvements += 1
            vnd.ImproveCallback(s, vnd.AlgState)
            stg = 0
        } else {
            stg += 1
        }
    }
    
    return improved
}

// Heuristic
//------------
type Solution [T any] interface {
    GetCost() float64
    Copy() T
}

type AlterSolutionMethod [T Solution[T]] func (s *T) float64

type Heuristic [T Solution[T]] struct {
    AlgState[T]
    AlterSolutionMethods [] AlterSolutionMethod[T]
}

type HeuristicInterface interface {
    AddAlteringStrategy()
}

func (h *Heuristic[T]) AddAlteringStrategy(method AlterSolutionMethod[T]) {
    h.AlterSolutionMethods = append(h.AlterSolutionMethods, method)
}

// ILSAlg struct
//--------------------------------
type ILSAlg [T Solution[T]] struct {
    Heuristic[T]
    MaxNonImprovingIter int
}

// Constructor
func ILS [T Solution[T]] () ILSAlg[T] {
    return ILSAlg[T] { MaxNonImprovingIter: 5 }
}

func (ils *ILSAlg[T]) Improve(s *T) {
    vnd := VND[T]()
    SetStrategiesEx(&vnd, ils.ImproveStrategiesEx)
    
    nonImprovingIter := 0
    
    best := (*s).Copy()
    
    for nonImprovingIter <= ils.MaxNonImprovingIter {
        for p := 0; p < nonImprovingIter; p++ {
            m := GetRandomInt(0, len(ils.AlterSolutionMethods)-1)
            ils.AlterSolutionMethods[m](s)
        }
        
        _ = vnd.Improve(s)
        
        if (*s).GetCost() < best.GetCost() {
            ils.Improvements++
            best = (*s).Copy()
            nonImprovingIter = 1
            ils.ImproveCallback(s, ils.AlgState)
        } else {
            *s = best.Copy()
            nonImprovingIter++
        }
    }
}

// SAAlg struct and interface
//--------------------------------
type SAAlg [T Solution[T]] struct {
    Heuristic[T]
    IterationsEachTemperature int
    InitialTemperature float64
    MinTemperature float64
    CoolingRate float64
}

func SA [T Solution[T]] () SAAlg[T] {
    return SAAlg[T] {
        IterationsEachTemperature: 1,
        InitialTemperature: 1000,
        MinTemperature: 0.001,
        CoolingRate: 0.999,
    }
}

func (sa *SAAlg[T]) Improve(s *T) {
    vnd := VND[T]()
    SetStrategiesEx(&vnd, sa.ImproveStrategiesEx)
    _ = vnd.Improve(s)
    
    temperature := sa.InitialTemperature
    best := (*s).Copy()

    for temperature > sa.MinTemperature {
        for i := 0; i < sa.IterationsEachTemperature; i++ {
            candidate := (*s).Copy()
            
            sa.CurrentStrategy = GetRandomInt(0, len(sa.AlterSolutionMethods)-1)
            costDiff := sa.AlterSolutionMethods[sa.CurrentStrategy](&candidate)
            
            if costDiff < 0.0 || rand.Float64() < math.Exp(-costDiff/temperature) {
                *s = candidate
            }
            
            if ((*s).GetCost() < best.GetCost()) {
                sa.Improvements++
                sa.ImproveCallback(s, sa.AlgState)
                best = *s
            }
        }
        
        temperature *= sa.CoolingRate
    }
    
    *s = best
    _ = vnd.Improve(s)
}

// TabuSearch
//--------------------------------


