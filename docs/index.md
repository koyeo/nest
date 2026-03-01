---
layout: home

hero:
  name: Nest
  text: Local CI/CD, Simplified.
  tagline: One YAML. One Command. Build, deploy, and manage — all from your terminal.
  image:
    src: /logo.png
    alt: Nest
  actions:
    - theme: brand
      text: Get Started →
      link: /guide/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/koyeo/nest

features:
  - icon: 📝
    title: Single YAML Config
    details: Define servers, tasks, and pipelines in one clean nest.yaml — no complex toolchains required.
  - icon: ⚡
    title: One Command Execution
    details: "Build, deploy, restart — run any task with <code>nest run &lt;task&gt;</code>. Chain multiple tasks seamlessly."
  - icon: 🚀
    title: Full Pipeline
    details: Test → Build → Deploy → Health Check. A complete CI/CD pipeline in a single YAML file.
  - icon: ☁️
    title: Cloud Storage Relay
    details: Upload artifacts to OSS/S3 then deploy via pre-signed URLs — bypass slow VPN connections.
  - icon: 🔄
    title: Multi-Environment
    details: Manage dev, staging, and production with separate config files using the --config flag.
  - icon: 🔒
    title: Secure by Default
    details: SSH key auth, AES-256 encrypted credentials, pre-signed URLs with 1-hour expiry.
---

<style>
:root {
  --vp-home-hero-name-color: transparent;
  --vp-home-hero-name-background: -webkit-linear-gradient(
    120deg,
    #4ade80 30%,
    #22d3ee
  );
  --vp-home-hero-image-background-image: linear-gradient(
    -45deg,
    #4ade8050 50%,
    #22d3ee50 50%
  );
  --vp-home-hero-image-filter: blur(44px);
}

.dark {
  --vp-home-hero-image-background-image: linear-gradient(
    -45deg,
    #4ade8030 50%,
    #22d3ee30 50%
  );
}

@media (min-width: 640px) {
  :root {
    --vp-home-hero-image-filter: blur(56px);
  }
}

@media (min-width: 960px) {
  :root {
    --vp-home-hero-image-filter: blur(68px);
  }
}

/* ─── Feature card animations ─── */
.VPFeatures .VPFeature {
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.VPFeatures .VPFeature:hover {
  transform: translateY(-4px);
  box-shadow:
    0 12px 32px rgba(0, 0, 0, 0.08),
    0 2px 6px rgba(0, 0, 0, 0.03);
}

.dark .VPFeatures .VPFeature:hover {
  box-shadow:
    0 12px 32px rgba(0, 0, 0, 0.3),
    0 2px 6px rgba(0, 0, 0, 0.2);
}

/* ─── Hero image pulse ─── */
.VPImage.image-src {
  animation: float 6s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(-12px); }
}

/* ─── Quick install section (now rendered via Vue component above features) ─── */
</style>

