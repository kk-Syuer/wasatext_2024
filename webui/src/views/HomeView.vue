<script>
import ErrorMsg from "../components/ErrorMsg.vue"
import {
  listConversations,
  listMessages,
  createConversation,
  sendText,
  sendFile,
  listUsers,
  listGroups
} from "../services/api"

export default {
  name: "HomeView",
  components: { ErrorMsg },
  data() {
    return {
      currentUser: localStorage.getItem("last_username") || "",
      errormsg: null,
      loading: false,

      activeTab: "chats", // 'chats' | 'contacts' | 'groups'

      conversations: [],
      contacts: [],
      groups: [],

      activeId: "",
      activeTitle: "Chat",
      messages: [],
      loadingMsgs: false,

      text: "",
      file: null
    }
  },
  computed: {
    centerItems() {
      if (this.activeTab === "contacts") return this.contacts
      if (this.activeTab === "groups") return this.groups
      return this.conversations
    }
  },
  methods: {
    async refresh() {
      this.errormsg = null; this.loading = true
      try {
        await Promise.all([
          this.loadConversations(),
          this.loadContacts(),
          this.loadGroups(),
        ])
        if (this.activeTab === "chats" && !this.activeId && this.conversations.length) {
          this.openConversation(this.conversations[0])
        }
      } catch (e) {
        this.errormsg = e?.message || String(e)
      } finally {
        this.loading = false
      }
    },
    async loadConversations() { this.conversations = (await listConversations()) || [] },
    async loadContacts() {
      try {
        const users = await listUsers()
        this.contacts = (users || []).filter(u => u !== this.currentUser)
      } catch (e) { console.warn(e) }
    },
    async loadGroups() {
      try { this.groups = (await listGroups()) || [] }
      catch (e) { console.warn(e) }
    },
    async selectTab(tab) {
      this.activeTab = tab
      if (tab === "chats" && !this.activeId && this.conversations.length) {
        this.openConversation(this.conversations[0])
      }
    },
    async openConversation(item) {
      try {
        if (this.activeTab === "contacts") {
          const conv = await createConversation(item, "Hello!")
          await this.loadConversations()
          const found = this.conversations.find(c => c.id === conv.id)
          this.activeId = conv.id
          this.activeTitle = this.convTitle(found || conv)
          await this.loadMessages(this.activeId)
          this.activeTab = "chats"
          return
        }
        if (this.activeTab === "groups") {
          const conv = this.conversations.find(
            c => c.type === "group" &&
            (c.group?.groupName === (item.groupName || item) || c.id === item.id)
          )
          if (conv) {
            this.activeId = conv.id
            this.activeTitle = this.convTitle(conv)
            await this.loadMessages(conv.id)
            this.activeTab = "chats"
          } else {
            await this.loadConversations()
          }
          return
        }
        // item is a conversation
        this.activeId = item.id
        this.activeTitle = this.convTitle(item)
        await this.loadMessages(item.id)
      } catch (e) {
        this.errormsg = e?.message || String(e)
      }
    },
    async loadMessages(conversationId) {
      if (!conversationId) return
      this.loadingMsgs = true; this.errormsg = null
      try { this.messages = (await listMessages(conversationId)) || [] }
      catch (e) { this.errormsg = e?.message || String(e) }
      finally { this.loadingMsgs = false }
    },
    async send() {
      if (!this.activeId) return
      try {
        let created = null
        if (this.file) {
          const kind = this.file.type?.includes("gif") ? "gif" : "image"
          created = await sendFile(this.activeId, this.file, kind)
          this.file = null
          this.$refs.fileInput && (this.$refs.fileInput.value = "")
        } else if (this.text.trim()) {
          created = await sendText(this.activeId, this.text.trim())
          this.text = ""
        }
        if (created) this.messages.unshift(created)
      } catch (e) {
        this.errormsg = e?.message || String(e)
      }
    },
    onFileChange(evt) { this.file = evt?.target?.files?.[0] || null },
    convTitle(c) {
      if (!c) return "Chat"
      if (c.type === "group") return c.group?.groupName || "Group"
      const names = (c.participants || []).map(p => p.username)
      const title = names.filter(n => n !== this.currentUser).join(", ")
      return title || "Chat"
    },
    avatar(text) { return (text || "?").slice(0, 1).toUpperCase() }
  },
  async mounted() { await this.refresh() }
}
</script>

<template>
  <div class="layout">
    <!-- LEFT -->
	<aside class="left">
	<div class="me"><div class="avatar">{{ avatar(currentUser) }}</div></div>
	<div class="lscroll">
		<div class="tabs">
		<button class="tab" :class="{active:activeTab==='chats'}"    @click="selectTab('chats')"    title="Chats">💬</button>
		<button class="tab" :class="{active:activeTab==='contacts'}" @click="selectTab('contacts')" title="Contacts">👥</button>
		<button class="tab" :class="{active:activeTab==='groups'}"   @click="selectTab('groups')"   title="Groups">🗂️</button>
		</div>
	</div>
	</aside>


    <!-- CENTER -->
    <section class="center">
      <div class="head">
        <h5 v-if="activeTab==='chats'">Chats</h5>
        <h5 v-else-if="activeTab==='contacts'">Contacts</h5>
        <h5 v-else>Groups</h5>
        <button class="btn sm ghost" @click="refresh">Refresh</button>
      </div>

      <ErrorMsg v-if="errormsg" :msg="errormsg" />

      <div class="list">
        <template v-if="activeTab==='chats'">
          <div v-for="c in centerItems" :key="c.id" class="row" :class="{active:c.id===activeId}" @click="openConversation(c)">
            <div class="badge">{{ avatar(convTitle(c)) }}</div>
            <div class="col">
              <div class="title">{{ convTitle(c) }}</div>
              <div class="sub">{{ c.lastMessage?.text || c.lastMessage?.contentType || 'No messages' }}</div>
            </div>
          </div>
          <div v-if="centerItems.length===0" class="empty">No conversations yet.</div>
        </template>

        <template v-else-if="activeTab==='contacts'">
          <div v-for="u in centerItems" :key="u" class="row" @click="openConversation(u)">
            <div class="badge">{{ avatar(u) }}</div>
            <div class="col">
              <div class="title">{{ u }}</div>
              <div class="sub">Start chat</div>
            </div>
          </div>
          <div v-if="centerItems.length===0" class="empty">No contacts found.</div>
        </template>

        <template v-else>
          <div v-for="g in centerItems" :key="g.groupName || g.id" class="row" @click="openConversation(g)">
            <div class="badge">G</div>
            <div class="col">
              <div class="title">{{ g.groupName || g.id }}</div>
              <div class="sub">Open group chat</div>
            </div>
          </div>
          <div v-if="centerItems.length===0" class="empty">No groups yet.</div>
        </template>
      </div>
    </section>

    <!-- RIGHT -->
    <section class="right">
      <div class="rhead"><div class="rtitle">{{ activeTitle }}</div></div>
      <div class="rbody">
        <div v-if="loadingMsgs" class="muted">Loading…</div>
        <div v-else-if="messages.length===0" class="muted">No messages</div>

        <div v-else v-for="m in messages" :key="m.id" class="msg">
          <div class="who">{{ m.sender?.username || m.senderUsername }}</div>
          <div class="bubble" :class="{me: m.sender?.username===currentUser}">
            <template v-if="m.contentType==='text'">{{ m.text }}</template>
            <img v-else-if="m.contentUrl" :src="m.contentUrl" class="img" />
            <span v-else>{{ m.contentType }}</span>
          </div>
          <div class="time">{{ m.timestamp }}</div>
        </div>
      </div>

      <form class="compose" @submit.prevent="send">
        <input v-model="text" class="inp" placeholder="Type a message…" />
        <input ref="fileInput" type="file" class="file" @change="onFileChange" />
        <button class="btn">Send</button>
      </form>
    </section>
  </div>
</template>

<style scoped>
/* --- GRID SKELETON ------------------------------------------------------ */
.layout {
  display: grid;
  grid-template-columns: 72px 320px 1fr;
  /* subtract your top bar height; adjust if your header height differs */
  height: calc(100vh - 48px);
}

/* IMPORTANT for independent scrolling inside CSS grid */
.left, .center, .right { min-height: 0; }

/* --- LEFT BAR (ROOT) ---------------------------------------------------- */
.left {
  border-right: 1px solid #e5e7eb;
  padding: 12px 8px 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}
/* make left pane scroll independently if items overflow */
.lscroll { overflow: auto; width: 100%; display: flex; flex-direction: column; align-items: center; }

.me .avatar {
  width: 42px; height: 42px; border-radius: 50%;
  background: #111827; color: #fff; display: grid; place-items: center; font-weight: 700;
}
.tabs { display: grid; gap: 8px; }
.tab {
  width: 44px; height: 44px; border-radius: 12px; border: 1px solid #e5e7eb; background: #fff;
  cursor: pointer; font-size: 18px;
}
.tab.active { border-color: #2563eb; box-shadow: 0 0 0 3px rgba(37,99,235,.15); }

/* --- CENTER (LIST LEVEL) ------------------------------------------------ */
.center {
  border-right: 1px solid #e5e7eb;
  display: flex; flex-direction: column; min-height: 0;
}
.head {
  position: sticky; top: 0; z-index: 1; /* sticky header */
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 12px; border-bottom: 1px solid #e5e7eb; background: #fff;
}
/* breadcrumb vibe: left(root) → middle(list) → right(detail) */
.head h5 { margin: 0; font-weight: 700; letter-spacing: .2px; }

.list { overflow: auto; min-height: 0; }
.row {
  display: grid; grid-template-columns: 44px 1fr; gap: 10px; align-items: center;
  padding: 10px 12px; border-bottom: 1px solid #f1f5f9; cursor: pointer;
}
.row.active { background: #eff6ff; }
.badge {
  width: 36px; height: 36px; border-radius: 50%; background: #e5e7eb;
  display: grid; place-items: center; font-weight: 700;
}
.title { font-weight: 600; }
.sub   { font-size: 12px; color: #6b7280; }
.empty { padding: 12px; color: #6b7280; }

/* --- RIGHT (DETAIL / CHAT) --------------------------------------------- */
.right { display: grid; grid-template-rows: auto 1fr auto; min-height: 0; }
.rhead {
  position: sticky; top: 0; z-index: 1;
  padding: 10px 14px; border-bottom: 1px solid #e5e7eb; background: #fff;
}
.rtitle { font-weight: 700; }
.rpath  { font-size: 12px; color: #6b7280; } /* optional breadcrumb line */

.rbody { overflow: auto; padding: 16px; display: grid; gap: 12px; min-height: 0; }
.msg .who  { font-size: 12px; color: #6b7280; margin-left: 2px; }
.bubble {
  display: inline-block; background: #f8fafc; border: 1px solid #e5e7eb; border-radius: 10px;
  padding: 8px 10px; max-width: 520px;
}
.bubble.me { background: #dbeafe; border-color: #bfdbfe; }
.img { max-width: 240px; max-height: 240px; border-radius: 8px; }
.time { font-size: 11px; color: #94a3b8; margin-left: 2px; }

.compose {
  display: grid; grid-template-columns: 1fr 240px auto; gap: 8px;
  padding: 10px; border-top: 1px solid #e5e7eb; background: #fff;
}
.inp  { padding: 8px 10px; border: 1px solid #d1d5db; border-radius: 8px; }
.file { border: 1px solid #d1d5db; border-radius: 8px; padding: 6px; }
.btn  { background: #2563eb; color:#fff; border:0; padding:8px 14px; border-radius:8px; cursor:pointer; }
.btn.sm { padding: 6px 10px; font-size: 12px; }
.btn.ghost { background:#fff; color:#374151; border:1px solid #d1d5db; }
.muted { color:#6b7280; }
</style>
