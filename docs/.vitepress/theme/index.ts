import DefaultTheme from 'vitepress/theme'
import HomeInstall from './HomeInstall.vue'
import { h } from 'vue'

export default {
    extends: DefaultTheme,
    Layout() {
        return h(DefaultTheme.Layout, null, {
            'home-features-before': () => h(HomeInstall),
        })
    },
}
