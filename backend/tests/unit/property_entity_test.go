package unit

import (
	"testing"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/google/uuid"
)

func TestNewProperty_Valid(t *testing.T) {
	p, err := entity.NewProperty("Farm Plot A", "FP-001", nil, nil, "LAND", nil, 500.0, uuid.New(), nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Name != "Farm Plot A" {
		t.Errorf("expected name 'Farm Plot A', got %q", p.Name)
	}
	if p.PropertyCode != "FP-001" {
		t.Errorf("expected code 'FP-001', got %q", p.PropertyCode)
	}
	if p.PropertyType != "LAND" {
		t.Errorf("expected type 'LAND', got %q", p.PropertyType)
	}
	if !p.IsActive {
		t.Error("expected IsActive true")
	}
	if !p.IsAvailableForRent {
		t.Error("expected IsAvailableForRent true")
	}
	if p.ID == (uuid.UUID{}) {
		t.Error("expected non-zero UUID")
	}
}

func TestNewProperty_DefaultType(t *testing.T) {
	p, err := entity.NewProperty("Plot", "P-001", nil, nil, "", nil, 100.0, uuid.New(), nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.PropertyType != "LAND" {
		t.Errorf("expected default type 'LAND', got %q", p.PropertyType)
	}
}

func TestNewProperty_NameRequired(t *testing.T) {
	_, err := entity.NewProperty("", "P-001", nil, nil, "LAND", nil, 100.0, uuid.New(), nil, nil)
	if err != entity.ErrPropertyNameRequired {
		t.Errorf("expected ErrPropertyNameRequired, got %v", err)
	}
}

func TestNewProperty_NameTooLong(t *testing.T) {
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := entity.NewProperty(string(longName), "P-001", nil, nil, "LAND", nil, 100.0, uuid.New(), nil, nil)
	if err != entity.ErrPropertyNameTooLong {
		t.Errorf("expected ErrPropertyNameTooLong, got %v", err)
	}
}

func TestNewProperty_CodeRequired(t *testing.T) {
	_, err := entity.NewProperty("Plot", "", nil, nil, "LAND", nil, 100.0, uuid.New(), nil, nil)
	if err != entity.ErrPropertyCodeRequired {
		t.Errorf("expected ErrPropertyCodeRequired, got %v", err)
	}
}

func TestNewProperty_SizeRequired(t *testing.T) {
	_, err := entity.NewProperty("Plot", "P-001", nil, nil, "LAND", nil, 0, uuid.New(), nil, nil)
	if err != entity.ErrPropertySizeRequired {
		t.Errorf("expected ErrPropertySizeRequired, got %v", err)
	}
}

func TestNewProperty_NegativeSize(t *testing.T) {
	_, err := entity.NewProperty("Plot", "P-001", nil, nil, "LAND", nil, -5.0, uuid.New(), nil, nil)
	if err != entity.ErrPropertySizeRequired {
		t.Errorf("expected ErrPropertySizeRequired, got %v", err)
	}
}

func TestNewProperty_WithAllFields(t *testing.T) {
	addr := entity.NewAddress("123 Farm Rd", "Guimba", "Nueva Ecija", "3115", "Philippines")
	geojson := `{"type":"Point","coordinates":[120.77,15.66]}`
	acres := 1.24
	rent := 5000.0
	desc := "Rice paddy"

	p, err := entity.NewProperty("Big Farm", "BF-001", addr, &geojson, "AGRICULTURAL", &acres, 5000.0, uuid.New(), &rent, &desc)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Address == nil || p.Address.City != "Guimba" {
		t.Error("address mismatch")
	}
	if *p.GeoJSONCoordinates != geojson {
		t.Error("geojson mismatch")
	}
	if *p.SizeInAcres != acres {
		t.Error("acres mismatch")
	}
	if *p.MonthlyRentAmount != rent {
		t.Error("rent mismatch")
	}
	if *p.Description != desc {
		t.Error("description mismatch")
	}
}
