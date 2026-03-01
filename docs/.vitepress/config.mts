import { defineConfig } from 'vitepress'

export default defineConfig({
    title: 'Nest',
    description: 'Lightweight local CI/CD tool for rapid integration and deployment',
    head: [
        ['link', { rel: 'icon', href: '/logo.png' }],
        ['link', { rel: 'preconnect', href: 'https://fonts.googleapis.com' }],
        ['link', { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' }],
        ['link', { href: 'https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap', rel: 'stylesheet' }],
    ],
    themeConfig: {
        logo: '/logo.png',
        nav: [
            { text: 'Guide', link: '/guide/getting-started' },
            { text: 'Reference', link: '/reference/cli' },
            {
                text: 'Links',
                items: [
                    { text: 'GitHub', link: 'https://github.com/koyeo/nest' },
                    { text: 'Changelog', link: 'https://github.com/koyeo/nest/releases' },
                ]
            }
        ],
        sidebar: {
            '/guide/': [
                {
                    text: 'Introduction',
                    items: [
                        { text: 'Getting Started', link: '/guide/getting-started' },
                    ]
                },
                {
                    text: 'Core Concepts',
                    items: [
                        { text: 'Configuration', link: '/guide/configuration' },
                        { text: 'Deployment', link: '/guide/deployment' },
                        { text: 'Cloud Storage', link: '/guide/cloud-storage' },
                        { text: 'Multi-Environment', link: '/guide/multi-environment' },
                    ]
                }
            ],
            '/reference/': [
                {
                    text: 'Reference',
                    items: [
                        { text: 'CLI Commands', link: '/reference/cli' },
                    ]
                }
            ]
        },
        socialLinks: [
            { icon: 'github', link: 'https://github.com/koyeo/nest' }
        ],
        footer: {
            message: 'Released under the MIT License.',
            copyright: 'Copyright © 2024-present Koyeo'
        },
        search: {
            provider: 'local'
        }
    }
})
