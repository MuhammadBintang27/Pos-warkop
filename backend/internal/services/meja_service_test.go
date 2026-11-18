package services

import (
	"reflect"
	"testing"
)

func TestNamaMeja(t *testing.T) {
	got := namaMeja("Meja", 1, 3, nil)
	if want := []string{"Meja 1", "Meja 2", "Meja 3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	got = namaMeja("Meja", 1, 4, map[string]bool{"Meja 2": true, "Meja 3": true, "Lesehan 1": true})
	if want := []string{"Meja 1", "Meja 4"}; !reflect.DeepEqual(got, want) {
		t.Errorf("nama yang sudah ada harus dilewati: got %v, want %v", got, want)
	}

	if got := namaMeja("Meja", 10, 2, nil); got[0] != "Meja 10" || got[1] != "Meja 11" {
		t.Errorf("mulai dari 10 salah: %v", got)
	}
}
