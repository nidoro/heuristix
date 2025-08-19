package main

import (
    "container/heap"
    "fmt"
)

// State represents the current configuration of the pegs
type State struct {
    Pegs [][]int
}

// PriorityQueue implements a priority queue for the A* algorithm
type PriorityQueue []*Item

type Item struct {
    state State
    cost  int
    index int
}

// Len returns the length of the priority queue
func (pq PriorityQueue) Len() int {
    return len(pq)
}

// Less compares two items based on their cost
func (pq PriorityQueue) Less(i, j int) bool {
    return pq[i].cost < pq[j].cost
}

// Swap swaps two items in the priority queue
func (pq PriorityQueue) Swap(i, j int) {
    pq[i], pq[j] = pq[j], pq[i]
    pq[i].index = i
    pq[j].index = j
}

// Push adds an item to the priority queue
func (pq *PriorityQueue) Push(x interface{}) {
    n := len(*pq)
    item := x.(*Item)
    item.index = n
    *pq = append(*pq, item)
}

// Pop removes the item with the lowest cost from the priority queue
func (pq *PriorityQueue) Pop() interface{} {
    old := *pq
    n := len(old)
    item := old[n-1]
    *pq = old[0 : n-1]
    return item
}

// IsGoal checks if the current state is the goal state
func (s State) IsGoal(totalDisks int) bool {
    return len(s.Pegs[2]) == totalDisks
}

// GetNeighbors generates all possible next states
func (s State) GetNeighbors() []State {
    var neighbors []State
    for i := 0; i < 3; i++ {
        if len(s.Pegs[i]) > 0 { // If peg i is not empty
            disk := s.Pegs[i][len(s.Pegs[i])-1] // Get the top disk
            for j := 0; j < 3; j++ {
                if i != j && (len(s.Pegs[j]) == 0 || s.Pegs[j][len(s.Pegs[j])-1] > disk) {
                    newPegs := make([][]int, 3)
                    for k := 0; k < 3; k++ {
                        newPegs[k] = append([]int{}, s.Pegs[k]...) // Copy current state
                    }
                    newPegs[i] = newPegs[i][:len(newPegs[i])-1] // Remove disk from peg i
                    newPegs[j] = append(newPegs[j], disk) // Add disk to peg j
                    neighbors = append(neighbors, State{Pegs: newPegs})
                }
            }
        }
    }
    return neighbors
}

// Heuristic estimates the cost to reach the goal
func Heuristic(s State, totalDisks int) int {
    return totalDisks - len(s.Pegs[2]) // Disks not in the target peg
}

// AStar implements the A* algorithm
func AStar(startState State, totalDisks int) bool {
    openList := &PriorityQueue{}
    heap.Init(openList)
    heap.Push(openList, &Item{state: startState, cost: 0})

    closedSet := make(map[string]struct{})

    for openList.Len() > 0 {
        currentItem := heap.Pop(openList).(*Item)
        currentState := currentItem.state

        if currentState.IsGoal(totalDisks) {
            return true // Goal reached
        }

        // Create a unique key for the current state
        key := fmt.Sprintf("%v", currentState.Pegs)
        closedSet[key] = struct{}{}

        for _, neighbor := range currentState.GetNeighbors() {
            // Check if the neighbor is already evaluated
            neighborKey := fmt.Sprintf("%v", neighbor.Pegs)
            if _, found := closedSet[neighborKey]; found {
                continue
            }

            gCost := currentItem.cost + 1
            hCost := Heuristic(neighbor, totalDisks)
            fCost := gCost + hCost

            heap.Push(openList, &Item{state: neighbor, cost: fCost})
        }
    }

    return false // No solution found
}

func main() {
    totalDisks := 3
    initialPegs := [][]int{{3, 2, 1}, {}, {}} // Start with all disks on peg 0
    startState := State{Pegs: initialPegs}

    result := AStar(startState, totalDisks)
    if result {
        fmt.Println("Solution found!")
    } else {
        fmt.Println("No solution found.")
    }
}
