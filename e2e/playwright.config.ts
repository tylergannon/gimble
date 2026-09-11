import { defineConfig, devices } from '@playwright/test';
import { defineBddConfig } from 'playwright-bdd';

const run = process.env.SKGO_E2E_RUN ?? 'run';
const artifacts = process.env.SKGO_E2E_ARTIFACTS ?? '.';
const testDir = defineBddConfig({
	features: 'features/**/*.feature',
	steps: 'steps/**/*.ts',
	outputDir: '.features-gen'
});

export default defineConfig({
	testDir,
	fullyParallel: false,
	workers: 1,
	retries: 0,
	reporter: [
		['list'],
		['html', { open: 'never', outputFolder: `${artifacts}/playwright-report/${run}` }]
	],
	outputDir: `${artifacts}/test-results/${run}`,
	use: {
		baseURL: process.env.BASE_URL,
		trace: 'retain-on-failure'
	},
	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }]
});
