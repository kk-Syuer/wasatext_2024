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

<style>
.app-plain { min-height: 100vh; }
.app-shell { display: grid; grid-template-columns: 260px 1fr; min-height: 100vh; }
.sidebar { border-right: 1px solid #e5e7eb; padding: 1rem; }
.content { padding: 1rem 2rem; }
</style>
