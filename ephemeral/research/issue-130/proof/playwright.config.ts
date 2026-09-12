import { defineConfig } from '@playwright/test'

export default defineConfig({
	testDir: '.',
	testMatch: 'browser.spec.ts',
	workers: 1,
	timeout: 30_000,
	use: { browserName: 'chromium', channel: 'chrome', headless: true, launchOptions: { ...(process.env.GIMBLE_PLAYWRIGHT_EXECUTABLE ? { executablePath: process.env.GIMBLE_PLAYWRIGHT_EXECUTABLE } : {}) } }
})
