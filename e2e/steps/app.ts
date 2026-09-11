import { expect } from '@playwright/test';
import { createBdd } from 'playwright-bdd';
import { hydrated, test } from './fixtures.js';

const { Given, When, Then } = createBdd(test);

Given('I open the generated application', async ({ page, browserState }) => {
	await page.goto('/');
	await hydrated(page);
	expect(browserState.documents).toBe(1);
});

Then(
	'the heading is {string} and the initial Go status is visible',
	async ({ page, browserState }, app: string) => {
		await expect(page.getByTestId('title')).toHaveText(app);
		await expect(page.getByTestId('answered-by')).toHaveText(/^Served by go1\./);
		await expect(page.getByTestId('greetings')).toHaveText('0');
		await expect(page.getByTestId('last-greeting')).toHaveText('Last greeting: (none yet)');
		expect(browserState.documents).toBe(1);
		expect(browserState.pageErrors).toEqual([]);
	}
);

When('I greet {string}', async ({ page, browserState }, name: string) => {
	browserState.remoteMark = browserState.remotes.length;
	await page.getByTestId('name').fill(name);
	await page.getByTestId('greet').click();
});

Then(
	'exactly one greeting from {string} is visible without a document reload',
	async ({ page, browserState }, name: string) => {
		await expect(page.getByTestId('greetings')).toHaveText('1');
		await expect(page.getByTestId('last-greeting')).toHaveText(`Last greeting: ${name}`);
		expect(browserState.documents).toBe(1);
		expect(browserState.remotes.slice(browserState.remoteMark)).toHaveLength(1);
		expect(browserState.pageErrors).toEqual([]);
	}
);

When('I follow the About link', async ({ page }) => {
	await page.getByRole('link', { name: 'About' }).click();
});

Then('About is visible without a document reload', async ({ page, browserState }) => {
	await expect(page).toHaveURL(/\/about$/);
	await expect(page.getByTestId('title')).toHaveText('About');
	expect(browserState.documents).toBe(1);
	expect(browserState.pageErrors).toEqual([]);
});

When('I load the About route directly', async ({ page }) => {
	await page.goto('/about');
});

Then('About is visible in a new document', async ({ page, browserState }) => {
	await expect(page.getByTestId('title')).toHaveText('About');
	expect(browserState.documents).toBe(2);
	expect(browserState.pageErrors).toEqual([]);
});
