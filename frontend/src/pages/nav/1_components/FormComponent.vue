<script lang="ts" setup>
import axios from 'axios'
import { onMounted, ref } from 'vue'

const props = defineProps<{
  config: {
    apiUrl: string
  }
}>()

const formData = ref({
  name: '',
  value: '',
})

onMounted(async () => {
  try {
    const response = await axios.get(props.config.apiUrl)
    formData.value = response.data
  }
  catch (error) {
    console.error('加载表单数据失败:', error)
  }
})

function submitForm() {
  console.log('提交表单:', formData.value)
}
</script>

<template>
  <el-form :model="formData" label-width="120px">
    <el-form-item label="名称">
      <el-input v-model="formData.name" />
    </el-form-item>
    <el-form-item label="值">
      <el-input v-model="formData.value" />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="submitForm">
        提交
      </el-button>
    </el-form-item>
  </el-form>
</template>
