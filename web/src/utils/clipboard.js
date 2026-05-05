/**
 * 复制文本到剪贴板（兼容 HTTP 环境）
 * navigator.clipboard 在非 HTTPS 非 localhost 环境下不可用，fallback 到 execCommand
 */
export async function copyToClipboard(text) {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(text)
    return
  }
  // fallback: 创建临时 textarea，始终挂到 body 上避免 Drawer/Modal 的 focus trap 干扰
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  textarea.style.top = '-9999px'
  textarea.style.opacity = '0'
  textarea.setAttribute('readonly', '')
  document.body.appendChild(textarea)
  // 临时移除 Drawer 的 aria-hidden，避免阻止 textarea 获取焦点
  const drawers = document.querySelectorAll('.n-drawer-container[aria-hidden="true"]')
  drawers.forEach(el => el.removeAttribute('aria-hidden'))
  textarea.focus()
  textarea.setSelectionRange(0, textarea.value.length)
  let ok = false
  try {
    ok = document.execCommand('copy')
  } finally {
    document.body.removeChild(textarea)
  }
  if (!ok) throw new Error('execCommand copy failed')
}
