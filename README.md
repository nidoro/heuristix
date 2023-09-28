# heuristix - Optimization Heuristics in Go

> **Warning**
> This is a work-in-progress. Use it at your own risk.

## Introduction

**heuristix** is a framework for implementing optimization heuristics
in Go. It implements the basic algorithm for various metaheuristics,
to which you can "plug" your own improving and diversigying strategies
to solve an optimization problem defined by you.

Implemented metaheuristics:

- VND - Variable Neighborhood Descent
- ILS - Iterated Local Search
- SA - Simulated Annealing
- TS - Tabu Search

## Basic Usage

See `/examples`

## Importing

    import github.com/nidoro/heuristix

## Installation

    go get github.com/nidoro/heuristix@latest
