<script setup>
import { onMounted, ref } from 'vue'
import { useCari } from '@/composables/useCari'
import { get, post, put } from '@/api/http'
import { useRun } from '@/composables/useRun'
import { useAuthStore } from '@/stores/auth'

const ROLES = [
  { label: 'Pelayan', value: 'pelayan' },
  { label: 'Kasir', value: 'kasir' },
  { label: 'Owner', value: 'owner' },
  { label: 'Tenant', value: 'tenant' },
]
const SEVERITY = { pelayan: 'secondary', kasir: 'info', owner: 'warn' }

const run = useRun()
const auth = useAuthStore()
const users = ref([])
const tenants = ref([])
const form = ref(null)

const muat = async () => {
  ;[users.value, tenants.value] = await Promise.all([get('/users'), get('/tenant')])
}
const namaTenant = (id) => tenants.value.find((t) => t.id === id)?.nama ?? '—'
const valid = () => {
  const f = form.value
  if (f?.role === 'tenant' && !f.tenantId) return false
  return f?.nama.trim() && (f.id ? !f.password || f.password.length >= 6 : /^[a-z0-9._-]{3,30}$/.test(f.username.trim().toLowerCase()) && f.password.length >= 6)
}

async function simpan() {
  const { id, nama, username, password, role, aktif } = form.value
  const tenantId = role === 'tenant' ? form.value.tenantId : 0
  const req = id
    ? () => put(`/users/${id}`, { nama: nama.trim(), role, aktif, password, tenantId })
    : () => post('/users', { nama: nama.trim(), username: username.trim(), password, role, tenantId })
  if (await run(req, 'Pengguna disimpan')) {
    form.value = null
    await run(muat)
  }
}

onMounted(() => run(muat))
const { cari: cariUser, hasil: usersTampil } = useCari(users, 'nama', 'username', 'role')
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-users" title="Pengguna" subtitle="Akun pelayan, kasir, owner, dan tenant">
      <button class="pill" @click="form = { nama: '', username: '', password: '', role: 'pelayan', aktif: true }">
        <i class="pi pi-plus" /> Tambah pengguna
      </button>
    </PageHeader>

    <section class="panel box">
      <CariInput v-model="cariUser" placeholder="Cari nama atau username…" />
      <DataTable paginator :rows="10" :rows-per-page-options="[10, 25, 50]" :always-show-paginator="false" :value="usersTampil" size="small" empty-message="Belum ada pengguna">
        <Column field="nama" header="Nama" />
        <Column field="username" header="Username" />
        <Column header="Role">
          <template #body="{ data }">
            <div class="role">
              <Tag :value="data.role" :severity="SEVERITY[data.role]" />
              <small v-if="data.role === 'tenant'" class="hint"><i class="pi pi-shop" /> {{ namaTenant(data.tenantId) }}</small>
            </div>
          </template>
        </Column>
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
        <template v-if="form.role === 'tenant'">
          <label for="tenant">Tenant</label>
          <Select v-model="form.tenantId" input-id="tenant" :options="tenants" option-label="nama" option-value="id" placeholder="Pilih tenant" fluid />
          <small class="hint">Akun ini hanya bisa melihat pesanan masuk dan rekapan tenant tersebut.</small>
        </template>
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
.role {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
}
.role .pi {
  font-size: 0.8rem;
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
