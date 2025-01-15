<script lang="ts" setup>
import axios from 'axios'
import * as echarts from 'echarts'
import { onMounted, ref } from 'vue'

const props = defineProps<{
  config: {
    apiUrl: string
  }
}>()

const chartRef = ref<HTMLElement | null>(null)

onMounted(async () => {
  try {
    const response = await axios.get(props.config.apiUrl)
    const chartData = response.data

    if (chartRef.value) {
      const chart = echarts.init(chartRef.value)
      chart.setOption({
        xAxis: {
          type: 'category',
          data: chartData.categories,
        },
        yAxis: {
          type: 'value',
        },
        series: [
          {
            data: chartData.values,
            type: 'bar',
          },
        ],
      })
    }
  }
  catch (error) {
    console.error('加载图表数据失败:', error)
  }
})
</script>

<template>
  <div ref="chartRef" style="width: 100%; height: 400px;" />
</template>
