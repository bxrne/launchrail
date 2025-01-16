package components_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs/components"
)

// TEST: GIVEN a nosecone component WHEN String is called THEN the type is returned.
func TestNosecone_String(t *testing.T) {
	nosecone := components.Nosecone{}

	expected := "Nosecone{Position: {0 0 0}, Radius: 0.00, Length: 0.00, Mass: 0.00, ShapeParameter: 0.00}"
	actual := nosecone.String()

	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TEST: GIVEN a nosecone component and a delta time WHEN Update is called THEN the component is updated.
func TestNosecone_Update(t *testing.T) {
	nosecone := components.Nosecone{}
	nosecone.Update(1.0)
}

// TEST: GIVEN nothing WHEN NewNosecone is called THEN a new Nosecone instance is returned.
func TestNewNosecone(t *testing.T) {
	nosecone := components.NewNosecone(1.0, 1.0, 1.0, 1.0)
	if nosecone == nil {
		t.Errorf("Expected Nosecone instance, got nil")
	}
}
