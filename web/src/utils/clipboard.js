/**
 * 复制文本到剪贴板（兼容 HTTP 环境）
 * navigator.clipboard 在非 HTTPS 非 localhost 环境下不可用，fallback 到 execCommand
 */
export async function copyToClipboard(text) {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(text)
    return
  }
  // fallback: 创建临时 textarea 复制
  const textarea = document.createElement('textarea')
  textarea.value = text
  // 放到当前活跃元素所在的容器中（兼容 Drawer/Modal 等 focus trap 场景）
  const container = document.activeElement?.closest('.n-drawer-content, .n-modal, .n-dialog') || document.body
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  textarea.style.opacity = '0'
  container.appendChild(textarea)
  textarea.focus()
  textarea.select()
  try {
    document.execCommand('copy')
  } finally {
    container.removeChild(textarea)
  }
}
