<script setup>
import { marked } from 'marked'
import Prism from 'prismjs'
import { onMounted, ref } from 'vue'
import 'prismjs/themes/prism-tomorrow.css' // 引入 Prism.js 主题

// 定义 props
const props = defineProps({
  markdownFile: {
    type: String,
    required: true,
  },
})

// 加载 Markdown 文件
const markdownContent = ref('')
const htmlContent = ref('')

// 解析 Markdown
async function parseMarkdown() {
  const response = await fetch(props.markdownFile) // 使用 props 中的文件名
  markdownContent.value = await response.text()
  htmlContent.value = marked(markdownContent.value, {
    highlight: (code, language) => {
      const validLanguage = Prism.languages[language] ? language : 'plaintext'
      return Prism.highlight(code, Prism.languages[validLanguage], validLanguage)
    },
  })
}

onMounted(() => {
  parseMarkdown()
})
</script>

<template>
  <div class="markdown-content" v-html="htmlContent" />
</template>

<style scoped>
.markdown-content {
  padding: 20px;
  font-family: Arial, sans-serif;
  text-align: left; /* 确保内容左对齐 */
}
</style>
