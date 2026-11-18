package types

import "regexp"

var (
	hexColor      = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	primaryColors = map[string]bool{
		"emerald": true, "green": true, "lime": true, "red": true, "orange": true, "amber": true,
		"yellow": true, "teal": true, "cyan": true, "sky": true, "blue": true, "indigo": true,
		"violet": true, "purple": true, "fuchsia": true, "pink": true, "rose": true, "slate": true,
	}
)

// ValidPrimary menerima nama palet PrimeVue atau hex 6 digit (#RRGGBB).
func ValidPrimary(s string) bool { return primaryColors[s] || hexColor.MatchString(s) }

type Role string

const (
	RolePelayan Role = "pelayan"
	RoleKasir   Role = "kasir"
	RoleOwner   Role = "owner"
	RoleTenant  Role = "tenant"     // hanya melihat pesanan masuk dan rekapan tenant miliknya
	RoleSuper   Role = "superadmin" // bukan peran pengguna klien, jadi tidak termasuk Valid()
)

var KasirUp = []Role{RoleKasir, RoleOwner}

type PesananStatus string

const (
	StatusOpen   PesananStatus = "open"
	StatusLunas  PesananStatus = "lunas"
	StatusBatal  PesananStatus = "batal"
	StatusGabung PesananStatus = "gabung"
)

type Metode string

const (
	MetodeTunai    Metode = "tunai"
	MetodeQris     Metode = "qris"
	MetodeTransfer Metode = "transfer"
	MetodeKasbon   Metode = "kasbon"
)

func (r Role) Valid() bool {
	switch r {
	case RolePelayan, RoleKasir, RoleOwner, RoleTenant:
		return true
	}
	return false
}

func (r Role) IsKasirUp() bool { return r == RoleKasir || r == RoleOwner }

func (m Metode) Valid() bool {
	switch m {
	case MetodeTunai, MetodeQris, MetodeTransfer, MetodeKasbon:
		return true
	}
	return false
}
