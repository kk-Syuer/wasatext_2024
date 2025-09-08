<script>
currentUser: localStorage.getItem('last_username') || ''

// and in LoginView.vue after successful login:
localStorage.setItem('last_username', username.value)

import ErrorMsg from "../components/ErrorMsg.vue"
import {
  listConversations,
  listMessages,
  createConversation,
  sendText,
  sendFile,
} from "../services/api"

export default {
  name: "HomeView",
  components: { ErrorMsg },
  data() {
    return {
      errormsg: null,
      loading: false,

      // 会话与消息
      conversations: [],
      activeId: "",
      messages: [],
      loadingMsgs: false,

      // 发送与新建
      text: "",
      file: null,
      recipient: "",

      // 其他（示例占位）
      some_data: null,
    }
  },
  methods: {
    // 顶部工具栏：刷新
    async refresh() {
      this.loading = true
      this.errormsg = null
      try {
        await this.loadConversations()
        if (this.activeId) {
          await this.loadMessages(this.activeId)
        }
      } catch (e) {
        this.errormsg = e?.message || String(e)
      } finally {
        this.loading = false
      }
    },

    // 顶部工具栏：导出（示例：导出当前会话消息为 JSON）
    exportList() {
      try {
        const blob = new Blob([JSON.stringify(this.messages, null, 2)], {
          type: "application/json",
        })
        const url = URL.createObjectURL(blob)
        const a = document.createElement("a")
        a.href = url
        a.download = `messages-${this.activeId || "none"}.json`
        a.click()
        URL.revokeObjectURL(url)
      } catch (e) {
        this.errormsg = e?.message || String(e)
      }
    },

    // 顶部工具栏：新建（示例：快速和某个用户开始私聊）
    async newItem() {
      if (!this.recipient) {
        this.errormsg = "Please enter a recipient username (left panel top input)."
        return
      }
      try {
        const conv = await createConversation(this.recipient, "Hello!")
        this.recipient = ""
        await this.loadConversations()
        this.activeId = conv.id
        await this.loadMessages(conv.id)
      } catch (e) {
        this.errormsg = e?.message || String(e)
      }
    },

    // 加载我的会话列表
    async loadConversations() {
      const convs = await listConversations()
      this.conversations = convs || []
      // 若没有选中会话，自动选中第一个
      if (!this.activeId && this.conversations.length > 0) {
        this.activeId = this.conversations[0].id
      }
    },

    // 加载指定会话的消息
    async loadMessages(conversationId) {
      if (!conversationId) return
      this.loadingMsgs = true
      this.errormsg = null
      try {
        const msgs = await listMessages(conversationId)
        this.messages = msgs || []
      } catch (e) {
        this.errormsg = e?.message || String(e)
      } finally {
        this.loadingMsgs = false
      }
    },

    // 点击左侧会话
    async openConversation(id) {
      this.activeId = id
      await this.loadMessages(id)
    },

    // 发送消息：文本或文件
    async send() {
      if (!this.activeId) {
        this.errormsg = "Select a conversation first."
        return
      }
      this.errormsg = null
      try {
        let created = null
        if (this.file) {
          const kind = this.file.type?.includes("gif") ? "gif" : "image"
          created = await sendFile(this.activeId, this.file, kind)
          this.file = null
          // 清空文件 input（见模板中的 @change）
          this.$refs.fileInput && (this.$refs.fileInput.value = "")
        } else if (this.text.trim()) {
          created = await sendText(this.activeId, this.text.trim())
          this.text = ""
        }
        if (created) {
          // 后端返回新消息，插到顶部（你的 API 是倒序）
          this.messages.unshift(created)
        }
      } catch (e) {
        this.errormsg = e?.message || String(e)
      }
    },

    // 输入框：文件选择
    onFileChange(evt) {
      const f = evt?.target?.files?.[0]
      this.file = f || null
    },

    // 小工具：展示会话标题
    convTitle(c) {
      if (!c) return "Chat"
      if (c.type === "group") return `[Group] ${c.group?.groupName || c.id}`
      const names = (c.participants || []).map(p => p.username)
      return names.join(", ")
    },
  },
  async mounted() {
    await this.refresh()
  },
}
</script>

<template>
  <div class="container-fluid p-0">
    <!-- 顶部工具栏（保留模板样式） -->
    <div
      class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
      <h1 class="h2">Home page</h1>
      <div class="btn-toolbar mb-2 mb-md-0">
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="refresh">
            Refresh
          </button>
          <button type="button" class="btn btn-sm btn-outline-secondary" @click="exportList">
            Export
          </button>
        </div>
        <div class="btn-group me-2">
          <button type="button" class="btn btn-sm btn-outline-primary" @click="newItem">
            New
          </button>
        </div>
      </div>
		<h5 class="text-muted">Hello, {{ currentUser }}</h5>

    </div>

    <ErrorMsg v-if="errormsg" :msg="errormsg" />

    <div class="row g-0">
      <!-- 左侧：会话列表 + 新建私聊 -->
      <div class="col-12 col-md-4 border-end" style="height: calc(100vh - 140px); overflow: auto">
        <div class="p-3 border-bottom d-flex align-items-center gap-2">
          <input v-model="recipient" class="form-control" placeholder="Start chat with user…" />
          <button class="btn btn-sm btn-primary" @click="newItem" :disabled="!recipient">Start</button>
        </div>

        <ul class="list-group list-group-flush">
          <li
            v-for="c in conversations"
            :key="c.id"
            class="list-group-item list-group-item-action"
            :class="{ active: c.id === activeId }"
            @click="openConversation(c.id)"
          >
            <div class="fw-semibold">{{ convTitle(c) }}</div>
            <small class="text-muted">
              {{ c.lastMessage?.text || c.lastMessage?.contentType || 'No messages' }}
            </small>
          </li>
        </ul>
      </div>

      <!-- 右侧：聊天区 -->
      <div class="col-12 col-md-8 d-flex flex-column" style="height: calc(100vh - 140px)">
        <div class="p-3 border-bottom">
          <h5 class="m-0">
            {{ convTitle(conversations.find(x => x.id === activeId)) }}
          </h5>
        </div>

        <div class="flex-grow-1 p-3" style="overflow: auto">
          <div v-if="loadingMsgs" class="text-muted">Loading…</div>
          <div v-else-if="messages.length === 0" class="text-muted">No messages</div>

          <div v-else v-for="m in messages" :key="m.id" class="mb-3">
            <div class="fw-semibold">{{ m.sender?.username || m.senderUsername }}</div>

            <div v-if="m.contentType === 'text'">{{ m.text }}</div>
            <div v-else>
              <img
                v-if="m.contentUrl"
                :src="m.contentUrl"
                style="max-width:200px; max-height:200px"
              />
              <span v-else>{{ m.contentType }}</span>
            </div>

            <small class="text-muted">{{ m.timestamp }}</small>
          </div>
        </div>

        <form class="p-3 border-top d-flex gap-2" @submit.prevent="send">
          <input v-model="text" class="form-control" placeholder="Type a message…" />
          <input ref="fileInput" type="file" class="form-control" style="max-width: 280px" @change="onFileChange" />
          <button class="btn btn-primary">Send</button>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.list-group-item.active {
  background-color: #0d6efd;
  color: #fff;
}
</style>
