import { ref, onMounted, onUnmounted } from 'vue'

export function useElementWidth(fallback = 600) {
  const el = ref(null)
  const width = ref(fallback)
  let observer = null

  onMounted(() => {
    if (!el.value) return
    width.value = el.value.clientWidth || fallback
    observer = new ResizeObserver(([entry]) => {
      width.value = Math.max(120, Math.floor(entry.contentRect.width))
    })
    observer.observe(el.value)
  })

  onUnmounted(() => observer?.disconnect())

  return { el, width }
}
