<template>
  <div class="timeline-view">
    <div class="timeline-container">
      <!-- 左侧时间轴 -->
      <div class="timeline-sidebar">
        <div class="timeline-header">
          <h3>时间轴</h3>
          <n-button
            v-if="currentViewLevel > 'year'"
            text
            size="small"
            type="primary"
            @click="handleBackToParent"
          >
            <template #icon>
              <n-icon><ChevronBackOutline /></n-icon>
            </template>
            返回
          </n-button>
        </div>

        <div class="timeline-scroll" ref="timelineTreeRef">
          <div class="timeline-track">
            <!-- 年级别 -->
            <template v-if="currentViewLevel === 'year'">
              <div
                v-for="yearData in timelineData?.years || []"
                :key="yearData.year"
                class="timeline-item"
                :class="{ 'is-active': isActiveYear(yearData.year) }"
                @click="handleYearClick(yearData)"
              >
                <div class="timeline-dot"></div>
                <div class="timeline-content">
                  <div class="timeline-label">{{ yearData.year }}年</div>
                  <div class="timeline-meta">{{ yearData.count }} 个</div>
                </div>
              </div>
            </template>

            <!-- 月级别 -->
            <template v-else-if="currentViewLevel === 'month'">
              <div
                class="timeline-item is-parent"
                @click="handleBackToYear"
              >
                <div class="timeline-dot is-parent"></div>
                <div class="timeline-content">
                  <div class="timeline-label">{{ selectedYear }}年</div>
                  <div class="timeline-meta">{{ getYearTotal(selectedYear) }} 个</div>
                </div>
              </div>

              <div
                v-for="monthData in getCurrentYearMonths()"
                :key="monthData.month"
                class="timeline-item"
                :class="{ 'is-active': isActiveMonth(monthData.month) }"
                @click="handleMonthClick(monthData)"
              >
                <div class="timeline-dot"></div>
                <div class="timeline-content">
                  <div class="timeline-label">{{ monthData.month }}月</div>
                  <div class="timeline-meta">{{ monthData.count }} 个</div>
                </div>
              </div>
            </template>

            <!-- 日级别 -->
            <template v-else-if="currentViewLevel === 'day'">
              <div
                class="timeline-item is-parent"
                @click="handleBackToMonth"
              >
                <div class="timeline-dot is-parent"></div>
                <div class="timeline-content">
                  <div class="timeline-label">{{ selectedYear }}-{{ selectedMonth }}</div>
                  <div class="timeline-meta">{{ getMonthTotal(selectedYear, selectedMonth) }} 个</div>
                </div>
              </div>

              <div
                v-for="dayData in getCurrentMonthDays()"
                :key="dayData.day"
                class="timeline-item"
                :class="{ 'is-active': isActiveDay(dayData.day) }"
                @click="handleDayClick(dayData)"
              >
                <div class="timeline-dot"></div>
                <div class="timeline-content">
                  <div class="timeline-label">{{ dayData.day }}日</div>
                  <div class="timeline-meta">{{ dayData.count }} 个</div>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>

      <!-- 右侧仓库列表 -->
      <div class="repos-main">
        <div class="repos-header">
          <n-breadcrumb>
            <n-breadcrumb-item v-if="selectedYear">{{ selectedYear }}年</n-breadcrumb-item>
            <n-breadcrumb-item v-if="selectedMonth">{{ selectedMonth }}月</n-breadcrumb-item>
            <n-breadcrumb-item v-if="selectedDay">{{ selectedDay }}日</n-breadcrumb-item>
          </n-breadcrumb>
          <n-tag type="info" size="small" round>
            {{ reposStore.dimensionRepos.length }} 个仓库
          </n-tag>
        </div>

        <div class="repos-scroll" ref="reposContainerRef">
          <n-spin :show="reposStore.dimensionLoading">
            <div v-if="reposStore.dimensionRepos.length > 0" class="repos-grid">
              <RepoCard
                v-for="repo in reposStore.dimensionRepos"
                :key="repo.id"
                :repo="repo"
                :repo-status="reposStore.repoStatuses[repo.id]"
                @click="$emit('repo-click', repo)"
                @view-code="$emit('view-code', repo)"
                @open-folder="$emit('open-folder', repo)"
                @clone="$emit('clone', repo)"
                @pull="$emit('pull', repo)"
                @delete="$emit('delete', repo)"
                @update-info="$emit('update-info', repo)"
              />
            </div>

            <div v-if="isLoadingMore" class="loading-trigger">
              <n-spin size="small" />
              <span>加载更多...</span>
            </div>

            <n-empty
              v-if="!reposStore.dimensionLoading && reposStore.dimensionRepos.length === 0"
              description="该时间段暂无仓库"
              size="large"
            >
              <template #icon>
                <n-icon><FolderOpenOutline /></n-icon>
              </template>
            </n-empty>
          </n-spin>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useMessage } from 'naive-ui'
import {
  ChevronBackOutline,
  FolderOpenOutline
} from '@vicons/ionicons5'
import RepoCard from '@/components/repo/RepoCard.vue'
import { useReposStore } from '@/stores/repos'

const message = useMessage()
const reposStore = useReposStore()

const props = defineProps({
  repos: { type: Array, default: () => [] },
  repoStatuses: { type: Object, default: () => ({}) },
  loading: { type: Boolean, default: false }
})

const emit = defineEmits([
  'repo-click', 'view-code', 'open-folder',
  'clone', 'pull', 'delete', 'update-info'
])

const timelineTreeRef = ref(null)
const reposContainerRef = ref(null)

const currentViewLevel = ref('year')
const selectedYear = ref(null)
const selectedMonth = ref(null)
const selectedDay = ref(null)

const isLoadingMore = ref(false)
const currentPageInPeriod = ref(1)
const hasMore = ref(true)

const timelineData = computed(() => reposStore.timelineDimensions)

const getCurrentYearMonths = () => {
  if (!timelineData.value?.years) return []
  const yearData = timelineData.value.years.find(y => y.year === selectedYear.value)
  return yearData?.months || []
}

const getCurrentMonthDays = () => {
  const months = getCurrentYearMonths()
  const monthData = months.find(m => m.month === selectedMonth.value)
  return monthData?.days || []
}

const getYearTotal = (year) => {
  const yearData = timelineData.value?.years?.find(y => y.year === year)
  return yearData?.count || 0
}

const getMonthTotal = (year, month) => {
  const months = getCurrentYearMonths()
  const monthData = months.find(m => m.month === month)
  return monthData?.count || 0
}

const isActiveYear = (year) => currentViewLevel.value === 'month' && selectedYear.value === year
const isActiveMonth = (month) => currentViewLevel.value === 'day' && selectedMonth.value === month
const isActiveDay = (day) => selectedDay.value === day

const handleYearClick = async (yearData) => {
  selectedYear.value = yearData.year
  selectedMonth.value = null
  selectedDay.value = null
  currentViewLevel.value = 'month'
  const months = yearData.months || []
  if (months.length > 0) {
    await loadPeriodData({ year: yearData.year, month: months[0].month })
  }
}

const handleMonthClick = async (monthData) => {
  selectedMonth.value = monthData.month
  selectedDay.value = null
  currentViewLevel.value = 'day'
  const days = monthData.days || []
  if (days.length > 0) {
    await loadPeriodData({
      year: selectedYear.value,
      month: monthData.month,
      day: days[0].day
    })
  }
}

const handleDayClick = async (dayData) => {
  selectedDay.value = dayData.day
  await loadPeriodData({
    year: selectedYear.value,
    month: selectedMonth.value,
    day: dayData.day
  })
}

const handleBackToParent = () => {
  if (currentViewLevel.value === 'day') handleBackToMonth()
  else if (currentViewLevel.value === 'month') handleBackToYear()
}

const handleBackToYear = () => {
  currentViewLevel.value = 'year'
  selectedYear.value = null
  selectedMonth.value = null
  selectedDay.value = null
  reposStore.dimensionRepos = []
}

const handleBackToMonth = () => {
  currentViewLevel.value = 'month'
  selectedDay.value = null
  loadPeriodData({ year: selectedYear.value, month: selectedMonth.value })
}

const loadPeriodData = async (periodParams, page = 1) => {
  try {
    if (page === 1) reposStore.dimensionLoading = true
    else isLoadingMore.value = true

    const result = await reposStore.fetchTimelinePeriod(periodParams, page, 12)
    hasMore.value = result.hasMore
    currentPageInPeriod.value = page
  } catch (error) {
    message.error('加载数据失败: ' + error.message)
  } finally {
    reposStore.dimensionLoading = false
    isLoadingMore.value = false
  }
}

const handleScroll = async () => {
  if (!reposContainerRef.value || isLoadingMore.value || !hasMore.value) return
  const container = reposContainerRef.value
  if (container.scrollTop + container.clientHeight >= container.scrollHeight - 100) {
    const period = {
      year: selectedYear.value,
      month: selectedMonth.value,
      day: selectedDay.value
    }
    currentPageInPeriod.value += 1
    await loadPeriodData(period, currentPageInPeriod.value)
  }
}

const initializeTimeline = async () => {
  await reposStore.fetchTimelineDimensions()
  if (timelineData.value?.years?.length > 0) {
    const latestYear = timelineData.value.years[0]
    await handleYearClick(latestYear)
  }
}

onMounted(async () => {
  await initializeTimeline()
  nextTick(() => {
    if (reposContainerRef.value) {
      reposContainerRef.value.addEventListener('scroll', handleScroll)
    }
  })
})

onUnmounted(() => {
  if (reposContainerRef.value) {
    reposContainerRef.value.removeEventListener('scroll', handleScroll)
  }
})
</script>

<style scoped>
.timeline-view {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.timeline-container {
  display: flex;
  gap: 1px;
  height: calc(100vh - 160px);
  background: #e5e7eb;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

/* 左侧时间轴 */
.timeline-sidebar {
  width: 200px;
  background: #fff;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.timeline-header {
  padding: 16px 12px;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f9fafb;
}

.timeline-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #111827;
}

.timeline-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.timeline-scroll::-webkit-scrollbar {
  width: 4px;
}

.timeline-scroll::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 2px;
}

.timeline-scroll::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}

.timeline-track {
  position: relative;
  padding-left: 20px;
}

.timeline-track::before {
  content: '';
  position: absolute;
  left: 19px;
  top: 8px;
  bottom: 8px;
  width: 2px;
  background: linear-gradient(180deg, #3b82f6 0%, #8b5cf6 100%);
  border-radius: 1px;
}

.timeline-item {
  position: relative;
  padding: 0 8px 0 24px;
  margin-bottom: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.timeline-item:hover {
  background: #f3f4f6;
}

.timeline-item.is-active {
  background: #eff6ff;
}

.timeline-dot {
  position: absolute;
  left: 11px;
  top: 50%;
  transform: translateY(-50%);
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #fff;
  border: 2px solid #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
  transition: all 0.15s ease;
  z-index: 1;
}

.timeline-dot.is-parent {
  width: 12px;
  height: 12px;
  left: 10px;
  border-color: #8b5cf6;
}

.timeline-item:hover .timeline-dot {
  transform: translateY(-50%) scale(1.15);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}

.timeline-item.is-active .timeline-dot {
  background: #3b82f6;
  transform: translateY(-50%) scale(1.2);
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.2);
}

.timeline-content {
  padding: 8px 6px;
}

.timeline-label {
  font-size: 13px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 2px;
}

.timeline-item.is-active .timeline-label {
  color: #2563eb;
  font-weight: 600;
}

.timeline-meta {
  font-size: 11px;
  color: #9ca3af;
}

.timeline-item.is-active .timeline-meta {
  color: #60a5fa;
}

/* 右侧仓库列表 */
.repos-main {
  flex: 1;
  background: #fff;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.repos-header {
  padding: 12px 16px;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f9fafb;
  flex-shrink: 0;
}

.repos-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.repos-scroll::-webkit-scrollbar {
  width: 6px;
}

.repos-scroll::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}

.repos-scroll::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}

.repos-scroll::-webkit-scrollbar-track {
  background: transparent;
}

.repos-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}

.loading-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 16px;
  color: #9ca3af;
  font-size: 13px;
}

/* 响应式 */
@media (max-width: 1024px) {
  .repos-grid {
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  }
}

@media (max-width: 768px) {
  .timeline-container {
    flex-direction: column;
    height: auto;
  }

  .timeline-sidebar {
    width: 100%;
    max-height: 250px;
  }

  .repos-grid {
    grid-template-columns: 1fr;
  }
}
</style>
