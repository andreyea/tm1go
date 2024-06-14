package tm1go_test

import (
	"testing"
)

func TestCubeService_Get_Success(t *testing.T) {
	cubeName := "Balance Sheet"
	cube, err := tm1ServiceT.CubeService.Get(cubeName)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if cube == nil {
		t.Fatal("Expected a cube, got nil")
	}
	if cube.Name != cubeName {
		t.Errorf("Expected cube name %v, got %v", cubeName, cube.Name)
	}
}

func TestCubeService_Get_NonExistentCube(t *testing.T) {
	cubeName := "NonExistentCube"
	_, err := tm1ServiceT.CubeService.Get(cubeName)
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}
