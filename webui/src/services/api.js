// webui/src/services/api.js
import http from './axios'

/* --------------------------- Session / Auth --------------------------- */

// POST /session  body: { username }  -> { identifier, username }
export async function doLogin(username) {
  const { data } = await http.post('/session', { username })
  return data
}

/* ----------------------------- Users -------------------------------- */

// GET /users -> { usernames: [...] } (unwrap to array)
export async function getAllUsers() {
  const { data } = await http.get('/users')
  return data.usernames || []
}

// PATCH /user/name { username } -> { username }
export async function setMyUserName(username) {
  const { data } = await http.patch('/user/name', { username })
  return data
}

// PATCH /user/photo  multipart: photo -> { photoUrl }
export async function setMyPhoto(file) {
  const fd = new FormData()
  fd.append('photo', file)
  const { data } = await http.patch('/user/photo', fd)
  return data
}

export async function getUser(username) {
    const { data } = await http.get(`/users/${encodeURIComponent(username)}`)
    return data
  }

// List all usernames (backend should return { usernames: string[] })
export async function listUsers() {
    try {
      const { data } = await http.get('/users')
      // accept a few shapes to be robust
      if (Array.isArray(data)) return data
      if (Array.isArray(data?.usernames)) return data.usernames
      if (Array.isArray(data?.users)) return data.users
      return []
    } catch (e) {
      // surface a clear message upward
      const msg = e?.response?.data?.error || e?.message || 'Failed to load users'
      throw new Error(msg)
    }
  } 
  

/* -------------------------- Conversations --------------------------- */

// POST /conversations { type:'individual', recipient, initialMessage } -> Conversation
export async function createConversation(recipient, initialMessage) {
  const { data } = await http.post('/conversations', {
    type: 'individual',
    recipient,
    initialMessage,
  })
  return data
}

// GET /conversations -> { conversations: [...] } (unwrap)
export async function listConversations() {
    const { data } = await http.get('/conversations')
    return (data.conversations || []).sort((a, b) =>
      new Date(b.updatedAt || b.UpdatedAt) - new Date(a.updatedAt || a.UpdatedAt)
    )
  }
  

// GET /conversations/:id -> Conversation
export async function getConversation(id) {
  const { data } = await http.get(`/conversations/${encodeURIComponent(id)}`)
  return data
}

// GET /conversations/:id/messages -> { messages: [...] } (unwrap)
// NOTE: exported as listMessages to match your HomeView imports
export async function listMessages(id) {
  const { data } = await http.get(`/conversations/${encodeURIComponent(id)}/messages`)
  return data.messages || []
}

// GET /conversations/:id/messages/status -> { statuses: [...] } (unwrap)
export async function messageStatuses(id) {
  const { data } = await http.get(`/conversations/${encodeURIComponent(id)}/messages/status`)
  return data.statuses || []
}

/* ----------------------------- Messages ------------------------------ */

// internal helper; used by sendText/sendFile
async function sendMessage({ conversationId, text, file, kind }) {
  const fd = new FormData()
  fd.append('conversationId', conversationId)
  if (file) {
    fd.append('contentType', kind || (file.type?.includes('gif') ? 'gif' : 'image'))
    fd.append('file', file)
  } else {
    fd.append('contentType', 'text')
    fd.append('text', text)
  }
  const { data } = await http.post('/messages', fd)
  return data
}

// Exported convenience functions matching your HomeView usage
export async function sendText(conversationId, text) {
  return sendMessage({ conversationId, text })
}
export async function sendFile(conversationId, file, kind) {
  return sendMessage({ conversationId, file, kind })
}

// DELETE /messages/:id
export async function deleteMessage(messageId) {
  await http.delete(`/messages/${encodeURIComponent(messageId)}`)
  return true
}

// POST /messages/:id/reply { text } -> Message
export async function replyMessage(messageId, text) {
  const { data } = await http.post(`/messages/${encodeURIComponent(messageId)}/reply`, { text })
  return data
}

// POST /messages/:id/reaction { emoji } -> Reaction
export async function addReaction(messageId, emoji) {
  const { data } = await http.post(
    `/messages/${encodeURIComponent(messageId)}/reaction`,
    { emoji }
  )
  return data
}

// DELETE /messages/{id}/reaction/{reactionId}
export async function removeReaction(messageId, reactionId = 'me') {
  await http.delete(`/messages/${encodeURIComponent(messageId)}/reaction/${encodeURIComponent(reactionId)}`)
  return true
}
  

// POST /messages/:id/forward { conversationId } -> Message
export async function forwardMessage(messageId, conversationId) {
    const { data } = await http.post(`/messages/${encodeURIComponent(messageId)}/forward`, { conversationId })
    return data
  }
  

/* -------------------------------- Groups ----------------------------- */

// POST /groups { groupName, members, initialMessage } -> Group
export async function createGroup({ groupName, members, initialMessage }) {
  const { data } = await http.post('/groups', { groupName, members, initialMessage })
  return data
}

// POST /groups/:groupName/members { username } -> { username }
export async function addToGroup(groupName, username) {
  const { data } = await http.post(
    `/groups/${encodeURIComponent(groupName)}/members`,
    { username }
  )
  return data
}

// POST /groups/:groupName/leave
export async function leaveGroup(groupName) {
  await http.post(`/groups/${encodeURIComponent(groupName)}/leave`)
  return true
}

// PATCH /groups/:groupName/photo  multipart: photo -> { photoUrl }
export async function setGroupPhoto(groupName, file) {
  const fd = new FormData()
  fd.append('photo', file)
  const { data } = await http.patch(
    `/groups/${encodeURIComponent(groupName)}/photo`,
    fd
  )
  return data
}
// List groups (adjust to your backend response)
export async function listGroups() {
    // If your API exposes /groups or /conversations filtered by group, use that.
    // For now assume GET /conversations then filter group ones on the client:
        const convs = await listConversations()
        return (convs || [])
          .filter(c => c.type === 'group')
          .map(c => c.group || { id: c.id, groupName: 'Group' })
      }
      
/* ----------------------------- Utilities ----------------------------- */

export function fullUrl(u) {
  if (!u) return u
  return /^https?:\/\//i.test(u) ? u : `${__API_URL__}${u.startsWith('/') ? '' : '/'}${u}`
}
