import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const owner = { roles: ['owner'] }
const tenant = { roles: ['tenant'] }

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { public: true } },
    {
      path: '/',
      component: () => import('@/layouts/AppLayout.vue'),
      children: [
        { path: '', name: 'kasir', component: () => import('@/views/KasirView.vue') },
        { path: 'admin', component: () => import('@/views/AdminView.vue'), meta: { roles: ['superadmin'] } },
        { path: 'admin/pengguna', component: () => import('@/views/AdminPenggunaView.vue'), meta: { roles: ['superadmin'] } },
        { path: 'denah', component: () => import('@/views/DenahView.vue') },
        { path: 'riwayat', component: () => import('@/views/RiwayatView.vue') },
        // satu route untuk semua tab laporan supaya halaman (dan data yang sudah dimuat) dipakai ulang saat pindah menu
        { path: 'laporan/:tab(waktu|karyawan|kasbon)?', component: () => import('@/views/LaporanView.vue'), meta: owner },
        { path: 'laporan/tenant', component: () => import('@/views/LaporanTenantView.vue'), meta: owner },
        { path: 'laba-rugi', component: () => import('@/views/LabaRugiView.vue'), meta: owner },
        { path: 'karyawan', component: () => import('@/views/KaryawanView.vue'), meta: owner },
        { path: 'ketersediaan', component: () => import('@/views/KetersediaanView.vue'), meta: { roles: ['kasir', 'owner'] } },
        { path: 'bahan', component: () => import('@/views/BahanView.vue'), meta: { roles: ['kasir', 'owner'] } },
        { path: 'menu', component: () => import('@/views/MenuView.vue'), meta: owner },
        { path: 'tenant', component: () => import('@/views/TenantView.vue'), meta: owner },
        { path: 'printer', component: () => import('@/views/PrinterView.vue'), meta: owner },
        { path: 'meja', component: () => import('@/views/MejaView.vue'), meta: owner },
        { path: 'pengguna', component: () => import('@/views/PenggunaView.vue'), meta: owner },
        { path: 'pengaturan', component: () => import('@/views/PengaturanView.vue'), meta: owner },
        { path: 'audit', component: () => import('@/views/AuditView.vue'), meta: owner },
        { path: 'tenant-saya', component: () => import('@/views/PesananMasukView.vue'), meta: tenant },
        { path: 'tenant-saya/rekap', component: () => import('@/views/RekapTenantView.vue'), meta: tenant },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.public) return auth.isLoggedIn ? '/' : true
  if (!auth.isLoggedIn) return '/login'
  if (auth.role === 'superadmin') return to.path.startsWith('/admin') ? true : '/admin'
  if (auth.role === 'tenant') return to.meta.roles?.includes('tenant') ? true : '/tenant-saya'
  if (to.meta.roles && !to.meta.roles.includes(auth.role)) return '/'
  return true
})

export default router
