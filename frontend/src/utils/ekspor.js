// Ekspor tabel ke .xlsx. exceljs dimuat hanya saat tombol ditekan supaya tidak membebani halaman lain.
// sheets: [{ nama, kolom: [{ header, key, lebar?, uang? }], baris: [{...}] }]
export async function unduhExcel(namaFile, sheets) {
  const { default: ExcelJS } = await import('exceljs')
  const wb = new ExcelJS.Workbook()
  for (const s of sheets) {
    const ws = wb.addWorksheet(s.nama.slice(0, 31)) // batas nama sheet Excel
    ws.columns = s.kolom.map((k) => ({ header: k.header, key: k.key, width: k.lebar ?? Math.max(12, k.header.length + 4) }))
    ws.addRows(s.baris)
    ws.getRow(1).font = { bold: true }
    ws.views = [{ state: 'frozen', ySplit: 1 }]
    s.kolom.forEach((k, i) => {
      if (k.uang) ws.getColumn(i + 1).numFmt = '"Rp"#,##0'
    })
  }
  const buf = await wb.xlsx.writeBuffer()
  const url = URL.createObjectURL(new Blob([buf], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }))
  const a = Object.assign(document.createElement('a'), { href: url, download: namaFile })
  a.click()
  URL.revokeObjectURL(url)
}
