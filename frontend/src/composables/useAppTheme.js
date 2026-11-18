import { definePreset } from '@primeuix/themes'
import Aura from '@primeuix/themes/aura'
import Lara from '@primeuix/themes/lara'
import Material from '@primeuix/themes/material'
import Nora from '@primeuix/themes/nora'
import { load, save } from '@/utils/storage'

export const THEMES = [
  { id: 'aura', label: 'Aura', base: Aura },
  { id: 'lara', label: 'Lara', base: Lara },
  { id: 'material', label: 'Material', base: Material },
  { id: 'nora', label: 'Nora', base: Nora },
]

export const PRIMARY_COLORS = [
  { name: 'amber', hex: '#f59e0b' },
  { name: 'orange', hex: '#f97316' },
  { name: 'red', hex: '#ef4444' },
  { name: 'rose', hex: '#f43f5e' },
  { name: 'pink', hex: '#ec4899' },
  { name: 'fuchsia', hex: '#d946ef' },
  { name: 'purple', hex: '#a855f7' },
  { name: 'violet', hex: '#8b5cf6' },
  { name: 'indigo', hex: '#6366f1' },
  { name: 'blue', hex: '#3b82f6' },
  { name: 'sky', hex: '#0ea5e9' },
  { name: 'cyan', hex: '#06b6d4' },
  { name: 'teal', hex: '#14b8a6' },
  { name: 'emerald', hex: '#10b981' },
  { name: 'green', hex: '#22c55e' },
  { name: 'lime', hex: '#84cc16' },
  { name: 'yellow', hex: '#eab308' },
  { name: 'slate', hex: '#64748b' },
]

export const DEFAULT_SETTINGS = {
  namaUsaha: 'Warkop',
  alamat: '',
  logo: '',
  printerKasirId: null,
  footerStruk: '',
  theme: 'aura',
  primary: 'amber',
  mode: 'light',
}

const STORAGE_KEY = 'warkopTheme'
const STEPS = [50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950]
const DARK_TEXT = '#1c1917'
const HEX = /^#[0-9a-f]{6}$/i

export const isHex = (s) => HEX.test(s)

const toRgb = (hex) => [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16))
const toHex = (rgb) => '#' + rgb.map((v) => v.toString(16).padStart(2, '0')).join('')
const mix = (a, b, t) => a.map((v, i) => Math.round(v + (b[i] - v) * t))

// ponytail: campuran RGB ke putih/hitam, bukan OKLCH; upgrade bila ada warna brand yang hasilnya kusam.
export function buildPalette(hex) {
  const base = toRgb(hex)
  const light = (t) => toHex(mix(base, [255, 255, 255], t))
  const dark = (t) => toHex(mix(base, [0, 0, 0], t))
  return {
    50: light(0.95), 100: light(0.88), 200: light(0.75), 300: light(0.55), 400: light(0.3),
    500: hex.toLowerCase(),
    600: dark(0.12), 700: dark(0.25), 800: dark(0.38), 900: dark(0.5), 950: dark(0.65),
  }
}

const luminance = (hex) => {
  const [r, g, b] = toRgb(hex).map((v) => {
    const c = v / 255
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

/** Warna teks di atas warna primer: putih atau gelap, mana yang kontrasnya lebih tinggi. */
export function contrastFor(hex) {
  const l = luminance(hex)
  return 1.05 / (l + 0.05) >= (l + 0.05) / (luminance(DARK_TEXT) + 0.05) ? '#ffffff' : DARK_TEXT
}

const themeById = (id) => THEMES.find((t) => t.id === id) || THEMES[0]

const hex500 = (primary) =>
  isHex(primary) ? primary : (PRIMARY_COLORS.find((c) => c.name === primary) || PRIMARY_COLORS[0]).hex

function build(s) {
  const named = !isHex(s.primary) && PRIMARY_COLORS.some((c) => c.name === s.primary)
  const palette = isHex(s.primary)
    ? buildPalette(s.primary)
    : Object.fromEntries(STEPS.map((n) => [n, `{${named ? s.primary : 'amber'}.${n}}`]))
  return definePreset(themeById(s.theme).base, {
    semantic: {
      primary: palette,
      colorScheme: {
        light: {
          primary: {
            color: '{primary.500}',
            contrastColor: contrastFor(hex500(s.primary)),
            hoverColor: '{primary.600}',
            activeColor: '{primary.700}',
          },
        },
        dark: {
          primary: {
            color: '{primary.400}',
            contrastColor: DARK_TEXT,
            hoverColor: '{primary.300}',
            activeColor: '{primary.200}',
          },
        },
      },
    },
  })
}

let current = { ...DEFAULT_SETTINGS }
let primeConfig = null

function setDark(mode) {
  const auto = typeof window !== 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches
  document.documentElement.classList.toggle('app-dark', mode === 'dark' || (mode === 'auto' && !!auto))
}

export const loadCachedSettings = () => ({ ...DEFAULT_SETTINGS, ...load(STORAGE_KEY, {}) })

export function initTheme() {
  current = loadCachedSettings()
  setDark(current.mode)
  if (typeof window !== 'undefined') {
    window.matchMedia?.('(prefers-color-scheme: dark)').addEventListener('change', () => {
      if (current.mode === 'auto') setDark('auto')
    })
  }
  return build(current)
}

export function bindConfig(config) {
  primeConfig = config
}

export function applyTheme(settings, { persist = true } = {}) {
  current = { ...DEFAULT_SETTINGS, ...settings }
  // lewat config.theme, bukan usePreset(): usePreset() bisa mengubah salinan @primeuix/styled yang salah
  if (primeConfig) primeConfig.theme = { ...primeConfig.theme, preset: build(current) }
  setDark(current.mode)
  document.title = current.namaUsaha
  if (persist) save(STORAGE_KEY, current)
}
