<template>
  <div v-if="$route.meta && $route.meta.plain" class="app-plain">
    <router-view />
  </div>

  <div v-else class="app-shell">
    <aside class="sidebar">Example App</aside>
    <main class="content">
      <router-view />
    </main>
  </div>
</template>

<style>
.app-plain { min-height: 100vh; }
.app-shell { display: grid; grid-template-columns: 260px 1fr; min-height: 100vh; }
.sidebar { border-right: 1px solid #e5e7eb; padding: 1rem; }
.content { padding: 1rem 2rem; }
</style>
