export const konfirmasi = (confirm, { header, message, label = 'Hapus', accept }) =>
  confirm.require({
    header,
    message,
    icon: 'pi pi-exclamation-triangle',
    acceptLabel: label,
    rejectLabel: 'Batal',
    acceptProps: { severity: 'danger' },
    rejectProps: { severity: 'secondary', text: true },
    accept,
  })
