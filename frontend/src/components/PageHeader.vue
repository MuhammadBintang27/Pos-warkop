<script setup>
defineProps({ icon: String, title: String, subtitle: String })
</script>

<template>
  <section class="band">
    <div class="icon"><i :class="icon" /></div>
    <div class="title">
      <h1>{{ title }}</h1>
      <small v-if="subtitle">{{ subtitle }}</small>
    </div>
    <div class="actions"><slot /></div>
  </section>
</template>

<style scoped>
.band {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 16px;
  padding: 12px 16px;
  border-radius: 14px;
  background: var(--p-primary-color);
  color: var(--p-primary-contrast-color);
  box-shadow: 0 4px 14px color-mix(in srgb, var(--p-primary-color) 30%, transparent);
}
.icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  font-size: 1.4rem;
  background: color-mix(in srgb, var(--p-primary-contrast-color) 18%, transparent);
}
.title h1 {
  margin: 0;
  font-size: 1.35rem;
  line-height: 1.2;
}
.title small {
  opacity: 0.85;
}
.actions {
  margin-left: auto;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
</style>
