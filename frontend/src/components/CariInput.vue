<script setup>
const model = defineModel({ type: String, default: '' })
defineProps({ placeholder: { type: String, default: 'Cari…' } })
</script>

<template>
  <IconField class="cari-input">
    <InputIcon class="pi pi-search" />
    <InputText v-model="model" :placeholder="placeholder" fluid />
  </IconField>
</template>

<style scoped>
.cari-input {
  width: 100%;
  max-width: 340px;
  margin-bottom: 10px;
}
</style>
