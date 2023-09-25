package hx

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
