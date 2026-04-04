# Forks Animation System Guide

**Spark (交互魔术师) - Chromatic Team**

> "Stillness is death, movement is life."

---

## Table of Contents

1. [Design Principles](#design-principles)
2. [Easing Functions](#easing-functions)
3. [Duration Guidelines](#duration-guidelines)
4. [Component Animations](#component-animations)
5. [Utility Classes](#utility-classes)
6. [Performance Optimization](#performance-optimization)
7. [Best Practices](#best-practices)

---

## Design Principles

### Core Philosophy

1. **Fluidity** - 200-500ms optimal range
2. **Natural Feel** - cubic-bezier easing
3. **Restraint** - animation serves function
4. **High Performance** - transform and opacity only
5. **Accessibility** - prefers-reduced-motion support

### Performance First

RECOMMENDED (transform and opacity):
```css
.element {
  transition: transform 0.3s ease, opacity 0.3s ease;
}
```

AVOID (triggers layout):
```css
.element {
  transition: width 0.3s ease, height 0.3s ease;
}
```

---

## Easing Functions

### Standard Easing

| Variable | Value | Use Case |
|----------|-------|----------|
| `--ease-smooth` | cubic-bezier(0.4, 0, 0.2, 1) | Most transitions |
| `--ease-in-smooth` | cubic-bezier(0.4, 0, 1, 1) | Enter animations |
| `--ease-out-smooth` | cubic-bezier(0, 0, 0.2, 1) | Exit animations |

### Elastic Effects

| Variable | Value | Use Case |
|----------|-------|----------|
| `--ease-bounce-smooth` | cubic-bezier(0.34, 1.56, 0.64, 1) | Modals |
| `--ease-bounce-subtle` | cubic-bezier(0.68, -0.55, 0.265, 1.55) | Micro-interactions |

### Quick Response

| Variable | Value | Use Case |
|----------|-------|----------|
| `--ease-snappy` | cubic-bezier(0.25, 0.46, 0.45, 0.94) | Buttons |
| `--ease-quick` | cubic-bezier(0.16, 1, 0.3, 1) | Hover effects |

---

## Duration Guidelines

| Scenario | Duration | CSS Variable |
|----------|----------|--------------|
| Instant Feedback | 50-75ms | --duration-instant |
| Buttons/Switches | 150ms | --duration-fast |
| Base Transition | 200ms | --duration-base |
| Cards/Panels | 300ms | --duration-medium |
| Modals | 400ms | --duration-slow |
| Page Transition | 400ms | --duration-page |
| Special Effects | 500ms | --duration-slower |

---

## Component Animations

### 1. Sidebar

Collapse/Expand:
```css
.sidebar {
  width: 240px;
  transition: width var(--duration-medium) var(--ease-smooth);
}

.sidebar.collapsed {
  width: 64px;
}
```

Menu Item Hover:
```css
.sidebar-item::before {
  width: 3px;
  background: var(--color-primary-600);
  transform: scaleY(0);
  transition: transform var(--duration-base) var(--ease-smooth);
}

.sidebar-item:hover::before {
  transform: scaleY(1);
}
```

### 2. Cards

Hover Effect:
```css
.card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.12);
  transition: var(--transition-hover);
}
```

### 3. Buttons

Click Feedback:
```css
.btn-primary:active::after {
  width: 300px;
  height: 300px;
  transition: width var(--duration-slow) var(--ease-smooth);
}
```

Press Effect:
```html
<button class="active-press">Click</button>
```

### 4. Modals

```css
.modal-enter-active {
  animation: scaleInBounce var(--duration-medium) var(--ease-bounce-smooth);
}
```

### 5. Page Transitions

```css
.page-enter-active {
  animation: fadeInUp var(--duration-page) var(--ease-smooth);
}
```

### 6. Loading

Skeleton:
```html
<div class="skeleton"></div>
```

Spinner:
```html
<div class="animate-spin">
  <svg><!-- spinner --></svg>
</div>
```

---

## Utility Classes

### Enter Animations

| Class | Effect |
|-------|--------|
| `.animate-fade-in` | Fade in |
| `.animate-fade-in-up` | Fade in + slide up |
| `.animate-scale-in` | Scale in |
| `.animate-scale-in-bounce` | Bounce scale in |

### Continuous Animations

| Class | Effect |
|-------|--------|
| `.animate-spin` | Continuous rotation |
| `.animate-pulse` | Pulse |
| `.animate-shake` | Shake |

### Hover Effects

| Class | Effect |
|-------|--------|
| `.hover-lift` | Lift + shadow |
| `.hover-scale` | Scale |
| `.active-press` | Press down |

---

## Performance Optimization

### GPU Acceleration

```css
.gpu-accelerated {
  transform: translateZ(0);
  backface-visibility: hidden;
  perspective: 1000px;
}
```

### Hint Browser

```css
.will-change-transform {
  will-change: transform;
}
```

---

## Best Practices

### Duration

RECOMMENDED (200-500ms):
```css
.element {
  transition: transform 0.3s ease;
}
```

AVOID (too fast/slow):
```css
.element {
  transition: transform 0.05s ease; /* Too fast */
  transition: transform 1.5s ease; /* Too slow */
}
```

### Easing

RECOMMENDED:
```css
.element {
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
```

AVOID:
```css
.element {
  transition: transform 0.3s linear; /* Too stiff */
}
```

### Properties

RECOMMENDED:
```css
.element {
  transition: transform 0.3s ease, opacity 0.3s ease;
}
```

AVOID:
```css
.element {
  transition: width 0.3s ease, height 0.3s ease;
}
```

---

## Quick Reference

| Scenario | Duration | Easing | Properties |
|----------|----------|--------|------------|
| Button hover | 150ms | ease-smooth | transform, opacity |
| Card hover | 300ms | ease-smooth | transform, shadow |
| Sidebar | 300ms | ease-smooth | width |
| Modal | 400ms | ease-bounce-smooth | opacity, transform |
| Page | 400ms | ease-smooth | opacity, transform |
| Form | 300ms | ease-smooth | border, shadow |

---

**Version**: 1.0.0  
**Updated**: 2025-02-20  
**Team**: Chromatic - Spark
