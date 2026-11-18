<script setup>
import { onMounted, ref } from 'vue'
import { get, post, put } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { useAuthStore } from '@/stores/auth'

const ROLES = [
  { label: 'Pelayan', value: 'pelayan' },
  { label: 'Kasir', value: 'kasir' },
  { label: 'Owner', value: 'owner' },
]
const SEVERITY = { pelayan: 'secondary', kasir: 'info', owner: 'warn' }

const run = useRun()
const auth = useAuthStore()
const users = ref([])
const form = ref(null)

const muat = async () => (users.value = await get('/users'))
const valid = () => {
  const f = form.value
  return f?.nama.trim() && (f.id ? !f.password || f.password.length >= 6 : /^[a-z0-9._-]{3,30}$/.test(f.username.trim().toLowerCase()) && f.password.length >= 6)
}

async function simpan() {
  const { id, nama, username, password, role, aktif } = form.value
  const req = id
    ? () => put(`/users/${id}`, { nama: nama.trim(), role, aktif, password })
    : () => post('/users', { nama: nama.trim(), username: username.trim(), password, role })
  if (await run(req, 'Pengguna disimpan')) {
    form.value = null
    await run(muat)
  }
}

onMounted(() => run(muat))
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-users" title="Pengguna" subtitle="Akun pelayan, kasir, dan owner">
      <button class="pill" @click="form = { nama: '', username: '', password: '', role: 'pelayan', aktif: true }">
        <i class="pi pi-plus" /> Tambah pengguna
      </button>
    </PageHeader>

    <section class="panel box">
      <DataTable :value="users" size="small" empty-message="Belum ada pengguna">
        <Column field="nama" header="Nama" />
        <Column field="username" header="Username" />
        <Column header="Role"><template #body="{ data }"><Tag :value="data.role" :severity="SEVERITY[data.role]" /></template></Column>
        <Column header="Status">
          <template #body="{ data }"><Tag :value="data.aktif ? 'Aktif' : 'Nonaktif'" :severity="data.aktif ? 'success' : 'danger'" /></template>
        </Column>
        <Column>
          <template #body="{ data }">
            <Button icon="pi pi-pencil" text severity="secondary" aria-label="Ubah" @click="form = { ...data, password: '' }" />
          </template>
        </Column>
      </DataTable>
    </section>

    <Dialog :visible="!!form" modal :header="form?.id ? 'Ubah pengguna' : 'Tambah pengguna'" :style="{ width: '26rem' }" @update:visible="form = null">
      <div v-if="form" class="stack">
        <label for="nama">Nama</label>
        <InputText id="nama" v-model="form.nama" maxlength="100" fluid />
        <label for="user">Username</label>
        <InputText v-if="form.id" id="user" :model-value="form.username" disabled fluid />
        <InputGroup v-else>
          <InputGroupAddon>{{ auth.user.klienId }}-</InputGroupAddon>
          <InputText id="user" v-model="form.username" maxlength="30" placeholder="nama.pengguna" />
        </InputGroup>
        <small v-if="!form.id" class="hint">Username final: <b>{{ auth.user.klienId }}-{{ form.username.trim().toLowerCase() || '…' }}</b>. Huruf kecil, angka, titik, garis bawah, strip.</small>
        <label for="pw">{{ form.id ? 'Password baru (kosongkan bila tidak diganti)' : 'Password (min. 6 karakter)' }}</label>
        <Password v-model="form.password" input-id="pw" :feedback="false" toggle-mask fluid />
        <label>Role</label>
        <Select v-model="form.role" :options="ROLES" option-label="label" option-value="value" fluid />
        <div v-if="form.id" class="sw"><ToggleSwitch v-model="form.aktif" input-id="aktif" /><label for="aktif">Akun aktif</label></div>
      </div>
      <template #footer>
        <Button label="Simpan" icon="pi pi-check" :disabled="!valid()" @click="simpan" />
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.gap {
  gap: 14px;
}
.box {
  padding: 14px;
}
.hint {
  color: var(--p-text-muted-color);
}
.sw {
  display: flex;
  align-items: center;
  gap: 10px;
}
.sw label {
  font-weight: 500;
}
</style>
