# hx &mdash; Optimization Heuristix in Go

> **Warning**
> This is a work-in-progress. There are currently no stable versions, and the latest code may be broken. Use it at your own risk.

## Introduction

The **hx** package provides a collection of optimization
metaheuristics implemented in Go. It is a heuristic framework
to which you can "plug" your own strategies &mdash; also known as operators &mdash;
and solution structures.

Implemented metaheuristics:

- Variable Neighborhood Descent (VND)
- Iterated Local Search (ILS)
- Simulated Annealing (SA)
- Tabu Search (TS)
- Genetic Algorithm (GA)

## Basic Usage

See `/examples`

## Installation

```shell
go get github.com/nidoro/heuristix@latest
```

## Importing
```go
import "github.com/nidoro/heuristix"
```
