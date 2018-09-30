package main

import (
	"sync"
	"testing"
)

var GlobalInt int
var wg sync.WaitGroup

var tests = []struct{
		t bool
		p bool
		m bool
}{

	{false, false, false},
	{false, false, true},
	{false, true, false},
	{false, true, true},
	{true, false, false},
	{true, false, true},
	{true, true, false},
	{true, true, true},
}

func BenchmarkAgent(b *testing.B) {
	wg.Add(1)
	ingredients := make(chan ingredientTypes)
	signal := make(chan bool)

	go agent(ingredients, signal, &wg)

	for i := 0; i < b.N; i++ {

	}
}


func BenchmarkHelper(b *testing.B) {
	wg.Add(1)
	ingCh := make(chan ingredientTypes)
	smokers := smokerTypes{
		tobacco: make(chan bool),
		matches: make(chan bool),
		paper: make(chan bool),
	}
	signal := make(chan bool)

	for _, test := range tests {
		ing := ingredientTypes{
			tobacco: test.t,
			matches: test.m,
			paper: test.p,
		}
		helper(ingCh <- ing, smokers, signal, &wg)
	}

}

