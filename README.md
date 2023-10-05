# heuristix - Optimization Heuristics in Go

> **Warning**
> This is a work-in-progress. Code may be broken. Use it at your own risk.

## Introduction

**heuristix**, or just **hx**, is a collection of optimization metaheuristics
implementations in Go. It is an heuristic framework
to which you can "plug" your own improving and diversifying strategies
to solve a given optimization problem.

Implemented metaheuristics:

- Variable Neighborhood Descent (VND)
- Iterated Local Search (ILS)
- Simulated Annealing (SA)
- Tabu Search (TS)
- Genetic Algorithm (GA)

## Basic Usage

See `/examples`

## Importing
```go
import "github.com/nidoro/heuristix"
```

## Installation

```shell
go get github.com/nidoro/heuristix@latest
```
