<template>
  <div class="auth-wrap">
    <div class="card">
      <h2 class="title">WASA Text</h2>
      <form @submit.prevent="onLogin">
        <label class="label">Username</label>
        <input
          v-model.trim="username"
          class="input"
          maxlength="16"
          required
          placeholder="Enter a username"
        />
        <button class="btn" :disabled="loading">
          {{ loading ? "Signing in…" : "Sign in" }}
        </button>
        <p v-if="error" class="error">{{ error }}</p>
      </form>
      <p class="hint">No account yet? Just pick a name — we’ll create it on first login.</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { doLogin } from '../services/api'

const TOKEN_KEY = 'wasa_token'
const router = useRouter()
const username = ref('')
const loading = ref(false)
const error = ref('')

async function onLogin() {
  error.value = ''
  loading.value = true
  try {
    // backend already does "create if not exists" on POST /session
    const data = await doLogin(username.value)
    localStorage.setItem(TOKEN_KEY, data.identifier)
    router.push({ name: 'home' })
  } catch (e) {
    error.value = e.message || 'Network Error'
    console.error(e)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-wrap {
  min-height: 100vh; display: grid; place-items: center; background: #f7f7fb;
}
.card {
  width: 420px; max-width: 92vw; background: #fff; border-radius: 14px;
  padding: 28px; box-shadow: 0 10px 30px rgba(0,0,0,.06);
}
.title { margin: 0 0 18px; font-size: 26px; text-align: center; }
.label { display:block; margin: 8px 0 6px; font-weight: 600; }
.input {
  width:100%; padding:10px 12px; border:1px solid #d7dae0; border-radius:10px; outline:none;
}
.input:focus { border-color:#2563eb; box-shadow:0 0 0 3px rgba(37,99,235,.15); }
.btn {
  width:100%; margin-top:14px; padding:10px 12px; border:0; border-radius:10px;
  background:#2563eb; color:#fff; font-weight:600; cursor:pointer;
}
.btn:disabled { opacity:.6; cursor:not-allowed; }
.error { color:#dc2626; margin-top:10px; }
.hint { color:#6b7280; font-size: 12px; margin-top: 14px; text-align:center; }
</style>
