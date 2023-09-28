# heuristix - Optimization Heuristics in Go

> **Warning**
> This is a work-in-progress. Use it at your own risk.

## Introduction

**heuristix**, or just **hx**, is a collection of optimization metaheuristics
implementations in Go. It is an heuristic implementation framework
to which you can "plug" your own improving and diversifying strategies
to solve a given optimization.

Implemented metaheuristics:

- VND - Variable Neighborhood Descent
- ILS - Iterated Local Search
- SA - Simulated Annealing
- TS - Tabu Search

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
