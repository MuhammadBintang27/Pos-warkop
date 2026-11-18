<script setup>
defineProps({ icon: String, label: String, value: String })
</script>

<template>
  <div class="panel stat">
    <div class="ico"><i :class="icon" /></div>
    <div>
      <small>{{ label }}</small>
      <strong>{{ value }}</strong>
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
</style>
