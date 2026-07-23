package adapters

import (
	"testing"
)

func BenchmarkGoGenerateProject(b *testing.B) {
	a := GoAdapter{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		artifacts := a.GenerateProject("bench-app")
		if len(artifacts) == 0 {
			b.Fatal("expected artifacts")
		}
	}
}

func BenchmarkGoGenerateModule(b *testing.B) {
	a := GoAdapter{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		artifacts := a.GenerateModule("core", "./core", "bench-app")
		if len(artifacts) == 0 {
			b.Fatal("expected artifacts")
		}
	}
}

func BenchmarkGoGenerateService(b *testing.B) {
	a := GoAdapter{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		artifacts := a.GenerateService("http-api", "http", 8080, "bench-app")
		if len(artifacts) == 0 {
			b.Fatal("expected artifacts")
		}
	}
}

func BenchmarkPythonGenerateProject(b *testing.B) {
	a := PythonAdapter{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		artifacts := a.GenerateProject("bench-app")
		if len(artifacts) == 0 {
			b.Fatal("expected artifacts")
		}
	}
}

func BenchmarkTypeScriptGenerateProject(b *testing.B) {
	a := TypeScriptAdapter{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		artifacts := a.GenerateProject("bench-app")
		if len(artifacts) == 0 {
			b.Fatal("expected artifacts")
		}
	}
}

func BenchmarkRustGenerateProject(b *testing.B) {
	a := RustAdapter{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		artifacts := a.GenerateProject("bench-app")
		if len(artifacts) == 0 {
			b.Fatal("expected artifacts")
		}
	}
}

func BenchmarkJavaGenerateProject(b *testing.B) {
	a := JavaAdapter{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		artifacts := a.GenerateProject("bench-app")
		if len(artifacts) == 0 {
			b.Fatal("expected artifacts")
		}
	}
}
