package main

import (
	"testing"

	"github.com/ralreegorganon/cddamap/internal/gen/metadata"
	"github.com/ralreegorganon/cddamap/internal/gen/render"
	"github.com/ralreegorganon/cddamap/internal/gen/save"
	"github.com/ralreegorganon/cddamap/internal/gen/world"
)

var gw world.World
var gs save.Save
var gm metadata.Overmap

func BenchmarkSaveBuild(b *testing.B) {
	var s save.Save
	for n := 0; n < b.N; n++ {
		s, _ = save.Build("/Users/jj/code/Cataclysm-DDA/save/Spenard")
	}
	gs = s
}

func BenchmarkMetadatadBuild(b *testing.B) {
	s, _ := save.Build("/Users/jj/code/Cataclysm-DDA/save/Spenard")
	b.ResetTimer()

	var m metadata.Overmap
	for n := 0; n < b.N; n++ {
		m, _ = metadata.Build(s, "/Users/jj/code/Cataclysm-DDA")
	}
	gm = m
}

func BenchmarkWorldBuild(b *testing.B) {
	s, _ := save.Build("/Users/jj/code/Cataclysm-DDA/save/Spenard")
	m, _ := metadata.Build(s, "/Users/jj/code/Cataclysm-DDA")
	b.ResetTimer()

	var w world.World
	for n := 0; n < b.N; n++ {
		w, _ = world.Build(m, s)
	}
	gw = w
}

func BenchmarkRenderTerrainToImages(b *testing.B) {
	s, _ := save.Build("/Users/jj/code/Cataclysm-DDA/save/Spenard")
	m, _ := metadata.Build(s, "/Users/jj/code/Cataclysm-DDA")
	w, _ := world.Build(m, s)
	l := []int{10}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		render.Image(w, "/Users/jj/Desktop/GoTest", l, true, false, false, true)
	}
}

func BenchmarkRenderSeenToImages(b *testing.B) {
	s, _ := save.Build("/Users/jj/code/Cataclysm-DDA/save/Spenard")
	m, _ := metadata.Build(s, "/Users/jj/code/Cataclysm-DDA")
	w, _ := world.Build(m, s)
	l := []int{10}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		render.Image(w, "/Users/jj/Desktop/GoTest", l, false, true, false, true)
	}
}

func BenchmarkRenderSeenSolidToImages(b *testing.B) {
	s, _ := save.Build("/Users/jj/code/Cataclysm-DDA/save/Spenard")
	m, _ := metadata.Build(s, "/Users/jj/code/Cataclysm-DDA")
	w, _ := world.Build(m, s)
	l := []int{10}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		render.Image(w, "/Users/jj/Desktop/GoTest", l, false, false, true, true)
	}
}

func BenchmarkRenderAllToImages(b *testing.B) {
	s, _ := save.Build("/Users/jj/code/Cataclysm-DDA/save/Spenard")
	m, _ := metadata.Build(s, "/Users/jj/code/Cataclysm-DDA")
	w, _ := world.Build(m, s)
	l := []int{10}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		render.Image(w, "/Users/jj/Desktop/GoTest", l, true, true, true, true)
	}
}
