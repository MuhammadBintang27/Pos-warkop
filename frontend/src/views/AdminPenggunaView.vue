<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { get, post, put } from '@/api/http'
import { useRun } from '@/composables/useRun'

const ROLES = [
  { label: 'Pelayan', value: 'pelayan' },
  { label: 'Kasir', value: 'kasir' },
  { label: 'Owner', value: 'owner' },
]
const SEVERITY = { pelayan: 'secondary', kasir: 'info', owner: 'warn' }

const run = useRun()
const route = useRoute()
const kliens = ref([])
const expanded = ref({})
const users = reactive({}) // klienId -> daftar pengguna, dimuat saat baris dibuka
const form = ref(null)

const muatUsers = async (id) => (users[id] = await get(`/admin/klien/${id}/users`))
const bukaBaris = (e) => run(() => muatUsers(e.data.id))
const base = (id) => `/admin/klien/${id}/users`

const valid = () => {
  const f = form.value
  return f?.nama.trim() && (f.id ? !f.password || f.password.length >= 6 : /^[a-z0-9._-]{3,30}$/.test(f.username.trim().toLowerCase()) && f.password.length >= 6)
}

function acak() {
  const abjad = 'abcdefghjkmnpqrstuvwxyz23456789'
  const b = crypto.getRandomValues(new Uint8Array(10))
  form.value.password = Array.from(b, (x) => abjad[x % abjad.length]).join('')
}

async function simpan() {
  const { id, klienId, nama, username, password, role, aktif } = form.value
  const req = id
    ? () => put(`${base(klienId)}/${id}`, { nama: nama.trim(), role, aktif, password })
    : () => post(base(klienId), { nama: nama.trim(), username: username.trim(), password, role })
  if (await run(req, 'Pengguna disimpan')) {
    form.value = null
    await run(() => muatUsers(klienId))
  }
}

async function ubahAktif(klienId, u) {
  if (await run(() => put(`${base(klienId)}/${u.id}`, { nama: u.nama, role: u.role, aktif: !u.aktif }))) await run(() => muatUsers(klienId))
}

onMounted(async () => {
  await run(async () => (kliens.value = await get('/admin/klien')))
  const dariUrl = kliens.value.find((k) => k.id === Number(route.query.klien))
  if (dariUrl) {
    expanded.value = { [dariUrl.id]: true }
    await run(() => muatUsers(dariUrl.id))
  }
})
</script>

<template>
  <div class="stack gap">
    <PageHeader icon="pi pi-users" title="Pengguna per klien" subtitle="Buka sebuah klien untuk melihat dan mengelola penggunanya" />

    <section class="panel box">
      <DataTable v-model:expanded-rows="expanded" :value="kliens" data-key="id" size="small" empty-message="Belum ada klien" @row-expand="bukaBaris">
        <Column expander style="width: 3rem" />
        <Column field="id" header="ID Klien" />
        <Column header="Nama usaha"><template #body="{ data }"><strong>{{ data.nama }}</strong></template></Column>
        <Column field="alamat" header="Alamat" />
        <Column field="pengguna" header="Pengguna aktif" />
        <Column header="Akses">
          <template #body="{ data }"><Tag :value="data.aktif ? 'Aktif' : 'Nonaktif'" :severity="data.aktif ? 'success' : 'danger'" /></template>
        </Column>

        <template #expansion="{ data: klien }">
          <div class="dalam">
            <div class="atas">
              <span>Awalan username: <b>{{ klien.id }}-</b></span>
              <Button
                label="Tambah pengguna"
                icon="pi pi-plus"
                size="small"
                @click="form = { klienId: klien.id, nama: '', username: '', password: '', role: 'pelayan', aktif: true }"
              />
            </div>
            <DataTable :value="users[klien.id] || []" size="small" empty-message="Belum ada pengguna">
              <Column field="nama" header="Nama user" />
              <Column field="username" header="Username" />
              <Column header="Role"><template #body="{ data }"><Tag :value="data.role" :severity="SEVERITY[data.role]" /></template></Column>
              <Column header="Status">
                <template #body="{ data }"><Tag :value="data.aktif ? 'Aktif' : 'Nonaktif'" :severity="data.aktif ? 'success' : 'danger'" /></template>
              </Column>
              <Column header="Aksi">
                <template #body="{ data }">
                  <div class="act">
                    <Button label="Ubah / reset password" icon="pi pi-key" size="small" text severity="secondary" @click="form = { ...data, klienId: klien.id, password: '' }" />
                    <Button :label="data.aktif ? 'Nonaktifkan' : 'Aktifkan'" size="small" text :severity="data.aktif ? 'danger' : 'success'" @click="ubahAktif(klien.id, data)" />
                  </div>
                </template>
              </Column>
            </DataTable>
          </div>
        </template>
      </DataTable>
    </section>

    <Dialog :visible="!!form" modal :header="form?.id ? 'Ubah pengguna' : 'Tambah pengguna'" :style="{ width: '26rem' }" @update:visible="form = null">
      <div v-if="form" class="stack">
        <label for="nama">Nama</label>
        <InputText id="nama" v-model="form.nama" maxlength="100" fluid />
        <label for="user">Username</label>
        <InputText v-if="form.id" id="user" :model-value="form.username" disabled fluid />
        <InputGroup v-else>
          <InputGroupAddon>{{ form.klienId }}-</InputGroupAddon>
          <InputText id="user" v-model="form.username" maxlength="30" placeholder="nama.pengguna" />
        </InputGroup>
        <label for="pw">{{ form.id ? 'Password baru (kosongkan bila tidak diganti)' : 'Password (min. 6 karakter)' }}</label>
        <InputGroup>
          <InputText id="pw" v-model="form.password" autocomplete="off" />
          <Button icon="pi pi-refresh" severity="secondary" aria-label="Buat password acak" @click="acak" />
        </InputGroup>
        <small class="hint">Password terlihat supaya bisa disalin dan diberikan ke pengguna.</small>
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
.dalam {
  padding: 4px 8px 12px 40px;
}
.atas {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
  color: var(--p-text-muted-color);
}
.act {
  display: flex;
  gap: 4px;
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
