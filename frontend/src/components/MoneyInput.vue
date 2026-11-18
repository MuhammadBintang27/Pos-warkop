<script setup>
const model = defineModel({ type: Number, default: 0 })
</script>

<template>
  <InputNumber v-model="model" mode="currency" currency="IDR" locale="id-ID" :max-fraction-digits="0" :min="0" fluid />
</template>
