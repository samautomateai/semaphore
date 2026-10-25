package projects

import (
	"runtime"
	"testing"
)

func TestIsValidInventoryPath(t *testing.T) {
	if !IsValidInventoryPath("..hidden/inventory") {
		t.Fatal("dot-dot-prefixed filename within cwd should be valid")
	}

	if IsValidInventoryPath("../outside/inventory") {
		t.Fatal("parent directory traversal must be rejected")
	}

	if !IsValidInventoryPath("inventories/test") {
		t.Fatal(" a path below the cwd should be valid")
	}

	if !IsValidInventoryPath("inventories/test/../prod") {
		t.Fatal(" a path below the cwd should be valid")
	}

	if IsValidInventoryPath("/test/../../../inventory") {
		t.Fatal(" a path out of the cwd should be invalid")
	}

	if IsValidInventoryPath("/test/inventory") {
		t.Fatal(" a path out of the cwd should be invalid")
	}

	if runtime.GOOS == "windows" && IsValidInventoryPath("c:\\test\\inventory") {
		t.Fatal(" a path out of the cwd should be invalid")
	}
}
