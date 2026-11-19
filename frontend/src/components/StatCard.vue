<script setup>
import { computed } from 'vue'

// delta: perubahan % dibanding periode sebelumnya (null = tidak ditampilkan); turunBaik untuk angka yang lebih baik bila turun
const props = defineProps({ icon: String, label: String, value: String, delta: { type: Number, default: null }, turunBaik: Boolean })
const baik = computed(() => (props.delta > 0) !== props.turunBaik)
</script>

<template>
  <div class="panel stat">
    <div class="ico"><i :class="icon" /></div>
    <div>
      <small>{{ label }}</small>
      <strong>{{ value }}</strong>
      <span v-if="delta !== null" class="delta" :class="delta === 0 ? 'datar' : baik ? 'naik' : 'turun'">
        <i :class="delta > 0 ? 'pi pi-arrow-up' : delta < 0 ? 'pi pi-arrow-down' : 'pi pi-minus'" />
        {{ Math.abs(delta) }}% <span class="vs">vs periode lalu</span>
      </span>
    </div>
  </div>
</template>

<style scoped>
.stat {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
}
.ico {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  font-size: 1.2rem;
  background: color-mix(in srgb, var(--p-primary-color) 14%, transparent);
  color: var(--accent-text);
}
small {
  display: block;
  color: var(--p-text-muted-color);
}
strong {
  font-size: 1.25rem;
}
.delta {
  display: block;
  font-size: 0.78rem;
  font-weight: 600;
}
.delta i {
  font-size: 0.7rem;
}
.naik {
  color: var(--p-green-500, #22c55e);
}
.turun {
  color: var(--p-red-500, #ef4444);
}
.datar,
.vs {
  color: var(--p-text-muted-color);
  font-weight: 400;
}
</style>
