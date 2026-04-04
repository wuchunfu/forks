/**
 * Forks Animation Composable
 * Spark (交互魔术师) - Chromatic Team
 */

import { ref, computed, onMounted, onUnmounted } from 'vue'

export const animationConfig = {
  easings: {
    smooth: 'cubic-bezier(0.4, 0, 0.2, 1)',
    in: 'cubic-bezier(0.4, 0, 1, 1)',
    out: 'cubic-bezier(0, 0, 0.2, 1)',
    bounce: 'cubic-bezier(0.34, 1.56, 0.64, 1)',
    snappy: 'cubic-bezier(0.25, 0.46, 0.45, 0.94)'
  },
  durations: {
    instant: 50,
    fast: 150,
    base: 200,
    medium: 300,
    slow: 400,
    page: 400
  }
}

export function useReducedMotion() {
  const prefersReducedMotion = ref(false)
  
  const updatePreference = () => {
    prefersReducedMotion.value = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  }
  
  onMounted(() => {
    updatePreference()
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
    mediaQuery.addEventListener('change', updatePreference)
    
    onUnmounted(() => {
      mediaQuery.removeEventListener('change', updatePreference)
    })
  })
  
  return { prefersReducedMotion }
}

export function usePageTransition() {
  const { prefersReducedMotion } = useReducedMotion()
  
  const duration = computed(() => 
    prefersReducedMotion.value ? 0 : animationConfig.durations.page
  )
  
  const beforeEnter = (el) => {
    el.style.opacity = '0'
    el.style.transform = 'translateY(20px)'
  }
  
  const enter = (el, done) => {
    const delay = 50
    
    requestAnimationFrame(() => {
      setTimeout(() => {
        el.style.transition = 'opacity ' + duration.value + 'ms ' + animationConfig.easings.smooth + ', transform ' + duration.value + 'ms ' + animationConfig.easings.smooth
        el.style.opacity = '1'
        el.style.transform = 'translateY(0)'
        setTimeout(done, duration.value)
      }, delay)
    })
  }
  
  const leave = (el, done) => {
    const leaveDuration = duration.value * 0.75
    el.style.transition = 'opacity ' + leaveDuration + 'ms ' + animationConfig.easings.smooth + ', transform ' + leaveDuration + 'ms ' + animationConfig.easings.smooth
    el.style.opacity = '0'
    el.style.transform = 'translateY(-20px)'
    setTimeout(done, leaveDuration)
  }
  
  return { beforeEnter, enter, leave }
}

export function useSidebarAnimation() {
  const isCollapsed = ref(false)
  const sidebarWidth = computed(() => isCollapsed.value ? 64 : 240)
  
  const toggle = () => {
    isCollapsed.value = !isCollapsed.value
  }
  
  const transitionStyle = computed(() => ({
    width: sidebarWidth.value + 'px',
    transition: 'width ' + animationConfig.durations.medium + 'ms ' + animationConfig.easings.smooth
  }))
  
  return { isCollapsed, sidebarWidth, toggle, transitionStyle }
}

export function useCardAnimation() {
  const isHovered = ref(false)
  
  const cardStyle = computed(() => ({
    transform: isHovered.value ? 'translateY(-4px)' : 'translateY(0)',
    boxShadow: isHovered.value ? '0 12px 40px rgba(0, 0, 0, 0.12)' : '0 1px 3px rgba(0, 0, 0, 0.1)',
    transition: 'transform ' + animationConfig.durations.medium + 'ms ' + animationConfig.easings.smooth + ', box-shadow ' + animationConfig.durations.medium + 'ms ' + animationConfig.easings.smooth
  }))
  
  return { isHovered, cardStyle }
}

export function useButtonAnimation() {
  const isPressed = ref(false)
  
  const buttonStyle = computed(() => ({
    transform: isPressed.value ? 'scale(0.97)' : 'scale(1)',
    transition: 'transform ' + animationConfig.durations.fast + 'ms ' + animationConfig.easings.snappy
  }))
  
  const handlePress = () => {
    isPressed.value = true
    setTimeout(() => {
      isPressed.value = false
    }, 150)
  }
  
  return { isPressed, buttonStyle, handlePress }
}

export function useModalAnimation() {
  const { prefersReducedMotion } = useReducedMotion()
  
  const duration = computed(() => prefersReducedMotion.value ? 0 : animationConfig.durations.slow)
  
  const beforeEnter = (el) => {
    el.style.opacity = '0'
    el.style.transform = 'scale(0.9)'
  }
  
  const enter = (el, done) => {
    requestAnimationFrame(() => {
      el.style.opacity = '1'
      el.style.transform = 'scale(1)'
      setTimeout(done, duration.value)
    })
  }
  
  const leave = (el, done) => {
    const leaveDuration = duration.value * 0.75
    el.style.transition = 'opacity ' + leaveDuration + 'ms ease, transform ' + leaveDuration + 'ms ease'
    el.style.opacity = '0'
    el.style.transform = 'scale(0.95)'
    setTimeout(done, leaveDuration)
  }
  
  return { beforeEnter, enter, leave }
}

export function useStaggerAnimation() {
  const getDelay = (index) => {
    return Math.min(index * 50, 300)
  }
  
  const getItemStyle = (index) => {
    return {
      animationDelay: getDelay(index) + 'ms'
    }
  }
  
  return { getDelay, getItemStyle }
}

export function useTransition(duration, easing) {
  const d = animationConfig.durations[duration] || animationConfig.durations.base
  const e = animationConfig.easings[easing] || animationConfig.easings.smooth
  
  const style = (properties) => {
    const props = Array.isArray(properties) ? properties : [properties]
    return props.map(p => p + ' ' + d + 'ms ' + e).join(', ')
  }
  
  return { style }
}
