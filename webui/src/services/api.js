// webui/src/services/api.js
import http from './axios'

/* --------------------------- Session / Auth --------------------------- */
/**
 * POST /session  body: { username }  -> { identifier, username }
 * We intentionally send a "simple request" (no headers) using fetch
 * so the browser does NOT send a CORS preflight (your CORS allows only x-example-header).
 * In dev (Vite on :5173) we try the same-origin path first; if that fails, fall back to __API_URL__.
 */

function myUser() {
  return (
    localStorage.getItem('wasa_username') ||
    localStorage.getItem('wasa_token') ||
    ''
  )
}

export async function doLogin(username) {
  const name = String(username || '').trim();
  if (!name) throw new Error('Username is required');

  // Use the same axios instance (baseURL = /api in dev, or VITE_API_BASE in prod)
  const { data } = await http.post('/session', { username: name });
  return data; // { identifier, username }
}


/* ----------------------------- Users -------------------------------- */

// GET /users -> { usernames: [...] } (unwrap to array)
export async function getAllUsers() {
  const { data } = await http.get('/users')
  return data?.usernames || []
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

// GET /users/:username -> { username, photoUrl, ... }
export async function getUser(username) {
  const { data } = await http.get(`/users/${encodeURIComponent(username)}`)
  return data
}

// List all usernames (robust to a few shapes)
export async function listUsers() {
  const { data } = await http.get('/users')
  if (Array.isArray(data)) return data
  if (Array.isArray(data?.usernames)) return data.usernames
  if (Array.isArray(data?.users)) return data.users
  return []
}

/* -------------------------- Conversations --------------------------- */

function myUsername() {
  // token == username in your app
  return localStorage.getItem('wasa_username') || localStorage.getItem('wasa_token') || ''
}

// Create a 1-to-1 conversation AND the first text message in one step.
// The backend REQUIRES this JSON body.
export async function createConversation(recipient, initialMessage) {
  const { data } = await http.post('/conversations', {
    type: 'individual',
    recipient,
    initialMessage,
  });
  return data; // => Conversation object (with id/participants/updatedAt)
}

// GET /conversations -> { conversations: [...] } (unwrap + sort by updatedAt desc)
export async function listConversations() {
  const { data } = await http.get('/conversations')
  const list = data?.conversations || []
  return list.sort((a, b) =>
    new Date(b.updatedAt || b.UpdatedAt) - new Date(a.updatedAt || a.UpdatedAt)
  )
}

// GET /conversations/:id -> Conversation
export async function getConversation(id) {
  const { data } = await http.get(`/conversations/${encodeURIComponent(id)}`)
  return data
}

// GET /conversations/:id/messages -> { messages: [...] }
export async function listMessages(id) {
  const { data } = await http.get(`/conversations/${encodeURIComponent(id)}/messages`)
  return data?.messages || []
}

// GET /conversations/:id/messages/status -> { statuses: [...] } (unwrap)
export async function messageStatuses(conversationId) {
  const { data } = await http.get(
    `/conversations/${encodeURIComponent(conversationId)}/messages/status`
  )
  return Array.isArray(data) ? data : (data?.statuses || [])
}

/* ----------------------------- Messages ------------------------------ */

// internal helper; used by sendText/sendFile
async function sendMessage({ conversationId, text, file, kind }) {
  const fd = new FormData()
  fd.append('conversationId', conversationId)
  if (file) {
    fd.append('contentType', kind || (file.type?.includes('gif') ? 'gif' : 'image'))
    fd.append('file', file)
    if (text && String(text).trim()) {
      // include caption alongside the image/gif
      fd.append('text', text);
    }
  } else {
    fd.append('contentType', 'text')
    fd.append('text', text)
  }
  const { data } = await http.post('/messages', fd)
  return data
}

export async function sendText(conversationId, text) {
  return sendMessage({ conversationId, text })
}
export async function sendFile(conversationId, file, kind, caption = '') {
  return sendMessage({ conversationId, file, kind, text: caption });
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
  await http.delete(
    `/messages/${encodeURIComponent(messageId)}/reaction/${encodeURIComponent(reactionId)}`
  )
  return true
}

// POST /messages/:id/forward { conversationId } -> Message
export async function forwardMessage(messageId, conversationId) {
  const { data } = await http.post(
    `/messages/${encodeURIComponent(messageId)}/forward`,
    { targetConversationId: conversationId }
  )
  return data
}

// Fetch a single message (used to hydrate reactions on demand)
export async function getMessage(id) {
  const { data } = await http.get(`/messages/${encodeURIComponent(id)}`);
  return data;
}


/* -------------------------------- Groups ----------------------------- */

// POST /groups { groupName, members, initialMessage } -> Group
export async function createGroup({ groupName, members = [], initialMessage }) {
  const name = String(groupName || '').trim();
  const initial = String(initialMessage || '').trim() || 'Group created';
  const uniq = Array.from(new Set((members || []).map(s => String(s || '').trim()).filter(Boolean)));
  if (!name) throw new Error('Group name is required');
  if (uniq.length < 1) throw new Error('Pick at least one member');
  const { data } = await http.post('/groups', { groupName: name, members: uniq, initialMessage: initial })
  return data
}

// GET /groups -> { groups: [...] } (unwrap)
export async function listGroups() {
  const { data } = await http.get('/groups')
  return Array.isArray(data) ? data : (data?.groups || [])
}

// POST /groups/:groupName/members { username } -> { username }
export async function addToGroup(groupName, username) {
  const { data } = await http.post(
    `/groups/${encodeURIComponent(groupName)}/members`,
    { username }
  )
  return data
}
export async function addGroupMember(name, username) {
  const { data } = await http.post(`/groups/${encodeURIComponent(name)}/members`, { username })
  return data
}
export async function removeGroupMember(name, username) {
  await http.delete(`/groups/${encodeURIComponent(name)}/members/${encodeURIComponent(username)}`)
}
export async function leaveGroup(name) {
  await http.post(`/groups/${encodeURIComponent(name)}/leave`)
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

// GET /groups/:groupName -> group details ({ groupName, conversationId, members, ... })
export async function getGroup(groupName) {
  const { data } = await http.get(`/groups/${encodeURIComponent(groupName)}`)
  return data
}


/* ----------------------------- Utilities ----------------------------- */

// Helper to absolutize backend-returned paths like "/uploads/xyz.jpg"
export function fullUrl(path, { cacheBust } = {}) {
  if (!path) return '';

  // leave already-absolute and safe URLs alone
  const lower = String(path).toLowerCase();
  if (
    lower.startsWith('http://') ||
    lower.startsWith('https://') ||
    lower.startsWith('data:') ||
    lower.startsWith('blob:') ||
    lower.startsWith('about:')
  ) {
    return path;
  }

  // Optional override for a different uploads origin (CDN, separate host)
  const origin = (import.meta?.env?.VITE_UPLOADS_ORIGIN && import.meta.env.VITE_UPLOADS_ORIGIN.trim())
    || window.location.origin;

  try {
    const url = new URL(path, origin);
    if (cacheBust) {
      url.searchParams.set('v', String(cacheBust));
    }
    return url.toString();
  } catch {
    return path;
  }
}
