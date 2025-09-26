<template>
  <div class="home-wrap">
    <header class="topbar">
      <span class="topbar-title">Welcome to WASAText</span>
    </header>
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
        <div class="searchbox" v-if="activeTab !== 'profile'">
          <input
            v-model.trim="q"
            type="text"
            :placeholder="activeTab === 'groups' ? 'Search groups…' : 'Search…'"
          />
          <button
            class="plus"
            v-if="activeTab === 'groups'"
            @click="openCreateGroup"
            title="Create group"
          >＋</button>
        </div>

        <!-- USERS -->
        <!-- SINGLE CHATS (middle pane) -->
        <div v-if="activeTab === 'users'" class="contacts-pane">
          <div class="pane-header">Contacts</div>

          <div class="contacts-scroll">
            <button
              v-for="item in singleContacts"
              :key="item.key"
              class="chat-item"
              @click="openOrCreate1to1(item.username)"
            >
              <!-- avatar -->
              <img v-if="item.photoUrl" :src="item.photoUrl" alt="" class="avatar" />
              <div v-else class="avatar avatar-fallback">{{ item.initial }}</div>

              <!-- main -->
              <div class="meta">
                <div class="row-1">
                  <span class="name">{{ item.display }}</span>
                  <time v-if="item.lastAt" class="time">{{ formatTime(item.lastAt) }}</time>
                </div>

                <div class="row-2">
                  <span v-if="item.lastType === 'text'" class="snippet">
                    <span v-if="item.lastSender">{{ item.lastSender }}: </span>{{ item.lastText || ' ' }}
                  </span>
                  <span v-else-if="item.lastType === 'image' || item.lastType === 'gif'" class="snippet dim">
                    <span v-if="item.lastSender">{{ item.lastSender }}: </span>[Photo]
                  </span>
                  <span v-else class="snippet dim">Start a chat</span>
                </div>
              </div>
            </button>
          </div>
        </div>

        <!-- GROUPS -->
        <div v-else-if="activeTab==='groups'" class="contacts-pane">
          <div class="pane-header">Groups</div>

          <div class="contacts-scroll">
            <button
              v-for="g in filteredGroups"
              :key="g.key"
              class="chat-item"
              type="button"
              @click="openGroup(g)"
              :title="g.display"
            >
              <!-- avatar -->
              <img v-if="g.photoUrl" :src="g.photoUrl" alt="" class="avatar" />
              <div v-else class="avatar avatar-fallback">{{ (g.display?.[0] || 'G').toUpperCase() }}</div>

              <!-- main -->
              <div class="meta">
                <div class="row-1">
                  <span class="name">{{ g.display }}</span>
                  <time v-if="g.lastAt" class="time">{{ formatTime(g.lastAt) }}</time>
                </div>

                <div class="row-2">
                  <span v-if="g.lastType === 'text'" class="snippet">
                    <span v-if="g.lastSender">{{ g.lastSender }}: </span>{{ g.lastText || ' ' }}
                  </span>
                  <span v-else-if="g.lastType === 'image' || g.lastType === 'gif'" class="snippet dim">
                    <span v-if="g.lastSender">{{ g.lastSender }}: </span>[Photo]
                  </span>
                  <span v-else class="snippet dim">No messages yet</span>
                </div>
              </div>
            </button>
          </div>
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
        <!-- show composer if we have a real conversation OR a pending draft peer -->
        <div v-else-if="currentConversationId || pendingPeer" class="chat">
          <!-- Header -->
          <div class="conv-head">
            <h3 class="conv-title">{{ currentTitle }}</h3>
            <button
              v-if="isGroupThread"
              class="conv-menu"
              @click="openGroupMgmt"
              title="Group options"
            >⋯</button>
          </div>

          <!-- Messages -->
          <div class="msg-list" ref="msgList">
            <div
              v-for="m in messages"
              :key="m.id || m.ID"
            >
              <div class="msg-row" :class="isMine(m) ? 'mine' : 'theirs'">
                <!-- NEW: sender name (groups only) -->
                <div
                  v-if="isGroupThread"
                  :class="['sender-line', isMine(m) ? 'sender--mine' : 'sender--theirs']"
                >
                  {{ senderOf(m) }}
                </div>
                <!-- text bubble -->
                <div
                  v-if="contentTypeOf(m) === 'text'"
                  class="bubble"
                  :class="isMine(m) ? 'bubble--mine' : 'bubble--theirs'"
                >
                  {{ msgText(m) }}
                </div>

                <!-- image bubble -->
                <div
                  v-else
                  class="bubble bubble--image"
                  :class="isMine(m) ? 'bubble--mine' : 'bubble--theirs'"
                >
                  <img
                    :src="fullUrl(msgImg(m))"
                    alt=""
                    @load="scrollToBottom"
                  />
                  <div v-if="msgText(m)" class="caption">{{ msgText(m) }}</div>
                </div>
                  <!-- Reactions row -->
                  <div class="reactions-row">
                    <!-- existing reactions as chips -->
                    <button
                      v-for="rx in aggregateReactions(m)"
                      :key="rx.emoji"
                      class="rx-chip"
                      :class="{ mine: rx.mine }"
                      @click.stop="toggleReaction(m, rx.emoji)"
                      :title="rx.mine ? 'Remove my reaction' : 'React'"
                    >
                      <span class="rx-emoji">{{ rx.emoji }}</span>
                      <span class="rx-count" v-if="rx.count > 1">{{ rx.count }}</span>
                    </button>

                    <!-- small “add reaction” button -->
                    <button class="rx-add" @click.stop="toggleReactionBar(m)" title="Add reaction">😊</button>

                    <!-- tiny popover with choices -->
                    <div
                      v-if="reactionBarForId === idForMessage(m)"
                      :class="['rx-pop', isMine(m) ? 'right' : 'left']"
                    >
                      <button v-for="e in reactionChoices" :key="e" class="rx-pick" @click.stop="toggleReaction(m, e)">{{ e }}</button>
                    </div>
                  </div>
              </div>

              <div class="meta-time" :class="isMine(m) ? 'meta--mine' : 'meta--theirs'">
                {{ prettyTime(m) }}
                <span v-if="isMine(m)" class="checks">
                  <span v-if="isRead(m)"   class="check check--double" title="Read"></span>
                  <span v-else-if="isDelivered(m)" class="check check--single" title="Delivered"></span>
                </span>
              </div>
            </div>

          </div>

          <!-- Composer -->
          <div class="composer">
            <button class="cbtn" title="Emoji" @click="toggleEmoji">😊</button>
            <input
              ref="fileInput"
              type="file"
              accept="image/*"
              hidden
              @change="onSelectFile"
            />
            <button class="cbtn" title="Attach image" @click="fileInput.click()">📎</button>

            <textarea
              v-model="draft"
              class="cinput"
              placeholder="Write a message"
              :disabled="!canSend"
              @keydown.enter.exact.prevent="onSendText"
              @keydown.enter.shift.stop
            ></textarea>

            <button class="sendbtn" :disabled="!canSend || sending || (!draft.trim() && !selectedFile)" @click="onSendText">
              Send
            </button>

            <!-- very small inline emoji popover (optional) -->
            <div v-if="showEmoji" class="emoji-pop">
              <button v-for="e in emojis" :key="e" @click="insertEmoji(e)">{{ e }}</button>
            </div>
          </div>
          <div v-if="selectedFile" style="padding: 0 12px 10px;">
            <span style="font-size:12px; background:#f4f4f5; border:1px solid #e5e7eb; border-radius:999px; padding:4px 10px;">
              {{ selectedFile.name }}
              <button type="button" @click="selectedFile=null" style="border:none; background:transparent; margin-left:6px; cursor:pointer;">✕</button>
            </span>
          </div>

        </div>
        <div v-else class="chat-empty">
          <div class="bubbles">💬</div>
          <div class="hint">Pick a user or group to start chatting</div>
        </div>
        <!-- Create Group Modal -->
        <div v-if="showCreateGroup" class="cg-backdrop" @click.self="closeCreateGroup">
          <div class="cg-modal">
            <div class="cg-head">
              <h3>Create group</h3>
              <button class="cg-x" @click="closeCreateGroup" :disabled="cgBusy">✕</button>
            </div>

            <!-- top: avatar + name -->
            <div class="cg-top">
              <div class="cg-avatar">
                <img v-if="cgPhotoPreview" :src="cgPhotoPreview" alt="group" />
                <div v-else class="cg-avatar-ph">G</div>
                <label class="cg-photo-btn">
                  <input type="file" accept="image/*" hidden @change="onPickGroupPhoto">
                  {{ cgPhotoPreview ? 'Change photo' : 'Upload photo' }}
                </label>
                <button v-if="cgPhotoPreview" class="cg-photo-clear" @click="clearGroupPhoto">Remove</button>
              </div>

              <label class="cg-field grow">
                <span class="cg-label">Group name</span>
                <input class="cg-input" v-model.trim="cgName" placeholder="e.g., Project A" />
              </label>
            </div>

            <!-- members -->
            <label class="cg-field">
              <span class="cg-label">Members</span>
              <div class="cg-members">
                <label v-for="u in otherUsers" :key="u" class="cg-pill">
                  <input type="checkbox" :value="u" v-model="cgMembers" />
                  <span>{{ u }}</span>
                </label>
              </div>
              <div class="cg-hint">You are included automatically.</div>
            </label>

            <!-- initial message -->
            <label class="cg-field">
              <span class="cg-label">Initial message </span>
              <textarea class="cg-textarea" v-model="cgMessage" placeholder="Say hi…"></textarea>
            </label>

            <p v-if="cgError" class="cg-error">{{ cgError }}</p>

            <div class="cg-actions">
              <button class="cg-btn" @click="closeCreateGroup" :disabled="cgBusy">Cancel</button>
              <button class="cg-btn cg-primary" @click="submitCreateGroup" :disabled="cgBusy || cgPhotoBusy">
                <span v-if="cgBusy || cgPhotoBusy">Creating…</span>
                <span v-else>Create</span>
              </button>
            </div>
          </div>
        </div>
        <!-- Group management drawer -->
      <div v-if="showGroupMgmt" class="gm-backdrop" @click.self="closeGroupMgmt">
        <div class="gm-drawer">
          <div class="gm-head">
            <h3>Group settings</h3>
            <button class="gm-x" @click="closeGroupMgmt" :disabled="gmBusy">✕</button>
          </div>

          <div class="gm-section">
            <div class="gm-label">Members ({{ gmMembers.length }})</div>
            <div class="gm-members">
              <div v-for="u in gmMembers" :key="u" class="gm-member">
                <div class="gm-user">
                  <img v-if="userPhotos[u]" :src="userPhotos[u]" class="gm-avatar" />
                  <div v-else class="gm-avatar ph">{{ u[0]?.toUpperCase() }}</div>
                  <span class="gm-name">{{ u }}</span>
                  <span v-if="u === me" class="gm-me">you</span>
                </div>
                <button
                  v-if="u !== me"
                  class="gm-remove"
                  :disabled="gmBusy"
                  @click="onRemoveMember(u)"
                  title="Remove"
                >−</button>
              </div>
            </div>
          </div>

          <div class="gm-section">
            <div class="gm-label">Add members</div>
            <div class="gm-candidates">
              <button
                v-for="u in gmCandidates"
                :key="u"
                class="gm-add"
                :disabled="gmBusy"
                @click="onAddMember(u)"
              >+ {{ u }}</button>
            </div>
          </div>

          <div class="gm-footer">
            <button class="gm-leave" :disabled="gmBusy" @click="onLeaveGroup">Leave group</button>
          </div>

          <p v-if="gmError" class="gm-error">{{ gmError }}</p>
        </div>
      </div>

      </section>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed, onUnmounted } from 'vue'
import { listMessages,  sendText, sendFile, listUsers, listGroups, createConversation, listConversations, getUser, setMyPhoto, setMyUserName, fullUrl, messageStatuses, getConversation, addReaction, removeReaction, getMessage, createGroup, getGroup, setGroupPhoto,addGroupMember, removeGroupMember, leaveGroup } from '@/services/api'
import { TOKEN_KEY, UNAUTHORIZED_EVENT } from '@/services/axios'
import { useRouter } from 'vue-router'
import { watch, nextTick } from 'vue'

const router = useRouter()
const me = ref(localStorage.getItem('wasa_username') || '')
const mePhotoUrl = ref('')
const activeTab = ref('users') // users | groups | profile
const q = ref('')
const singleContacts = ref([])               // [{ username, display, photoUrl, lastAt, lastType, lastText, unread }]
const photoCache = new Map()                 // username -> photoUrl ('' if none)
const users = ref([])

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
const selectedFile = ref(null)

// Map username -> absolute photo URL (or '' if none)
const userPhotos = ref({})  // Record<string, string>
const USERNAME_RE = /^[A-Za-z0-9-]{3,16}$/;
const success = ref('')
let messagesTimer = null    // polling timer for messages
// Who I'm about to chat with if no conversation exists yet
const pendingPeer = ref('')   // username we’re composing to (no conversation yet)
let contactsTicker = null
let groupMembersTicker = null

// Create Group modal state
const showCreateGroup = ref(false)
const cgName = ref('')
const cgMembers = ref([])          // array of usernames
const cgMessage = ref('')
const cgBusy = ref(false)
const cgError = ref('')
// Group image
const cgPhotoFile = ref(null)
const cgPhotoPreview = ref('')
const cgPhotoBusy = ref(false)
const isGroupThread  = computed(() => String(currentConvType.value).toLowerCase() === 'group')
const currentConvType = ref('')  
let groupsTicker   = null

// Normalized groups for the middle pane
const groupItems = ref([])   // [{ key, groupName, display, conversationId, photoUrl, lastAt, lastType, lastText, lastSender }]
// All users but me (already loaded in `users`)
const otherUsers = computed(() =>
  alphabeticalUsers.value.filter(u => u && u !== me.value)
)
// Filtered by search
const filteredGroups = computed(() => {
  const needle = (q.value || '').toLowerCase()
  const src = (groupItems.value || []).slice()
  src.sort((a, b) => {
    if (a.lastAt && b.lastAt) return new Date(b.lastAt) - new Date(a.lastAt)
    if (a.lastAt && !b.lastAt) return -1
    if (!a.lastAt && b.lastAt) return 1
    return String(a.display).localeCompare(String(b.display))
  })
  return needle ? src.filter(g => String(g.display).toLowerCase().includes(needle)) : src
})
// Derived users list
const alphabeticalUsers = computed(() => {
  const src = Array.isArray(users.value) ? users.value.slice() : [];
  src.sort((a, b) => String(a).localeCompare(String(b)));
  return src;
});

const filteredUsers = computed(() => {
  const mine = (me.value || '').toLowerCase();
  const needle = (q.value || '').toLowerCase();

  return alphabeticalUsers.value
    .filter(u => String(u).toLowerCase() !== mine)
    .filter(u => !needle || String(u).toLowerCase().includes(needle));
});

const canSend = computed(() => {
  // draft 1:1 (no conversation yet) is allowed
  if (!currentConversationId.value) return !!pendingPeer.value
  // 1:1 conversations allowed
  if ((participants.value || []).length <= 2) return true
  // groups: only members
  return (participants.value || []).includes(me.value)
})

// Current group info (set when opening a group)
const currentGroupName = ref('')

// Group management drawer
const showGroupMgmt = ref(false)
const gmMembers = ref([])          // current list of members (strings)
const gmBusy = ref(false)
const gmError = ref('')
let   gmTicker = null

// Candidates = all users not already in the group (and not me)
const gmCandidates = computed(() =>
  alphabeticalUsers.value.filter(u => u !== me.value && !gmMembers.value.includes(u))
)


function stopAllPollers() {
  clearInterval(statusTimer);   statusTimer = null;
  clearInterval(messagesTimer); messagesTimer = null;
  clearInterval(contactsTicker);contactsTicker = null;
  clearInterval(groupsTicker);  groupsTicker = null;
}


function isNearBottom() {
  const el = msgList.value
  if (!el) return true
  return el.scrollHeight - el.scrollTop - el.clientHeight < 120
}

function lastMsgId(list) {
  const m = list?.[list.length - 1]
  return m?.id ?? m?.ID ?? m?.messageId ?? m?.MessageID
}
async function pollMessages() {
  const id = currentConversationId.value
  if (!id) return
  try {
    const arr = await listMessages(id)
    // keep ASC so newest is at the bottom
    arr.sort((a, b) => {
      const ta = new Date(a.timestamp ?? a.Timestamp ?? 0).getTime()
      const tb = new Date(b.timestamp ?? b.Timestamp ?? 0).getTime()
      return ta - tb
    })

    const wasNear = isNearBottom()
    const newLast = lastMsgId(arr)
    const oldLast = lastMsgId(messages.value)

    // 1) If the thread shape changed, replace wholesale
    if (newLast !== oldLast || arr.length !== messages.value.length) {
      messages.value = normalizeReactionsField(arr)
      await nextTick()
      if (wasNear) scrollToBottom()
      // Update the group middle pane (last line + time) on incoming messages
      const last = arr[arr.length - 1]
      upsertGroupFromMessage(currentConversationId.value, last)
    } else {
      // 2) Same messages → reconcile reactions per message
      const byId = new Map(messages.value.map(m => [idForMessage(m), m]))
      let anyPatched = false

      for (const fresh of arr) {
        const mid = idForMessage(fresh)
        const old = byId.get(mid)
        if (!old) continue

        // compare reactions only (ignore other fields)
        if (reactionsKey(old) !== reactionsKey(fresh)) {
          const idx = messages.value.findIndex(x => idForMessage(x) === mid)
          if (idx !== -1) {
            messages.value.splice(idx, 1, fresh)   // patch in place
            anyPatched = true
          }
        }
      }

      if (anyPatched && wasNear) {
        await nextTick(); scrollToBottom()
      }
    }

    // If any message still lacks a reactions array (older pages, etc.), hydrate it
    await ensureReactions(arr)
  } catch {
    /* ignore transient errors */
  }
}

// Normalize id and check for reactions
function hasReactionsField(m) {
  const rx = m?.reactions ?? m?.Reactions;
  return Array.isArray(rx);              // true if field exists (even empty)
}

// Pull reactions for messages that don't have the field yet, then patch in place
async function ensureReactions(list) {
  const idsToFetch = list
    .filter(m => !hasReactionsField(m))
    .map(idForMessage)
    .filter(Boolean);

  if (!idsToFetch.length) return;

  const fresh = await mapWithLimit(idsToFetch, 4, async mid => {
    try { return await getMessage(mid); } catch { return null; }
  });

  const byId = new Map(fresh.filter(Boolean).map(m => [idForMessage(m), m]));

  // Replace messages in-place so Vue updates but scroll stays stable
  messages.value = messages.value.map(m => byId.get(idForMessage(m)) || m);
}

async function refreshOneMessage(mid) {
  try {
    const fresh = await getMessage(mid);
    const fid = idForMessage(fresh);
    const i = messages.value.findIndex(m => idForMessage(m) === fid);
    if (i !== -1) messages.value.splice(i, 1, fresh);
  } catch {}
}

// Load lists
async function loadUsersAndGroups() {
  loading.value = true
  error.value = ''
  try {
    const [u] = await Promise.all([listUsers()])
    users.value = Array.isArray(u) ? u : []
    await hydrateGroups()                                 // ← build middle pane groups
    hydrateUserPhotos(users.value)
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || 'Failed to load'
  } finally {
    loading.value = false
  }
}


// Open existing 1:1 if it exists, otherwise enter "draft" (no creation yet)
async function openOrCreate1to1(username) {
  try {
    const convs = await listConversations()

    const existing = (convs || []).find(c => {
      const p   = partsOf(c)
      const typ = String(c.type || c.Type || '').toLowerCase()
      // strictly a 1:1 conversation
      return typ === 'individual' && p.length === 2 && p.includes(me.value) && p.includes(username)
    })

    if (existing) {
      currentConvType.value = 'individual'
      pendingPeer.value = ''                      // leave draft mode
      selectConversation(existing)                // sets currentConversationId + title
      await loadMessages(idOf(existing))          // show existing messages immediately
      return
    }

    // No existing conversation → draft mode (composer visible, no messages yet)
    currentConvType.value = 'individual'
    pendingPeer.value = username
    currentConversationId.value = ''
    currentTitle.value = username
    participants.value = [me.value, username]
    messages.value = []
    draft.value = ''
  } catch (e) {
    console.error('open 1:1 failed', e)
  }
}


function sortContacts(arr) {
  arr.sort((a, b) => {
    if (a.lastAt && b.lastAt) return new Date(b.lastAt) - new Date(a.lastAt)
    if (a.lastAt && !b.lastAt) return -1
    if (!a.lastAt && b.lastAt) return 1
    return a.username.localeCompare(b.username)
  })
}


function upsertContactFromMessage(peer, msg) {
  const meta = {
    lastAt:     msg.timestamp     || msg.Timestamp || null,
    lastType:   msg.contentType   || msg.ContentType || null,
    lastText:   msg.text          || msg.Text || '',
    lastSender: msg.senderUsername|| msg.SenderUsername || msg.sender || '',
  }

  const idx = singleContacts.value.findIndex(i => i.username === peer)
  if (idx >= 0) {
    Object.assign(singleContacts.value[idx], meta)
  } else {
    singleContacts.value.push({
      key: peer, username: peer, display: peer,
      initial: peer?.[0]?.toUpperCase() || '?',
      photoUrl: userPhotos.value[peer] || '',
      unread: false,
      ...meta
    })
  }

  // keep the same ordering rule: recent first, then A–Z
  singleContacts.value.sort((a, b) => {
    if (a.lastAt && b.lastAt) return new Date(b.lastAt) - new Date(a.lastAt)
    if (a.lastAt && !b.lastAt) return -1
    if (!a.lastAt && b.lastAt) return 1
    return a.username.localeCompare(b.username)
  })
}



function ensureContactExists(username) {
  const idx = singleContacts.value.findIndex(c => c.username === username)
  if (idx >= 0) return
  toContactItem(username, { lastAt: null })
    .then(item => {
      singleContacts.value.push(item)
      sortContacts(singleContacts.value)
    })
    .catch(() => {})
}

function selectConversation(c) {
  pendingPeer.value = ''
  currentConversationId.value = c.id || c.ID
  const parts = c.participants || c.Participants || []
  currentTitle.value = parts?.find(p => p !== me.value) || 'Conversation'
}



function selectGroup(g) {
  currentConversationId.value = g.conversationId || g.id || g.ID
  currentTitle.value = g.groupName || g.name || 'Group'
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

// Messages state
const messages = ref([])
const draft = ref('')
const sending = ref(false)
const fileInput = ref(null)
const msgList = ref(null)

// tiny emoji picker
const showEmoji = ref(false)
const emojis = ['😀','😁','😂','😊','😍','👍','🙏','🎉','🔥','❤️']

function toggleEmoji() { showEmoji.value = !showEmoji.value }
function insertEmoji(e) { draft.value += e; showEmoji.value = false }

// Load messages when conversation changes
async function loadMessages(id) {
  try {
    const arr = await listMessages(id)
    // sort ASC so latest is at the bottom
    arr.sort((a, b) => {
      const ta = new Date(a.timestamp ?? a.Timestamp ?? 0).getTime()
      const tb = new Date(b.timestamp ?? b.Timestamp ?? 0).getTime()
      return ta - tb
    })
    messages.value = normalizeReactionsField(arr)
    await nextTick()
    scrollToBottom()
    await ensureReactions(arr); 
  } catch (e) {
    // optionally surface error
  }
}


watch(currentConversationId, async (id) => {
  // tear down previous polls
  clearInterval(statusTimer);   statusTimer = null
  clearInterval(messagesTimer); messagesTimer = null

  // reset per-thread state
  statusMap.value = new Map()
  messages.value  = []

  // nothing to do if no convo, not authed, or component unmounted
  if (!id) return
  if (!localStorage.getItem(TOKEN_KEY)) return
  if (!alive) return

  // load participants, history, and initial statuses
  await loadConvMeta(id); if (!alive) return
  await loadMessages(id); if (!alive) return
  await pollStatuses();   if (!alive) return

  // start polls (guarded by `alive`)
  statusTimer   = setInterval(() => { if (alive) pollStatuses() }, 2500)
  messagesTimer = setInterval(() => { if (alive) pollMessages() }, 2000)
})

watch(activeTab, (tab) => {
  // leaving a mismatched thread? clear the right pane
  if (tab === 'users'  && currentConversationId.value && (participants.value.length !== 2)) {
    currentConversationId.value = ''; currentTitle.value = ''; messages.value = []; pendingPeer.value = ''; currentConvType.value = ''
  }
  if (tab === 'groups' && currentConversationId.value && (participants.value.length === 2)) {
    currentConversationId.value = ''; currentTitle.value = ''; messages.value = []; pendingPeer.value = ''; currentConvType.value = ''
  }
})

watch([currentConversationId, currentConvType], async ([cid, typ]) => {
  // Keep currentGroupName in sync with the selected thread
  if (String(typ).toLowerCase() === 'group') {
    currentGroupName.value = findGroupNameByConv(cid) || ''
    if (showGroupMgmt.value) {
      gmMembers.value = []
      await loadGroupMembers()
    }
  } else {
    currentGroupName.value = ''
  }
})



// Scroll to bottom of thread
function scrollToBottom() {
  const el = msgList.value
  if (!el) return
  el.scrollTop = el.scrollHeight
}

// Send text
async function onSendText() {
  const caption  = draft.value.trim()
  const hasFile  = !!selectedFile.value
  const isNew1to1 = !currentConversationId.value && !!pendingPeer.value

  // nothing to send (and not creating a new 1:1)
  if (!caption && !hasFile && !isNew1to1) return

  let usedCaptionAsInitial = false

  // If this is a brand-new 1:1, create the conversation first.
  if (isNew1to1) {
    sending.value = true
    try {
      // Backend requires a non-empty initialMessage for 1:1 creation.
      // If the user typed text, use it; otherwise use a placeholder for image-first sends.
      const initial = caption ? caption : '📷 Photo'
      usedCaptionAsInitial = !!caption

      const conv = await createConversation(pendingPeer.value, initial)
      pendingPeer.value = ''
      selectConversation(conv)
      await nextTick()
    } catch (e) {
      sending.value = false
      return
    }

    // If we already used the typed caption as the initial message AND there’s no file to send,
    // we're done (avoid sending the same text again).
    if (!hasFile && usedCaptionAsInitial) {
      draft.value = ''
      sending.value = false
      refreshSingleContacts()
      return
    }
    // keep going below to send the image (with *no* duplicate caption)
  }

  if (!currentConversationId.value) return

  sending.value = true
  try {
    let msg
    if (hasFile) {
      // If the caption was already used as the initial text, do not attach it again to the image.
      const capForFile = usedCaptionAsInitial ? '' : caption
      msg = await sendFile(currentConversationId.value, selectedFile.value, undefined, capForFile)
    } else {
      // Only happens when initial text was NOT used (e.g., user typed after opening an existing convo)
      msg = await sendText(currentConversationId.value, caption)
    }

    messages.value.push(msg)
    await nextTick(); scrollToBottom()

    // clear inputs
    draft.value = ''
    selectedFile.value = null

    // update side panes
    if ((participants.value || []).length > 2) {
      // group
      upsertGroupFromMessage(currentConversationId.value, msg)
    } else {
      // 1:1
      const peer = participants.value.find(p => p !== me.value) || currentTitle.value
      upsertContactFromMessage(peer, msg)
      refreshSingleContacts()
    }
    refreshStatusesSoon()
  } finally {
    sending.value = false
  }
}




// Attach image
function onSelectFile(e) {
  const f = e.target.files?.[0]
  e.target.value = '' // allow re-selecting same file later
  if (!f) return
  selectedFile.value = f
}

// BEFORE (yours likely missed camelCase)
function isMine(m) {
  const s = m.senderUsername ?? m.sender_username ?? m.sender ?? m.Sender
  return s === me.value
}

function contentTypeOf(m) {
  return String(m.contentType ?? m.ContentType ?? '').toLowerCase()
}
function msgText(m) {
  return m.text ?? m.Text ?? ''
}
function msgImg(m) {
  return m.contentUrl ?? m.ContentURL ?? m.content_url ?? ''
}
function prettyTime(m) {
  const raw = m.timestamp ?? m.Timestamp
  const d = raw ? new Date(raw) : null
  if (!d || isNaN(+d)) return ''
  // yyyy-mm-dd hh:mm
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')} ` +
         `${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
}


onMounted(async () => {
  const token = localStorage.getItem(TOKEN_KEY);
  if (!token) return;                 // not logged in → don't start anything

  await Promise.all([loadUsersAndGroups(), loadMyProfile()]);
  await refreshSingleContacts();

  // lightweight middle-pane refresh
  if (!contactsTicker) {
    contactsTicker = setInterval(refreshSingleContacts, 5000);
  }
  if (!groupsTicker)   groupsTicker   = setInterval(refreshGroupSummaries, 5000);
});

// ================== CHECKMARK STATUS ==================
const statusMap = ref(new Map())      // mid -> { delivered: bool, read: bool }
const participants = ref([])          // usernames in current conversation (incl. me)
let   statusTimer = null

// Resolve message id regardless of backend casing
function idForMessage(m) {
  return (
    m?.id ??
    m?.ID ??
    m?.messageId ??
    m?.MessageID ??
    m?.messageID ??
    m?.message_id ??
    ''
  )
}
function idForRow(r) {
  return (
    r?.messageId ??
    r?.MessageID ??
    r?.messageID ??
    r?.id ??
    r?.message_id ??
    ''
  )
}

// Read status for a message
function statusOf(m) {
  const id = idForMessage(m)
  return statusMap.value.get(id) || { delivered: false, read: false }
}
function isDelivered(m) { return statusOf(m).delivered }
function isRead(m)      { return statusOf(m).read }

// Load conversation members (so we know who counts)
async function loadConvMeta(id) {
  const conv = await getConversation(id).catch(() => null)
  const typ  = String(conv?.type ?? conv?.Type ?? '').toLowerCase()
  currentConvType.value = typ
  const ppl =
    conv?.participants || conv?.Participants ||
    conv?.usernames    || conv?.Usernames   ||
    conv?.members      || conv?.Members     ||
    conv?.users        || conv?.Users       || []
  participants.value = Array.isArray(ppl) ? ppl : []
}

// Aggregate status rows into booleans for each message
function aggregateStatuses(rows) {
  // who counts toward ✓/✓✓ : everyone but me
  const others = new Set((participants.value || []).filter(u => u && u !== me.value))
  const need = others.size

  // mid -> Set(recipients) that have that status
  const deliveredBy = new Map()
  const readBy      = new Map()

  const add = (map, mid, who) => {
    let s = map.get(mid); if (!s) { s = new Set(); map.set(mid, s) }
    s.add(who)
  }

  for (const r of rows) {
    const mid = idForRow(r)
    if (!mid) continue

    // backend uses Recipient (keep aliases too)
    const who =
      r.Recipient ?? r.recipient ??
      r.username  ?? r.Username  ??
      r.user      ?? r.User      ?? ''
    if (!who || who === me.value || !others.has(who)) continue

    const st = String(r.Status ?? r.status ?? '').toLowerCase()

    // consider 'sent'/'received'/'delivered' as delivered; 'read' implies delivered too
    if (st === 'sent' || st === 'received' || st === 'delivered' || st === 'read') {
      add(deliveredBy, mid, who)
    }
    if (st === 'read') {
      add(readBy, mid, who)
    }
  }

  // Convert sets to booleans: all other participants must have that status
  const out = new Map()
  const mids = new Set([...deliveredBy.keys(), ...readBy.keys()])
  for (const mid of mids) {
    const d = deliveredBy.get(mid)?.size ?? 0
    const r = readBy.get(mid)?.size ?? 0
    out.set(mid, {
      delivered: need > 0 && d >= need,
      read:      need > 0 && r >= need,
    })
  }
  return out
}

async function pollStatuses() {
  const cid = currentConversationId.value
  if (!cid) return
  try {
    const rows = await messageStatuses(cid) || []
    if (!alive) return

    // dev aid (remove after verifying)
    if (import.meta.env.DEV && rows.length) console.debug('[statuses sample]', rows[0])

    statusMap.value = aggregateStatuses(rows)
  } catch (e) {
    if (e?.response?.status === 401) stopAllPollers()
  }
}
// ======================================================


// 🔸 life-cycle guard used in async code (pollers, loads, sends)
let alive = true
onUnmounted(() => { 
  alive = false; stopAllPollers() 
  if (groupMembersTicker) { clearInterval(groupMembersTicker); groupMembersTicker = null }
  if (gmTicker) { clearInterval(gmTicker); gmTicker = null}
})

// stop timers immediately when axios broadcasts a global 401
function onUnauthorized() { stopAllPollers() }
window.addEventListener(UNAUTHORIZED_EVENT, onUnauthorized)
window.addEventListener('click', () => { reactionBarForId.value = '' })

onUnmounted(() => window.removeEventListener(UNAUTHORIZED_EVENT, onUnauthorized))

// (optional) ensure logout also stops everything
function logout() {
  try {
    stopAllPollers()
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem('wasa_username')
  } finally {
    router.replace({ name: 'login' })
  }
}

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts)
  const now = new Date()
  const sameDay = d.toDateString() === now.toDateString()
  const opt = sameDay
    ? { hour: '2-digit', minute: '2-digit' }
    : { year: 'numeric', month: '2-digit', day: '2-digit' }
  return new Intl.DateTimeFormat(undefined, opt).format(d)
}

// pick the newest message in an array (defensive if API order changes)
function pickNewestMessage(msgs) {
  if (!Array.isArray(msgs) || !msgs.length) return null
  let newest = msgs[0]
  let newestTs = new Date(newest.createdAt || newest.CreatedAt || 0).getTime()
  for (let i = 1; i < msgs.length; i++) {
    const m = msgs[i]
    const t = new Date(m.createdAt || m.CreatedAt || 0).getTime()
    if (t > newestTs) { newest = m; newestTs = t }
  }
  return newest
}

// Build one contact row object
async function toContactItem(username, meta = {}) {
  let photoUrl = photoCache.get(username)
  if (photoUrl === undefined) {
    try {
      const u = await getUser(username)
      photoUrl = u?.photoUrl ? fullUrl(u.photoUrl) : ''
    } catch { photoUrl = '' }
    photoCache.set(username, photoUrl)
  }
  return {
    key: username,
    username,
    display: username,
    initial: username?.[0]?.toUpperCase() || '?',
    photoUrl,
    lastAt: meta.lastAt || null,
    lastType: meta.lastType || null,     // 'text' | 'image' | 'gif'
    lastText: meta.lastText || '',
    lastSender: meta.lastSender || '',   // <-- NEW
    // keep unread if you still have it in your model; otherwise it’s fine to omit
    unread: !!meta.unread,
  }
}



// small helper: limit concurrency so we don’t blast the API
async function mapWithLimit(items, limit, task) {
  const ret = []
  let i = 0
  const running = new Set()
  async function launch(idx) {
    const p = task(items[idx]).then((v) => { ret[idx] = v }).finally(() => running.delete(p))
    running.add(p)
    await p
  }
  while (i < items.length || running.size) {
    while (i < items.length && running.size < limit) { await launch(i++) }
    if (running.size) await Promise.race(running)
  }
  return ret
}

async function refreshSingleContacts() {
  const my = me.value
  const [usernames, convs] = await Promise.all([listUsers(), listConversations()])

const oneToOne = (convs || []).filter((c) => {
  const parts = c.participants || c.Participants || []
  const typ = String(c.type || c.Type || '').toLowerCase()
  // strictly 1:1
  return typ === 'individual' && parts.length === 2 && parts.includes(my)
  })

  const convMeta = await mapWithLimit(oneToOne, 4, async (c) => {
    const convId = c.id || c.ID
    const parts = c.participants || c.Participants || []
    const peer = parts.find((p) => p !== my)

    let lastAt = c.updatedAt || c.UpdatedAt || null
    let lastType = null
    let lastText = ''
    let lastSender = ''

    try {
      const msgs = await listMessages(convId)
      // backend returns DESC; be defensive anyway:
      const last = pickNewestMessage(msgs) || msgs[0]
      if (last) {
        lastAt     = last.timestamp     || last.Timestamp || lastAt
        lastType   = last.contentType   || last.ContentType || null
        lastText   = last.text          || last.Text || ''
        lastSender = last.senderUsername|| last.SenderUsername || last.sender ||  last.Sender ||''
      }
    } catch { /* ignore */ }

    return { peer, lastAt, lastType, lastText, lastSender }
  })

  const recent = new Map()
  for (const m of convMeta) recent.set(m.peer, m)

  const recentPeers = Array.from(recent.keys())
  const untouched = (usernames || []).filter((u) => u && u !== my && !recentPeers.includes(u))

  const items = []
  for (const u of recentPeers) items.push(await toContactItem(u, recent.get(u)))
  for (const u of untouched)  items.push(await toContactItem(u, { lastAt: null }))

  items.sort((a, b) => {
    if (a.lastAt && b.lastAt) return new Date(b.lastAt) - new Date(a.lastAt)
    if (a.lastAt && !b.lastAt) return -1
    if (!a.lastAt && b.lastAt) return 1
    return a.username.localeCompare(b.username)
  })

  singleContacts.value = items
}

// After sending something, refresh once quickly so the ✓ appears fast
function refreshStatusesSoon() { setTimeout(pollStatuses, 500) }
function partsOf(c) {
  return (
    c.participants || c.Participants ||
    c.usernames    || c.Usernames    ||
    c.members      || c.Members      ||
    c.users        || c.Users        ||
    []
  )
}
function idOf(c) { return c.id || c.ID || c.conversationId || c.ConversationID || '' }

// Normalize reactions array from a message to: [{ emoji, username }]
function rawReactions(m) {
  const arr = m?.reactions ?? m?.Reactions ?? m?.replies ?? []
  if (!Array.isArray(arr)) return []
  // Accept shapes like {emoji, username} or {Emoji, Username}
  return arr.map(r => ({
    emoji:     r.emoji ?? r.Emoji ?? r.reaction ?? r.Reaction ?? '',
    username:  r.username ?? r.Username ?? r.user ?? r.User ?? ''
  })).filter(r => r.emoji && r.username)
}

// Aggregate per emoji for display, mark if I reacted
function aggregateReactions(m) {
  const mine = me.value
  const byEmoji = new Map()
  for (const r of rawReactions(m)) {
    const rec = byEmoji.get(r.emoji) || { emoji: r.emoji, count: 0, mine: false }
    rec.count++
    if (r.username === mine) rec.mine = true
    byEmoji.set(r.emoji, rec)
  }
  return Array.from(byEmoji.values()) // [{emoji, count, mine}]
}

// quick check: did I react with this emoji?
function iReactedWith(m, emoji) {
  const want = String(emoji || '')
  return rawReactions(m).some(r => String(r.username) === me.value && String(r.emoji) === want)
}

// Local optimistic update helpers (keeps UI snappy while polling catches up)
function addLocalReaction(m, emoji) {
  if (!m.reactions) m.reactions = []
  m.reactions.push({ emoji, username: me.value })
}
function removeLocalReaction(m, emoji) {
  const arr = rawReactions(m)
  const idx = arr.findIndex(r => r.username === me.value && r.emoji === emoji)
  if (idx >= 0) {
    // remove the matching item from the original array (whatever casing it uses)
    const raw = m.reactions ?? m.Reactions ?? []
    // find by comparing fields defensively
    const j = raw.findIndex(x =>
      (x.emoji ?? x.Emoji) === emoji &&
      (x.username ?? x.Username) === me.value
    )
    if (j >= 0) raw.splice(j, 1)
    if (m.reactions) m.reactions = raw
    if (m.Reactions) m.Reactions = raw
  }
}

// Reaction bar state
const reactionBarForId = ref('')                       // messageId that has the bar open
const reactionChoices = ['👍','❤️','😂','😮','😢','🙏']  // pick your set

async function toggleReactionBar(m) {
  const mid = idForMessage(m)
  reactionBarForId.value = (reactionBarForId.value === mid) ? '' : mid
  await refreshOneMessage(mid)
}

async function toggleReaction(m, emoji) {
  const mid = idForMessage(m)
  if (!mid) return
  const mine = iReactedWith(m, emoji)

  try {
    if (mine) {
      removeLocalReaction(m, emoji)
      await removeReaction(mid, 'me')
    } else {
      addLocalReaction(m, emoji)
      await addReaction(mid, emoji)
    }
    // NEW: hydrate from server (persists across next poll)
    await refreshOneMessage(mid)
  } catch (e) {
    if (mine)  addLocalReaction(m, emoji)
    else       removeLocalReaction(m, emoji)
    console.warn('reaction error', e)
  } finally {
    if (reactionBarForId.value === mid) reactionBarForId.value = ''
  }
}

function normalizeReactionsField(arr) {
  for (const m of arr) {
    // If backend didn’t include reactions, keep it as an empty array (so UI works)
    const rx = m.reactions ?? m.Reactions
    if (!Array.isArray(rx)) m.reactions = []
  }
  return arr
}

function reactionsKey(m) {
  // stable, order-independent signature of reactions
  const arr = rawReactions(m).map(r => [String(r.emoji), String(r.username)]);
  arr.sort((a, b) => (a[0]+a[1]).localeCompare(b[0]+b[1]));
  return JSON.stringify(arr);
}

function closeCreateGroup() {
  showCreateGroup.value = false
  if (groupMembersTicker) { clearInterval(groupMembersTicker); groupMembersTicker = null }
}
async function submitCreateGroup() {
  cgError.value = ''

  const name = (cgName.value || '').trim()
  const unique = Array.from(new Set([me.value, ...cgMembers.value]))
  if (!name)          { cgError.value = 'Group name is required.'; return }
  if (unique.length < 2) { cgError.value = 'Pick at least 1 other member.'; return }

  cgBusy.value = true
  try {
    // 1) create group (server creates the conversation; if initialMessage is non-empty it will also create it)
    await createGroup({
      groupName: name,
      members: unique,
      initialMessage: (cgMessage.value || '').trim(),
    })

    // 2) (optional) upload photo if you added that flow here, then…

    // 3) fetch full details to get conversationId and check messages
    const g = await getGroup(name)  // { groupName, conversationId, photoUrl?, ... }
    const convId =
      g?.conversationId ?? g?.ConversationId ?? g?.conversationID ?? g?.ConversationID ?? ''

    if (convId) {
      // 4) make sure there is at least one message; if not and we had a cgMessage, seed it
      const msgs = await listMessages(convId)
      if ((!Array.isArray(msgs) || msgs.length === 0) && (cgMessage.value || '').trim()) {
        await sendText(convId, (cgMessage.value || '').trim())
      }
      // 5) refresh the groups middle pane and jump into the new thread
      await hydrateGroups()
      const row = (groupItems.value || []).find(x => x.groupName === name)
      if (row) await openGroup(row)
    } else {
      // Fallback: still refresh list so the new group appears
      await hydrateGroups()
    }

    // close + reset
    showCreateGroup.value = false
    cgName.value = ''; cgMembers.value = []; cgMessage.value = ''
  } catch (e) {
    cgError.value = e?.response?.data?.error || e?.message || 'Failed to create group.'
  } finally {
    cgBusy.value = false
  }
}

function onPickGroupPhoto(e) {
  const f = e.target.files?.[0]
  if (!f) return
  cgPhotoFile.value = f
  const rd = new FileReader()
  rd.onload = () => { cgPhotoPreview.value = String(rd.result || '') }
  rd.readAsDataURL(f)
}
function clearGroupPhoto() {
  cgPhotoFile.value = null
  cgPhotoPreview.value = ''
}
function openCreateGroup() {
  cgName.value = ''
  cgMembers.value = []
  cgMessage.value = ''
  cgError.value = ''
  clearGroupPhoto()
  showCreateGroup.value = true
  refreshUsers() 
  if (!groupMembersTicker) {
    groupMembersTicker = setInterval(refreshUsers, 4000) // every ~4s
  }
}
async function toGroupItem(groupInput) {
  const groupName = typeof groupInput === 'string'
    ? groupInput
    : (groupInput?.name || groupInput?.groupName || '');
  try {
    const g = groupName ? await getGroup(groupName) : (groupInput || {}); // { groupName, conversationId, photoUrl, ... }
    // keep only groups where I'm a member
    const members = g?.members || g?.Members || []
    if (!Array.isArray(members) || !members.includes(me.value)) {
      return null
    }
    // Be liberal about backend casing
    const convId =
      g?.conversationId ??
      g?.ConversationId ??
      g?.conversationID ??
      g?.ConversationID ??
      g?.id ??
      g?.ID ??
      '';

    // Try to fetch last message
    let lastAt = null, lastType = null, lastText = '', lastSender = '';
    if (convId) {
      try {
        const msgs = await listMessages(convId);
        // Normalize and pick newest by timestamp
        const arr = Array.isArray(msgs) ? msgs.slice() : [];
        arr.sort((a,b) =>
          new Date(a.timestamp ?? a.Timestamp ?? 0) - new Date(b.timestamp ?? b.Timestamp ?? 0)
        );
        const last = arr[arr.length - 1];
        if (last) {
          lastAt     = last.timestamp      ?? last.Timestamp ?? null;
          lastType   = (last.contentType   ?? last.ContentType ?? '').toLowerCase();
          lastText   = last.text           ?? last.Text ?? '';
          lastSender = last.senderUsername ?? last.SenderUsername ?? last.sender ?? last.Sender ?? '';
        }
      } catch { /* ignore */ }
    }

    const photoUrl = g?.photoUrl ? fullUrl(g.photoUrl) : '';

    return {
      key: groupName,
      groupName,
      display: groupName,
      conversationId: convId,
      photoUrl,
      lastAt, lastType, lastText, lastSender,
    };
  } catch {
    // fallback if GET /groups/{name} fails
    return { key: groupName, groupName, display: groupName, conversationId: '', photoUrl: '' };
  }
}

async function hydrateGroups() {
  try {
    const list = await listGroups() // array of names OR array of objects
    const items = await mapWithLimit(list, 4, toGroupItem)
    groupItems.value = (items || []).filter(Boolean)
  } catch {}
}

async function openGroup(item) {
  if (!item) return;

  // Ensure we have a conversationId
  if (!item.conversationId) {
    currentGroupName.value = item.groupName || item.display || ''
    currentConvType.value  = 'group'
    const g = await getGroup(item.groupName).catch(() => null);
    item.conversationId =
      g?.conversationId ?? g?.ConversationId ?? g?.conversationID ?? g?.ConversationID ?? item.conversationId ?? '';
  }

  pendingPeer.value = '';
  currentConversationId.value = item.conversationId || '';
  currentTitle.value = item.display || item.groupName || 'Group';

  if (currentConversationId.value) {
    // show history immediately
    await loadConvMeta(currentConversationId.value);
    await loadMessages(currentConversationId.value);
    await pollStatuses();
  }
}
function sortGroupItems() {
  groupItems.value.sort((a, b) => {
    if (a.lastAt && b.lastAt) return new Date(b.lastAt) - new Date(a.lastAt)
    if (a.lastAt && !b.lastAt) return -1
    if (!a.lastAt && b.lastAt) return 1
    return String(a.display).localeCompare(String(b.display))
  })
}
async function refreshGroupSummaries() {
  try {
    const convs = await listConversations()
    const my = me.value
    const groupConvs = (convs || []).filter(c =>
      String(c.type || c.Type || '').toLowerCase() === 'group' &&
      (partsOf(c) || []).includes(my)
    )

    const byId = new Map((groupItems.value || []).map(g => [g.conversationId, g]))
    let needFullHydrate = false

    await mapWithLimit(groupConvs, 3, async (c) => {
      const cid = idOf(c)
      const updatedAt = c.updatedAt || c.UpdatedAt || null
      const row = byId.get(cid)

      // If we don't know this group yet, do a full hydrate later
      if (!row) { needFullHydrate = true; return }

      // If the server says it's newer, fetch its last message
      if (updatedAt && (!row.lastAt || new Date(updatedAt) > new Date(row.lastAt))) {
        const msgs = await listMessages(cid).catch(() => [])
        const arr = Array.isArray(msgs) ? msgs.slice() : []
        arr.sort((a,b) => new Date(a.timestamp ?? a.Timestamp ?? 0) - new Date(b.timestamp ?? b.Timestamp ?? 0))
        const last = arr[arr.length - 1]
        if (last) upsertGroupFromMessage(cid, last)
      }
    })

    if (needFullHydrate) await hydrateGroups()
    sortGroupItems()
  } catch {/* ignore */}
}
function upsertGroupFromMessage(conversationId, msg) {
  if (!conversationId || !msg) return

  const i = (groupItems.value || []).findIndex(g => g.conversationId === conversationId)
  const patch = {
    lastAt:     msg.timestamp      ?? msg.Timestamp ?? null,
    lastType:   (msg.contentType   ?? msg.ContentType ?? '').toLowerCase(),
    lastText:   msg.text           ?? msg.Text ?? '',
    lastSender: msg.senderUsername ?? msg.SenderUsername ?? msg.sender ?? msg.Sender ?? '',
  }

  if (i >= 0) {
    Object.assign(groupItems.value[i], patch)
  } else {
    // If we somehow don’t have this group yet, create a minimal row.
    groupItems.value.push({
      key: conversationId,
      groupName: '',
      display: 'Group',
      conversationId,
      photoUrl: '',
      ...patch,
    })
  }
  sortGroupItems()
}
function senderOf(m) {
  return (
    m.senderUsername ??
    m.SenderUsername ??
    m.sender ??
    m.Sender ??
    ''
  )
}

async function refreshUsers() {
  try {
    const u = await listUsers()
    users.value = Array.isArray(u) ? u : []
    // keep only existing users and never allow selecting myself
    cgMembers.value = cgMembers.value.filter(x => users.value.includes(x) && x !== me.value)
    // (optional) refresh cached photos
    hydrateUserPhotos(users.value)
  } catch {/* ignore */}
}
function findGroupNameByConv(cid) {
  const row = (groupItems.value || []).find(g => g.conversationId === cid)
  return row?.groupName || ''
}

async function loadGroupMembers() {
  if (!currentGroupName.value) return
  try {
    const g = await getGroup(currentGroupName.value)
    gmMembers.value = Array.isArray(g?.members) ? g.members.slice() : []
    hydrateUserPhotos(gmMembers.value) // avatars in the drawer
  } catch { /* ignore */ }
}

function openGroupMgmt() {
  if (!isGroupThread.value) return
  gmError.value = ''

  // Always compute from the current conversation (do NOT keep the old value)
  currentGroupName.value = findGroupNameByConv(currentConversationId.value) || currentGroupName.value || ''

  // Clear stale content so UI doesn't show previous members for a split second
  gmMembers.value = []

  showGroupMgmt.value = true
  loadGroupMembers()

  if (gmTicker) clearInterval(gmTicker)
  gmTicker = setInterval(loadGroupMembers, 4000)
}


function closeGroupMgmt() {
  showGroupMgmt.value = false
  if (gmTicker) { clearInterval(gmTicker); gmTicker = null }
}

async function onAddMember(u) {
  gmBusy.value = true; gmError.value = ''
  try {
    await addGroupMember(currentGroupName.value, u)
    await loadGroupMembers()
    await hydrateGroups()                // keep middle pane fresh
    await loadConvMeta(currentConversationId.value) // participants[]
  } catch (e) {
    gmError.value = e?.response?.data?.error || e?.message || 'Failed to add member'
  } finally { gmBusy.value = false }
}

async function onRemoveMember(u) {
  gmBusy.value = true; gmError.value = ''
  try {
    await removeGroupMember(currentGroupName.value, u)
    await loadGroupMembers()
    await hydrateGroups()
    await loadConvMeta(currentConversationId.value)

    // If group would have < 2 members -> eliminate it for me
    if ((gmMembers.value || []).length < 2) {
      alert('Not enough members. This group will be eliminated.')
      // If I'm still in, leave it so it disappears from my lists
      if ((gmMembers.value || []).includes(me.value)) {
        await leaveGroup(currentGroupName.value)
      }
      closeGroupMgmt()
      // Clear the chat pane if we were in this conversation
      if (currentConversationId.value) {
        currentConversationId.value = ''
        currentTitle.value = ''
        messages.value = []
      }
      await hydrateGroups()
    }
  } catch (e) {
    gmError.value = e?.response?.data?.error || e?.message || 'Failed to remove member'
  } finally { gmBusy.value = false }
}

async function onLeaveGroup() {
  gmBusy.value = true; gmError.value = ''
  try {
    await leaveGroup(currentGroupName.value)
    closeGroupMgmt()
    if (currentConversationId.value) {
      currentConversationId.value = ''
      currentTitle.value = ''
      messages.value = []
    }
    await hydrateGroups()
  } catch (e) {
    gmError.value = e?.response?.data?.error || e?.message || 'Failed to leave group'
  } finally { gmBusy.value = false }
}

</script>

<style scoped>
:global(html, body, #app){ height:100%; margin:0; overflow:hidden; }
/* App shell becomes a column: topbar + content grid */
.home-wrap{
  width:100vw; height:100vh; background:#fff;
  display:flex; flex-direction:column; overflow:hidden;
}

/* New top bar */
.topbar{
  flex:0 0 48px;                 /* fixed height */
  display:flex; align-items:center;
  padding:0 16px;
  background:#0f172a;            /* slate-900 */
  color:#fff;
  border-bottom:1px solid rgba(255,255,255,.06);
  z-index:10;
}
.topbar-title{
  font-weight:700; letter-spacing:.02em;
}

/* The main 3-column grid fills the remaining space (no page scroll) */
.home-card{
  flex:1 1 auto;                 /* fill below the topbar */
  width:100%; height:auto;       /* no hard 100vh here */
  display:grid; grid-template-columns:200px 360px 1fr;
  overflow:hidden; background:#fff; border-radius:0; box-shadow:none;
}

/* Make sure only inner areas scroll */
.leftbar, .middle, .right{ min-height:0; overflow:hidden; }
.middle{ display:flex; flex-direction:column; }
.contacts-scroll{ flex:1 1 auto; min-height:0; overflow-y:auto; }
.right{ display:grid; grid-template-rows:auto 1fr auto; min-height:0; overflow:hidden; }
.msg-list{ overflow:auto; min-height:0; }

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
.right { 
  position: relative; 
  display: flex; 
  flex-direction: column; 
  min-height: 0;           /* <-- important */
  overflow: hidden;        /* keeps the layout tidy */
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

/* Right pane chat layout */
.chat {
  display: grid;
  grid-template-rows: auto 1fr auto; /* header, messages, composer */
  height: 100%;
  min-height: 0;         
}

.chat-empty {
  margin: auto; text-align: center; color: #9aa4b2;
}

/* Header */
.conv-title {
  margin: 0;
  padding: 14px 16px;
  border-bottom: 1px solid #eef0f4;
  font-size: 18px;
  font-weight: 700;
  color: #1f2937;
}

/* Messages area */
.msg-list {
  overflow: auto;
  padding: 12px 16px;
  background: #fafbfe;
  min-height: 0;          
  -webkit-overflow-scrolling: touch;  /* smooth on iOS */
}

.msg {
  max-width: 70%;
  margin: 8px 0;
  display: grid;
  gap: 4px;
}
.msg.me {
  margin-left: auto;
}
.bubble {
  background: #ffffff;
  border: 1px solid #e6eaf2;
  border-radius: 14px;
  padding: 10px 12px;
  line-height: 1.35;
  color: #111827;
  word-break: break-word;
}
.msg.me .bubble {
  background: #2563eb;
  border-color: #2563eb;
  color: #fff;
}
.bubble.image {
  padding: 0;
  overflow: hidden;
}
.bubble.image img {
  display: block;
  max-width: 360px;
  border-radius: 12px;
}

/* caption inside the same bubble */
.bubble--image .caption {
  padding: 8px 10px;
  font-size: 14px;
  line-height: 1.35;
  word-break: break-word;
  border-top: 1px solid rgba(0,0,0,0.06);
  border-radius: 0 0 14px 14px; /* bottom rounded */
}

/* match sent (mine) / received (theirs) colors */
.bubble--mine.bubble--image .caption {
  background: #2563eb;
  color: #fff;
  border-top-color: rgba(255,255,255,0.25);
}
.bubble--theirs.bubble--image .caption {
  background: #fff;
  color: #111827;
  border-top-color: #e6eaf2;
}
.meta-time {
  font-size: 12px;
  color: #6b7280;
}
.empty-thread {
  text-align: center;
  color: #6b7280;
  margin-top: 30px;
}

/* Composer */
.composer {
  position: relative;    
  flex-shrink: 0; 
  display: grid;
  grid-template-columns: auto auto 1fr auto;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-top: 1px solid #eef0f4;
  background: #fff;
}
.cbtn {
  border: 1px solid #e5e9f2;
  background: #fff;
  border-radius: 10px;
  padding: 8px 10px;
  font-size: 18px;
  cursor: pointer;
}
.cbtn:hover { background: #f8fafc; }

.cinput {
  min-height: 42px;
  max-height: 120px;
  border: 1px solid #e5e9f2;
  border-radius: 10px;
  padding: 10px 12px;
  resize: vertical;
  outline: none;
}

.sendbtn {
  border: none;
  border-radius: 10px;
  padding: 10px 16px;
  background: #2563eb;
  color: #fff;
  font-weight: 700;
  cursor: pointer;
}
.sendbtn:disabled { opacity: .6; cursor: default; }

/* Simple emoji popover */
.emoji-pop {
  position: absolute;
  bottom: 64px;
  left: 16px;
  background: #fff;
  border: 1px solid #e5e9f2;
  border-radius: 10px;
  padding: 6px;
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 4px;
}
.emoji-pop button {
  border: none;
  background: transparent;
  font-size: 20px;
  padding: 6px;
  cursor: pointer;
}
.emoji-pop button:hover { background: #f3f4f6; border-radius: 8px; }
/* Layout for messages */
.msg-list {
  overflow: auto;
  padding: 12px 16px;
  background: #fafbfe;
}

.msg-row {
  display: flex;
  gap: 8px;
  margin: 8px 0;
}
.msg-row.mine   { justify-content: flex-end; }
.msg-row.theirs { justify-content: flex-start; }

.bubble {
  max-width: 68%;
  padding: 10px 14px;
  border-radius: 18px;
  line-height: 1.35;
  word-break: break-word;
  box-shadow: 0 1px 2px rgba(0,0,0,.04);
}

/* Received (left) */
.bubble--theirs {
  background: #fff;
  color: #111827;
  border: 1px solid #e6eaf2;
  border-top-left-radius: 6px;   /* subtle “tail” effect */
}

/* Sent (right) */
.bubble--mine {
  background: #2563eb;
  color: #fff;
  border: 1px solid transparent;
  border-top-right-radius: 6px;  /* subtle “tail” effect */
}

/* Images */
.bubble--image { padding: 0; overflow: hidden; }
.bubble--image img {
  display: block;
  max-width: 320px;
  border-radius: 14px;
}

/* Timestamps under each message block */
.meta-time {
  font-size: 12px;
  color: #9ca3af;
  margin: 2px 4px 6px;
}
.meta--mine   { text-align: right; }
.meta--theirs { text-align: left; }

.checks { display: inline-flex; gap: 2px; margin-left: 6px; vertical-align: middle; }
.check { font-size: 12px; line-height: 1; }
.check--single::before { content: "✓";  color: #60a5fa; }  /* blue-400 */
.check--double::before { content: "✓✓"; color: #2563eb; letter-spacing: -2px; } /* blue-600 */
/* Middle pane layout */
.contacts-pane {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.pane-header {
  padding: 6px 10px 4px 10px;
  font-weight: 700;
  font-size: 12px;
  color: #6b7280;            /* slate-500 */
  text-transform: uppercase;
  letter-spacing: .04em;
}

.contacts-scroll {
  flex: 1 1 auto;
  min-height: 0;             /* allow child to scroll */
  overflow-y: auto;
  padding: 4px 0;
}

/* Row */
.chat-item {
  width: 100%;
  display: grid;
  grid-template-columns: 40px 1fr;
  gap: 10px;
  align-items: center;
  padding: 8px 10px;
  border: 0;
  background: transparent;
  border-radius: 10px;
  text-align: left;
  cursor: pointer;
}
.chat-item:hover { background: #f5f7fb; }

/* Avatar */
.avatar {
  width: 38px; height: 38px;
  border-radius: 50%;
  object-fit: cover;
  background: #e5e7eb;       /* slate-200 */
}
.avatar-fallback {
  display: grid; place-items: center;
  color: #374151;            /* slate-700 */
  font-weight: 700;
  font-size: 14px;
}

/* Text blocks */
.meta { min-width: 0; }
.row-1 {
  display: flex; align-items: baseline; justify-content: space-between;
  gap: 8px;
}
.name { font-weight: 600; color: #111827; }
.time { font-size: 12px; color: #9ca3af; white-space: nowrap; }

.row-2 {
  display: flex; align-items: center; gap: 6px;
  min-width: 0;
}
.snippet {
  flex: 1 1 auto; min-width: 0;
  font-size: 13px; color: #6b7280;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.snippet.dim { color: #9ca3af; }

.reactions-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 4px;
  margin-left: 4px;
}

.rx-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: 1px solid #e5e7eb;
  background: #fff;
  border-radius: 12px;
  padding: 2px 6px;
  font-size: 12px;
  cursor: pointer;
}
.rx-chip.mine {
  border-color: #2563eb;
  background: #eff6ff;
}
.rx-emoji { line-height: 1; }
.rx-count { color: #6b7280; }

.rx-add {
  border: 1px dashed #e5e7eb;
  background: #fff;
  border-radius: 12px;
  padding: 2px 6px;
  font-size: 12px;
  cursor: pointer;
  opacity: .85;
}

/* Make a positioning context for the popover */
.msg-row { position: relative; }
.reactions-row { position: relative; }

/* Popover */
.rx-pop {
  position: absolute;
  bottom: 28px;              /* sit above the row */
  z-index: 20;
  padding: 4px;
  border: 1px solid #e5e9f2;
  background: #fff;
  border-radius: 10px;
  display: flex;
  gap: 4px;
  box-shadow: 0 6px 16px rgba(0,0,0,.08);
}
.rx-pop.left  { left: 0;  }
.rx-pop.right { right: 0; }  /* anchor to the right for my messages */

.rx-pick {
  border: none;
  background: transparent;
  font-size: 18px;
  cursor: pointer;
  padding: 2px 4px;
}
.rx-pick:hover { background: #f3f4f6; border-radius: 8px; }
.cg-backdrop {
  position: fixed; inset: 0; background: rgba(15,23,42,.35);
  display: grid; place-items: center; z-index: 60;
  backdrop-filter: blur(1px);
}
.cg-modal {
  width: min(600px, 94vw);
  background: #fff; border-radius: 16px;
  box-shadow: 0 25px 60px rgba(0,0,0,.25);
  padding: 16px; display: grid; gap: 14px;
}
.cg-head { display: flex; align-items: center; justify-content: space-between; }
.cg-head h3 { margin: 0; font-size: 18px; font-weight: 700; color: #111827; }
.cg-x {
  border: none; background: transparent; font-size: 18px; line-height: 1;
  cursor: pointer; color: #6b7280; padding: 4px 6px; border-radius: 8px;
}
.cg-x:hover { background: #f3f4f6; }

.cg-top { display: flex; gap: 14px; align-items: center; }
.cg-avatar { display: grid; gap: 6px; justify-items: center; }
.cg-avatar img, .cg-avatar-ph {
  width: 72px; height: 72px; border-radius: 50%; object-fit: cover;
  background: #e5e7eb; display: grid; place-items: center; font-weight: 800; color: #374151;
}
.cg-photo-btn, .cg-photo-clear {
  border: 1px solid #e5e7eb; background: #fff; border-radius: 999px;
  padding: 4px 10px; font-size: 12px; cursor: pointer;
}
.cg-photo-clear { background: #fff5f5; border-color: #fecaca; color: #b91c1c; }
.grow { flex: 1; }

.cg-field { display: grid; gap: 6px; }
.cg-label { font-size: 13px; color: #4b5563; }
.cg-input, .cg-textarea {
  width: 90%; border: 1px solid #e5e7eb; border-radius: 10px; padding: 10px 12px; outline: none;
}
.cg-input:focus, .cg-textarea:focus { border-color: #93c5fd; box-shadow: 0 0 0 3px rgba(37,99,235,.08); }
.cg-textarea { min-height: 80px; resize: vertical; }
.cg-members { display: flex; flex-wrap: wrap; gap: 8px; }
.cg-pill {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 10px; border: 1px solid #e5e7eb; border-radius: 999px; background: #fff;
  font-size: 13px; cursor: pointer;
}
.cg-hint { font-size: 12px; color: #9ca3af; }

.cg-actions { display: flex; gap: 8px; justify-content: flex-end; margin-top: 4px; }
.cg-btn {
  border: 1px solid #e5e7eb; background: #fff; border-radius: 10px; padding: 8px 14px; cursor: pointer;
}
.cg-primary { border-color: #2563eb; background: #2563eb; color: #fff; }
.cg-error { color: #dc2626; font-size: 13px; }
.sender-line {
  font-size: 12px;
  font-weight: 600;
  color: #6b7280;      /* slate-500 */
  margin: 2px 6px 4px;
}
.sender--mine   { text-align: right; }
.sender--theirs { text-align: left;  }
/* ===== Group management drawer (right side) ===== */
.gm-backdrop{
  position: fixed; inset: 0;
  background: rgba(15,23,42,.35);
  backdrop-filter: blur(2px);
  z-index: 1000;            /* above everything */
  display: grid;
  place-items: stretch;
}
.gm-drawer{
  margin-left: auto;
  width: min(380px, 92vw);
  height: 100%;
  background: #fff;
  border-left: 1px solid #eef0f4;
  box-shadow: -16px 0 40px rgba(0,0,0,.14);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overscroll-behavior: contain;
}
.gm-head{
  display:flex; align-items:center; justify-content:space-between;
  padding-bottom: 6px; border-bottom:1px solid #eef0f4;
}
.gm-head h3{ margin:0; font-size:16px; font-weight:700; color:#111827; }
.gm-x{
  border:none; background:transparent; font-size:18px; line-height:1;
  cursor:pointer; color:#6b7280; padding:4px 6px; border-radius:8px;
}
.gm-x:hover{ background:#f3f4f6; }

.gm-section{ display:grid; gap:8px; }
.gm-label{
  font-size:12px; font-weight:700; color:#6b7280;
  text-transform:uppercase; letter-spacing:.04em;
}

.gm-members{
  display:grid; gap:8px;
  max-height: 38vh; overflow:auto;
  padding-right:4px;
}
.gm-member{
  display:flex; align-items:center; justify-content:space-between;
  gap:10px; padding:6px 8px; border-radius:10px;
  background:#fafbfe; border:1px solid #eef0f4;
}
.gm-user{ display:flex; align-items:center; gap:8px; min-width:0; }
.gm-avatar{
  width:28px; height:28px; border-radius:50%;
  object-fit:cover; background:#e5e7eb; flex:0 0 auto;
}
.gm-avatar.ph{ display:grid; place-items:center; font-weight:700; color:#374151; }
.gm-name{ font-weight:600; color:#111827; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.gm-me{ color:#9ca3af; font-size:12px; margin-left:4px; }

.gm-remove{
  border:1px solid #fecaca; background:#fff5f5; color:#b91c1c;
  border-radius:8px; padding:4px 10px; cursor:pointer;
}
.gm-remove:disabled{ opacity:.6; cursor:default; }

.gm-candidates{
  display:flex; flex-wrap:wrap; gap:6px;
  max-height: 28vh; overflow:auto; padding-right:4px;
}
.gm-add{
  border:1px solid #e5e7eb; background:#fff; color:#111827;
  border-radius:999px; padding:6px 10px; font-size:13px; cursor:pointer;
}
.gm-add:hover{ background:#f3f4f6; }

.gm-footer{ margin-top:auto; display:flex; justify-content:flex-end; gap:8px; }
.gm-leave{
  border:1px solid #fecaca; background:#fff5f5; color:#b91c1c;
  border-radius:10px; padding:8px 12px; cursor:pointer;
}
.gm-error{ color:#dc2626; font-size:13px; }
/* Header row */
.conv-head{
  display:flex; align-items:center; gap:8px;
  padding:12px 16px; border-bottom:1px solid #eef0f4;
}
.conv-title{
  margin:0; font-size:18px; font-weight:700; color:#1f2937;
  flex:1 1 auto; min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
}

/* Vertical three-dots button, right-aligned */
.conv-menu{
  margin-left:auto;
  width:32px; height:32px;
  display:grid; place-items:center;
  border:none; background:transparent; cursor:pointer;
  color:#6b7280; border-radius:8px;
}
.conv-menu:hover{ background:#f3f4f6; }

</style>
