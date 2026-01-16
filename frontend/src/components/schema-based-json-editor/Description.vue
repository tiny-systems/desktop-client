<template>
  <p v-if="message" :class="error ? theme.errorDescription : theme.description">{{message}}</p>
</template>
<script lang="ts">
export default {
  props: ['theme', 'message', 'error'],
}
</script>
