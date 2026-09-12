import { expect } from '@playwright/test';
import { createBdd } from 'playwright-bdd';
import { hydrated, test } from './fixtures.js';

const { Given, When, Then } = createBdd(test);

Given('I open the Gimble guide', async ({ page, browserState }) => {
	await page.goto('/');
	await hydrated(page);
	expect(browserState.documents).toBe(1);
});

Then('the guide says Gimble runs agent workflows in Go', async ({ page, browserState }) => {
	await expect(page.getByTestId('title')).toHaveText('Agent workflows in Go');
	await expect(page.getByText('Gimble is a Go library for running agent work from ordinary Go code.')).toBeVisible();
	expect(browserState.documents).toBe(1);
	expect(browserState.pageErrors).toEqual([]);
});

Then('the guide says what Gimble does not do', async ({ page }) => {
	await expect(page.getByRole('heading', { name: 'What it does not do' })).toBeVisible();
	await expect(page.getByText('Gimble does not choose your process for you.')).toBeVisible();
});

When('I follow the About link', async ({ page }) => {
	await page.getByRole('link', { name: 'About' }).click();
});

Then('About is visible without a document reload', async ({ page, browserState }) => {
	await expect(page).toHaveURL(/\/about$/);
	await expect(page.getByTestId('title')).toHaveText('About Gimble');
	expect(browserState.documents).toBe(1);
	expect(browserState.pageErrors).toEqual([]);
});

When('I load the About route directly', async ({ page }) => {
	await page.goto('/about');
});

Then('About is visible in a new document', async ({ page, browserState }) => {
	await expect(page.getByTestId('title')).toHaveText('About Gimble');
	expect(browserState.documents).toBe(2);
	expect(browserState.pageErrors).toEqual([]);
});
