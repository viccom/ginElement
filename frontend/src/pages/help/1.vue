<script setup>
import { ref } from 'vue'
import MarkdownViewer from '~/components/MarkdownViewer.vue'

const markdownFilePath = ref('/md/iothelp.md') // 你可以根据需要动态修改这个路径
</script>

<template>
  <div>
    <MarkdownViewer :markdown-file="markdownFilePath" />
  </div>
</template>
