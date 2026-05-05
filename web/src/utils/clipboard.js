/**
 * 复制文本到剪贴板（兼容 HTTP 环境和 Drawer/Modal 焦点陷阱）
 */
export async function copyToClipboard(text) {
  // 优先使用 Clipboard API（需要 HTTPS 或 localhost）
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text)
      return
    } catch {
      // Clipboard API 失败，走 fallback
    }
  }

  // fallback: 创建临时 textarea，放入焦点陷阱容器内避免被抢焦点
  const container =
    document.activeElement?.closest('.n-drawer-content, .n-modal, .n-dialog') ||
    document.body

  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.cssText = 'position:fixed;left:-9999px;top:-9999px;opacity:0;'
  textarea.setAttribute('readonly', '')
  container.appendChild(textarea)
  textarea.focus()
  textarea.setSelectionRange(0, textarea.value.length)

  let ok = false
  try {
    ok = document.execCommand('copy')
  } finally {
    container.removeChild(textarea)
  }
  if (!ok) throw new Error('execCommand copy failed')
}
