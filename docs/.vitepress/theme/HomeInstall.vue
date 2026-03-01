<script setup>
import { ref } from 'vue'

const copied = ref(false)
const command = 'curl -fsSL https://raw.githubusercontent.com/koyeo/nest/master/scripts/install.sh | bash'

function copy() {
  navigator.clipboard.writeText(command)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}
</script>

<template>
  <div class="install-section" v-if="$frontmatter?.layout === 'home'">
    <div class="install-inner">
      <h2 class="install-title">⚡ Quick Install</h2>
      <p class="install-desc">Auto-detects your OS and architecture. One line to install.</p>
      <div class="install-box" @click="copy">
        <code class="install-code">$ {{ command }}</code>
        <button class="copy-btn" :class="{ copied }">
          <span v-if="copied">✓ Copied</span>
          <span v-else>Copy</span>
        </button>
      </div>
      <p class="install-alt">
        or via Go: <code>go install github.com/koyeo/nest@latest</code>
      </p>
    </div>
  </div>
</template>

<style scoped>
.install-section {
  max-width: 960px;
  margin: -20px auto 0;
  padding: 0 24px 48px;
}

.install-inner {
  text-align: center;
}

.install-title {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
  background: linear-gradient(120deg, #4ade80, #22d3ee);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.install-desc {
  color: var(--vp-c-text-2);
  margin-bottom: 16px;
  font-size: 15px;
}

.install-box {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--vp-c-bg-soft);
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  padding: 16px 20px;
  cursor: pointer;
  transition: border-color 0.25s, box-shadow 0.25s;
}

.install-box:hover {
  border-color: var(--vp-c-brand-1);
  box-shadow: 0 2px 12px rgba(74, 222, 128, 0.1);
}

.install-code {
  flex: 1;
  text-align: left;
  font-family: var(--vp-font-family-mono);
  font-size: 13px;
  color: var(--vp-c-text-1);
  word-break: break-all;
  line-height: 1.6;
}

.copy-btn {
  flex-shrink: 0;
  padding: 6px 14px;
  border-radius: 8px;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg);
  color: var(--vp-c-text-2);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.25s;
}

.copy-btn:hover {
  color: var(--vp-c-brand-1);
  border-color: var(--vp-c-brand-1);
}

.copy-btn.copied {
  color: #4ade80;
  border-color: #4ade80;
}

.install-alt {
  margin-top: 12px;
  font-size: 13px;
  color: var(--vp-c-text-3);
}

.install-alt code {
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-text-2);
}
</style>
