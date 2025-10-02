<template>
  <!-- pages that want a plain layout (like /login) -->
  <div v-if="$route.meta && $route.meta.plain" class="app-plain">
    <router-view />
  </div>

  <!-- default app layout -->
  <div v-else class="app-shell">
    <main class="content">
      <router-view />
    </main>
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>
