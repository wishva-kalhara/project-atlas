import { test, expect } from '@playwright/test';

test('test', async ({ page }) => {

// images upload
await page.getByRole('button', { name: 'Add ingredients to the prompt' }).click();
await page.getByRole('button', { name: 'Upload media' }).click();
await page.locator('input[type="file"]').setInputFiles('image_11-22-1.png');
await page.getByRole('progressbar', { name: 'Loading preview' }).click();
wait for 5s
await page.getByRole('button', { name: 'Add to prompt' }).click();

// enter the prompt
await page.getByRole('textbox', { name: 'What do you want to create?' }).getByRole('paragraph').click();

// select video config
await page.getByRole('button', { name: 'Settings trigger' }).click();
await page.getByRole('radio', { name: 'Video' }).click();
await page.getByRole('radio', { name: ':16' }).click();
await page.getByRole('button', { name: 'Select model family' }).click();
await page.getByRole('menuitem', { name: 'Omni 1.1 Flash Leaving Dec' }).click();
await page.getByRole('radio', { name: '10s' }).click();

// click on generate
await page.getByRole('button', { name: 'Start generation' }).click();

// wait until this element is visible
await page.getByLabel('Generated video').click();

// Download video to {{EXTRACT_BASE_PATH_FROM_`data-harness-gemini-in`_images[0]}}.\\data-workflow-googleflow-out\\{{EXTRACT_BASE_PATH_FROM `data-harness-gemini-in`.tag}}.mp4
await page.getByRole('button', { name: 'Download media' }).click();
const downloadPromise = page.waitForEvent('download');
await page.getByRole('menuitem', { name: '720p Original size' }).click();
const download = await downloadPromise;
});
