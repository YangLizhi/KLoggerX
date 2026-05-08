import type { Directive } from 'vue'

const PLACEHOLDER =
  'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxIiBoZWlnaHQ9IjEiLz4='

const lazyImgDirective: Directive<HTMLImageElement, string> = {
  mounted(el, binding) {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            el.src = binding.value
            observer.unobserve(el)
          }
        })
      },
      { rootMargin: '100px' }
    )

    // 先设置 1x1 透明 SVG 占位
    el.src = PLACEHOLDER
    observer.observe(el)
  },
  updated(el, binding) {
    // 当绑定值变化时直接更新 src
    if (binding.value !== binding.oldValue) {
      el.src = binding.value
    }
  },
}

export default lazyImgDirective
