<template>
  <div class="apple-bg">
    <div class="apple-card">
      <h1 class="apple-title">WASA Text</h1>

      <form class="apple-form" @submit.prevent="onLogin">
       

        <!-- centered input inside the white card -->
        <div class="field-wrap">
          <label class="apple-label">Username</label>
          <input
            v-model.trim="username"
            class="apple-input"
            maxlength="16"
            required
            placeholder="Your name"
            autocomplete="off"
          />
        </div>

        <button class="apple-btn" :disabled="loading">
          {{ loading ? "Signing in…" : "Sign in" }}
        </button>

        <p v-if="error" class="apple-error">{{ error }}</p>
        <p class="apple-hint">
          No account? Just pick a name — it will be created on first login.
        </p>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { doLogin } from '../services/api'
import { TOKEN_KEY } from '@/services/axios'

const router = useRouter()
const username = ref('')
const loading = ref(false)
const error = ref('')

async function onLogin() {
  error.value = ''
  loading.value = true
  try {
    const data = await doLogin(username.value)   // POST /session
    localStorage.setItem(TOKEN_KEY, data.identifier)
    localStorage.setItem('last_username', username.value)
    router.push({ name: 'home' })
  } catch (e) {
    error.value = e.message || 'Network Error'
    console.error(e)
  } finally {
    loading.value = false
  }
}


async function submit() {
  const data = await doLogin(username.value)
  // data.identifier is the token; we keep storing it as before
  localStorage.setItem(TOKEN_KEY, data.identifier)
  // NEW: remember my username so HomeView can exclude it.
  localStorage.setItem('wasa_username', username.value.trim())
  router.push('/')
}
</script>

<style scoped>
/* Background and centering */
.apple-bg {
  min-height: 100vh;
  display: grid;
  place-items: center;
  background:
    radial-gradient(1200px 700px at 50% -200px, #ffffff 0, #f5f6fb 60%, #f1f2f7 100%);
}

/* Card */
.apple-card {
  width: 560px;
  max-width: 92vw;
  background: #fff;
  border-radius: 18px;
  padding: 34px 34px 28px;
  box-shadow:
    0 25px 60px rgba(0, 0, 0, 0.08),
    0 6px 18px rgba(0, 0, 0, 0.05);
}

/* Title */
.apple-title {
  margin: 0 0 22px;
  text-align: center;
  font-size: 34px;
  font-weight: 700;
  letter-spacing: 0.3px;
  color: #111827; /* neutral-900 */
}

/* Form */
.apple-form { margin: 0; }
.apple-label {
  font-size: 20px;
  display: block;
  margin: 8px 0 10px;
  font-weight: 600;
  color: #1f2937; /* neutral-800 */
}

/* Input wrapper to keep field perfectly centered */
.field-wrap {
  display: grid;
  place-items: center;
}

/* Input */
.apple-input {
  width: 100%;
  max-width: 460px;
  font-size: 16px;
  padding: 12px 14px;
  border: 1px solid #d7dae0;
  border-radius: 12px;
  outline: none;
  background: #fff;
  transition: box-shadow .15s, border-color .15s;
}
.apple-input:focus {
  border-color: #3b82f6; /* blue-500 */
  box-shadow: 0 0 0 4px rgba(59, 130, 246, .15);
}

/* Button */
.apple-btn {
  display: block;
  width: 100%;
  max-width: 460px;
  margin: 14px auto 0;
  padding: 12px 14px;
  border: 0;
  border-radius: 12px;
  font-weight: 700;
  font-size: 16px;
  color: #fff;
  background: linear-gradient(180deg, #2563eb 0%, #1e54d7 100%); /* blue-600 gradient */
  cursor: pointer;
  transition: transform .05s ease, filter .2s ease;
}
.apple-btn:hover { filter: brightness(1.03); }
.apple-btn:active { transform: translateY(1px); }
.apple-btn:disabled { opacity: .6; cursor: not-allowed; }

/* Feedback */
.apple-error {
  color: #dc2626; /* red-600 */
  margin: 10px 0 0;
  text-align: center;
}
.apple-hint {
  margin: 12px 0 0;
  color: #6b7280; /* neutral-500 */
  text-align: center;
  font-size: 14px;
}
</style>
