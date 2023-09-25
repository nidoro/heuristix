package hx

// AlgState struct and interface
//--------------------------------
type ImproveStrategy [T any] func (s *T) bool
type ImproveCallbackFunc [T any] func (ils *ILSAlg [T], s *T)

type AlgState [T any] struct {
    ImproveStrategies []ImproveStrategy [T]
    ImproveCallback ImproveCallbackFunc [T]
    
    Iter int
    Improvements int
    CurrentStrategy int
}

type AlgStateInterface [T any] interface {
    Improve(solution *T)
    AddImproveStrategy(strategy ImproveStrategy [T])
    SetImproveCallback(callback ImproveCallbackFunc [T])
}

func (alg *AlgState [T]) AddImproveStrategy(strategy ImproveStrategy [T]) {
    alg.ImproveStrategies = append(alg.ImproveStrategies, strategy)
}

func (alg *AlgState [T]) SetImproveCallback(callback ImproveCallbackFunc[T]) {
    alg.ImproveCallback = callback
}

// ILSAlg struct and interface
//--------------------------------
type ILSAlg [T any] struct {
    AlgState[T]
    MaxNonImprovingIter int
}

type ILSAlgInterface [T any] interface {
    SetMaxNonImprovingIter(value int)
}

// Constructor
func ILS [T any] () ILSAlg[T] {
    return ILSAlg[T] {
        MaxNonImprovingIter: 5,
    }
}

func (ils *ILSAlg [T]) SetMaxNonImprovingIter(value int) {
    ils.MaxNonImprovingIter = value
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
                ils.ImproveCallback(ils, s)
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
