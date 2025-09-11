<template>
  <div class="home-wrap">
    <div class="home-card">
      <!-- LEFT: 3-button vertical tab bar -->
      <aside class="leftbar">
        <!-- 1) Profile -->
        <button class="tabbtn" :class="{ active: activeTab==='profile' }" @click="activeTab='profile'">
          <span class="iconwrap">
            <img v-if="mePhotoUrl" :src="mePhotoUrl" alt="me" class="avatar" />
            <span v-else class="avatar placeholder">{{ meInitial }}</span>
          </span>
          <span class="label">Profile</span>
        </button>

        <!-- 2) Single Chats -->
        <button class="tabbtn" :class="{ active: activeTab==='users' }" @click="activeTab='users'">
          <span class="icon">👤</span>
          <span class="label">Single Chats</span>
        </button>

        <!-- 3) Group Chats -->
        <button class="tabbtn" :class="{ active: activeTab==='groups' }" @click="activeTab='groups'">
          <span class="icon">👥</span>
          <span class="label">Group Chats</span>
        </button>
      </aside>

      <!-- MIDDLE: list (users | groups | profile) -->
      <section class="middle">
        <!-- Hide search on Profile -->
        <div class="searchbox" v-if="activeTab!=='profile'">
          <input v-model.trim="q" type="text" placeholder="Search…" />
          <button class="plus" v-if="activeTab==='users'" @click="startNewConversation">＋</button>
        </div>

        <!-- USERS -->
        <div v-if="activeTab==='users'" class="list">
          <div
            v-for="u in filteredUsers"
            :key="u"
            class="row"
            @click="openOrCreate1to1(u)"
          >
            <span class="iconwrap">
              <img v-if="userPhotos[u]" :src="userPhotos[u]" alt="" class="avatar" />
              <span v-else class="avatar placeholder">{{ u.slice(0,1).toUpperCase() }}</span>
            </span>
            <div class="meta">
              <div class="title">{{ u }}</div>
            </div>
          </div>

          <div v-if="!loading && filteredUsers.length===0" class="empty">No users found</div>
          <LoadingSpinner v-if="loading" />
          <ErrorMsg v-if="error" :msg="error" />
        </div>


        <!-- GROUPS -->
        <div v-else-if="activeTab==='groups'" class="list">
          <div v-for="g in groups" :key="g.groupName || g.name || g.id" class="row" @click="selectGroup(g)">
            <div class="circle">G</div>
            <div class="meta">
              <div class="title">{{ g.groupName || g.name || 'Group' }}</div>
            </div>
          </div>
          <div v-if="!loading && groups.length===0" class="empty">No groups yet</div>
          <LoadingSpinner v-if="loading" />
          <ErrorMsg v-if="error" :msg="error" />
        </div>

        <!-- PROFILE -->
        <div v-else class="profilepane">
          <div class="section-title">Profile</div>
          <div class="profile-list">
            <!-- Username row -->
            <button class="profile-row" @click="selectProfileAction('username')">
              <div class="left">
                <div class="circle">{{ meInitial }}</div>
                <div class="meta">
                  <div class="title">Username</div>
                  <div class="sub">{{ me }}</div>
                </div>
              </div>
              <div class="chev">›</div>
            </button>

            <!-- Profile image row -->
            <button class="profile-row" @click="selectProfileAction('photo')">
              <div class="left">
                <div class="circle">
                  <img v-if="mePhotoUrl" :src="mePhotoUrl" alt="me" class="mini-avatar" />
                  <span v-else>{{ meInitial }}</span>
                </div>
                <div class="meta">
                  <div class="title">Profile image</div>
                  <div class="sub">{{ mePhotoUrl ? 'Tap to change' : 'No image yet' }}</div>
                </div>
              </div>
              <div class="chev">›</div>
            </button>
          </div>
		  <div class="profile-actions">
			<button class="logout-btn" @click="logout">
				◦ Log out
			</button>
		  </div>

        </div>
      </section>

      <!-- RIGHT: editors / chat pane -->
      <section class="right">
        <!-- PROFILE EDITORS TAKE PRIORITY -->
        <div v-if="activeTab==='profile' && selectedProfileAction==='photo'" class="editor-pane">
          <h3 class="conv-title">Change profile image</h3>

          <div class="uploader">
            <div class="preview">
              <img v-if="photoPreview" :src="photoPreview" alt="preview" />
              <img v-else-if="mePhotoUrl" :src="mePhotoUrl" alt="current" />
              <div v-else class="preview-placeholder">{{ meInitial }}</div>
            </div>

            <label class="file-btn">
              <input type="file" accept="image/*" @change="onPickImage" hidden />
              Choose image
            </label>

            <div class="actions">
              <button class="save" :disabled="!photoFile || saving" @click="savePhoto">
				{{ saving ? 'Saving…' : 'Save' }}
			  </button>

              <button class="cancel" :disabled="saving" @click="cancelPhotoEdit">Cancel</button>
            </div>

            <ErrorMsg v-if="error" :msg="error" />
          </div>
        </div>

        <!-- (Optional) username editor placeholder; keep UI consistent -->
        <div v-else-if="activeTab==='profile' && selectedProfileAction==='username'" class="editor-pane">
          <h3 class="conv-title">Change username</h3>
          <div class="uploader">
            <div class="current-line">Current: <strong>{{ me }}</strong></div>
            <input class="text-input" v-model.trim="pendingUsername" placeholder="New username" />
            <div class="actions">
              <button class="save" :disabled="!pendingUsername || saving" @click="saveUsername">
                {{ saving ? 'Saving…' : 'Save' }}
              </button>
              <button class="cancel" :disabled="saving" @click="cancelUsernameEdit">Cancel</button>
              
            </div>
            <p v-if="success" class="hint-success">{{ success }}</p>
            <ErrorMsg v-if="error" :msg="error" />
          </div>
        </div>

        <!-- CHAT/EMPTY fallbacks -->
        <div v-else-if="currentConversationId" class="chat">
          <h3 class="conv-title">{{ currentTitle }}</h3>
          <!-- … your messages UI here … -->
        </div>
        <div v-else class="chat-empty">
          <div class="bubbles">💬</div>
          <div class="hint">Pick a user or group to start chatting</div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue'
import { listUsers, getAllUsers, listGroups, createConversation, listConversations, getUser, setMyPhoto, setMyUserName, fullUrl } from '@/services/api'
import { TOKEN_KEY } from '@/services/axios'
import { useRouter } from 'vue-router'

const router = useRouter()
const me = ref(localStorage.getItem('wasa_username') || '')
const mePhotoUrl = ref('')
const activeTab = ref('users') // users | groups | profile
const q = ref('')

const users = ref([])
const groups = ref([])
const loading = ref(false)
const error = ref('')

const currentConversationId = ref('')
const currentTitle = ref('')

const meInitial = computed(() => (me.value ? me.value[0].toUpperCase() : '?'))

// Profile editor state
const selectedProfileAction = ref('') // '', 'username', 'photo'
const photoFile = ref(null)
const photoPreview = ref('')
const saving = ref(false)
const pendingUsername = ref('')
// Map username -> absolute photo URL (or '' if none)
const userPhotos = ref({})  // Record<string, string>
const USERNAME_RE = /^[A-Za-z0-9-]{3,16}$/;
const success = ref('')

// Derived users list
const alphabeticalUsers = computed(() =>
  [...users.value].sort((a, b) => a.localeCompare(b))
)
const filteredUsers = computed(() => {
  const mine = me.value.toLowerCase()
  const needle = q.value.toLowerCase()
  return alphabeticalUsers.value
    .filter(u => u.toLowerCase() !== mine)
    .filter(u => !needle || u.toLowerCase().includes(needle))
})


function logout() {
  try {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem('wasa_username')
  } finally {
    router.replace({ name: 'login' }) // immediate redirect
  }
}

// Load lists
async function loadUsersAndGroups() {
  loading.value = true
  error.value = ''
  try {
    const [u, g] = await Promise.all([listUsers(), listGroups()])
    users.value = Array.isArray(u) ? u : []
    groups.value = Array.isArray(g) ? g : []
    hydrateUserPhotos(users.value)  // fetch avatars in background
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || 'Failed to load'
  } finally {
    loading.value = false
  }
}

// Open or create 1-to-1
async function openOrCreate1to1(username) {
  try {
    const convs = await listConversations()
    const existing = (convs || []).find(c => {
      const parts = c.participants || c.Participants || []
      return parts.length === 2 && parts.includes(me.value) && parts.includes(username)
    })
    if (existing) return selectConversation(existing)
    const conv = await createConversation(username, '👋')
    selectConversation(conv)
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || 'Cannot open conversation'
  }
}

function selectConversation(c) {
  currentConversationId.value = c.id || c.ID
  const parts = c.participants || c.Participants || []
  currentTitle.value = parts?.find(p => p !== me.value) || 'Conversation'
}

function selectGroup(g) {
  currentConversationId.value = g.conversationId || g.id || g.ID
  currentTitle.value = g.groupName || g.name || 'Group'
}

function startNewConversation() {
  // Optional: open a dialog to type a username
}

// Fetch my profile (name/photo) once
async function loadMyProfile() {
  try {
    const u = await getUser(me.value)
    mePhotoUrl.value = u?.photoUrl ? fullUrl(u.photoUrl) : ''
  } catch {
    // non-fatal
  }
}

function selectProfileAction(action) {
  selectedProfileAction.value = action
  if (action === 'photo') {
    photoFile.value = null
    photoPreview.value = ''
  }
  if (action === 'username') {
    pendingUsername.value = me.value
  }
}

function onPickImage(e) {
  const f = e.target.files?.[0]
  if (!f) return
  photoFile.value = f
  // Just for preview (any method is fine)
  const reader = new FileReader()
  reader.onload = () => { photoPreview.value = String(reader.result || '') }
  reader.readAsDataURL(f)
}

async function savePhoto() {
  if (!photoFile.value) return
  saving.value = true
  error.value = ''
  try {
      const { photoUrl } = await setMyPhoto(photoFile.value)   // backend returns { photoUrl }
      // Build absolute URL and add cache-buster so you see the fresh image immediately
      mePhotoUrl.value = fullUrl(photoUrl) + `?t=${Date.now()}`
      selectedProfileAction.value = ''
      photoFile.value = null
      photoPreview.value = ''
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || 'Failed to save photo'
  } finally {
    saving.value = false
  }
}

async function fetchUserPhoto(u) {
  if (userPhotos.value[u] !== undefined) return
  try {
    const prof = await getUser(u)
    userPhotos.value = {
      ...userPhotos.value,
      [u]: prof?.photoUrl ? fullUrl(prof.photoUrl) : ''
    }
  } catch {
    userPhotos.value = { ...userPhotos.value, [u]: '' }
  }
}

async function hydrateUserPhotos(usernames) {
  // fire-and-forget to keep UI snappy
  Promise.all(usernames.map((u) => fetchUserPhoto(u))).catch(() => {})
}


function cancelPhotoEdit() {
  selectedProfileAction.value = ''
  photoFile.value = null
  photoPreview.value = ''
}

async function saveUsername() {
  const newName = String(pendingUsername.value || '').trim()

  // reset messages
  error.value = ''
  success.value = ''

  // no change → close editor
  if (!newName || newName === me.value) {
    selectedProfileAction.value = ''
    return
  }

  // client-side validation
  if (!USERNAME_RE.test(newName)) {
    error.value = 'Username must be 3–16 characters (letters, numbers, hyphen).'
    return
  }

  saving.value = true
  try {
    const { username } = await setMyUserName(newName)

    // show success hint in-place
    success.value = `Username changed to “${username}”. You will be logged out to sign in again.`
    selectedProfileAction.value = 'username' // keep the editor open so they see the hint

    // optional blocking alert (uncomment if you prefer a popup)
    // alert(`Username changed to "${username}". Please log in again.`)

    // logout after a short delay so users can read the hint
    setTimeout(() => {
      logout()
    }, 1500)
  } catch (e) {
    const status = e?.response?.status
    const serverMsg = e?.response?.data?.error
    if (status === 409) {
      error.value = 'That username is already taken.'
    } else if (status === 400) {
      error.value = serverMsg || 'Invalid username.'
    } else if (status === 404) {
      error.value = 'User not found.'
    } else {
      error.value = 'Failed to change username. Please try again.'
    }
  } finally {
    saving.value = false
  }
}
function cancelUsernameEdit() { selectedProfileAction.value = '' }

onMounted(() => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (!token) return
  loadUsersAndGroups()
  loadMyProfile()
})
</script>

<style scoped>
/* page background and centered white card */
.home-wrap {
  min-height: 100vh;
  min-width: 100vw;
  display: flex;
  justify-content: center;
  align-items: center;
  background: #fff;
  margin: 0;
  padding: 0;
}

/* centered card */
.home-card {
  width: min(1100px, 96vw);
  height: min(720px, 88vh);
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 10px 30px rgba(0,0,0,.08);
  display: grid;
  grid-template-columns: 180px 340px 1fr;
  overflow: hidden;
}

/* left bar */
.leftbar {
  background: #fafbfc;
  border-right: 1px solid #eef0f4;
  padding: 10px 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  align-items: stretch;
}

/* row-like buttons with icon + text */
.tabbtn {
  height: 44px;
  border-radius: 10px;
  border: none;
  background: transparent;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 10px;
  cursor: pointer;
  text-align: left;
}
.tabbtn:hover { background: #eef2f7; }
.tabbtn.active { background: #e6f0ff; }
.icon { font-size: 20px; }
.iconwrap { width: 32px; height: 32px; display: grid; place-items: center; }
.avatar {
  width: 32px; height: 32px; border-radius: 50%;
  object-fit: cover;
}
.avatar.placeholder {
  width: 32px; height: 32px; border-radius: 50%;
  display: grid; place-items: center;
  background: #dbe2f3; color: #2b3a67; font-weight: 700;
}

/* middle list */
.middle {
  border-right: 1px solid #eef0f4;
  display: flex; flex-direction: column;
}
.searchbox {
  display: flex; gap: 8px; align-items: center;
  padding: 10px; border-bottom: 1px solid #eef0f4;
}
.searchbox input {
  flex: 1; height: 34px; border-radius: 10px; border: 1px solid #e5e9f2;
  padding: 0 10px; background: #fff;
}
.searchbox .plus {
  height: 34px; padding: 0 10px; border-radius: 8px; border: 1px solid #e5e9f2; background: #fff;
  cursor: pointer;
}
.list { overflow: auto; padding: 6px; }
.row {
  display: grid; grid-template-columns: 36px 1fr; gap: 10px;
  padding: 10px 8px; border-radius: 10px; cursor: pointer;
}
.row:hover { background: #f5f7fb; }
.circle {
  width: 36px; height: 36px; border-radius: 50%; display: grid; place-items: center;
  background: #e8eefc; font-weight: 700;
}
.meta .title { font-size: 14px; color: #1f2633; }

/* right pane */
.right { position: relative; display: flex; flex-direction: column; }
.chat-empty {
  margin: auto; text-align: center; color: #9aa4b2;
}
.bubbles { font-size: 40px; margin-bottom: 8px; }
.conv-title { padding: 12px 16px; border-bottom: 1px solid #eef0f4; }
.empty { color: #9aa4b2; padding: 16px; }

/* profile */
.profilepane { padding: 16px; }
.section-title {
  font-weight: 600;
  padding: 12px 12px 6px;
  color: #2b2f36;
}
.profile-list {
  display: flex; flex-direction: column;
  padding: 0 6px 10px;
}
.profile-row {
  display: flex; align-items: center; justify-content: space-between;
  gap: 10px; width: 100%; padding: 10px 8px;
  border-radius: 10px; border: none; background: transparent;
  cursor: pointer; text-align: left;
}
.profile-row:hover { background: #f5f7fb; }
.profile-row .left { display: flex; align-items: center; gap: 10px; }
.profile-row .circle {
  width: 36px; height: 36px; border-radius: 50%;
  display: grid; place-items: center; background: #e8eefc; font-weight: 700;
  overflow: hidden;
}
.mini-avatar { width: 100%; height: 100%; object-fit: cover; }
.profile-row .meta .title { font-size: 14px; color: #1f2633; }
.profile-row .meta .sub { font-size: 12px; color: #7a8596; }
.chev { color: #9aa4b2; font-size: 18px; }

/* Right-pane editor */
.editor-pane { display: flex; flex-direction: column; }
.uploader { padding: 16px; display: grid; gap: 12px; }
.preview {
  width: 160px; height: 160px; border-radius: 16px;
  overflow: hidden; border: 1px solid #eef0f4; display: grid; place-items: center;
}
.preview img { width: 100%; height: 100%; object-fit: cover; }
.preview-placeholder {
  width: 100%; height: 100%;
  display: grid; place-items: center; background: #e8eefc; color: #2b3a67; font-size: 48px; font-weight: 700;
}
.file-btn {
  display: inline-block; border: 1px solid #e5e9f2; padding: 8px 12px;
  border-radius: 8px; cursor: pointer; background: #fff;
}
.actions { display: flex; gap: 10px; }
.actions .save {
  border: none; padding: 8px 14px; border-radius: 8px; cursor: pointer;
  background: #2563eb; color: #fff;
}
.actions .save:disabled { opacity: .6; cursor: default; }
.actions .cancel {
  border: 1px solid #e5e9f2; padding: 8px 14px; border-radius: 8px;
  background: #fff; cursor: pointer;
}
.current-line { color: #546075; padding: 12px 0 4px; }
.text-input {
  height: 36px; border-radius: 10px; border: 1px solid #e5e9f2; padding: 0 10px; width: 260px;
}

.profile-actions {
  padding: 12px 12px 0;
  border-top: 1px solid #eef0f4;
  margin-top: 8px;
}
.logout-btn {
  width: 100%;
  background: #fff;
  border: 1px solid #e5e9f2;
  border-radius: 10px;
  padding: 10px 12px;
  cursor: pointer;
  color: #c0392b;
  font-weight: 600;
}
.logout-btn:hover {
  background: #fff5f4;
  border-color: #f3d0cd;
}
.hint-success {
  margin-top: 10px;
  color: #16a34a; /* green-600 */
  font-size: 14px;
}

.hint-error {
  margin-top: 10px;
  color: #dc2626; /* red-600 */
  font-size: 14px;
}
</style>
